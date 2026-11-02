package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap/commands"
)

const (
	cookieName      = "schaden_sitzung"
	sitzungsDauer   = 180 * 24 * time.Hour
	imapFrist       = 15 * time.Second
	imapBefehlFrist = 5 * time.Second
)

var (
	// errAbgelehnt heißt: Der Mailserver hat geantwortet und die Zugangsdaten abgelehnt. Nur das
	// zählt als Fehlversuch.
	errAbgelehnt = errors.New("zugangsdaten abgelehnt")
	// errMailserverWeg trennt den Ausfall vom falschen Passwort. Sonst sperrte ein Ausfall des
	// Mailservers nach ein paar Versuchen das ganze Kollegium aus.
	errMailserverWeg = errors.New("mailserver nicht erreichbar")
)

type benutzer struct {
	Email string `json:"email"`
	Rolle string `json:"rolle"`
}

// darfBearbeiten: Status setzen und antworten. darfSehen: das Backend mit allen Meldungen und wer
// gemeldet hat. Das Kollegium sieht nur das Meldeformular.
func (b benutzer) darfBearbeiten() bool { return b.Rolle == rolleHausmeister || b.Rolle == rolleAdmin }
func (b benutzer) darfSehen() bool      { return b.darfBearbeiten() || b.Rolle == rolleLeitung }

// ── Mailserver ───────────────────────────────────────────────────────────────

// pruefeIMAP meldet sich mit den Zugangsdaten am Mailserver der Schule an, über TLS ab der ersten
// Zeile. Das Passwort geht nur dorthin und wird nirgends gespeichert.
func (a *app) pruefeIMAP(ctx context.Context, email, passwort string) error {
	if a.konf.IMAPHost == "mock" {
		// konfiguration.pruefe lässt mock nur mit APP_ENV=local zu.
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, imapFrist)
	defer cancel()

	adresse := net.JoinHostPort(a.konf.IMAPHost, a.konf.IMAPPort)
	verbinder := &tls.Dialer{Config: &tls.Config{ServerName: a.konf.IMAPHost, MinVersion: tls.VersionTLS12}}
	verbindung, err := verbinder.DialContext(ctx, "tcp", adresse)
	if err != nil {
		slog.Error("Mailserver nicht erreichbar", "adresse", adresse, "fehler", err)
		return errMailserverWeg
	}
	// Endet die Frist, schließt diese Goroutine die Verbindung; das löst jedes hängende Lesen.
	fertig := make(chan struct{})
	defer close(fertig)
	go func() {
		select {
		case <-ctx.Done():
			_ = verbindung.Close()
		case <-fertig:
		}
	}()

	c, err := client.New(verbindung)
	if err != nil {
		_ = verbindung.Close()
		slog.Error("Mailserver ohne Begrüßung", "fehler", err)
		return errMailserverWeg
	}
	c.Timeout = imapBefehlFrist
	defer func() { _ = c.Logout() }()

	// Execute statt Login: Nur so lässt sich „Server lehnt ab“ (NO, BAD) von „Server ist weg“ trennen.
	status, err := c.Execute(&commands.Login{Username: email, Password: passwort}, nil)
	switch {
	case err != nil:
		slog.Warn("Anmeldung am Mailserver abgebrochen", "fehler", err)
		return errMailserverWeg
	case status == nil:
		return errMailserverWeg
	case status.Type == imap.StatusRespOk:
		c.SetState(imap.AuthenticatedState, nil)
		return nil
	default:
		return errAbgelehnt
	}
}

// ── Fehlversuchsbremse ───────────────────────────────────────────────────────

// bremse sperrt einen Schlüssel (Konto oder Adresse) für ein Zeitfenster, wenn darin zu viele
// Fehlversuche auflaufen. Sie bremst das Durchprobieren von Passwörtern über diese App.
type bremse struct {
	mu        sync.Mutex
	eintraege map[string]*versuche
	grenze    int
	fenster   time.Duration
}

type versuche struct {
	anzahl      int
	seit        time.Time
	gesperrtBis time.Time
}

