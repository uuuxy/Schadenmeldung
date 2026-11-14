package main

import (
	"context"
	"net/http"
	"strconv"
	"testing"
)

// Zwei Hausmeister: Hat einer erledigt, öffnet ein Tipp vom überholten Stand des anderen (alte Seite oder
// Ausgang im Funkloch) die Meldung nicht wieder. Bewusst wieder öffnen geht, Dringend und alte Seiten
// ohne vorher auch.
func TestUeberholterStandOeffnetNichtWieder(t *testing.T) {
	u := neueTestUmgebung(t)
	if err := u.app.sp.setzeRolle(context.Background(), "hm2@schule.de", rolleHausmeister, "test"); err != nil {
		t.Fatal(err)
	}
	lehrer, hm, hm2 := u.nutzer("lehrer"), u.nutzer("hm"), u.nutzer("hm2")
	_, m := lehrer.melde(map[string]string{"ort": "Keller", "kategorie": "wasser", "was": "Rohr tropft"}, nil)
	pfad := "/api/meldungen/" + strconv.Itoa(int(m["id"].(float64))) + "/status"
	if status, _ := hm.json("POST", pfad, map[string]string{"status": "erledigt", "vorher": "offen"}); status != http.StatusOK {
		t.Fatalf("Erledigt: Status %d", status)
	}
	for name, aenderung := range map[string]map[string]any{
		"„In Arbeit“ aus dem Funkloch":      {"status": "in_arbeit", "vorher": "offen"},
		"„Rückgängig“ vom überholten Stand": {"status": "offen", "vorher": "in_arbeit"},
	} {
		if status, antwort := hm2.json("POST", pfad, aenderung); status != http.StatusConflict || antwort["fehler"] != ueberholtText {
			t.Errorf("%s: Status %d, %v, erwartet 409", name, status, antwort)
		}
	}
	if status, neu := hm2.json("POST", pfad, map[string]any{"dringend": true, "vorher": "offen"}); status != http.StatusOK || neu["status"] != "erledigt" {
		t.Errorf("Dringend ändert den Stand nicht und geht immer: Status %d, %v", status, neu)
	}
	if status, neu := hm2.json("POST", pfad, map[string]string{"status": "offen", "vorher": "erledigt"}); status != http.StatusOK || neu["status"] != "offen" {
		t.Errorf("Bewusst wieder öffnen: Status %d, %v", status, neu)
	}
	hm.json("POST", pfad, map[string]string{"status": "erledigt"})
	if status, neu := hm2.json("POST", pfad, map[string]string{"status": "in_arbeit"}); status != http.StatusOK || neu["status"] != "in_arbeit" {
		t.Errorf("Alte Seite ohne vorher wie bisher: Status %d, %v", status, neu)
	}
}
