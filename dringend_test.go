package main

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"
)

// Dringend setzt und nimmt jede Person mit der Rolle Hausmeister, auch ein zweiter, bei jeder Meldung,
// ob selbst gemeldet oder nicht; Kollegium und Verwaltung nicht. Stand und Vermerk bleiben dabei stehen.
func TestDringendSetzenUndWegnehmen(t *testing.T) {
	u := neueTestUmgebung(t)
	for email, rolle := range map[string]string{"hm2@schule.de": rolleHausmeister, "sekretariat@schule.de": rolleVerwaltung} {
		if err := u.app.sp.setzeRolle(context.Background(), email, rolle, "test"); err != nil {
			t.Fatal(err)
		}
	}
	lehrer, hm, hm2, sek := u.nutzer("lehrer"), u.nutzer("hm"), u.nutzer("hm2"), u.nutzer("sekretariat")
	_, eigene := hm.melde(map[string]string{"ort": "Aula", "kategorie": "licht", "was": "Lampe defekt"}, nil)
	u.uhr = u.uhr.Add(time.Minute)
	_, fremde := lehrer.melde(map[string]string{"ort": "Raum 204", "kategorie": "technik", "was": "Beamer/Display aus"}, nil)
	pfad := func(m map[string]any) string {
		return "/api/meldungen/" + strconv.Itoa(int(m["id"].(float64))) + "/status"
	}
	ersteOffene := func() any {
		_, alle := hm.json("GET", "/api/meldungen", nil)
		return eintraege(alle, "offen")[0]["was"]
	}

	for name, n := range map[string]*testNutzer{"Kollegium": lehrer, "Verwaltung": sek} {
		if status, _ := n.json("POST", pfad(fremde), map[string]bool{"dringend": true}); status != http.StatusForbidden {
			t.Errorf("%s setzt Dringend: Status %d, erwartet 403", name, status)
		}
	}
	hm.json("POST", pfad(eigene), map[string]string{"status": "in_arbeit", "antwort": "Leiter fehlt"})
	if status, neu := hm2.json("POST", pfad(eigene), map[string]bool{"dringend": true}); status != 200 ||
		neu["dringend"] != true || neu["status"] != "in_arbeit" || neu["antwort"] != "Leiter fehlt" {
		t.Fatalf("Zweiter Hausmeister setzt Dringend: %d, %v", status, neu)
	}
	if was := ersteOffene(); was != "Lampe defekt" {
		t.Errorf("Dringendes steht nicht oben, sondern %v", was)
	}
	if status, neu := hm.json("POST", pfad(eigene), map[string]bool{"dringend": false}); status != 200 || neu["dringend"] != false {
		t.Fatalf("Dringend wegnehmen: %d, %v", status, neu)
	}
	if was := ersteOffene(); was != "Beamer/Display aus" {
		t.Errorf("Ohne Dringend steht die neueste oben, nicht %v", was)
	}
}
