package main

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// testSpeicher legt ein frisches Schema an, spielt die Migrationen ein und räumt danach auf.
func testSpeicher(t *testing.T) *speicher {
	t.Helper()
	// Ohne Datenbank schlagen die Tests fehl statt still grün zu überspringen; wer bewusst ohne
	// testet, setzt OHNE_DB=1.
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		if os.Getenv("OHNE_DB") == "1" {
			t.Skip("OHNE_DB=1: Test mit Datenbank übersprungen")
		}
		t.Fatal("TEST_DATABASE_URL fehlt. Start: docker compose -f docker-compose.local.yml up -d, dann " +
			"TEST_DATABASE_URL=postgres://schaden:schaden@localhost:5435/schaden?sslmode=disable go test ./...")
	}
	ctx := context.Background()
	schema := "test_" + zufallsHex(6)
	verwaltung, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := verwaltung.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal(err)
	}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	sp, err := oeffneMitKonfiguration(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		sp.schliesse()
		_, _ = verwaltung.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE")
		_ = verwaltung.Close(ctx)
	})
	return sp
}

type testUmgebung struct {
	t    *testing.T
	app  *app
	srv  *httptest.Server
	uhr  time.Time
	post *testPost
}

// testPost fängt die Mails ab, statt sie zu verschicken.
type testPost struct {
	mu    sync.Mutex
	mails []testMail
}

type testMail struct{ An, Betreff, Text, HTML string }

func (p *testPost) sende(_ context.Context, an, betreff, text, html string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.mails = append(p.mails, testMail{an, betreff, text, html})
	return nil
}

func (p *testPost) alle() []testMail {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]testMail(nil), p.mails...)
}

var berlin, _ = time.LoadLocation("Europe/Berlin")

// neueTestUmgebung: Montag, 2. November 2026, 9 Uhr in Berlin. hm@schule.de ist Hausmeister,
// admin@schule.de fest Verwaltung (ADMINS), Mail ist an.
func neueTestUmgebung(t *testing.T) *testUmgebung {
	t.Helper()
	konf := konfiguration{
		Lokal:               true,
		IMAPHost:            "mock",
		Domains:             []string{"schule.de"},
		Admins:              map[string]bool{"admin@schule.de": true},
		FotoTage:            30,
		AufbewahrungTage:    365,
		SMTPHost:            "smtp.test",
		MailAbsender:        "schaden@schule.de",
		OeffentlicheAdresse: "https://schaden.schule.de",
		SammelZeiten:        []uhrzeit{{7, 0}, {12, 0}},
		Zeitzone:            berlin,
	}
	statisch := fstest.MapFS{"index.html": {Data: []byte("<!doctype html><title>Schadenmeldung</title>")}}
	a := neueApp(konf, testSpeicher(t), statisch)
	u := &testUmgebung{t: t, app: a, uhr: time.Date(2026, 11, 2, 9, 0, 0, 0, berlin), post: &testPost{}}
	a.jetzt = func() time.Time { return u.uhr }
	a.versender = u.post
	if err := a.sp.setzeRolle(context.Background(), "hm@schule.de", rolleHausmeister, "test"); err != nil {
		t.Fatal(err)
	}
	u.srv = httptest.NewServer(a.routen())
	t.Cleanup(u.srv.Close)
	return u
}

// anonym ruft ohne Anmeldung auf, wie ein Knopf aus der Mail.
func (u *testUmgebung) anonym() *testNutzer {
	jar, _ := cookiejar.New(nil)
	return &testNutzer{u: u, client: &http.Client{Jar: jar}}
}

type testNutzer struct {
	u      *testUmgebung
	client *http.Client
}

