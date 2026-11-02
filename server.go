package main

import (
	"context"
	"io/fs"
	"net/http"
	"sync"
	"time"
)

type app struct {
	konf         konfiguration
	sp           *speicher
	statisch     fs.FS
	pruefeZugang func(ctx context.Context, email, passwort string) error
	kontoBremse  *bremse
	ipBremse     *bremse
	jetzt        func() time.Time
	versender    versender
	verteiler    *verteiler
	// hintergrund zählt laufende Mails, damit Tests und das Herunterfahren auf sie warten können.
	hintergrund sync.WaitGroup
}

func neueApp(konf konfiguration, sp *speicher, statisch fs.FS) *app {
	a := &app{
		konf:        konf,
		sp:          sp,
		statisch:    statisch,
		kontoBremse: neueBremse(5, 15*time.Minute),
		ipBremse:    neueBremse(20, 15*time.Minute),
		jetzt:       time.Now,
		verteiler:   neuerVerteiler(),
	}
	a.pruefeZugang = a.pruefeIMAP
	if konf.mailAktiv() {
		a.versender = smtpVersender{k: konf}
	}
	return a
}

func (a *app) routen() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/einstellungen", a.einstellungen)
	mux.HandleFunc("POST /api/anmelden", a.anmelden)
	mux.HandleFunc("POST /api/abmelden", a.abmelden)
	mux.HandleFunc("GET /api/ich", a.mitBenutzer(a.ich))
	mux.HandleFunc("GET /api/raeume", a.mitBenutzer(a.raeume))
	mux.HandleFunc("POST /api/raeume", a.mitBenutzer(a.nurAdmin(a.raeumeAnlegen)))
	mux.HandleFunc("DELETE /api/raeume/{id}", a.mitBenutzer(a.nurAdmin(a.raumLoeschen)))
	mux.HandleFunc("GET /api/meldungen", a.mitBenutzer(a.meldungen))
	mux.HandleFunc("POST /api/meldungen", a.mitBenutzer(a.meldungAnlegen))
	mux.HandleFunc("POST /api/meldungen/{id}/status", a.mitBenutzer(a.nurBearbeiter(a.statusSetzen)))
	mux.HandleFunc("DELETE /api/meldungen/{id}", a.mitBenutzer(a.nurAdmin(a.meldungLoeschen)))
	mux.HandleFunc("GET /api/ereignisse", a.mitBenutzer(a.nurBackend(a.ereignisse)))
	mux.HandleFunc("GET /api/rollen", a.mitBenutzer(a.nurAdmin(a.rollenListe)))
	mux.HandleFunc("POST /api/rollen", a.mitBenutzer(a.nurAdmin(a.rolleSetzen)))
	mux.HandleFunc("DELETE /api/rollen/{email}", a.mitBenutzer(a.nurAdmin(a.rolleLoeschen)))
	mux.HandleFunc("GET /api/benachrichtigung", a.mitBenutzer(a.nurBackend(a.benachrichtigungLesen)))
	mux.HandleFunc("POST /api/benachrichtigung", a.mitBenutzer(a.nurBackend(a.benachrichtigungSetzen)))
	// Die Knöpfe aus der Mail: ohne Anmeldung, der Link selbst berechtigt.
	mux.HandleFunc("GET /api/aktion/{token}", a.aktionLesen)
	mux.HandleFunc("POST /api/aktion/{token}", a.aktionAusfuehren)
	mux.HandleFunc("GET /api/aktion/{token}/foto", a.aktionFoto)
	mux.HandleFunc("GET /api/", func(w http.ResponseWriter, _ *http.Request) {
		fehler(w, http.StatusNotFound, "Unbekannte Adresse.")
	})
	mux.HandleFunc("GET /fotos/{name}", a.mitBenutzer(a.foto))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("GET /", a.oberflaeche())
	return a.sicherheit(mux)
}

// sicherheit setzt die Schutzköpfe und verlangt bei schreibenden Anfragen den Kopf X-Schaden. Ein
// eigener Kopf lässt sich von fremden Seiten nicht ohne CORS-Vorabfrage setzen; zusammen mit dem
// SameSite-Cookie hält das Formulare anderer Seiten von den schreibenden Türen fern.
func (a *app) sicherheit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' blob: data:; style-src 'self' 'unsafe-inline'; "+
			"font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'; object-src 'none'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "same-origin")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Permissions-Policy", "geolocation=(), microphone=()")
		if r.Method != http.MethodGet && r.Method != http.MethodHead && r.Header.Get("X-Schaden") != "1" {
			fehler(w, http.StatusForbidden, "Anfrage abgelehnt.")
			return
		}
		next.ServeHTTP(w, r)
	})
}