func neueBremse(grenze int, fenster time.Duration) *bremse {
	return &bremse{eintraege: map[string]*versuche{}, grenze: grenze, fenster: fenster}
}

// gesperrt liefert die verbleibende Sperrzeit, sonst 0.
func (b *bremse) gesperrt(schluessel string, jetzt time.Time) time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()
	if v := b.eintraege[schluessel]; v != nil && jetzt.Before(v.gesperrtBis) {
		return v.gesperrtBis.Sub(jetzt)
	}
	return 0
}

func (b *bremse) fehlversuch(schluessel string, jetzt time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.eintraege) > 10000 {
		b.vergiss(jetzt)
	}
	v := b.eintraege[schluessel]
	if v == nil || jetzt.Sub(v.seit) > b.fenster {
		v = &versuche{seit: jetzt}
		b.eintraege[schluessel] = v
	}
	v.anzahl++
	if v.anzahl >= b.grenze {
		v.gesperrtBis = jetzt.Add(b.fenster)
		v.anzahl = 0
		v.seit = jetzt
	}
}

func (b *bremse) erfolg(schluessel string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.eintraege, schluessel)
}

func (b *bremse) vergiss(jetzt time.Time) {
	for k, v := range b.eintraege {
		if jetzt.Sub(v.seit) > b.fenster && jetzt.After(v.gesperrtBis) {
			delete(b.eintraege, k)
		}
	}
}

// ── Sitzungen ────────────────────────────────────────────────────────────────

func neuesToken() (token, hash string) {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	token = base64.RawURLEncoding.EncodeToString(b)
	return token, tokenHash(token)
}

// tokenHash: In der Datenbank steht nur der Hash. Wer die Datei liest, kann sich damit nicht anmelden.
func tokenHash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (a *app) setzeCookie(w http.ResponseWriter, token string, ablauf time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		Expires:  ablauf,
		MaxAge:   int(ablauf.Sub(a.jetzt()).Seconds()),
		HttpOnly: true,
		Secure:   !a.konf.Lokal,
		SameSite: http.SameSiteLaxMode,
	})
}

func (a *app) loescheCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: !a.konf.Lokal, SameSite: http.SameSiteLaxMode})
}

// angemeldet liest die Sitzung aus dem Cookie. Wer die App nutzt, bleibt angemeldet: Ist der letzte
// Besuch mehr als einen Tag her, beginnt die Frist von 180 Tagen neu.
func (a *app) angemeldet(w http.ResponseWriter, r *http.Request) (benutzer, bool) {
	c, err := r.Cookie(cookieName)
	if err != nil || c.Value == "" {
		return benutzer{}, false
	}
	hash := tokenHash(c.Value)
	email, ablauf, err := a.sp.sitzung(r.Context(), hash)
	jetzt := a.jetzt()
	if err != nil || jetzt.After(ablauf) {
		return benutzer{}, false
	}
	if neu := jetzt.Add(sitzungsDauer); neu.Sub(ablauf) > 24*time.Hour {
		if err := a.sp.verlaengereSitzung(r.Context(), hash, neu); err == nil {
			a.setzeCookie(w, c.Value, neu)
		}
	}
	rolle, err := a.rolle(r.Context(), email)
	if err != nil {
		slog.Error("Rolle nicht gelesen", "fehler", err)
		return benutzer{}, false
	}
	return benutzer{Email: email, Rolle: rolle}, true
}

type benutzerHandler func(http.ResponseWriter, *http.Request, benutzer)

func (a *app) mitBenutzer(h benutzerHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		b, ok := a.angemeldet(w, r)
		if !ok {
			fehler(w, http.StatusUnauthorized, "Bitte zuerst anmelden.")
			return
		}
		h(w, r, b)
	}
}

func (a *app) nurBackend(h benutzerHandler) benutzerHandler {
	return func(w http.ResponseWriter, r *http.Request, b benutzer) {
		if !b.darfSehen() {
			fehler(w, http.StatusForbidden, "Das sehen nur Hausmeister, Leitung und Verwaltung.")
			return
		}
		h(w, r, b)
	}
}

