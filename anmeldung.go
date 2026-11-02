package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"
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
