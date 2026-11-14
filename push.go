package main

import (
	"context"
	"crypto/ecdh"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Mitteilungen: Wer aus dem Backend sie auf einem Gerät einschaltet, bekommt zu jeder neuen
// Meldung eine, neben der Mail oder statt ihr, dazu die Zahl der neuen Meldungen am App-Symbol. Wie die
// Mail hetzt sie nicht: eine je Meldung, keine Erinnerungen. Sie zeigt was und wo, keinen Namen.
// Verschlüsselung und Versand: webpush.go.

// pushDienste: Der Server schickt nur an die Push-Dienste der Browser, nie an beliebige Adressen.
var pushDienste = []string{"fcm.googleapis.com", "push.services.mozilla.com", "push.apple.com", "notify.windows.com"}

// dienstVon nennt fürs Log nur den Push-Dienst eines Abos, gekürzt, denn die Adresse kommt vom Gerät.
func dienstVon(endpunkt string) string {
	u, err := url.Parse(endpunkt)
	if err != nil || u.Hostname() == "" {
		return "unlesbar"
	}
	return strings.ToValidUTF8(u.Hostname()[:min(len(u.Hostname()), 80)], "")
}

func pushDienstErlaubt(endpunkt string) bool {
	u, err := url.Parse(endpunkt)
	if err != nil || u.Scheme != "https" || u.User != nil || len(endpunkt) > 1000 {
		return false
	}
	host := u.Hostname()
	for _, d := range pushDienste {
		if host == d || strings.HasSuffix(host, "."+d) {
			return true
		}
	}
	return false
}

// gueltigesAbo prüft Adresse und Schlüssel des Geräts: ein Punkt auf P-256 und 16 Byte Geheimnis.
func gueltigesAbo(abo pushAbo) bool {
	p256dh, err1 := b64.DecodeString(abo.P256dh)
	auth, err2 := b64.DecodeString(abo.Auth)
	if err1 != nil || err2 != nil || len(auth) != 16 || !pushDienstErlaubt(abo.Endpunkt) {
		return false
	}
	_, err := ecdh.P256().NewPublicKey(p256dh)
	return err == nil
}

// pushLesen gibt den öffentlichen Schlüssel des Servers, mit dem das Gerät sein Abo anlegt.
func (a *app) pushLesen(w http.ResponseWriter, r *http.Request, _ benutzer) {
	schluessel, err := a.vapidSchluessel(r.Context())
	if err != nil {
		intern(w, err)
		return
	}
	oeffentlich, err := oeffentlicherSchluessel(schluessel)
	if err != nil {
		intern(w, err)
		return
	}
	schreibeJSON(w, http.StatusOK, map[string]string{"schluessel": oeffentlich})
}

// pushAnfrage ist ein Abo, wie es der Browser liefert (PushSubscription.toJSON).
type pushAnfrage struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		P256dh string `json:"p256dh"`
		Auth   string `json:"auth"`
	} `json:"keys"`
}

// pushAnmelden merkt sich das Abo dieses Geräts. Die App schickt es bei jedem Öffnen wieder: Das hält
// es frisch und setzt die Zahl am App-Symbol zurück.
func (a *app) pushAnmelden(w http.ResponseWriter, r *http.Request, b benutzer) {
	var ein pushAnfrage
	if !leseJSON(w, r, &ein) {
		return
	}
	abo := pushAbo{Endpunkt: ein.Endpoint, Email: b.Email,
		P256dh: strings.TrimRight(ein.Keys.P256dh, "="), Auth: strings.TrimRight(ein.Keys.Auth, "=")}
	if !gueltigesAbo(abo) {
		// Nur der Dienst, keine Person: So zeigt das Log, wenn ein Browser einen Push-Dienst nutzt, der
		// noch nicht in pushDienste steht.
		slog.Warn("Push: Gerät abgelehnt", "dienst", dienstVon(abo.Endpunkt))
		fehler(w, http.StatusBadRequest, "Dieses Gerät kann keine Mitteilungen empfangen.")
		return
	}
	if err := a.sp.merkePushAbo(r.Context(), abo, a.jetzt()); err != nil {
		intern(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *app) pushAbmelden(w http.ResponseWriter, r *http.Request, b benutzer) {
	var ein pushAnfrage
	if !leseJSON(w, r, &ein) {
		return
	}
	if err := a.sp.loeschePushAbo(r.Context(), ein.Endpoint, b.Email); err != nil {
		intern(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type pushNachricht struct {
	Titel string `json:"titel"`
	Text  string `json:"text"`
	Weg   string `json:"weg"`
	Tag   string `json:"tag"`
	Zahl  int    `json:"zahl"`
}

// pushVerschicken läuft nach dem Speichern einer Meldung, neben der Mail. Wer selbst meldet, bekommt
// keine Mitteilung. Zugestellt wird im Hintergrund; die Antwort an die Meldenden wartet nicht darauf.
func (a *app) pushVerschicken(ctx context.Context, m meldung) {
	personen, err := a.backendPersonen(ctx)
	if err != nil {
		slog.Error("Push: Empfänger nicht gelesen", "fehler", err)
		return
	}
	emails := make([]string, 0, len(personen))
	for _, p := range personen {
		if p.Email != m.GemeldetVon {
			emails = append(emails, p.Email)
		}
	}
	ziele, err := a.sp.pushZiele(ctx, emails, m.ID)
	if err != nil {
		slog.Error("Push: Abos nicht gelesen", "fehler", err)
		return
	}
	n := pushNachricht{Titel: "Neue Meldung", Text: m.Was + " – " + m.Ort,
		Weg: fmt.Sprintf("/meldung/%d", m.ID), Tag: fmt.Sprintf("meldung-%d", m.ID)}
	if m.Dringend {
		n.Titel = "Dringende Meldung"
	}
	for _, z := range ziele {
		n.Zahl = z.Neu
		inhalt, err := json.Marshal(n)
		if err != nil {
			continue
		}
		a.pushLaeuft.Go(func() { a.stellePushZu(ctx, z.pushAbo, inhalt) })
	}
}

// stellePushZu versucht es dreimal. Ein erloschenes Abo fällt weg; der sichere Weg bleibt die Mail.
func (a *app) stellePushZu(ctx context.Context, abo pushAbo, inhalt []byte) {
	pause := a.pushPause
	for versuch := 1; versuch <= 3; versuch++ {
		erloschen, err := a.sendePush(ctx, abo, inhalt)
		if erloschen {
			if err := a.sp.loeschePushAbo(ctx, abo.Endpunkt, ""); err != nil {
				slog.Warn("Push: erloschenes Abo nicht gelöscht", "fehler", err)
			}
			slog.Info("Push: erloschenes Abo gelöscht", "dienst", dienstVon(abo.Endpunkt))
			return
		}
		if err == nil {
			return
		}
		slog.Warn("Push nicht zugestellt", "versuch", versuch, "fehler", err)
		if versuch < 3 {
			time.Sleep(pause)
			pause *= 5
		}
	}
}

// raeumePushAbosAuf: Abos, die 180 Tage nicht offen waren, und Abos ohne Backend-Rolle fallen weg.
func (a *app) raeumePushAbosAuf(ctx context.Context, jetzt time.Time) error {
	personen, err := a.backendPersonen(ctx)
	if err != nil {
		return err
	}
	emails := make([]string, 0, len(personen))
	for _, p := range personen {
		emails = append(emails, p.Email)
	}
	return a.sp.loescheAltePushAbos(ctx, jetzt.Add(-sitzungsDauer), emails)
}