// nutzer meldet sich mit dem Namen vor dem @ an; die Domain ergänzt der Server.
func (u *testUmgebung) nutzer(name string) *testNutzer {
	u.t.Helper()
	jar, _ := cookiejar.New(nil)
	n := &testNutzer{u: u, client: &http.Client{Jar: jar}}
	if status, antwort := n.json("POST", "/api/anmelden", map[string]string{"email": name, "passwort": "geheim"}); status != 200 {
		u.t.Fatalf("Anmeldung %s: Status %d, %v", name, status, antwort)
	}
	return n
}

func (n *testNutzer) anfrage(methode, pfad string, body io.Reader, typ string) (int, map[string]any) {
	n.u.t.Helper()
	req, err := http.NewRequest(methode, n.u.srv.URL+pfad, body)
	if err != nil {
		n.u.t.Fatal(err)
	}
	req.Header.Set("X-Schaden", "1")
	if typ != "" {
		req.Header.Set("Content-Type", typ)
	}
	resp, err := n.client.Do(req)
	if err != nil {
		n.u.t.Fatal(err)
	}
	defer resp.Body.Close()
	var aus map[string]any
	roh, _ := io.ReadAll(resp.Body)
	if len(roh) > 0 && strings.Contains(resp.Header.Get("Content-Type"), "json") {
		if roh[0] == '{' {
			_ = json.Unmarshal(roh, &aus)
		} else {
			aus = map[string]any{"liste": nil}
			var liste []any
			_ = json.Unmarshal(roh, &liste)
			aus["liste"] = liste
		}
	}
	return resp.StatusCode, aus
}

func (n *testNutzer) json(methode, pfad string, ein any) (int, map[string]any) {
	var body io.Reader
	if ein != nil {
		roh, _ := json.Marshal(ein)
		body = bytes.NewReader(roh)
	}
	return n.anfrage(methode, pfad, body, "application/json")
}

func (n *testNutzer) melde(felder map[string]string, foto []byte) (int, map[string]any) {
	n.u.t.Helper()
	var puffer bytes.Buffer
	mw := multipart.NewWriter(&puffer)
	for k, v := range felder {
		_ = mw.WriteField(k, v)
	}
	if foto != nil {
		fw, _ := mw.CreateFormFile("foto", "foto.jpg")
		_, _ = fw.Write(foto)
	}
	_ = mw.Close()
	return n.anfrage("POST", "/api/meldungen", &puffer, mw.FormDataContentType())
}

func eintraege(antwort map[string]any, schluessel string) []map[string]any {
	var aus []map[string]any
	roh, _ := antwort[schluessel].([]any)
	for _, e := range roh {
		aus = append(aus, e.(map[string]any))
	}
	return aus
}

func TestAnmeldungErgaenztDomainUndVergibtRollen(t *testing.T) {
	u := neueTestUmgebung(t)
	for name, rolle := range map[string]string{"lehrer": "kollegium", "hm": "hausmeister", "Admin": "admin"} {
		n := u.nutzer(name)
		status, ich := n.json("GET", "/api/ich", nil)
		if status != 200 || ich["rolle"] != rolle || !strings.HasSuffix(ich["email"].(string), "@schule.de") {
			t.Errorf("%s: Status %d, %v; erwartet Rolle %s", name, status, ich, rolle)
		}
	}
}

func TestFremdeDomainKommtNichtHinein(t *testing.T) {
	u := neueTestUmgebung(t)
	jar, _ := cookiejar.New(nil)
	n := &testNutzer{u: u, client: &http.Client{Jar: jar}}
	status, _ := n.json("POST", "/api/anmelden", map[string]string{"email": "x@gmail.com", "passwort": "p"})
	if status != http.StatusForbidden {
		t.Fatalf("Status %d, erwartet 403", status)
	}
	if status, _ := n.json("GET", "/api/ich", nil); status != http.StatusUnauthorized {
		t.Fatalf("ohne Anmeldung: Status %d, erwartet 401", status)
	}
}

