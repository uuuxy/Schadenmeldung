package main

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

// testPushDienst spielt Google, Apple und Co.: Die App erreicht ihn über einen umgebogenen Transport,
// die Adressen der Abos bleiben echte Push-Adressen.
type testPushDienst struct {
	mu       sync.Mutex
	status   int
	erhalten []testPush
}

type testPush struct {
	Endpunkt string
	Kopf     http.Header
	Rumpf    []byte
}

func (d *testPushDienst) RoundTrip(r *http.Request) (*http.Response, error) {
	rumpf, _ := io.ReadAll(r.Body)
	d.mu.Lock()
	defer d.mu.Unlock()
	d.erhalten = append(d.erhalten, testPush{Endpunkt: r.URL.String(), Kopf: r.Header.Clone(), Rumpf: rumpf})
	return &http.Response{StatusCode: d.status, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}, Request: r}, nil
}

func mitPushDienst(u *testUmgebung, status int) *testPushDienst {
	d := &testPushDienst{status: status}
	u.app.pushClient = &http.Client{Transport: d}
	u.app.pushPause = 0
	return d
}

// testGeraet ist ein Handy mit eigenem Schlüsselpaar, wie es der Browser beim Abonnieren anlegt.
type testGeraet struct {
	schluessel *ecdh.PrivateKey
	auth       []byte
	endpunkt   string
}

func neuesTestGeraet(t *testing.T, endpunkt string) testGeraet {
	t.Helper()
	k, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	auth := make([]byte, 16)
	_, _ = rand.Read(auth)
	return testGeraet{schluessel: k, auth: auth, endpunkt: endpunkt}
}

func (g testGeraet) abo() map[string]any {
	return map[string]any{"endpoint": g.endpunkt, "keys": map[string]string{
		"p256dh": b64.EncodeToString(g.schluessel.PublicKey().Bytes()), "auth": b64.EncodeToString(g.auth)}}
}

func TestPushNurFuerBackend(t *testing.T) {
	u := neueTestUmgebung(t)
	lehrer := u.nutzer("lehrer")
	if status, _ := lehrer.json("GET", "/api/push", nil); status != http.StatusForbidden {
		t.Errorf("Kollegium liest den Schlüssel: Status %d", status)
	}
	if status, _ := lehrer.json("POST", "/api/push", neuesTestGeraet(t, "https://fcm.googleapis.com/fcm/send/l").abo()); status != http.StatusForbidden {
		t.Errorf("Kollegium legt ein Abo an: Status %d", status)
	}
}

// Der Server schickt nur an die Push-Dienste der Browser, sonst könnte ein Abo ihn auf beliebige
// Adressen lenken, auch im eigenen Netz.
func TestPushNurAnPushDienste(t *testing.T) {
	u := neueTestUmgebung(t)
	hm := u.nutzer("hm")
	for _, adresse := range []string{"https://boese.example/fcm", "http://fcm.googleapis.com/fcm/send/x",
		"https://fcm.googleapis.com.boese.example/x", "https://jemand@fcm.googleapis.com/x", "https://localhost/x"} {
		if status, _ := hm.json("POST", "/api/push", neuesTestGeraet(t, adresse).abo()); status != http.StatusBadRequest {
			t.Errorf("%s: Status %d, erwartet 400", adresse, status)
		}
	}
}

