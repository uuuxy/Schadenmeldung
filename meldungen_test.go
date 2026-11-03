package main

import (
	"bytes"
	"context"
	"image/jpeg"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSchreibendeAnfrageBrauchtKopf(t *testing.T) {
	u := neueTestUmgebung(t)
	resp, err := http.Post(u.srv.URL+"/api/anmelden", "application/json",
		strings.NewReader(`{"email":"lehrer","passwort":"p"}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("ohne X-Schaden: Status %d, erwartet 403", resp.StatusCode)
	}
}

func TestMeldenAbhakenWiederOeffnen(t *testing.T) {
	u := neueTestUmgebung(t)
	lehrer, hm := u.nutzer("lehrer"), u.nutzer("hm")

	status, m := lehrer.melde(map[string]string{"ort": "Raum 204", "kategorie": "technik", "was": "Beamer/Display aus", "dringend": "1"}, nil)
	if status != http.StatusCreated {
		t.Fatalf("Melden: Status %d, %v", status, m)
	}
	if m["gemeldetVon"] != nil {
		t.Errorf("Kollegium sieht gemeldetVon: %v", m["gemeldetVon"])
	}
	if gemeldet, _ := time.Parse(time.RFC3339, m["gemeldetAm"].(string)); !gemeldet.Equal(u.uhr) {
		t.Errorf("gemeldetAm %v, erwartet die Uhr des Servers %v", m["gemeldetAm"], u.uhr)
	}
	id := int(m["id"].(float64))
	pfad := "/api/meldungen/" + strconv.Itoa(id) + "/status"

	if status, _ := lehrer.json("POST", pfad, map[string]string{"status": "erledigt"}); status != http.StatusForbidden {
		t.Fatalf("Kollegium hakt ab: Status %d, erwartet 403", status)
	}
	if status, neu := hm.json("POST", pfad, map[string]string{"status": "bestellt", "antwort": "  Lampe kommt Freitag "}); status != 200 ||
		neu["status"] != "bestellt" || neu["antwort"] != "Lampe kommt Freitag" || neu["gemeldetVon"] != "lehrer@schule.de" {
		t.Fatalf("Status setzen: %d, %v", status, neu)
	}

	u.uhr = u.uhr.Add(2 * time.Hour)
	if status, neu := hm.json("POST", pfad, map[string]string{"status": "erledigt"}); status != 200 || neu["erledigtAm"] == nil || neu["antwort"] != "Lampe kommt Freitag" {
		t.Fatalf("Erledigt: %d, %v", status, neu)
	}
	_, alle := hm.json("GET", "/api/meldungen", nil)
	if len(eintraege(alle, "offen")) != 0 || len(eintraege(alle, "erledigt")) != 1 {
		t.Fatalf("nach Erledigt: %v", alle)
	}

	if status, neu := hm.json("POST", pfad, map[string]string{"status": "offen"}); status != 200 || neu["erledigtAm"] != nil {
		t.Fatalf("Wieder öffnen: %d, %v", status, neu)
	}
}

func TestMeldungBrauchtOrtArtUndBeschreibung(t *testing.T) {
	u := neueTestUmgebung(t)
	lehrer := u.nutzer("lehrer")
	for _, felder := range []map[string]string{
		{"kategorie": "licht", "was": "Lampe defekt"},
		{"ort": "204", "kategorie": "erfunden", "was": "x"},
		{"ort": "204", "kategorie": "licht", "was": "  "},
		{"raumId": "999", "kategorie": "licht", "was": "Lampe defekt"},
	} {
		if status, _ := lehrer.melde(felder, nil); status < 400 {
			t.Errorf("%v: Status %d, erwartet Ablehnung", felder, status)
		}
	}
}

func TestFotoVerliertMetadatenUndWirdKleiner(t *testing.T) {
	u := neueTestUmgebung(t)
	lehrer := u.nutzer("lehrer")
	status, m := lehrer.melde(map[string]string{"ort": "WC 1. OG", "kategorie": "wasser", "was": "Abfluss verstopft"}, fotoMitExif(t, 2400, 1200))
	if status != http.StatusCreated {
		t.Fatalf("Melden mit Foto: %d, %v", status, m)
	}
	url := m["foto"].(string)
	gespeichert, err := u.app.sp.foto(context.Background(), strings.TrimPrefix(url, "/fotos/"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(gespeichert, []byte("Exif")) || bytes.Contains(gespeichert, []byte("GPS")) {
		t.Error("gespeichertes Foto enthält noch Metadaten")
	}
	if cfg, err := jpeg.DecodeConfig(bytes.NewReader(gespeichert)); err != nil || cfg.Width != 1600 || cfg.Height != 800 {
		t.Errorf("Maße %dx%d (%v), erwartet 1600x800", cfg.Width, cfg.Height, err)
	}

	// Fotos gibt es nur mit Anmeldung.
	resp, err := http.Get(u.srv.URL + url)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Foto ohne Anmeldung: Status %d, erwartet 401", resp.StatusCode)
	}
}

func TestAufraeumenSetztLoeschfristenUm(t *testing.T) {
	u := neueTestUmgebung(t)
	lehrer, hm := u.nutzer("lehrer"), u.nutzer("hm")
	_, m := lehrer.melde(map[string]string{"ort": "Aula", "kategorie": "moebel", "was": "Stuhl kaputt"}, fotoMitExif(t, 100, 100))
	_, offenBleibt := lehrer.melde(map[string]string{"ort": "Aula", "kategorie": "licht", "was": "Lampe defekt"}, fotoMitExif(t, 100, 100))
	id := int(m["id"].(float64))
	hm.json("POST", "/api/meldungen/"+strconv.Itoa(id)+"/status", map[string]string{"status": "erledigt"})
	fotoName := strings.TrimPrefix(m["foto"].(string), "/fotos/")

	u.uhr = u.uhr.AddDate(0, 0, 31)
	u.app.raeumeAuf(context.Background())
	if _, err := u.app.sp.foto(context.Background(), fotoName); err != errNichtGefunden {
		t.Errorf("Foto der erledigten Meldung nach 31 Tagen: %v, erwartet gelöscht", err)
	}
	if alt, err := u.app.sp.meldung(context.Background(), int64(id)); err != nil || alt.Foto != "" {
		t.Errorf("Meldung nach 31 Tagen: %v, %v", alt, err)
	}

	u.uhr = u.uhr.AddDate(0, 0, 335)
	u.app.raeumeAuf(context.Background())
	if _, err := u.app.sp.meldung(context.Background(), int64(id)); err != errNichtGefunden {
		t.Errorf("erledigte Meldung nach 366 Tagen: %v, erwartet gelöscht", err)
	}
	if offen, err := u.app.sp.meldung(context.Background(), int64(offenBleibt["id"].(float64))); err != nil || offen.Foto == "" {
		t.Errorf("offene Meldung darf nicht aufgeräumt werden: %v, %v", offen, err)
	}
}

func TestOberflaecheBeantwortetEigeneWege(t *testing.T) {
	u := neueTestUmgebung(t)
	for pfad, erwartet := range map[string]int{"/": 200, "/r/abcd1234": 200, "/uebersicht": 200, "/fehlt.js": 404} {
		resp, err := http.Get(u.srv.URL + pfad)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != erwartet {
			t.Errorf("%s: Status %d, erwartet %d", pfad, resp.StatusCode, erwartet)
		}
	}
	resp, err := http.Get(u.srv.URL + "/api/gibtesnicht")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 404 || !strings.Contains(resp.Header.Get("Content-Type"), "json") {
		t.Errorf("unbekannte API: Status %d, Typ %s", resp.StatusCode, resp.Header.Get("Content-Type"))
	}
}

func TestZeitpunktAusDemAusgang(t *testing.T) {
	u := neueTestUmgebung(t)
	lehrer := u.nutzer("lehrer")
	for name, fall := range map[string]struct {
		erstellt time.Time
		gilt     bool
	}{
		"vor 40 Minuten im Funkloch": {u.uhr.Add(-40 * time.Minute), true},
		"vor drei Tagen":             {u.uhr.AddDate(0, 0, -3), false},
		"in der Zukunft":             {u.uhr.Add(time.Hour), false},
	} {
		_, m := lehrer.melde(map[string]string{"ort": "Aula", "kategorie": "licht", "was": "Lampe defekt",
			"erstellt": fall.erstellt.Format(time.RFC3339)}, nil)
		gemeldet, _ := time.Parse(time.RFC3339, m["gemeldetAm"].(string))
		erwartet := u.uhr
		if fall.gilt {
			erwartet = fall.erstellt
		}
		if !gemeldet.Equal(erwartet) {
			t.Errorf("%s: gemeldetAm %v, erwartet %v", name, gemeldet, erwartet)
		}
	}
}