func TestBremseSperrtNachFehlversuchenUndZaehltAusfallNicht(t *testing.T) {
	u := neueTestUmgebung(t)
	var antwort error
	u.app.pruefeZugang = func(context.Context, string, string) error { return antwort }
	jar, _ := cookiejar.New(nil)
	n := &testNutzer{u: u, client: &http.Client{Jar: jar}}
	anmelden := func() int {
		status, _ := n.json("POST", "/api/anmelden", map[string]string{"email": "lehrer", "passwort": "falsch"})
		return status
	}

	antwort = errMailserverWeg
	for range 6 {
		if s := anmelden(); s != http.StatusServiceUnavailable {
			t.Fatalf("Ausfall: Status %d, erwartet 503", s)
		}
	}
	antwort = errAbgelehnt
	for i := range 5 {
		if s := anmelden(); s != http.StatusUnauthorized {
			t.Fatalf("Fehlversuch %d: Status %d, erwartet 401", i+1, s)
		}
	}
	antwort = nil
	if s := anmelden(); s != http.StatusTooManyRequests {
		t.Fatalf("nach 5 Fehlversuchen: Status %d, erwartet 429 auch mit richtigem Passwort", s)
	}
	u.uhr = u.uhr.Add(16 * time.Minute)
	if s := anmelden(); s != http.StatusOK {
		t.Fatalf("nach Ablauf der Sperre: Status %d, erwartet 200", s)
	}
}

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
	_, alle := lehrer.json("GET", "/api/meldungen", nil)
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

func TestRaeumeNurDurchVerwaltung(t *testing.T) {
	u := neueTestUmgebung(t)
	admin, lehrer := u.nutzer("admin"), u.nutzer("lehrer")
	text := "204; Hauptgebäude · 2. OG\nTurnhalle\tSport\n\n204; doppelt\n"
	if status, _ := lehrer.json("POST", "/api/raeume", map[string]string{"text": text}); status != http.StatusForbidden {
		t.Fatalf("Kollegium legt Räume an: Status %d", status)
	}
	status, antwort := admin.json("POST", "/api/raeume", map[string]string{"text": text})
	if status != 200 || antwort["neu"].(float64) != 2 {
		t.Fatalf("Anlegen: %d, %v", status, antwort)
	}
	raeume := eintraege(antwort, "raeume")
	if raeume[0]["name"] != "204" || raeume[0]["bereich"] != "Hauptgebäude · 2. OG" || len(raeume[0]["code"].(string)) != 8 {
		t.Fatalf("Räume: %v", raeume)
	}

	// Eine Meldung behält ihren Ort, auch wenn der Raum später gelöscht wird.
	raumID := int(raeume[0]["id"].(float64))
	if status, m := lehrer.melde(map[string]string{"raumId": strconv.Itoa(raumID), "kategorie": "licht", "was": "Lampe defekt"}, nil); status != 201 || m["ort"] != "204" || m["bereich"] != "Hauptgebäude · 2. OG" {
		t.Fatalf("Meldung mit Raum: %d, %v", status, m)
	}
	if status, _ := admin.json("DELETE", "/api/raeume/"+strconv.Itoa(raumID), nil); status != http.StatusNoContent {
		t.Fatalf("Löschen: Status %d", status)
	}
	_, alle := lehrer.json("GET", "/api/meldungen", nil)
	if offen := eintraege(alle, "offen"); len(offen) != 1 || offen[0]["ort"] != "204" || offen[0]["raumId"] != nil {
		t.Fatalf("nach Löschen des Raums: %v", offen)
	}
}

