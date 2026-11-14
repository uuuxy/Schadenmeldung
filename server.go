package main

import (
	"context"
	"crypto/ecdsa"
	"io/fs"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type app struct {
	konf         konfiguration
	sp           *speicher
	statisch     fs.FS
	pruefeZugang func(ctx context.Context, email, passwort string) error
	// pruefeVersand schickt mit einem neu eingetragenen Postfach eine Testmail (versandpostfach.go).
	pruefeVersand  func(ctx context.Context, z zugang, an string) error
	postfachBremse *bremse
	kontoBremse    *bremse
	ipBremse       *bremse
	meldeBremse    *bremse
	// imapPlaetze begrenzt die gleichzeitigen Prüfungen beim Mailserver der Schule.
	imapPlaetze chan struct{}
	jetzt       func() time.Time
	versender   versender
	verteiler   *verteiler
	// postWecker stößt den Postausgang an, sobald eine Meldung eingeht (Puffer 1, nichts geht verloren).
	postWecker chan struct{}
	// Mitteilungen (push.go): der Schlüssel des Servers, einmal geladen, und der Versand.
	pushMu         sync.Mutex
	pushSchluessel *ecdsa.PrivateKey
	pushClient     *http.Client
	pushPause      time.Duration
	pushLaeuft     sync.WaitGroup
	// mailStoerung: seit wann Mails scheitern, in Unix-Sekunden; 0, solange alles läuft (mailstoerung.go).
	mailStoerung atomic.Int64
}

func neueApp(konf konfiguration, sp *speicher, statisch fs.FS) *app {
	a := &app{
		konf:        konf,
		sp:          sp,
		statisch:    statisch,
		kontoBremse: neueBremse(5, 15*time.Minute),
		// Drei Versuche je Viertelstunde: Der Mailserver der Schule sperrt nach ein paar falschen
		// Passwörtern die ganze App, auch die Anmeldung aller.
		postfachBremse: neueBremse(3, 15*time.Minute),
		// Im Schul-WLAN teilen sich alle eine Adresse: Melden sich am ersten Tag viele zugleich an,
		// sperrten 20 Tippfehler alle aus. Den einzelnen Zugang schützt weiter die kontoBremse.
		ipBremse:    neueBremse(60, 15*time.Minute),
		meldeBremse: neueBremse(20, time.Hour),
		imapPlaetze: make(chan struct{}, 4),
		jetzt:       time.Now,
		verteiler:   neuerVerteiler(),
		postWecker:  make(chan struct{}, 1),
		// Keine Weiterleitungen: Der Server spricht nur mit der geprüften Adresse des Push-Dienstes.
		pushClient: &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}},
		pushPause: time.Second,
	}
	a.pruefeZugang = a.pruefeIMAP
	a.pruefeVersand = a.pruefeVersandSMTP
	if konf.SMTPHost != "" {
		a.versender = smtpVersender{k: konf, zugang: a.versandZugang}
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
	mux.HandleFunc("POST /api/raeume", a.mitBenutzer(a.nurBearbeiter(a.raeumeAnlegen)))
	mux.HandleFunc("POST /api/raeume/lesen", a.mitBenutzer(a.nurBearbeiter(a.raeumeLesen)))
	mux.HandleFunc("PUT /api/raeume/{id}", a.mitBenutzer(a.nurBearbeiter(a.raumAendern)))
	mux.HandleFunc("DELETE /api/raeume/{id}", a.mitBenutzer(a.nurBearbeiter(a.raumLoeschen)))
	mux.HandleFunc("GET /api/meldungen", a.mitBenutzer(a.meldungen))
	mux.HandleFunc("POST /api/meldungen", a.mitBenutzer(a.meldungAnlegen))
	mux.HandleFunc("POST /api/meldungen/{id}/status", a.mitBenutzer(a.nurBearbeiter(a.statusSetzen)))
	mux.HandleFunc("DELETE /api/meldungen/{id}", a.mitBenutzer(a.nurBearbeiter(a.meldungLoeschen)))
	mux.HandleFunc("GET /api/ereignisse", a.mitBenutzer(a.nurBackend(a.ereignisse)))
	mux.HandleFunc("POST /api/arten", a.mitBenutzer(a.nurBearbeiter(a.artSchalten)))
	mux.HandleFunc("GET /api/export.xlsx", a.mitBenutzer(a.nurBackend(a.exportExcel)))
	mux.HandleFunc("GET /api/export.pdf", a.mitBenutzer(a.nurBackend(a.exportPDF)))
	mux.HandleFunc("GET /api/druck.pdf", a.mitBenutzer(a.nurBackend(a.druckPDF)))
	mux.HandleFunc("GET /api/rollen", a.mitBenutzer(a.nurBearbeiter(a.rollenListe)))
	mux.HandleFunc("POST /api/rollen", a.mitBenutzer(a.nurBearbeiter(a.rolleSetzen)))
	mux.HandleFunc("DELETE /api/rollen/{email}", a.mitBenutzer(a.nurBearbeiter(a.rolleLoeschen)))
	mux.HandleFunc("GET /api/postfach", a.mitBenutzer(a.nurBearbeiter(a.postfachLesen)))
	mux.HandleFunc("PUT /api/postfach", a.mitBenutzer(a.nurBearbeiter(a.postfachSetzen)))
	mux.HandleFunc("DELETE /api/postfach", a.mitBenutzer(a.nurBearbeiter(a.postfachLoeschen)))
	mux.HandleFunc("GET /api/benachrichtigung", a.mitBenutzer(a.nurBackend(a.benachrichtigungLesen)))
	mux.HandleFunc("POST /api/benachrichtigung", a.mitBenutzer(a.nurBackend(a.benachrichtigungSetzen)))
	mux.HandleFunc("GET /api/push", a.mitBenutzer(a.nurBackend(a.pushLesen)))
	mux.HandleFunc("POST /api/push", a.mitBenutzer(a.nurBackend(a.pushAnmelden)))
	mux.HandleFunc("DELETE /api/push", a.mitBenutzer(a.nurBackend(a.pushAbmelden)))
	// Die Knöpfe aus der Mail: ohne Anmeldung, der Link selbst berechtigt.
	mux.HandleFunc("GET /api/aktion/{token}", a.aktionLesen)
	mux.HandleFunc("POST /api/aktion/{token}", a.aktionAusfuehren)
	mux.HandleFunc("GET /api/aktion/{token}/foto", a.aktionFoto)
	mux.HandleFunc("GET /api/", func(w http.ResponseWriter, _ *http.Request) {
		fehler(w, http.StatusNotFound, "Unbekannte Adresse.")
	})
	mux.HandleFunc("GET /fotos/{name}", a.mitBenutzer(a.nurBackend(a.foto)))
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
