package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

// Die Übersicht hört auf /api/ereignisse und erfährt jede Änderung sofort. Das Kollegium bekommt
// keinen Strom, und beim Herunterfahren endet er, sonst wartete Shutdown vergeblich.
func TestEreignisstromMeldetJedeAenderung(t *testing.T) {
	u := neueTestUmgebung(t)
	t.Cleanup(u.app.verteiler.schliesse)
	lehrer, hm := u.nutzer("lehrer"), u.nutzer("hm")

	if status, _ := lehrer.json("GET", "/api/ereignisse", nil); status != http.StatusForbidden {
		t.Fatalf("Kollegium: Status %d, erwartet 403", status)
	}

	zeilen := hm.strom("/api/ereignisse")
	warteAuf := func(erwartet string) {
		t.Helper()
		frist := time.After(2 * time.Second)
		for {
			select {
			case z, offen := <-zeilen:
				if !offen {
					t.Fatalf("Strom zu, bevor %q kam", erwartet)
				}
				if z == erwartet {
					return
				}
			case <-frist:
				t.Fatalf("%q kam nicht binnen 2 Sekunden", erwartet)
			}
		}
	}
	warteAuf(": verbunden")

	status, m := lehrer.melde(map[string]string{"ort": "Raum 204", "kategorie": "licht", "was": "Lampe defekt"}, nil)
	if status != http.StatusCreated {
		t.Fatalf("Melden: Status %d, %v", status, m)
	}
	warteAuf("event: meldungen")

	pfad := fmt.Sprintf("/api/meldungen/%v/status", m["id"])
	if status, antwort := hm.json("POST", pfad, map[string]string{"status": "gesehen"}); status != http.StatusOK {
		t.Fatalf("Gesehen: Status %d, %v", status, antwort)
	}
	warteAuf("event: meldungen")

	u.app.verteiler.schliesse()
	frist := time.After(2 * time.Second)
	for {
		select {
		case _, offen := <-zeilen:
			if !offen {
				return
			}
		case <-frist:
			t.Fatal("Strom blieb nach dem Herunterfahren offen")
		}
	}
}

// strom öffnet einen Ereignisstrom und liefert seine Zeilen, bis der Server ihn schließt.
func (n *testNutzer) strom(pfad string) <-chan string {
	n.u.t.Helper()
	ctx, abbrechen := context.WithCancel(context.Background())
	n.u.t.Cleanup(abbrechen)
	req, err := http.NewRequestWithContext(ctx, "GET", n.u.srv.URL+pfad, nil)
	if err != nil {
		n.u.t.Fatal(err)
	}
	resp, err := n.client.Do(req)
	if err != nil {
		n.u.t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK || !strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		n.u.t.Fatalf("Strom: Status %d, Typ %q", resp.StatusCode, resp.Header.Get("Content-Type"))
	}
	zeilen := make(chan string, 16)
	go func() {
		defer close(zeilen)
		defer resp.Body.Close()
		leser := bufio.NewScanner(resp.Body)
		for leser.Scan() {
			zeilen <- leser.Text()
		}
	}()
	return zeilen
}
