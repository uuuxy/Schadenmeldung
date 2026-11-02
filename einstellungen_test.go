package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"
)

func TestArtenAbUndAnschalten(t *testing.T) {
	u := neueTestUmgebung(t)
	artenAus := func(antwort map[string]any) string { return fmt.Sprint(antwort["artenAus"]) }

	// Medientechnik meldet das Kollegium über das Ticketsystem der IT: zu Beginn aus.
	if _, e := u.anonym().json("GET", "/api/einstellungen", nil); artenAus(e) != "[technik]" {
		t.Fatalf("zu Beginn: %v", e)
	}
	hm := u.nutzer("hm")
	for _, schritt := range []struct {
		kategorie string
		sichtbar  bool
		soll      string
	}{
		{"technik", true, "[]"},
		{"reinigung", false, "[reinigung]"},
		{"reinigung", false, "[reinigung]"},
		{"technik", false, "[reinigung technik]"},
	} {
		status, a := hm.json("POST", "/api/arten", map[string]any{"kategorie": schritt.kategorie, "sichtbar": schritt.sichtbar})
		if status != http.StatusOK || artenAus(a) != schritt.soll {
			t.Fatalf("%s sichtbar=%v: Status %d, %v, erwartet %s", schritt.kategorie, schritt.sichtbar, status, a, schritt.soll)
		}
	}
	for _, unerlaubt := range []string{"sonstiges", "garten"} {
		if status, _ := hm.json("POST", "/api/arten", map[string]any{"kategorie": unerlaubt, "sichtbar": false}); status != http.StatusBadRequest {
			t.Errorf("%s abschalten: Status %d, erwartet 400", unerlaubt, status)
		}
	}
	lehrer := u.nutzer("lehrer")
	if status, _ := lehrer.json("POST", "/api/arten", map[string]any{"kategorie": "licht", "sichtbar": false}); status != http.StatusForbidden {
		t.Errorf("Kollegium schaltet Arten: Status %d", status)
	}
	// Was im Funkloch schon im Ausgang lag, geht nicht verloren, auch wenn die Art inzwischen aus ist.
	if status, m := lehrer.melde(map[string]string{"ort": "Flur", "kategorie": "reinigung", "was": "Mülleimer voll"}, nil); status != http.StatusCreated {
		t.Errorf("Meldung einer abgeschalteten Art: Status %d, %v", status, m)
	}
}

// Namen stehen in Server und Oberfläche; Excel, PDF und die Kacheln sollen gleich heißen.
func TestNamenWieInDerOberflaeche(t *testing.T) {
	js, err := os.ReadFile("frontend/src/lib/kategorien.js")
	if err != nil {
		t.Fatal(err)
	}
	kacheln := regexp.MustCompile(`key: '(\w+)',\s*name: '([^']+)'`).FindAllStringSubmatch(string(js), -1)
	if len(kacheln) != len(arten) {
		t.Fatalf("%d Arten in kategorien.js, %d im Server", len(kacheln), len(arten))
	}
	for i, k := range kacheln {
		if arten[i].schluessel != k[1] || arten[i].name != k[2] {
			t.Errorf("Art %d: Oberfläche %s „%s“, Server %s „%s“", i+1, k[1], k[2], arten[i].schluessel, arten[i].name)
		}
	}
	js, err = os.ReadFile("frontend/src/lib/status.js")
	if err != nil {
		t.Fatal(err)
	}
	stati := regexp.MustCompile(`(\w+): \{ name: '([^']+)'`).FindAllStringSubmatch(string(js), -1)
	if len(stati) != len(statusNamen) {
		t.Fatalf("%d Stati in status.js, %d im Server", len(stati), len(statusNamen))
	}
	for _, s := range stati {
		if statusNamen[s[1]] != s[2] {
			t.Errorf("Status %s: Oberfläche „%s“, Server „%s“", s[1], s[2], statusNamen[s[1]])
		}
	}
}
