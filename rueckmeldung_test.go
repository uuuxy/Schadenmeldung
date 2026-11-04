package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"testing"
)

// Wer meldet, erfährt, dass die eigene Meldung erledigt ist; fremde erledigte Meldungen sieht niemand.
func TestRueckmeldungAnMeldende(t *testing.T) {
	u := neueTestUmgebung(t)
	lehrer := u.nutzer("lehrer")
	status, m := lehrer.melde(map[string]string{"ort": "Raum 204", "kategorie": "technik", "was": "Beamer/Display aus"}, nil)
	if status != http.StatusCreated {
		t.Fatalf("Meldung: Status %d", status)
	}
	pfad := "/api/meldungen/" + strconv.Itoa(int(m["id"].(float64))) + "/status"
	if status, _ := u.nutzer("hm").json("POST", pfad, map[string]string{"status": "erledigt"}); status != http.StatusOK {
		t.Fatalf("Erledigen: Status %d", status)
	}
	_, sicht := lehrer.json("GET", "/api/meldungen", nil)
	fertig, _ := sicht["meineErledigt"].([]any)
	if len(fertig) != 1 {
		t.Fatalf("eigene erledigte Meldungen: %v", sicht["meineErledigt"])
	}
	if e := fertig[0].(map[string]any); e["was"] != "Beamer/Display aus" || e["ort"] != "Raum 204" || e["erledigtAm"] == nil {
		t.Errorf("Rückmeldung: %v", e)
	}
	_, fremd := u.nutzer("kollegin").json("GET", "/api/meldungen", nil)
	if l, _ := fremd["meineErledigt"].([]any); len(l) != 0 {
		t.Errorf("fremde erledigte Meldung sichtbar: %v", l)
	}
}

// Scheitert der Versand, sieht das Backend, seit wann; klappt er wieder, verschwindet der Hinweis.
func TestMailStoerungSichtbar(t *testing.T) {
	u := neueTestUmgebung(t)
	ctx := context.Background()
	hm := u.nutzer("hm")
	stoerung := func(n *testNutzer) any {
		_, s := n.json("GET", "/api/meldungen", nil)
		return s["mailStoerungSeit"]
	}
	if s := stoerung(hm); s != nil {
		t.Fatalf("Störung ohne Grund: %v", s)
	}
	lehrer := u.nutzer("lehrer")
	u.post.stoere(errors.New("Postfach gesperrt"))
	lehrer.melde(map[string]string{"ort": "Aula", "kategorie": "licht", "was": "Lampe flackert"}, nil)
	u.app.versendeAusstehende(ctx)
	if stoerung(hm) == nil {
		t.Fatal("gescheiterter Versand nicht gemeldet")
	}
	if stoerung(lehrer) != nil {
		t.Error("Kollegium sieht die Störung")
	}
	u.post.stoere(nil)
	u.app.versendeAusstehende(ctx)
	if s := stoerung(hm); s != nil {
		t.Errorf("Hinweis bleibt nach geglücktem Versand: %v", s)
	}
}