func TestPushBeiNeuerMeldung(t *testing.T) {
	u := neueTestUmgebung(t)
	dienst := mitPushDienst(u, http.StatusCreated)
	hm := u.nutzer("hm")
	status, antwort := hm.json("GET", "/api/push", nil)
	if status != http.StatusOK || len(aus64(t, fmt.Sprint(antwort["schluessel"]))) != 65 {
		t.Fatalf("Schlüssel des Servers: Status %d, %v", status, antwort)
	}
	g := neuesTestGeraet(t, "https://fcm.googleapis.com/fcm/send/hausmeister")
	if status, _ := hm.json("POST", "/api/push", g.abo()); status != http.StatusNoContent {
		t.Fatalf("Abo anlegen: Status %d", status)
	}
	u.uhr = u.uhr.Add(time.Minute)
	lehrer := u.nutzer("lehrer")
	if status, _ := lehrer.melde(map[string]string{"ort": "Raum 204", "kategorie": "technik", "was": "Beamer/Display aus", "dringend": "1"}, nil); status != http.StatusCreated {
		t.Fatalf("Meldung: Status %d", status)
	}
	u.app.pushLaeuft.Wait()
	if len(dienst.erhalten) != 1 {
		t.Fatalf("%d Mitteilungen, erwartet 1", len(dienst.erhalten))
	}
	p := dienst.erhalten[0]
	if p.Endpunkt != g.endpunkt || p.Kopf.Get("Content-Encoding") != "aes128gcm" || p.Kopf.Get("TTL") != "86400" ||
		!strings.HasPrefix(p.Kopf.Get("Authorization"), "vapid t=") {
		t.Errorf("Anfrage an den Push-Dienst: %s %v", p.Endpunkt, p.Kopf)
	}
	klar := entschluessele(t, p.Rumpf, g.schluessel, g.auth)
	var n pushNachricht
	if err := json.Unmarshal(klar, &n); err != nil {
		t.Fatal(err)
	}
	if n.Titel != "Dringende Meldung" || n.Text != "Beamer/Display aus – Raum 204" || !strings.HasPrefix(n.Weg, "/meldung/") || n.Zahl != 1 {
		t.Errorf("Inhalt: %+v", n)
	}
	if strings.Contains(string(klar), "lehrer") {
		t.Errorf("Name der meldenden Person in der Mitteilung: %s", klar)
	}
	// Wer selbst meldet, bekommt keine Mitteilung.
	hm.melde(map[string]string{"ort": "Aula", "kategorie": "licht", "was": "Lampe flackert"}, nil)
	u.app.pushLaeuft.Wait()
	if len(dienst.erhalten) != 1 {
		t.Errorf("Mitteilung zur eigenen Meldung: %d", len(dienst.erhalten))
	}
}

func TestPushErloschenesAboFaelltWeg(t *testing.T) {
	u := neueTestUmgebung(t)
	mitPushDienst(u, http.StatusGone)
	hm := u.nutzer("hm")
	if status, _ := hm.json("POST", "/api/push", neuesTestGeraet(t, "https://web.push.apple.com/QGhtZ2VyYWV0").abo()); status != http.StatusNoContent {
		t.Fatalf("Abo anlegen: Status %d", status)
	}
	u.nutzer("lehrer").melde(map[string]string{"ort": "Aula", "kategorie": "licht", "was": "Lampe flackert"}, nil)
	u.app.pushLaeuft.Wait()
	if ziele, err := u.app.sp.pushZiele(context.Background(), []string{"hm@schule.de"}); err != nil || len(ziele) != 0 {
		t.Errorf("erloschenes Abo noch da: %v %v", ziele, err)
	}
}

// Abos ohne Öffnen der App seit 180 Tagen und Abos ohne Backend-Rolle fallen weg; wer sich überall
// abmeldet, verliert alle.
func TestPushAbosRaeumenSichAuf(t *testing.T) {
	u := neueTestUmgebung(t)
	ctx := context.Background()
	abo := func(name, email string) pushAbo {
		return pushAbo{Endpunkt: "https://fcm.googleapis.com/fcm/send/" + name, Email: email, P256dh: "x", Auth: "y"}
	}
	for _, e := range []struct {
		abo     pushAbo
		gesehen time.Time
	}{
		{abo("alt", "hm@schule.de"), u.uhr.Add(-181 * 24 * time.Hour)},
		{abo("frisch", "hm@schule.de"), u.uhr},
		{abo("ohneRolle", "lehrer@schule.de"), u.uhr},
	} {
		if err := u.app.sp.merkePushAbo(ctx, e.abo, e.gesehen); err != nil {
			t.Fatal(err)
		}
	}
	if err := u.app.raeumePushAbosAuf(ctx, u.uhr); err != nil {
		t.Fatal(err)
	}
	ziele, err := u.app.sp.pushZiele(ctx, []string{"hm@schule.de", "lehrer@schule.de"})
	if err != nil || len(ziele) != 1 || !strings.HasSuffix(ziele[0].Endpunkt, "/frisch") {
		t.Fatalf("nach dem Aufräumen: %v %v", ziele, err)
	}
	hm := u.nutzer("hm")
	if status, _ := hm.json("POST", "/api/abmelden?ueberall=1", nil); status != http.StatusNoContent {
		t.Fatalf("Abmelden: Status %d", status)
	}
	if ziele, _ := u.app.sp.pushZiele(ctx, []string{"hm@schule.de"}); len(ziele) != 0 {
		t.Errorf("nach dem Abmelden überall noch %d Abos", len(ziele))
	}
}
