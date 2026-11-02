package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

// Die Rechte je Rolle, wie der Nutzer sie festgelegt hat: Hausmeister, Leitung und Admin dürfen
// alles, die Verwaltung (Sekretariat) sieht nur, das Kollegium meldet nur. Geprüft an den Türen des
// Servers, denn die Oberfläche allein schützt nichts.
func TestRechteJeRolle(t *testing.T) {
	u := neueTestUmgebung(t)
	for email, rolle := range map[string]string{"chefin@schule.de": rolleLeitung, "sekretariat@schule.de": rolleVerwaltung} {
		if err := u.app.sp.setzeRolle(context.Background(), email, rolle, "test"); err != nil {
			t.Fatal(err)
		}
	}
	melder := u.nutzer("lehrer")
	for _, fall := range []struct {
		name           string
		sieht, aendert bool
	}{
		{"lehrer", false, false},
		{"sekretariat", true, false},
		{"hm", true, true},
		{"chefin", true, true},
		{"admin", true, true},
	} {
		n := u.nutzer(fall.name)
		_, m := melder.melde(map[string]string{"ort": "Flur", "kategorie": "licht", "was": "Lampe defekt"}, nil)
		id := fmt.Sprint(m["id"])
		pruefe := func(tuer string, darf bool, status, erlaubt int) {
			t.Helper()
			soll := http.StatusForbidden
			if darf {
				soll = erlaubt
			}
			if status != soll {
				t.Errorf("%s: %s gab %d, erwartet %d", fall.name, tuer, status, soll)
			}
		}
		status, _ := n.json("GET", "/api/benachrichtigung", nil)
		pruefe("Übersicht", fall.sieht, status, http.StatusOK)
		status, _, _ = n.holeDatei("/api/export.xlsx")
		pruefe("Excel-Export", fall.sieht, status, http.StatusOK)
		status, _, _ = n.holeDatei("/api/export.pdf")
		pruefe("PDF-Export", fall.sieht, status, http.StatusOK)
		status, _ = n.json("POST", "/api/arten", map[string]any{"kategorie": "technik", "sichtbar": true})
		pruefe("Arten schalten", fall.aendert, status, http.StatusOK)
		status, _ = n.json("POST", "/api/meldungen/"+id+"/status", map[string]string{"status": "gesehen"})
		pruefe("Stand setzen", fall.aendert, status, http.StatusOK)
		status, _ = n.json("POST", "/api/raeume", map[string]string{"text": "Raum " + fall.name})
		pruefe("Räume anlegen", fall.aendert, status, http.StatusOK)
		status, _ = n.json("GET", "/api/rollen", nil)
		pruefe("Rollen lesen", fall.aendert, status, http.StatusOK)
		status, _ = n.json("DELETE", "/api/meldungen/"+id, nil)
		pruefe("Meldung löschen", fall.aendert, status, http.StatusNoContent)
	}
}
