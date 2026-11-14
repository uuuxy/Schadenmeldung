package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestJedeMeldungGehtSofortRaus(t *testing.T) {
	u := neueTestUmgebung(t)
	lehrer, hm := u.nutzer("lehrer"), u.nutzer("hm")
	lehrer.melde(map[string]string{"ort": "Aula", "kategorie": "licht", "was": "Lampe defekt"}, nil)
	lehrer.melde(map[string]string{"ort": "WC 1. OG", "kategorie": "wasser", "was": "Wasser läuft aus", "dringend": "1"}, nil)
	hm.melde(map[string]string{"ort": "Keller", "kategorie": "sonstiges", "was": "Eigene Notiz"}, nil)
	u.app.versendeAusstehende(context.Background())
	u.app.versendeAusstehende(context.Background()) // eine zweite Runde schickt nichts doppelt

	mails := u.post.alle()
	erwartet := []string{"Neue Schadensmeldung: Lampe defekt – Aula", "Dringend: Wasser läuft aus – WC 1. OG"}
	if !slices.Equal(betreffe(mails), erwartet) {
		t.Fatalf("Mails %v, erwartet je Meldung eine eigene (nicht die selbst gemeldete): %v", betreffe(mails), erwartet)
	}
	for _, m := range mails {
		if m.AntwortAn != "lehrer@schule.de" {
			t.Errorf("Antworten geht an %q, nicht an die meldende Person", m.AntwortAn)
		}
		if m.An != "hm@schule.de" || !strings.Contains(m.HTML, "https://schaden.schule.de/aktion/") {
			t.Errorf("Mail an %q ohne Knöpfe", m.An)
		}
		for _, tun := range []string{"in_arbeit", "fachfirma", "erledigt"} {
			if !strings.Contains(m.HTML, "?tun="+tun) || !strings.Contains(m.Text, "?tun="+tun) {
				t.Errorf("Mail an %q ohne Knopf %q", m.An, tun)
			}
		}
	}
}

// Antwortet der Mailserver nicht, geht nichts verloren: Der Postausgang schickt später nach, aber nur,
// was noch offen ist, und höchstens einen Tag lang.
func TestGescheiterteMailKommtNach(t *testing.T) {
	u := neueTestUmgebung(t)
	ctx := context.Background()
	lehrer, hm := u.nutzer("lehrer"), u.nutzer("hm")
	ausfall := errors.New("Mailserver antwortet nicht")

	u.post.stoere(ausfall)
	lehrer.melde(map[string]string{"ort": "Aula", "kategorie": "licht", "was": "Lampe defekt"}, nil)
	u.app.versendeAusstehende(ctx)
	u.post.stoere(nil)
	u.uhr = u.uhr.Add(nachholTakt)
	u.app.versendeAusstehende(ctx)
	u.app.versendeAusstehende(ctx)
	if b := betreffe(u.post.alle()); len(b) != 1 {
		t.Fatalf("nach dem Ausfall: %v, erwartet genau eine Mail", b)
	}

	u.post.stoere(ausfall)
	lehrer.melde(map[string]string{"ort": "204", "kategorie": "fenster", "was": "Tür klemmt"}, nil)
	_, fenster := lehrer.melde(map[string]string{"ort": "205", "kategorie": "fenster", "was": "Fenster offen"}, nil)
	u.app.versendeAusstehende(ctx)
	hm.json("POST", fmt.Sprintf("/api/meldungen/%v/status", fenster["id"]), map[string]string{"status": "in_arbeit"})
	u.post.stoere(nil)
	u.uhr = u.uhr.Add(time.Hour)
	u.app.versendeAusstehende(ctx)
	if b := betreffe(u.post.alle()); len(b) != 2 || !strings.Contains(b[1], "Tür klemmt") {
		t.Fatalf("eine Stunde später: %v, erwartet nur „Tür klemmt“ dazu (das Fenster ist schon in Arbeit)", b)
	}

	u.post.stoere(ausfall)
	lehrer.melde(map[string]string{"ort": "101", "kategorie": "heizung", "was": "Zu kalt"}, nil)
	u.app.versendeAusstehende(ctx)
	u.post.stoere(nil)
	u.uhr = u.uhr.Add(nachholFrist + time.Minute)
	u.app.versendeAusstehende(ctx)
	if b := betreffe(u.post.alle()); len(b) != 2 {
		t.Fatalf("nach mehr als einem Tag kam noch: %v", b)
	}
}

// Jede Person schaltet ihre Mail selbst; wer nur sehen darf (Verwaltung), bekommt sie ohne Knöpfe.
func TestMailNurAnWerSieWill(t *testing.T) {
	u := neueTestUmgebung(t)
	if err := u.app.sp.setzeRolle(context.Background(), "sekretariat@schule.de", rolleVerwaltung, "test"); err != nil {
		t.Fatal(err)
	}
	hm, sek := u.nutzer("hm"), u.nutzer("sekretariat")
	if _, m := hm.json("POST", "/api/benachrichtigung", map[string]bool{"mail": false}); m["mail"] != false {
		t.Fatalf("Hausmeister schaltet ab: %v", m)
	}
	if _, m := sek.json("GET", "/api/benachrichtigung", nil); m["mail"] != false || m["mailAktiv"] != true {
		t.Fatalf("Verwaltung ohne eigene Wahl: %v, erwartet aus", m)
	}
	sek.json("POST", "/api/benachrichtigung", map[string]bool{"mail": true})
	// Ein alter Stand der Oberfläche schickt die früheren Felder; das ändert nichts.
	if _, m := sek.json("POST", "/api/benachrichtigung", map[string]bool{"dringendSofort": false}); m["mail"] != true {
		t.Fatalf("alte Felder haben die Wahl verändert: %v", m)
	}

	u.nutzer("lehrer").melde(map[string]string{"ort": "Aula", "kategorie": "licht", "was": "Lampe defekt"}, nil)
	u.app.versendeAusstehende(context.Background())
	mails := u.post.alle()
	if len(mails) != 1 || mails[0].An != "sekretariat@schule.de" {
		t.Fatalf("Mails an %v, erwartet nur die Verwaltung", mails)
	}
	if strings.Contains(mails[0].Text, "/aktion/") {
		t.Error("Die Verwaltung bekommt Knöpfe, obwohl sie nichts ändern darf")
	}
}

