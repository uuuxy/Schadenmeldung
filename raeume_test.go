package main

import (
	"fmt"
	"net/http"
	"testing"
)

// Umbenennen behält den QR-Code (der Aufkleber an der Tür gilt weiter) und zieht die Meldungen nach.
func TestRaumUmbenennen(t *testing.T) {
	u := neueTestUmgebung(t)
	hm, lehrer := u.nutzer("hm"), u.nutzer("lehrer")
	_, antwort := hm.json("POST", "/api/raeume", map[string]string{"text": "Bio1; Neubau\nChemie; Neubau"})
	raeume := eintraege(antwort, "raeume")
	bio, code := raeume[0]["id"], raeume[0]["code"]
	lehrer.melde(map[string]string{"raumId": fmt.Sprint(bio), "kategorie": "licht", "was": "Lampe defekt"}, nil)

	pfad := fmt.Sprintf("/api/raeume/%v", bio)
	status, neu := hm.json("PUT", pfad, map[string]string{"name": " Bio 1 ", "bereich": "Neubau · NaWi"})
	liste, _ := neu["liste"].([]any)
	if status != http.StatusOK || len(liste) != 2 {
		t.Fatalf("Umbenennen: Status %d, %v", status, neu)
	}
	if r := liste[0].(map[string]any); r["name"] != "Bio 1" || r["bereich"] != "Neubau · NaWi" || r["code"] != code {
		t.Errorf("nach dem Umbenennen: %v (Code vorher %v)", r, code)
	}
	if _, alle := hm.json("GET", "/api/meldungen", nil); eintraege(alle, "offen")[0]["ort"] != "Bio 1" {
		t.Errorf("Die Meldung trägt noch den alten Namen: %v", alle)
	}
	if status, a := hm.json("PUT", pfad, map[string]string{"name": "chemie"}); status != http.StatusBadRequest {
		t.Errorf("doppelter Name: Status %d, %v", status, a)
	}
	if status, _ := lehrer.json("PUT", pfad, map[string]string{"name": "X"}); status != http.StatusForbidden {
		t.Errorf("Kollegium benennt um: Status %d", status)
	}
	if status, _ := hm.json("PUT", "/api/raeume/999999", map[string]string{"name": "Y"}); status != http.StatusNotFound {
		t.Errorf("unbekannter Raum: Status %d", status)
	}
}
