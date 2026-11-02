package main

import (
	"context"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"testing"
	"time"
)

func TestAnmeldungErgaenztDomainUndVergibtRollen(t *testing.T) {
	u := neueTestUmgebung(t)
	for name, rolle := range map[string]string{"lehrer": "kollegium", "hm": "hausmeister", "Admin": "admin"} {
		n := u.nutzer(name)
		status, ich := n.json("GET", "/api/ich", nil)
		if status != 200 || ich["rolle"] != rolle || !strings.HasSuffix(ich["email"].(string), "@schule.de") {
			t.Errorf("%s: Status %d, %v; erwartet Rolle %s", name, status, ich, rolle)
		}
	}
}

func TestFremdeDomainKommtNichtHinein(t *testing.T) {
	u := neueTestUmgebung(t)
	jar, _ := cookiejar.New(nil)
	n := &testNutzer{u: u, client: &http.Client{Jar: jar}}
	status, _ := n.json("POST", "/api/anmelden", map[string]string{"email": "x@gmail.com", "passwort": "p"})
	if status != http.StatusForbidden {
		t.Fatalf("Status %d, erwartet 403", status)
	}
	if status, _ := n.json("GET", "/api/ich", nil); status != http.StatusUnauthorized {
		t.Fatalf("ohne Anmeldung: Status %d, erwartet 401", status)
	}
}

// Die Schul-IT nennt die volle Adresse als Benutzernamen, im Kollegium ist die kurze geläufig.
func TestKurzdomainWirdUmgeschrieben(t *testing.T) {
	k := konfiguration{Domains: []string{"schule.example"}, Aliase: []string{"kurz.example"}}
	for ein, erwartet := range map[string]string{
		"p.muster":                        "p.muster@schule.example",
		" p.muster@kurz.example ":           "p.muster@schule.example",
		"P.Muster@Kurz.Example":             "P.Muster@schule.example",
		"p.muster@schule.example": "p.muster@schule.example",
		"p.muster@gmail.com":              "p.muster@gmail.com",
		"p.muster@neben.kurz.example":       "p.muster@neben.kurz.example",
		"@kurz.example":                     "@kurz.example",
		"":                                "",
	} {
		if aus := k.vervollstaendige(ein); aus != erwartet {
			t.Errorf("vervollstaendige(%q) = %q, erwartet %q", ein, aus, erwartet)
		}
	}
}

func TestKurzadresseMeldetMitDerVollenAn(t *testing.T) {
	u := neueTestUmgebung(t)
	var amMailserver string
	u.app.pruefeZugang = func(_ context.Context, email, _ string) error {
		amMailserver = email
		return nil
	}
	n := u.anonym()
	status, ich := n.json("POST", "/api/anmelden", map[string]string{"email": "Lehrer@Kurz.de", "passwort": "p"})
	if status != 200 || ich["email"] != "lehrer@schule.de" || amMailserver != "lehrer@schule.de" {
		t.Fatalf("Status %d, %v, am Mailserver %q; erwartet lehrer@schule.de", status, ich, amMailserver)
	}
	// Auch die Verwaltung darf die Kurzform eintragen; die Rolle hängt an der vollen Adresse.
	v := u.nutzer("admin")
	if status, antwort := v.json("POST", "/api/rollen", map[string]string{"email": "lehrer@kurz.de", "rolle": "leitung"}); status != 200 {
		t.Fatalf("Rolle setzen: Status %d, %v", status, antwort)
	}
	if _, ich := n.json("GET", "/api/ich", nil); ich["rolle"] != "leitung" {
		t.Errorf("Rolle nach Eintrag in Kurzform: %v, erwartet leitung", ich["rolle"])
	}
}

