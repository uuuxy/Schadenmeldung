package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

// Ohne Mail auf dem Server steht die Rückmeldung beim nächsten Öffnen im Formular; fremde erledigte
// Meldungen sieht niemand.
func TestRueckmeldungOhneMail(t *testing.T) {
	u := neueTestUmgebung(t)
	u.app.konf.SMTPHost = ""
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

// Mit Mail bekommt, wer gemeldet hat, eine Mail, sobald die Meldung zehn Minuten erledigt ist: einmal
// je Meldung und ohne den internen Vermerk. „Rückgängig“ in der Zeit hält sie auf, wer die eigene
// Meldung selbst erledigt, bekommt keine, und die Karte im Formular bleibt leer.
func TestRueckmeldungPerMail(t *testing.T) {
	u := neueTestUmgebung(t)
	ctx := context.Background()
	lehrer, hm := u.nutzer("lehrer"), u.nutzer("hm")
	melde := func(n *testNutzer, ort, was string) string {
		t.Helper()
		status, m := n.melde(map[string]string{"ort": ort, "kategorie": "licht", "was": was}, nil)
		if status != http.StatusCreated {
			t.Fatalf("Meldung: Status %d", status)
		}
		return "/api/meldungen/" + strconv.Itoa(int(m["id"].(float64))) + "/status"
	}
	stand := func(pfad string, ein map[string]string) {
		t.Helper()
		if status, _ := hm.json("POST", pfad, ein); status != http.StatusOK {
			t.Fatalf("Stand %v: Status %d", ein, status)
		}
	}
	erledigtMails := func() (l []testMail) {
		for _, m := range u.post.alle() {
			if strings.HasPrefix(m.Betreff, "Erledigt") {
				l = append(l, m)
			}
		}
		return l
	}

	lampe := melde(lehrer, "Raum 204", "Lampe flackert")
	versehen := melde(lehrer, "Aula", "Licht aus")
	eigene := melde(hm, "Keller", "Lampe kaputt")
	stand(lampe, map[string]string{"status": "erledigt", "antwort": "Vorschaltgerät getauscht"})
	stand(versehen, map[string]string{"status": "erledigt"})
	stand(versehen, map[string]string{"status": "offen"}) // „Rückgängig“
	stand(eigene, map[string]string{"status": "erledigt"})
	u.app.versendeAusstehende(ctx)
	if l := erledigtMails(); len(l) != 0 {
		t.Fatalf("Mail vor Ablauf der Wartezeit: %s", l[0].Betreff)
	}

	u.uhr = u.uhr.Add(rueckmeldungWartezeit)
	u.app.versendeAusstehende(ctx)
	u.app.versendeAusstehende(ctx)
	l := erledigtMails()
	if len(l) != 1 || l[0].An != "lehrer@schule.de" || l[0].Betreff != "Erledigt: Lampe flackert – Raum 204" {
		t.Fatalf("Rückmeldungen: %+v", l)
	}
	if !strings.Contains(l[0].Text, "gemeldet heute, 09:00") || strings.Contains(l[0].Text+l[0].HTML, "Vorschaltgerät") ||
		strings.Contains(l[0].Text+l[0].HTML, "Abbestellen") {
		t.Errorf("Inhalt der Rückmeldung:\n%s", l[0].Text)
	}
	_, sicht := lehrer.json("GET", "/api/meldungen", nil)
	if k, _ := sicht["meineErledigt"].([]any); len(k) != 0 {
		t.Errorf("Karte im Formular trotz Mail: %v", k)
	}

	// Wieder geöffnet und erneut erledigt: Es bleibt bei einer Mail je Meldung.
	stand(lampe, map[string]string{"status": "offen"})
	stand(lampe, map[string]string{"status": "erledigt"})
	u.uhr = u.uhr.Add(rueckmeldungWartezeit)
	u.app.versendeAusstehende(ctx)
	if l := erledigtMails(); len(l) != 1 {
		t.Errorf("zweite Rückmeldung zur selben Meldung: %d Mails", len(l))
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
