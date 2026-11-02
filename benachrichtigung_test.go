package main

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestDringendesGehtSofortRaus(t *testing.T) {
	u := neueTestUmgebung(t)
	lehrer := u.nutzer("lehrer")
	lehrer.melde(map[string]string{"ort": "Aula", "kategorie": "licht", "was": "Lampe defekt"}, nil)
	lehrer.melde(map[string]string{"ort": "WC 1. OG", "kategorie": "wasser", "was": "Wasser läuft aus", "dringend": "1"}, nil)
	u.app.hintergrund.Wait()

	mails := u.post.alle()
	if len(mails) != 1 {
		t.Fatalf("%d Mails, erwartet genau eine (nur die dringende)", len(mails))
	}
	m := mails[0]
	if m.An != "hm@schule.de" || m.Betreff != "Dringend: Wasser läuft aus – WC 1. OG" {
		t.Errorf("Mail an %q mit Betreff %q", m.An, m.Betreff)
	}
	if !strings.Contains(m.HTML, "https://schaden.schule.de/aktion/") || !strings.Contains(m.Text, "?tun=erledigt") {
		t.Error("Mail ohne Knöpfe „Gesehen“/„Erledigt“")
	}
	// Die schon verschickte dringende Meldung kommt in der Sammelmail nicht noch einmal.
	neu, err := u.app.sp.ungesendeteFuer(context.Background(), "hm@schule.de", u.uhr)
	if err != nil || len(neu) != 1 || neu[0].Was != "Lampe defekt" {
		t.Errorf("für die Sammelmail übrig: %v, %v", neu, err)
	}
}

func TestSammelmailZuDenTerminenOhneDoppelte(t *testing.T) {
	u := neueTestUmgebung(t)
	ctx := context.Background()
	u.uhr = time.Date(2026, 11, 2, 6, 0, 0, 0, berlin)
	u.app.pruefeSammelmail(ctx) // erster Start: zählt ab jetzt

	lehrer, hm := u.nutzer("lehrer"), u.nutzer("hm")
	u.uhr = u.uhr.Add(10 * time.Minute)
	lehrer.melde(map[string]string{"ort": "204", "kategorie": "technik", "was": "Kein Ton"}, nil)
	lehrer.melde(map[string]string{"ort": "Aula", "kategorie": "licht", "was": "Lampe defekt"}, nil)
	hm.melde(map[string]string{"ort": "Keller", "kategorie": "sonstiges", "was": "Eigene Notiz"}, nil)

	u.uhr = time.Date(2026, 11, 2, 7, 1, 0, 0, berlin)
	u.app.pruefeSammelmail(ctx)
	u.app.pruefeSammelmail(ctx)
	mails := u.post.alle()
	if len(mails) != 1 || mails[0].Betreff != "2 neue Schadensmeldungen" {
		t.Fatalf("um 7 Uhr: %d Mails %v, erwartet eine mit zwei Meldungen (nicht die eigene)", len(mails), betreffe(mails))
	}

	u.uhr = time.Date(2026, 11, 2, 12, 1, 0, 0, berlin)
	u.app.pruefeSammelmail(ctx)
	if len(u.post.alle()) != 1 {
		t.Fatal("um 12 Uhr kam eine Mail ohne neue Meldungen")
	}

	// Samstag: keine Sammelmail; was am Wochenende kommt, steht Montag früh in der Mail.
	u.uhr = time.Date(2026, 11, 7, 10, 0, 0, 0, berlin)
	lehrer.melde(map[string]string{"ort": "Turnhalle", "kategorie": "fenster", "was": "Schloss defekt"}, nil)
	u.uhr = time.Date(2026, 11, 7, 12, 30, 0, 0, berlin)
	u.app.pruefeSammelmail(ctx)
	if len(u.post.alle()) != 1 {
		t.Fatal("am Samstag kam eine Sammelmail")
	}
	u.uhr = time.Date(2026, 11, 9, 7, 0, 30, 0, berlin)
	u.app.pruefeSammelmail(ctx)
	if mails := u.post.alle(); len(mails) != 2 || !strings.Contains(mails[1].Betreff, "Schloss defekt") {
		t.Fatalf("Montag 7 Uhr: %v", betreffe(mails))
	}
}

func TestKnopfAusDerMail(t *testing.T) {
	u := neueTestUmgebung(t)
	u.nutzer("lehrer").melde(map[string]string{"ort": "WC 1. OG", "kategorie": "wasser", "was": "WC verstopft", "dringend": "1"}, nil)
	u.app.hintergrund.Wait()
	treffer := linkMuster.FindStringSubmatch(u.post.alle()[0].Text)
	if treffer == nil {
		t.Fatal("kein Erledigt-Link in der Mail")
	}
	pfad := "/api/aktion/" + treffer[1]
	gast := u.anonym()

	if status, m := gast.json("GET", pfad, nil); status != 200 || m["was"] != "WC verstopft" || m["gemeldetVon"] != "lehrer@schule.de" {
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
	if status, m := gast.json("POST", pfad, map[string]string{"status": "erledigt"}); status != 200 || m["status"] != "erledigt" {
		t.Fatalf("Erledigt aus der Mail: %d, %v", status, m)
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

func TestLetzterTermin(t *testing.T) {
	zeiten := []uhrzeit{{7, 0}, {12, 0}}
	for name, fall := range map[string]struct {
		jetzt time.Time
		ok    bool
		um    int
	}{
		"Montag vor 7":   {time.Date(2026, 11, 2, 6, 59, 0, 0, berlin), false, 0},
		"Montag 9 Uhr":   {time.Date(2026, 11, 2, 9, 0, 0, 0, berlin), true, 7},
		"Montag abends":  {time.Date(2026, 11, 2, 20, 0, 0, 0, berlin), true, 12},
		"Samstag mittag": {time.Date(2026, 11, 7, 12, 30, 0, 0, berlin), false, 0},
	} {
		termin, ok := letzterTermin(fall.jetzt, zeiten)
		if ok != fall.ok || (ok && termin.Hour() != fall.um) {
			t.Errorf("%s: %v %v, erwartet %v um %d Uhr", name, termin, ok, fall.ok, fall.um)
		}
	}
}
