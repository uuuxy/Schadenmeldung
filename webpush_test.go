package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"math/big"
	"strings"
	"testing"
	"time"
)

func aus64(t *testing.T, s string) []byte {
	t.Helper()
	b, err := b64.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// Das Beispiel aus RFC 8291, Abschnitt 5: Nur wenn jedes Byte stimmt, können echte Browser lesen.
func TestVerschluesselungWieRFC8291(t *testing.T) {
	einmal, err := ecdh.P256().NewPrivateKey(aus64(t, "yfWPiYE-n46HLnH0KqZOF1fJJU3MYrct3AELtAQ-oRw"))
	if err != nil {
		t.Fatal(err)
	}
	if got := b64.EncodeToString(einmal.PublicKey().Bytes()); got != "BP4z9KsN6nGRTbVYI_c7VJSPQTBtkgcy27mlmlMoZIIgDll6e3vCYLocInmYWAmS6TlzAC8wEqKK6PBru3jl7A8" {
		t.Fatalf("öffentlicher Schlüssel des Servers: %s", got)
	}
	geraet := "BCVxsr7N_eNgVRqvHtD0zTZsEc6-VV-JvLexhqUzORcxaOzi6-AYWXvTBHm4bjyPjs7Vd8pZGH6SRpkNtoIAiw4"
	rumpf, err := verschluessele([]byte("When I grow up, I want to be a watermelon"),
		aus64(t, geraet), aus64(t, "BTBZMqHH6r4Tts7J_aSIgg"), aus64(t, "DGv6ra1nlYgDCS1FRnbzlw"), einmal)
	if err != nil {
		t.Fatal(err)
	}
	erwartet := "DGv6ra1nlYgDCS1FRnbzlwAAEABBBP4z9KsN6nGRTbVYI_c7VJSPQTBtkgcy27mlmlMoZIIgDll6e3vCYLocInmYWAmS6TlzAC8wEqKK6PBru3jl7A_yl95bQpu6cVPTpK4Mqgkf1CXztLVBSt2Ks3oZwbuwXPXLWyouBWLVWGNWQexSgSxsj_Qulcy4a-fN"
	if got := b64.EncodeToString(rumpf); got != erwartet {
		t.Errorf("Rumpf weicht vom RFC ab:\n%s\n%s", got, erwartet)
	}
	geraetGeheim, err := ecdh.P256().NewPrivateKey(aus64(t, "q1dXpw3UpT5VOmu_cf_v6ih07Aems3njxI-JWgLcM94"))
	if err != nil {
		t.Fatal(err)
	}
	if klar := entschluessele(t, rumpf, geraetGeheim, aus64(t, "BTBZMqHH6r4Tts7J_aSIgg")); string(klar) != "When I grow up, I want to be a watermelon" {
		t.Errorf("entschlüsselt: %q", klar)
	}
}

// entschluessele ist die Gegenseite, das Gerät (RFC 8291), nur für die Tests.
func entschluessele(t *testing.T, rumpf []byte, geraet *ecdh.PrivateKey, auth []byte) []byte {
	t.Helper()
	if len(rumpf) < 21+65+17 || rumpf[20] != 65 {
		t.Fatalf("Kopf der Nachricht stimmt nicht (%d Byte)", len(rumpf))
	}
	salz, absender, geheim := rumpf[:16], rumpf[21:86], rumpf[86:]
	oeffentlich, err := ecdh.P256().NewPublicKey(absender)
	if err != nil {
		t.Fatal(err)
	}
	geteilt, err := geraet.ECDH(oeffentlich)
	if err != nil {
		t.Fatal(err)
	}
	ikm, _ := hkdf.Key(sha256.New, geteilt, auth, "WebPush: info\x00"+string(geraet.PublicKey().Bytes())+string(absender), 32)
	cek, _ := hkdf.Key(sha256.New, ikm, salz, "Content-Encoding: aes128gcm\x00", 16)
	nonce, _ := hkdf.Key(sha256.New, ikm, salz, "Content-Encoding: nonce\x00", 12)
	block, _ := aes.NewCipher(cek)
	gcm, _ := cipher.NewGCM(block)
	klar, err := gcm.Open(nil, nonce, geheim, nil)
	if err != nil {
		t.Fatalf("nicht zu entschlüsseln: %v", err)
	}
	klar = bytes.TrimRight(klar, "\x00")
	if len(klar) == 0 || klar[len(klar)-1] != 2 {
		t.Fatal("Trennzeichen am Ende fehlt")
	}
	return klar[:len(klar)-1]
}

// Der Ausweis (VAPID): richtig signiert, für den Dienst der Adresse, zwölf Stunden gültig.
func TestVapidAusweis(t *testing.T) {
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	jetzt := time.Unix(1_800_000_000, 0)
	ausweis, err := vapid("https://fcm.googleapis.com/fcm/send/abc", k, "https://schaden.schule.de", jetzt)
	if err != nil {
		t.Fatal(err)
	}
	t1, k1, ok := strings.Cut(strings.TrimPrefix(ausweis, "vapid t="), ", k=")
	teile := strings.Split(t1, ".")
	if !ok || len(teile) != 3 {
		t.Fatalf("Ausweis falsch gebaut: %s", ausweis)
	}
	signatur := aus64(t, teile[2])
	summe := sha256.Sum256([]byte(teile[0] + "." + teile[1]))
	if len(signatur) != 64 || !ecdsa.Verify(&k.PublicKey, summe[:], new(big.Int).SetBytes(signatur[:32]), new(big.Int).SetBytes(signatur[32:])) {
		t.Error("Signatur stimmt nicht")
	}
	var ansprueche struct {
		Aud, Sub string
		Exp      int64
	}
	if err := json.Unmarshal(aus64(t, teile[1]), &ansprueche); err != nil {
		t.Fatal(err)
	}
	if ansprueche.Aud != "https://fcm.googleapis.com" || ansprueche.Sub != "https://schaden.schule.de" || ansprueche.Exp != jetzt.Unix()+12*3600 {
		t.Errorf("Angaben im Ausweis: %+v", ansprueche)
	}
	if oeffentlich, _ := oeffentlicherSchluessel(k); k1 != oeffentlich {
		t.Errorf("k= nennt nicht den öffentlichen Schlüssel: %s", k1)
	}
}
