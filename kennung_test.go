package main

import (
	"net/http"
	"strings"
	"testing"
)

// Schickt das Handy dieselbe Meldung zweimal (Antwort im Funkloch verloren), entsteht sie nur einmal.
// Eine andere Person mit derselben Kennung meldet trotzdem: Die Kennung gilt je Person.
func TestDoppeltGeschickteMeldungZaehltEinmal(t *testing.T) {
	u := neueTestUmgebung(t)
	lehrer, hm := u.nutzer("lehrer"), u.nutzer("hm")
	felder := map[string]string{"ort": "Flur", "kategorie": "licht", "was": "Lampe defekt", "kennung": strings.Repeat("ab", 16)}
	status1, erste := lehrer.melde(felder, nil)
	status2, zweite := lehrer.melde(felder, nil)
	if status1 != http.StatusCreated || status2 != http.StatusOK || erste["id"] != zweite["id"] {
		t.Fatalf("erste: %d %v, zweite: %d %v", status1, erste["id"], status2, zweite["id"])
	}
	if status, _ := u.nutzer("kollegin").melde(felder, nil); status != http.StatusCreated {
		t.Fatalf("andere Person mit derselben Kennung: Status %d", status)
	}
	_, alle := hm.json("GET", "/api/meldungen", nil)
	if n := len(eintraege(alle, "offen")); n != 2 {
		t.Fatalf("%d offene Meldungen, erwartet 2", n)
	}
}
