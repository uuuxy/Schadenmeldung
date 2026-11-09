package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"
	"testing"
)

func TestGeheimVerschluesseln(t *testing.T) {
	schluessel := bytes.Repeat([]byte{7}, schluesselLaenge)
	geheim, err := verschluesseleGeheim(schluessel, []byte("Sommer2026!"))
	if err != nil || bytes.Contains(geheim, []byte("Sommer2026!")) {
		t.Fatalf("verschlüsselt: %v, Klartext enthalten: %v", err, bytes.Contains(geheim, []byte("Sommer2026!")))
	}
	if klar, err := entschluesseleGeheim(schluessel, geheim); err != nil || string(klar) != "Sommer2026!" {
		t.Fatalf("zurück: %q, %v", klar, err)
	}
	if _, err := entschluesseleGeheim(bytes.Repeat([]byte{8}, schluesselLaenge), geheim); err == nil {
		t.Fatal("ein anderer Schlüssel darf nichts lesen")
	}
	geheim[len(geheim)-1] ^= 1
	if _, err := entschluesseleGeheim(schluessel, geheim); err == nil {
		t.Fatal("veränderte Daten müssen auffallen")
	}
	if _, err := verschluesseleGeheim(nil, []byte("x")); err == nil {
		t.Fatal("ohne SCHLUESSEL kein Speichern")
	}
	if _, err := schluesselAus("zu-kurz"); err == nil {
		t.Fatal("ein falscher SCHLUESSEL muss beim Start auffallen")
	}
}

// postfachUmgebung: kein MAIL_ABSENDER in der .env, also kommt das Postfach aus den Einstellungen;
// die Testmail kommt nur mit dem Passwort „richtig“ durch.
func postfachUmgebung(t *testing.T) (*testUmgebung, *[]string) {
	u := neueTestUmgebung(t)
	u.app.konf.MailAbsender = ""
	u.app.konf.Schluessel = bytes.Repeat([]byte{1}, schluesselLaenge)
	testmails := &[]string{}
	u.app.pruefeVersand = func(_ context.Context, z zugang, an string) error {
		if z.passwort != "richtig" {
			return &textproto.Error{Code: 535, Msg: "authentication failed"}
		}
		*testmails = append(*testmails, an)
		return nil
	}
	return u, testmails
}

func TestPostfachEintragen(t *testing.T) {
	u, testmails := postfachUmgebung(t)
	ctx := context.Background()
	hm := u.nutzer("hm")
	richtig := map[string]string{"adresse": "hm", "passwort": "richtig"}

	if status, _ := u.nutzer("lehrer").json("PUT", "/api/postfach", richtig); status != http.StatusForbidden {
		t.Fatalf("Kollegium: Status %d, erwartet 403", status)
	}
	if u.app.mailAktiv(ctx) {
		t.Fatal("ohne Postfach darf keine Mail rausgehen")
	}
	status, m := hm.json("PUT", "/api/postfach", map[string]string{"adresse": "hm", "passwort": "falsch"})
	if status != http.StatusBadRequest || !strings.Contains(fmt.Sprint(m["fehler"]), "Passwort") {
		t.Fatalf("falsches Passwort: %d, %v", status, m)
	}
	status, m = hm.json("PUT", "/api/postfach", richtig)
	if status != 200 || m["adresse"] != "hm@schule.de" || m["passwort"] != nil {
		t.Fatalf("eingetragen: %d, %v", status, m)
	}
	if len(*testmails) != 1 || (*testmails)[0] != "hm@schule.de" {
		t.Fatalf("Testmail an die eintragende Person: %v", *testmails)
	}
	if _, geheim, _, err := u.app.sp.versandPostfach(ctx); err != nil || bytes.Contains(geheim, []byte("richtig")) {
		t.Fatalf("gespeichert nur verschlüsselt: %v", err)
	}
	if z, err := u.app.versandZugang(ctx); err != nil || z.passwort != "richtig" || z.absender != "hm@schule.de" {
		t.Fatalf("Zugang für den Versand: %+v, %v", z, err)
	}
	if !u.app.mailAktiv(ctx) {
		t.Fatal("mit Postfach geht Mail raus")
	}
	if status, m := hm.json("DELETE", "/api/postfach", nil); status != 200 || m["adresse"] != "" {
		t.Fatalf("entfernt: %d, %v", status, m)
	}
	if u.app.mailAktiv(ctx) {
		t.Fatal("nach dem Entfernen keine Mail mehr")
	}
}

func TestPostfachBremseUndFest(t *testing.T) {
	u, _ := postfachUmgebung(t)
	hm := u.nutzer("hm")
	falsch := map[string]string{"adresse": "hm", "passwort": "falsch"}
	// Nach drei Fehlversuchen ist für eine Viertelstunde Schluss, bevor der Mailserver sperrt.
	for i := 0; i < 3; i++ {
		if status, _ := hm.json("PUT", "/api/postfach", falsch); status != http.StatusBadRequest {
			t.Fatalf("Versuch %d: Status %d, erwartet 400", i+1, status)
		}
	}
	if status, _ := hm.json("PUT", "/api/postfach", falsch); status != http.StatusTooManyRequests {
		t.Fatalf("vierter Versuch: Status %d, erwartet 429", status)
	}
	// Steht das Postfach fest in der .env, lässt es sich in der App nicht ändern.
	u.app.konf.MailAbsender = "schaden@schule.de"
	if status, m := hm.json("GET", "/api/postfach", nil); status != 200 || m["fest"] != true || m["adresse"] != "schaden@schule.de" {
		t.Fatalf("fest: %d, %v", status, m)
	}
	if status, _ := hm.json("PUT", "/api/postfach", falsch); status != http.StatusConflict {
		t.Fatalf("fest und trotzdem geändert: Status %d, erwartet 409", status)
	}
}
