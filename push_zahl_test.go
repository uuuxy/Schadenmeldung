package main

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

// Die Zahl am App-Symbol zählt je Gerät, was seit dem letzten Öffnen neu ist. Eine Meldung aus dem
// Funkloch trägt die Zeit, zu der sie getippt wurde; kommt sie erst nach dem Öffnen an, zählt sie trotzdem,
// denn sonst käme eine Mitteilung ohne Zahl.
func TestPushZahlJeGeraet(t *testing.T) {
	u := neueTestUmgebung(t)
	dienst := mitPushDienst(u, http.StatusCreated)
	hm, lehrer := u.nutzer("hm"), u.nutzer("lehrer")
	g := neuesTestGeraet(t, "https://fcm.googleapis.com/fcm/send/hausmeister")
	if status, _ := hm.json("POST", "/api/push", g.abo()); status != http.StatusNoContent {
		t.Fatalf("Abo anlegen: Status %d", status)
	}
	zahl := func() int {
		u.app.pushLaeuft.Wait()
		var n pushNachricht
		if err := json.Unmarshal(entschluessele(t, dienst.erhalten[len(dienst.erhalten)-1].Rumpf, g.schluessel, g.auth), &n); err != nil {
			t.Fatal(err)
		}
		return n.Zahl
	}
	beginn := u.uhr
	for _, was := range []string{"Lampe flackert", "Lampe aus"} {
		u.uhr = u.uhr.Add(time.Minute)
		lehrer.melde(map[string]string{"ort": "Aula", "kategorie": "licht", "was": was}, nil)
	}
	if z := zahl(); z != 2 {
		t.Errorf("Zwei neue Meldungen: Zahl %d, erwartet 2", z)
	}
	// Öffnen der App: Das Gerät meldet sich beim Server, die Zählung beginnt neu.
	u.uhr = u.uhr.Add(time.Minute)
	hm.json("POST", "/api/push", g.abo())
	u.uhr = u.uhr.Add(time.Minute)
	lehrer.melde(map[string]string{"ort": "Keller", "kategorie": "wasser", "was": "Rohr tropft",
		"erstellt": beginn.Add(90 * time.Second).Format(time.RFC3339)}, nil)
	if z := zahl(); z != 1 {
		t.Errorf("Aus dem Funkloch, getippt vor dem Öffnen: Zahl %d, erwartet 1", z)
	}
}