// Wer sich selbst in der Kurzform in ADMINS einträgt, darf sich damit nicht aussperren.
func TestAdminsDuerfenKurzGeschriebenSein(t *testing.T) {
	for _, name := range []string{"SMTP_HOST", "FOTO_TAGE", "AUFBEWAHRUNG_TAGE", "SAMMELMAIL_ZEITEN", "ZEITZONE"} {
		t.Setenv(name, "")
	}
	t.Setenv("DATABASE_URL", "postgres://schaden@localhost/schaden")
	t.Setenv("IMAP_HOST", "mail.schule.example")
	t.Setenv("ANMELDE_DOMAINS", "schule.example")
	t.Setenv("ALIAS_DOMAINS", "kurz.example")
	t.Setenv("ADMINS", "Chefin@kurz.example, verwaltung")
	k, err := ladeKonfiguration()
	if err != nil {
		t.Fatal(err)
	}
	for _, adresse := range []string{"chefin@schule.example", "verwaltung@schule.example"} {
		if !k.Admins[adresse] {
			t.Errorf("%s fehlt in ADMINS: %v", adresse, k.Admins)
		}
	}
}

func TestBremseSperrtNachFehlversuchenUndZaehltAusfallNicht(t *testing.T) {
	u := neueTestUmgebung(t)
	var antwort error
	u.app.pruefeZugang = func(context.Context, string, string) error { return antwort }
	jar, _ := cookiejar.New(nil)
	n := &testNutzer{u: u, client: &http.Client{Jar: jar}}
	anmelden := func() int {
		status, _ := n.json("POST", "/api/anmelden", map[string]string{"email": "lehrer", "passwort": "falsch"})
		return status
	}

	antwort = errMailserverWeg
	for range 6 {
		if s := anmelden(); s != http.StatusServiceUnavailable {
			t.Fatalf("Ausfall: Status %d, erwartet 503", s)
		}
	}
	antwort = errAbgelehnt
	for i := range 5 {
		if s := anmelden(); s != http.StatusUnauthorized {
			t.Fatalf("Fehlversuch %d: Status %d, erwartet 401", i+1, s)
		}
	}
	antwort = nil
	if s := anmelden(); s != http.StatusTooManyRequests {
		t.Fatalf("nach 5 Fehlversuchen: Status %d, erwartet 429 auch mit richtigem Passwort", s)
	}
	u.uhr = u.uhr.Add(16 * time.Minute)
	if s := anmelden(); s != http.StatusOK {
		t.Fatalf("nach Ablauf der Sperre: Status %d, erwartet 200", s)
	}
}

func TestKonfigurationVerweigertOffeneTueren(t *testing.T) {
	const db = "postgres://schaden@localhost/schaden"
	for name, fall := range map[string]struct {
		k  konfiguration
		ok bool
	}{
		"mock auf dem Server":   {konfiguration{DatenbankURL: db, IMAPHost: "mock"}, false},
		"ohne Mailserver":       {konfiguration{DatenbankURL: db}, false},
		"ohne Datenbank":        {konfiguration{IMAPHost: "mail.schule.de"}, false},
		"mock lokal":            {konfiguration{DatenbankURL: db, IMAPHost: "mock", Lokal: true}, true},
		"Mailserver der Schule": {konfiguration{DatenbankURL: db, IMAPHost: "mail.schule.de"}, true},
		"Kurzdomain ohne Ziel":  {konfiguration{DatenbankURL: db, IMAPHost: "mail.schule.de", Aliase: []string{"kurz.de"}}, false},
		"Kurzdomain, zwei Ziele": {konfiguration{DatenbankURL: db, IMAPHost: "mail.schule.de",
			Domains: []string{"schule.de", "andere.de"}, Aliase: []string{"kurz.de"}}, false},
		"Kurzdomain mit Ziel": {konfiguration{DatenbankURL: db, IMAPHost: "mail.schule.de",
			Domains: []string{"schule.de"}, Aliase: []string{"kurz.de"}}, true},
	} {
		if err := fall.k.pruefe(); (err == nil) != fall.ok {
			t.Errorf("%s: pruefe() = %v, erwartet ok=%v", name, err, fall.ok)
		}
	}
}