func (a *app) nurBearbeiter(h benutzerHandler) benutzerHandler {
	return func(w http.ResponseWriter, r *http.Request, b benutzer) {
		if !b.darfBearbeiten() {
			fehler(w, http.StatusForbidden, "Das darf nur der Hausmeister.")
			return
		}
		h(w, r, b)
	}
}

func (a *app) nurAdmin(h benutzerHandler) benutzerHandler {
	return func(w http.ResponseWriter, r *http.Request, b benutzer) {
		if b.Rolle != rolleAdmin {
			fehler(w, http.StatusForbidden, "Das darf nur die Verwaltung.")
			return
		}
		h(w, r, b)
	}
}

// ── Türen ────────────────────────────────────────────────────────────────────

func (a *app) anmelden(w http.ResponseWriter, r *http.Request) {
	var ein struct {
		Email    string `json:"email"`
		Passwort string `json:"passwort"`
	}
	if !leseJSON(w, r, &ein) {
		return
	}
	email := a.konf.vervollstaendige(ein.Email)
	konto := strings.ToLower(email)
	if konto == "" || ein.Passwort == "" {
		fehler(w, http.StatusBadRequest, "Bitte E-Mail und Passwort eingeben.")
		return
	}
	ip, jetzt := a.clientIP(r), a.jetzt()
	if rest := max(a.kontoBremse.gesperrt(konto, jetzt), a.ipBremse.gesperrt(ip, jetzt)); rest > 0 {
		fehler(w, http.StatusTooManyRequests,
			fmt.Sprintf("Zu viele Fehlversuche. Bitte in %d Minuten erneut versuchen.", int(rest.Minutes())+1))
		return
	}
	// Vor dem Mailserver geprüft, damit die App nicht als Prüfstelle für fremde Postfächer dient.
	if !a.konf.domainErlaubt(konto) {
		a.ipBremse.fehlversuch(ip, jetzt)
		fehler(w, http.StatusForbidden, "Anmelden geht nur mit einem Postfach der Schule.")
		return
	}
	switch err := a.pruefeZugang(r.Context(), email, ein.Passwort); {
	case errors.Is(err, errAbgelehnt):
		a.kontoBremse.fehlversuch(konto, jetzt)
		a.ipBremse.fehlversuch(ip, jetzt)
		fehler(w, http.StatusUnauthorized, "E-Mail oder Passwort stimmt nicht.")
		return
	case err != nil:
		fehler(w, http.StatusServiceUnavailable, "Der Mailserver ist gerade nicht erreichbar. Bitte später erneut versuchen.")
		return
	}
	a.kontoBremse.erfolg(konto)

	token, hash := neuesToken()
	ablauf := jetzt.Add(sitzungsDauer)
	if err := a.sp.legeSitzungAn(r.Context(), hash, konto, ablauf); err != nil {
		intern(w, err)
		return
	}
	a.setzeCookie(w, token, ablauf)
	rolle, err := a.rolle(r.Context(), konto)
	if err != nil {
		intern(w, err)
		return
	}
	schreibeJSON(w, http.StatusOK, benutzer{Email: konto, Rolle: rolle})
}

func (a *app) abmelden(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(cookieName); err == nil && c.Value != "" {
		if err := a.sp.loescheSitzung(r.Context(), tokenHash(c.Value)); err != nil {
			slog.Warn("Sitzung nicht gelöscht", "fehler", err)
		}
	}
	a.loescheCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (a *app) ich(w http.ResponseWriter, _ *http.Request, b benutzer) {
	schreibeJSON(w, http.StatusOK, b)
}

// clientIP nimmt hinter Caddy die Adresse aus X-Real-IP. Ohne PROXY_VERTRAUEN gilt die Verbindung,
// weil sonst jeder den Kopf fälschen und der Bremse ausweichen könnte.
func (a *app) clientIP(r *http.Request) string {
	if a.konf.ProxyVertrauen {
		if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
