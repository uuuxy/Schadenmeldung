package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
)

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

func (s *speicher) legeSitzungAn(ctx context.Context, hash, email string, ablauf time.Time) error {
	_, err := s.db.Exec(ctx, "INSERT INTO sitzungen (token_hash, email, laeuft_ab) VALUES ($1, $2, $3)", hash, email, ablauf)
	return err
}

func (s *speicher) sitzung(ctx context.Context, hash string) (email string, ablauf time.Time, err error) {
	err = s.db.QueryRow(ctx, "SELECT email, laeuft_ab FROM sitzungen WHERE token_hash = $1", hash).Scan(&email, &ablauf)
	return email, ablauf, nichtGefunden(err)
}

func (s *speicher) verlaengereSitzung(ctx context.Context, hash string, ablauf time.Time) error {
	_, err := s.db.Exec(ctx, "UPDATE sitzungen SET laeuft_ab = $2 WHERE token_hash = $1", hash, ablauf)
	return err
}

func (s *speicher) loescheSitzung(ctx context.Context, hash string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM sitzungen WHERE token_hash = $1", hash)
	return err
}

func (s *speicher) loescheAbgelaufeneSitzungen(ctx context.Context, jetzt time.Time) error {
	_, err := s.db.Exec(ctx, "DELETE FROM sitzungen WHERE laeuft_ab < $1", jetzt)
	return err
}