// Im Betrieb verschickt der Takt: Eine neue Meldung weckt ihn, ohne fünf Minuten zu warten.
func TestPostausgangWachtBeiNeuerMeldungAuf(t *testing.T) {
	u := neueTestUmgebung(t)
	ctx, stop := context.WithCancel(context.Background())
	beendet := make(chan struct{})
	go func() { u.app.postausgangTakt(ctx); close(beendet) }()
	defer func() { stop(); <-beendet }()

	u.nutzer("lehrer").melde(map[string]string{"ort": "Aula", "kategorie": "licht", "was": "Lampe defekt"}, nil)
	for frist := time.Now().Add(5 * time.Second); len(u.post.alle()) == 0 && time.Now().Before(frist); {
		time.Sleep(20 * time.Millisecond)
	}
	if b := betreffe(u.post.alle()); len(b) != 1 {
		t.Fatalf("Mails nach der Meldung: %v, erwartet sofort eine", b)
	}
}

func TestKnopfAusDerMail(t *testing.T) {
	u := neueTestUmgebung(t)
	u.nutzer("lehrer").melde(map[string]string{"ort": "WC 1. OG", "kategorie": "wasser", "was": "WC verstopft", "dringend": "1"}, nil)
	u.app.versendeAusstehende(context.Background())
	treffer := linkMuster.FindStringSubmatch(u.post.alle()[0].Text)
	if treffer == nil {
		t.Fatal("kein Erledigt-Link in der Mail")
	}
	pfad := "/api/aktion/" + treffer[1]
	gast := u.anonym()

	if status, m := gast.json("GET", pfad, nil); status != 200 || m["was"] != "WC verstopft" || m["gemeldetVon"] != nil {
		t.Fatalf("Link öffnen: %d, %v", status, m)
	}
	resp, err := http.Post(u.srv.URL+pfad, "application/json", strings.NewReader(`{"status":"erledigt"}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("ohne X-Schaden: Status %d, erwartet 403 (Schutz vor fremden Formularen)", resp.StatusCode)
	}
	// Zurück auf „offen“ geht nur in der App, nicht mit einem weitergeleiteten Link.
	if status, _ := gast.json("POST", pfad, map[string]string{"status": "offen"}); status != http.StatusBadRequest {
		t.Fatalf("offen aus der Mail: Status %d, erwartet 400", status)
	}
	if status, m := gast.json("POST", pfad, map[string]string{"status": "erledigt", "vorher": "offen"}); status != 200 || m["status"] != "erledigt" {
		t.Fatalf("Erledigt aus der Mail: %d, %v", status, m)
	}

	// Der zweite Hausmeister tippt auf einer Seite, die noch „offen“ zeigte: Die Meldung bleibt erledigt.
	if status, _ := gast.json("POST", pfad, map[string]string{"status": "in_arbeit", "vorher": "offen"}); status != http.StatusConflict {
		t.Fatalf("In Arbeit auf überholter Seite: Status %d, erwartet 409", status)
	}
	if _, m := gast.json("GET", pfad, nil); m["status"] != "erledigt" {
		t.Fatalf("nach dem abgelehnten Tipp: %v, erwartet weiter erledigt", m["status"])
	}
	// Noch einmal „Erledigt“ ändert nichts und ist kein Fehler.
	if status, m := gast.json("POST", pfad, map[string]string{"status": "erledigt", "vorher": "offen"}); status != 200 || m["status"] != "erledigt" {
		t.Fatalf("zweimal Erledigt: %d, %v", status, m)
	}
	// Wer „Schon erledigt“ gesehen hat und bewusst ändert, darf wieder öffnen.
	if status, m := gast.json("POST", pfad, map[string]string{"status": "in_arbeit", "vorher": "erledigt"}); status != 200 || m["status"] != "in_arbeit" {
		t.Fatalf("bewusst wieder geöffnet: %d, %v", status, m)
	}

	// Wer nur noch lesen darf (Verwaltung), kann auch mit einem alten Link nichts mehr ändern.
	if err := u.app.sp.setzeRolle(context.Background(), "hm@schule.de", rolleVerwaltung, "test"); err != nil {
		t.Fatal(err)
	}
	if status, _ := gast.json("POST", pfad, map[string]string{"status": "offen"}); status != http.StatusForbidden {
		t.Fatalf("nach Rollenwechsel: Status %d, erwartet 403", status)
	}
	u.uhr = u.uhr.Add(15 * 24 * time.Hour)
	if status, _ := gast.json("GET", pfad, nil); status != http.StatusGone {
		t.Fatalf("nach 15 Tagen: Status %d, erwartet 410", status)
	}
}