// fotoMitExif baut ein JPEG mit einem EXIF-Block, wie ihn Handykameras samt GPS-Ort schreiben.
func fotoMitExif(t *testing.T, breite, hoehe int) []byte {
	t.Helper()
	bild := image.NewRGBA(image.Rect(0, 0, breite, hoehe))
	for x := range breite {
		bild.Set(x, hoehe/2, color.RGBA{R: 200, A: 255})
	}
	var jpg bytes.Buffer
	if err := jpeg.Encode(&jpg, bild, nil); err != nil {
		t.Fatal(err)
	}
	nutzlast := append([]byte("Exif\x00\x00"), []byte("GPS 50.2213N 8.6198E")...)
	segment := append([]byte{0xFF, 0xE1, byte((len(nutzlast) + 2) >> 8), byte(len(nutzlast) + 2)}, nutzlast...)
	roh := jpg.Bytes()
	return append(append(append([]byte{}, roh[:2]...), segment...), roh[2:]...)
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

func TestKonfigurationVerweigertOffeneTueren(t *testing.T) {
	const db = "postgres://schaden@localhost/schaden"
	for name, fall := range map[string]struct {
		k  konfiguration
		ok bool
	}{
		"mock auf dem Server":   {konfiguration{DatenbankURL: db, IMAPHost: "mock"}, false},
		"ohne Mailserver":       {konfiguration{DatenbankURL: db}, false},
		"ohne Datenbank":        {konfiguration{IMAPHost: "mail.schule.de"}, false},
		"mock lokal":            {konfiguration{DatenbankURL: db, IMAPHost: "mock", Lokal: true}, true},
		"Mailserver der Schule": {konfiguration{DatenbankURL: db, IMAPHost: "mail.schule.de"}, true},
	} {
		if err := fall.k.pruefe(); (err == nil) != fall.ok {
			t.Errorf("%s: pruefe() = %v, erwartet ok=%v", name, err, fall.ok)
		}
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

func TestDringendesGehtSofortRaus(t *testing.T) {
	u := neueTestUmgebung(t)
	lehrer := u.nutzer("lehrer")
	lehrer.melde(map[string]string{"ort": "Aula", "kategorie": "licht", "was": "Lampe defekt"}, nil)
	lehrer.melde(map[string]string{"ort": "WC 1. OG", "kategorie": "wasser", "was": "Wasser läuft aus", "dringend": "1"}, nil)
	u.app.hintergrund.Wait()

	mails := u.post.alle()
	if len(mails) != 1 {
		t.Fatalf("%d Mails, erwartet genau eine (nur die dringende)", len(mails))
	}
	m := mails[0]
	if m.An != "hm@schule.de" || m.Betreff != "Dringend: Wasser läuft aus – WC 1. OG" {
		t.Errorf("Mail an %q mit Betreff %q", m.An, m.Betreff)
	}
	if !strings.Contains(m.HTML, "https://schaden.schule.de/aktion/") || !strings.Contains(m.Text, "?tun=erledigt") {
		t.Error("Mail ohne Knöpfe „Gesehen“/„Erledigt“")
	}
	// Die schon verschickte dringende Meldung kommt in der Sammelmail nicht noch einmal.
	neu, err := u.app.sp.ungesendeteFuer(context.Background(), "hm@schule.de", u.uhr)
	if err != nil || len(neu) != 1 || neu[0].Was != "Lampe defekt" {
		t.Errorf("für die Sammelmail übrig: %v, %v", neu, err)
	}
}

func TestSammelmailZuDenTerminenOhneDoppelte(t *testing.T) {
	u := neueTestUmgebung(t)
	ctx := context.Background()
	u.uhr = time.Date(2026, 11, 2, 6, 0, 0, 0, berlin)
	u.app.pruefeSammelmail(ctx) // erster Start: zählt ab jetzt

	lehrer, hm := u.nutzer("lehrer"), u.nutzer("hm")
	u.uhr = u.uhr.Add(10 * time.Minute)
	lehrer.melde(map[string]string{"ort": "204", "kategorie": "technik", "was": "Kein Ton"}, nil)
	lehrer.melde(map[string]string{"ort": "Aula", "kategorie": "licht", "was": "Lampe defekt"}, nil)
	hm.melde(map[string]string{"ort": "Keller", "kategorie": "sonstiges", "was": "Eigene Notiz"}, nil)

	u.uhr = time.Date(2026, 11, 2, 7, 1, 0, 0, berlin)
	u.app.pruefeSammelmail(ctx)
	u.app.pruefeSammelmail(ctx)
	mails := u.post.alle()
	if len(mails) != 1 || mails[0].Betreff != "2 neue Schadensmeldungen" {
		t.Fatalf("um 7 Uhr: %d Mails %v, erwartet eine mit zwei Meldungen (nicht die eigene)", len(mails), betreffe(mails))
	}

	u.uhr = time.Date(2026, 11, 2, 12, 1, 0, 0, berlin)
	u.app.pruefeSammelmail(ctx)
	if len(u.post.alle()) != 1 {
		t.Fatal("um 12 Uhr kam eine Mail ohne neue Meldungen")
	}

	// Samstag: keine Sammelmail; was am Wochenende kommt, steht Montag früh in der Mail.
	u.uhr = time.Date(2026, 11, 7, 10, 0, 0, 0, berlin)
	lehrer.melde(map[string]string{"ort": "Turnhalle", "kategorie": "fenster", "was": "Schloss defekt"}, nil)
	u.uhr = time.Date(2026, 11, 7, 12, 30, 0, 0, berlin)
	u.app.pruefeSammelmail(ctx)
	if len(u.post.alle()) != 1 {
		t.Fatal("am Samstag kam eine Sammelmail")
	}
	u.uhr = time.Date(2026, 11, 9, 7, 0, 30, 0, berlin)
	u.app.pruefeSammelmail(ctx)
	if mails := u.post.alle(); len(mails) != 2 || !strings.Contains(mails[1].Betreff, "Schloss defekt") {
		t.Fatalf("Montag 7 Uhr: %v", betreffe(mails))
	}
}

func betreffe(mails []testMail) []string {
	var liste []string
	for _, m := range mails {
		liste = append(liste, m.Betreff)
	}
	return liste
}

var linkMuster = regexp.MustCompile(`/aktion/([A-Za-z0-9_-]+)\?tun=erledigt`)

func TestKnopfAusDerMail(t *testing.T) {
	u := neueTestUmgebung(t)
	u.nutzer("lehrer").melde(map[string]string{"ort": "WC 1. OG", "kategorie": "wasser", "was": "WC verstopft", "dringend": "1"}, nil)
	u.app.hintergrund.Wait()
	treffer := linkMuster.FindStringSubmatch(u.post.alle()[0].Text)
	if treffer == nil {
		t.Fatal("kein Erledigt-Link in der Mail")
	}
	pfad := "/api/aktion/" + treffer[1]
	gast := u.anonym()

	if status, m := gast.json("GET", pfad, nil); status != 200 || m["was"] != "WC verstopft" || m["gemeldetVon"] != "lehrer@schule.de" {
		t.Fatalf("Link öffnen: %d, %v", status, m)
	}
	resp, err := http.Post(u.srv.URL+pfad, "application/json", strings.NewReader(`{"status":"erledigt"}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("ohne X-Schaden: Status %d, erwartet 403 (Schutz vor fremden Formularen)", resp.StatusCode)
	}
	if status, m := gast.json("POST", pfad, map[string]string{"status": "erledigt"}); status != 200 || m["status"] != "erledigt" {
		t.Fatalf("Erledigt aus der Mail: %d, %v", status, m)
	}

	// Wer die Rolle verliert, kann auch mit einem alten Link nichts mehr ändern.
	if err := u.app.sp.setzeRolle(context.Background(), "hm@schule.de", rolleLeitung, "test"); err != nil {
		t.Fatal(err)
	}
	if status, _ := gast.json("POST", pfad, map[string]string{"status": "offen"}); status != http.StatusForbidden {
		t.Fatalf("nach Rollenwechsel: Status %d, erwartet 403", status)
	}
	u.uhr = u.uhr.Add(15 * 24 * time.Hour)
	if status, _ := gast.json("GET", pfad, nil); status != http.StatusGone {
		t.Fatalf("nach 15 Tagen: Status %d, erwartet 410", status)
	}
}

func TestRollenInDerApp(t *testing.T) {
	u := neueTestUmgebung(t)
	admin, lehrer := u.nutzer("admin"), u.nutzer("lehrer")
	lehrer.melde(map[string]string{"ort": "Aula", "kategorie": "licht", "was": "Lampe defekt"}, nil)

	if status, _ := lehrer.json("POST", "/api/rollen", map[string]string{"email": "lehrer", "rolle": "admin"}); status != http.StatusForbidden {
		t.Fatalf("Kollegium vergibt Rollen: Status %d", status)
	}
	status, antwort := admin.json("POST", "/api/rollen", map[string]string{"email": "direktorin", "rolle": "leitung"})
	if status != 200 || len(eintraege(antwort, "liste")) != 3 {
		t.Fatalf("Rolle vergeben: %d, %v", status, antwort)
	}

	leitung := u.nutzer("direktorin")
	_, ich := leitung.json("GET", "/api/ich", nil)
	_, alle := leitung.json("GET", "/api/meldungen", nil)
	offen := eintraege(alle, "offen")
	if ich["rolle"] != "leitung" || len(offen) != 1 || offen[0]["gemeldetVon"] != "lehrer@schule.de" {
		t.Fatalf("Leitung sieht: %v, %v", ich, offen)
	}
	id := strconv.Itoa(int(offen[0]["id"].(float64)))
	if status, _ := leitung.json("POST", "/api/meldungen/"+id+"/status", map[string]string{"status": "erledigt"}); status != http.StatusForbidden {
		t.Fatalf("Leitung ändert Status: %d, erwartet 403 (nur Einsicht)", status)
	}

	if status, _ := admin.json("POST", "/api/rollen", map[string]string{"email": "admin", "rolle": "leitung"}); status != http.StatusBadRequest {
		t.Fatalf("feste Verwaltung umgestuft: Status %d", status)
	}
	admin.json("POST", "/api/rollen", map[string]string{"email": "zweit", "rolle": "admin"})
	if status, _ := u.nutzer("zweit").json("DELETE", "/api/rollen/zweit@schule.de", nil); status != http.StatusBadRequest {
		t.Fatalf("eigene Rolle gelöscht: Status %d, erwartet 400", status)
	}
	if status, _ := admin.json("DELETE", "/api/rollen/direktorin@schule.de", nil); status != 200 {
		t.Fatalf("Rolle löschen: %d", status)
	}
	if _, ich := leitung.json("GET", "/api/ich", nil); ich["rolle"] != "kollegium" {
		t.Fatalf("nach dem Löschen: %v, erwartet kollegium", ich)
	}
}

func TestLetzterTermin(t *testing.T) {
	zeiten := []uhrzeit{{7, 0}, {12, 0}}
	for name, fall := range map[string]struct {
		jetzt time.Time
		ok    bool
		um    int
	}{
		"Montag vor 7":   {time.Date(2026, 11, 2, 6, 59, 0, 0, berlin), false, 0},
		"Montag 9 Uhr":   {time.Date(2026, 11, 2, 9, 0, 0, 0, berlin), true, 7},
		"Montag abends":  {time.Date(2026, 11, 2, 20, 0, 0, 0, berlin), true, 12},
		"Samstag mittag": {time.Date(2026, 11, 7, 12, 30, 0, 0, berlin), false, 0},
	} {
		termin, ok := letzterTermin(fall.jetzt, zeiten)
		if ok != fall.ok || (ok && termin.Hour() != fall.um) {
			t.Errorf("%s: %v %v, erwartet %v um %d Uhr", name, termin, ok, fall.ok, fall.um)
		}
	}
}
