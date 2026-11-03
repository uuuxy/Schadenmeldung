package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
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

// testPost fängt die Mails ab, statt sie zu verschicken. Mit stoerung spielt sie einen Mailserver,
// der nicht antwortet.
type testPost struct {
	mu       sync.Mutex
	mails    []testMail
	stoerung error
}

type testMail struct{ An, Betreff, Text, HTML string }

func (p *testPost) sende(_ context.Context, an, betreff, text, html string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stoerung != nil {
		return p.stoerung
	}
	p.mails = append(p.mails, testMail{an, betreff, text, html})
	return nil
}

func (p *testPost) stoere(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.stoerung = err
}

func (p *testPost) alle() []testMail {
	p.mu.Lock()
	defer p.mu.Unlock()
	return append([]testMail(nil), p.mails...)
}

var berlin, _ = time.LoadLocation("Europe/Berlin")

// neueTestUmgebung: Montag, 2. November 2026, 9 Uhr in Berlin. hm@schule.de ist Hausmeister,
// admin@schule.de fest Verwaltung (ADMINS), kurz.de die Kurzform von schule.de, Mail ist an.
func neueTestUmgebung(t *testing.T) *testUmgebung {
	t.Helper()
	konf := konfiguration{
		Lokal:               true,
		IMAPHost:            "mock",
		Domains:             []string{"schule.de"},
		Aliase:              []string{"kurz.de"},
		Admins:              map[string]bool{"admin@schule.de": true},
		FotoTage:            30,
		AufbewahrungTage:    365,
		SMTPHost:            "smtp.test",
		MailAbsender:        "schaden@schule.de",
		OeffentlicheAdresse: "https://schaden.schule.de",
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
