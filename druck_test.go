package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"testing"
)

func TestDruckJeMeldungEineSeiteMitFoto(t *testing.T) {
	u := neueTestUmgebung(t)
	if err := u.app.sp.setzeRolle(context.Background(), "sekretariat@schule.de", rolleVerwaltung, "test"); err != nil {
		t.Fatal(err)
	}
	lehrer, hm, sek := u.nutzer("lehrer"), u.nutzer("hm"), u.nutzer("sekretariat")
	_, lampe := lehrer.melde(map[string]string{"ort": "Aula", "kategorie": "licht", "was": "Lampe defekt", "notiz": "flackert seit Montag"}, fotoMitExif(t, 800, 600))
	_, hahn := lehrer.melde(map[string]string{"ort": "Schulhof", "kategorie": "wasser", "was": "Hahn tropft", "dringend": "1"}, nil)
	hm.json("POST", fmt.Sprintf("/api/meldungen/%v/status", hahn["id"]), map[string]string{"status": "in_arbeit", "antwort": "Dichtung bestellt"})

	// Zwei gewählt, in der Reihenfolge der Liste: der Hahn zuerst.
	status, kopf, roh := hm.holeDatei(fmt.Sprintf("/api/druck.pdf?ids=%v,%v", hahn["id"], lampe["id"]))
	if status != http.StatusOK || !bytes.HasPrefix(roh, []byte("%PDF-")) || !strings.Contains(kopf.Get("Content-Disposition"), `"Schadenmeldungen-2026-11-02.pdf"`) {
		t.Fatalf("Druck zweier Meldungen: Status %d, %v", status, kopf)
	}
	if seiten := len(regexp.MustCompile(`/Type /Page\b`).FindAll(roh, -1)); seiten != 2 {
		t.Errorf("Druck zweier Meldungen hat %d Seiten statt 2", seiten)
	}
	text := pdfInhalt(t, roh)
	for _, soll := range []string{"(Hahn tropft)", "(DRINGEND)", "(Dichtung bestellt)", "(In Arbeit)", "(Lampe defekt)", "(flackert seit Montag)", fmt.Sprintf("(Schadenmeldung Nr. %v", lampe["id"])} {
		if !strings.Contains(text, soll) {
			t.Errorf("Druck ohne %s", soll)
		}
	}
	if strings.Index(text, "(Hahn tropft)") > strings.Index(text, "(Lampe defekt)") {
		t.Error("Druck hält die Reihenfolge der Liste nicht ein")
	}
	if strings.Contains(text, "lehrer") || !bytes.Contains(roh, []byte("/Subtype /Image")) {
		t.Error("Druck: Name der meldenden Person oder fehlendes Foto")
	}

	// Eine einzelne heißt nach ihrer Nummer; die Verwaltung darf drucken, das Kollegium nicht.
	_, kopf, _ = sek.holeDatei(fmt.Sprintf("/api/druck.pdf?ids=%v", lampe["id"]))
	if !strings.Contains(kopf.Get("Content-Disposition"), fmt.Sprintf(`"Schadenmeldung-%v-2026-11-02.pdf"`, lampe["id"])) {
		t.Errorf("Druck einer Meldung durch die Verwaltung: %v", kopf)
	}
	if status, _, _ := lehrer.holeDatei(fmt.Sprintf("/api/druck.pdf?ids=%v", lampe["id"])); status != http.StatusForbidden {
		t.Errorf("Kollegium druckt: Status %d", status)
	}
	viele := zuVieleIDs()
	for pfad, soll := range map[string]int{"": http.StatusBadRequest, "?ids=abc": http.StatusBadRequest, "?ids=" + viele: http.StatusBadRequest, "?ids=999999": http.StatusNotFound} {
		if status, _, _ := hm.holeDatei("/api/druck.pdf" + pfad); status != soll {
			t.Errorf("/api/druck.pdf%s: Status %d statt %d", pfad[:min(len(pfad), 20)], status, soll)
		}
	}
}

func TestDruckIDs(t *testing.T) {
	viele := strings.TrimSuffix(strings.Repeat("7,", druckHoechstens+1), ",")
	for text, soll := range map[string][]int64{"3,1,2": {3, 1, 2}, " 4 , 5 ": {4, 5}, "2,2,1": {2, 1}, "": nil, "1,x": nil, "0": nil, "-3": nil, "1,,2": nil} {
		ids, ok := druckIDs(text)
		if !slices.Equal(ids, soll) || ok != (soll != nil) {
			t.Errorf("druckIDs(%q) = %v, %v; soll %v", text, ids, ok, soll)
		}
	}
	// Doppelte zählen einmal: 51-mal dieselbe ist eine Meldung, also erlaubt; 51 verschiedene sind zu viele.
	if ids, ok := druckIDs(viele); !ok || len(ids) != 1 {
		t.Errorf("druckIDs mit %d gleichen IDs: %v, %v", druckHoechstens+1, ids, ok)
	}
	if _, ok := druckIDs(zuVieleIDs()); ok {
		t.Errorf("druckIDs nimmt %d verschiedene IDs an", druckHoechstens+1)
	}
}

// zuVieleIDs: eine verschiedene ID mehr, als auf einmal gedruckt werden darf.
func zuVieleIDs() string {
	teile := make([]string, druckHoechstens+1)
	for i := range teile {
		teile[i] = fmt.Sprint(i + 1)
	}
	return strings.Join(teile, ",")
}
