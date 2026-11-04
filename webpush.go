package main

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Web Push mit der Standardbibliothek: Verschlüsselung nach RFC 8291, Ausweis des Servers nach RFC 8292
// (VAPID). Der Push-Dienst des Browsers (Google, Apple, Mozilla, Microsoft) stellt nur zu; lesen kann
// den Inhalt allein das Gerät, das ihn abonniert hat.

var b64 = base64.RawURLEncoding

// verschluessele baut den Rumpf einer Push-Nachricht: ein einziger Datensatz in aes128gcm. Salz und das
// Schlüsselpaar nur für diese Nachricht kommen von außen, damit der Test das Beispiel aus RFC 8291
// nachrechnen kann.
func verschluessele(klartext, p256dh, auth, salz []byte, einmal *ecdh.PrivateKey) ([]byte, error) {
	geraet, err := ecdh.P256().NewPublicKey(p256dh)
	if err != nil {
		return nil, err
	}
	geteilt, err := einmal.ECDH(geraet)
	if err != nil {
		return nil, err
	}
	absender := einmal.PublicKey().Bytes()
	ikm, err := hkdf.Key(sha256.New, geteilt, auth, "WebPush: info\x00"+string(p256dh)+string(absender), 32)
	if err != nil {
		return nil, err
	}
	cek, err := hkdf.Key(sha256.New, ikm, salz, "Content-Encoding: aes128gcm\x00", 16)
	if err != nil {
		return nil, err
	}
	nonce, err := hkdf.Key(sha256.New, ikm, salz, "Content-Encoding: nonce\x00", 12)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(cek)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	// Kopf: Salz, Datensatzgröße 4096, Länge und Wert des öffentlichen Schlüssels (P-256: immer 65 Byte).
	kopf := append(append([]byte{}, salz...), 0, 0, 0x10, 0, 65)
	kopf = append(kopf, absender...)
	// 0x02 schließt den letzten, hier einzigen Datensatz ab.
	return gcm.Seal(kopf, nonce, append(append([]byte{}, klartext...), 2), nil), nil
}

// vapid baut den Ausweis des Servers für einen Push-Dienst: ein JWT (ES256), zwölf Stunden gültig.
func vapid(endpunkt string, schluessel *ecdsa.PrivateKey, absender string, jetzt time.Time) (string, error) {
	u, err := url.Parse(endpunkt)
	if err != nil {
		return "", err
	}
	ansprueche, err := json.Marshal(map[string]any{"aud": u.Scheme + "://" + u.Host, "exp": jetzt.Add(12 * time.Hour).Unix(), "sub": absender})
	if err != nil {
		return "", err
	}
	unsigniert := b64.EncodeToString([]byte(`{"typ":"JWT","alg":"ES256"}`)) + "." + b64.EncodeToString(ansprueche)
	summe := sha256.Sum256([]byte(unsigniert))
	r, s, err := ecdsa.Sign(rand.Reader, schluessel, summe[:])
	if err != nil {
		return "", err
	}
	signatur := make([]byte, 64)
	r.FillBytes(signatur[:32])
	s.FillBytes(signatur[32:])
	oeffentlich, err := oeffentlicherSchluessel(schluessel)
	if err != nil {
		return "", err
	}
	return "vapid t=" + unsigniert + "." + b64.EncodeToString(signatur) + ", k=" + oeffentlich, nil
}

// oeffentlicherSchluessel: der Punkt unkomprimiert in Base64url, so wie ihn Browser und Push-Dienste erwarten.
func oeffentlicherSchluessel(k *ecdsa.PrivateKey) (string, error) {
	p, err := k.PublicKey.ECDH()
	if err != nil {
		return "", err
	}
	return b64.EncodeToString(p.Bytes()), nil
}

// vapidSchluessel lädt das Schlüsselpaar des Servers einmal aus der Datenbank (oder legt es dort an).
func (a *app) vapidSchluessel(ctx context.Context) (*ecdsa.PrivateKey, error) {
	a.pushMu.Lock()
	defer a.pushMu.Unlock()
	if a.pushSchluessel != nil {
		return a.pushSchluessel, nil
	}
	der, err := a.sp.pushSchluessel(ctx, func() ([]byte, error) {
		k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, err
		}
		return x509.MarshalPKCS8PrivateKey(k)
	})
	if err != nil {
		return nil, err
	}
	k, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, err
	}
	ek, ok := k.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New("push-schlüssel ist kein ECDSA-Schlüssel")
	}
	a.pushSchluessel = ek
	return ek, nil
}

// sendePush stellt einem Gerät eine Nachricht zu. erloschen meldet, dass der Push-Dienst das Abo nicht
// mehr kennt (404, 410); dann löscht es der Aufrufer.
func (a *app) sendePush(ctx context.Context, abo pushAbo, inhalt []byte) (erloschen bool, err error) {
	rumpf, ausweis, err := a.pushVorbereiten(ctx, abo, inhalt)
	if err != nil {
		return false, err
	}
	// Die Adresse stammt aus einem geprüften Abo: nur Push-Dienste der Browser (pushDienstErlaubt).
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, abo.Endpunkt, bytes.NewReader(rumpf)) // #nosec G107
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", ausweis)
	req.Header.Set("Content-Encoding", "aes128gcm")
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("TTL", "86400") // Ist das Handy aus, hält der Dienst die Nachricht einen Tag bereit.
	req.Header.Set("Urgency", "high")
	resp, err := a.pushClient.Do(req)
	if err != nil {
		// Ohne die Adresse ins Log: Sie kennzeichnet ein Gerät.
		var ue *url.Error
		if errors.As(err, &ue) {
			err = ue.Err
		}
		return false, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	switch {
	case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone:
		return true, nil
	case resp.StatusCode/100 == 2:
		return false, nil
	}
	return false, fmt.Errorf("push-dienst antwortet mit %d", resp.StatusCode)
}

func (a *app) pushVorbereiten(ctx context.Context, abo pushAbo, inhalt []byte) (rumpf []byte, ausweis string, err error) {
	schluessel, err := a.vapidSchluessel(ctx)
	if err != nil {
		return nil, "", err
	}
	p256dh, err1 := b64.DecodeString(abo.P256dh)
	auth, err2 := b64.DecodeString(abo.Auth)
	einmal, err3 := ecdh.P256().GenerateKey(rand.Reader)
	if err := errors.Join(err1, err2, err3); err != nil {
		return nil, "", err
	}
	salz := make([]byte, 16)
	_, _ = rand.Read(salz) // crypto/rand scheitert nicht (Go 1.24).
	if rumpf, err = verschluessele(inhalt, p256dh, auth, salz, einmal); err != nil {
		return nil, "", err
	}
	ausweis, err = vapid(abo.Endpunkt, schluessel, a.pushAbsender(), a.jetzt())
	return rumpf, ausweis, err
}

// pushAbsender nennt den Push-Diensten, wer schickt; Apple verlangt eine https- oder mailto-Adresse.
func (a *app) pushAbsender() string {
	if strings.HasPrefix(a.konf.OeffentlicheAdresse, "https://") || a.konf.MailAbsender == "" {
		return a.konf.OeffentlicheAdresse
	}
	return "mailto:" + a.konf.MailAbsender
}
