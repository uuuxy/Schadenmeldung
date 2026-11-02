package main

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
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
	}
	a.pruefeZugang = a.pruefeIMAP
	if konf.mailAktiv() {
		a.versender = smtpVersender{k: konf}
	}
	return a
}

var (
	kategorien = map[string]bool{"licht": true, "heizung": true, "wasser": true, "fenster": true,
		"moebel": true, "technik": true, "reinigung": true, "sonstiges": true}
	stati    = map[string]bool{"offen": true, "gesehen": true, "bestellt": true, "fachfirma": true, "erledigt": true}
	fotoName = regexp.MustCompile(`^[0-9a-f]{32}\.jpg$`)
)

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

// ── Einstellungen und Räume ──────────────────────────────────────────────────

func (a *app) einstellungen(w http.ResponseWriter, _ *http.Request) {
	domain := ""
	if len(a.konf.Domains) == 1 {
		domain = a.konf.Domains[0]
	}
	schreibeJSON(w, http.StatusOK, map[string]any{
		"schulname":        a.konf.Schulname,
		"domain":           domain,
		"fotoTage":         a.konf.FotoTage,
		"aufbewahrungTage": a.konf.AufbewahrungTage,
	})
}

func (a *app) raeume(w http.ResponseWriter, r *http.Request, _ benutzer) {
	liste, err := a.sp.raeume(r.Context())
	if err != nil {
		intern(w, err)
		return
	}
	schreibeJSON(w, http.StatusOK, liste)
}

// raeumeAnlegen nimmt eine Liste mit einer Zeile je Raum: „Name; Bereich“. Ein Tabulator trennt
// ebenso, damit sich Spalten aus einer Tabelle direkt einfügen lassen.
func (a *app) raeumeAnlegen(w http.ResponseWriter, r *http.Request, _ benutzer) {
	var ein struct {
		Text string `json:"text"`
	}
	if !leseJSON(w, r, &ein) {
		return
	}
	var neue []raum
	for _, zeile := range strings.Split(ein.Text, "\n") {
		name, bereich, _ := strings.Cut(strings.ReplaceAll(zeile, "\t", ";"), ";")
		name, bereich = strings.TrimSpace(name), strings.TrimSpace(bereich)
		if name == "" {
			continue
		}
		if utf8.RuneCountInString(name) > 60 || utf8.RuneCountInString(bereich) > 80 {
			fehler(w, http.StatusBadRequest, "Raumname höchstens 60, Bereich höchstens 80 Zeichen: "+name)
			return
		}
		neue = append(neue, raum{Name: name, Bereich: bereich})
	}
	if len(neue) == 0 || len(neue) > 1000 {
		fehler(w, http.StatusBadRequest, "Bitte zwischen 1 und 1000 Räume eingeben, einen je Zeile.")
		return
	}
	n, err := a.sp.legeRaeumeAn(r.Context(), neue)
	if err != nil {
		intern(w, err)
		return
	}
	liste, err := a.sp.raeume(r.Context())
	if err != nil {
		intern(w, err)
		return
	}
	schreibeJSON(w, http.StatusOK, map[string]any{"neu": n, "raeume": liste})
}

func (a *app) raumLoeschen(w http.ResponseWriter, r *http.Request, _ benutzer) {
	id, ok := idAusPfad(w, r)
	if !ok {
		return
	}
	if err := a.sp.loescheRaum(r.Context(), id); err != nil {
		nichtGefundenOderIntern(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ── Meldungen ────────────────────────────────────────────────────────────────

// fuerAnsicht bereitet eine Meldung für die Antwort auf. Wer gemeldet hat, sehen nur Hausmeister,
// Leitung und Verwaltung, für Rückfragen.
func fuerAnsicht(m meldung, b benutzer) meldung {
	if !b.darfSehen() {
		m.GemeldetVon = ""
	}
	if m.Foto != "" {
		m.Foto = "/fotos/" + m.Foto
	}
	return m
}

func (a *app) meldungen(w http.ResponseWriter, r *http.Request, b benutzer) {
	offen, err := a.sp.offeneMeldungen(r.Context())
	if err != nil {
		intern(w, err)
		return
	}
	erledigt, err := a.sp.erledigteMeldungen(r.Context(), a.jetzt().AddDate(0, 0, -30))
	if err != nil {
		intern(w, err)
		return
	}
	for i := range offen {
		offen[i] = fuerAnsicht(offen[i], b)
	}
	for i := range erledigt {
		erledigt[i] = fuerAnsicht(erledigt[i], b)
	}
	schreibeJSON(w, http.StatusOK, map[string]any{"offen": offen, "erledigt": erledigt})
}

// meldungAnlegen nimmt ein Formular, weil ein Foto dabei sein kann. Datum und meldende Person
// setzt der Server, damit niemand etwas eintippen muss.
func (a *app) meldungAnlegen(w http.ResponseWriter, r *http.Request, b benutzer) {
	r.Body = http.MaxBytesReader(w, r.Body, fotoHoechstGroesse+1<<20)
	if err := r.ParseMultipartForm(4 << 20); err != nil {
		fehler(w, http.StatusBadRequest, "Das Foto ist zu groß oder die Meldung unvollständig.")
		return
	}
	defer func() { _ = r.MultipartForm.RemoveAll() }()

	m := meldung{
		Kategorie:   r.FormValue("kategorie"),
		Was:         strings.TrimSpace(r.FormValue("was")),
		Notiz:       strings.TrimSpace(r.FormValue("notiz")),
		Dringend:    r.FormValue("dringend") == "1",
		GemeldetVon: b.Email,
		GemeldetAm:  a.jetzt(),
	}
	// Im Funkloch gespeicherte Meldungen bringen ihren Zeitpunkt mit. Er gilt, wenn er plausibel ist
	// (nicht in der Zukunft, höchstens zwei Tage alt); sonst zählt die Uhr des Servers.
	if t, err := time.Parse(time.RFC3339, r.FormValue("erstellt")); err == nil &&
		t.Before(m.GemeldetAm.Add(5*time.Minute)) && t.After(m.GemeldetAm.Add(-48*time.Hour)) {
		m.GemeldetAm = t
	}
	if !kategorien[m.Kategorie] {
		fehler(w, http.StatusBadRequest, "Bitte eine Art des Schadens wählen.")
		return
	}
	if m.Was == "" || utf8.RuneCountInString(m.Was) > 160 || utf8.RuneCountInString(m.Notiz) > 600 {
		fehler(w, http.StatusBadRequest, "Bitte kurz beschreiben, was kaputt ist (höchstens 160 Zeichen, Notiz 600).")
		return
	}
	if roh := r.FormValue("raumId"); roh != "" {
		id, err := strconv.ParseInt(roh, 10, 64)
		if err != nil {
			fehler(w, http.StatusBadRequest, "Unbekannter Raum.")
			return
		}
		rm, err := a.sp.raum(r.Context(), id)
		if err != nil {
			nichtGefundenOderIntern(w, err)
			return
		}
		m.RaumID, m.Ort = &rm.ID, rm.Name
	} else {
		m.Ort = strings.TrimSpace(r.FormValue("ort"))
		if m.Ort == "" || utf8.RuneCountInString(m.Ort) > 80 {
			fehler(w, http.StatusBadRequest, "Bitte angeben, wo der Schaden ist.")
			return
		}
	}

	var foto []byte
	if datei, _, err := r.FormFile("foto"); err == nil {
		foto, err = bereiteFotoAuf(datei)
		_ = datei.Close()
		if err != nil {
			fehler(w, http.StatusBadRequest, "Das Foto ließ sich nicht lesen. Bitte ein JPG- oder PNG-Foto nehmen.")
			return
		}
		m.Foto = zufallsHex(16) + ".jpg"
	}

	neu, err := a.sp.legeMeldungAn(r.Context(), m, foto)
	if err != nil {
		intern(w, err)
		return
	}
	slog.Info("Neue Meldung", "id", neu.ID, "kategorie", neu.Kategorie, "dringend", neu.Dringend)
	a.benachrichtigeSofort(neu)
	schreibeJSON(w, http.StatusCreated, fuerAnsicht(neu, b))
}

func (a *app) statusSetzen(w http.ResponseWriter, r *http.Request, b benutzer) {
	id, ok := idAusPfad(w, r)
	if !ok {
		return
	}
	var ein struct {
		Status  *string `json:"status"`
		Antwort *string `json:"antwort"`
	}
	if !leseJSON(w, r, &ein) {
		return
	}
	if ein.Status != nil && !stati[*ein.Status] {
		fehler(w, http.StatusBadRequest, "Unbekannter Status.")
		return
	}
	if ein.Antwort != nil {
		t := strings.TrimSpace(*ein.Antwort)
		if utf8.RuneCountInString(t) > 300 {
			fehler(w, http.StatusBadRequest, "Die Antwort darf höchstens 300 Zeichen lang sein.")
			return
		}
		ein.Antwort = &t
	}
	neu, err := a.sp.setzeStatus(r.Context(), id, ein.Status, ein.Antwort, a.jetzt())
	if err != nil {
		nichtGefundenOderIntern(w, err)
		return
	}
	schreibeJSON(w, http.StatusOK, fuerAnsicht(neu, b))
}

func (a *app) meldungLoeschen(w http.ResponseWriter, r *http.Request, _ benutzer) {
	id, ok := idAusPfad(w, r)
	if !ok {
		return
	}
	if err := a.sp.loescheMeldung(r.Context(), id); err != nil {
		nichtGefundenOderIntern(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *app) foto(w http.ResponseWriter, r *http.Request, _ benutzer) {
	name := r.PathValue("name")
	if !fotoName.MatchString(name) {
		http.NotFound(w, r)
		return
	}
	a.liefereFoto(w, r, name)
}

func (a *app) liefereFoto(w http.ResponseWriter, r *http.Request, name string) {
	// Der Name ist zufällig und ändert sich nie; das Handy darf das Bild deshalb lange behalten.
	etag := `"` + name + `"`
	w.Header().Set("Cache-Control", "private, max-age=604800, immutable")
	w.Header().Set("ETag", etag)
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	daten, err := a.sp.foto(r.Context(), name)
	if err != nil {
		if errors.Is(err, errNichtGefunden) {
			http.NotFound(w, r)
			return
		}
		intern(w, err)
		return
	}
	w.Header().Set("Content-Type", "image/jpeg")
	_, _ = w.Write(daten)
}

// ── Oberfläche ───────────────────────────────────────────────────────────────

// oberflaeche liefert die gebaute Svelte-App. Pfade ohne Datei (/r/…, /uebersicht) bekommen die
// index.html, die ihren Weg selbst liest; die gehashten Dateien unter /assets/ dürfen ewig im Cache liegen.
func (a *app) oberflaeche() http.Handler {
	dateien := http.FileServerFS(a.statisch)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pfad := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if pfad != "" && pfad != "index.html" {
			if info, err := fs.Stat(a.statisch, pfad); err == nil && !info.IsDir() {
				if strings.HasPrefix(pfad, "assets/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				} else {
					w.Header().Set("Cache-Control", "no-cache")
				}
				dateien.ServeHTTP(w, r)
				return
			}
			if path.Ext(pfad) != "" {
				http.NotFound(w, r)
				return
			}
		}
		index, err := fs.ReadFile(a.statisch, "index.html")
		if err != nil {
			http.Error(w, "Die Oberfläche ist nicht gebaut: im Ordner frontend »npm run build« ausführen.",
				http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(index)
	})
}

// ── Hilfen ───────────────────────────────────────────────────────────────────

func schreibeJSON(w http.ResponseWriter, status int, wert any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(wert); err != nil {
		slog.Warn("Antwort nicht geschrieben", "fehler", err)
	}
}

func fehler(w http.ResponseWriter, status int, text string) {
	schreibeJSON(w, status, map[string]string{"fehler": text})
}

func intern(w http.ResponseWriter, err error) {
	slog.Error("Interner Fehler", "fehler", err)
	fehler(w, http.StatusInternalServerError, "Da ist etwas schiefgegangen. Bitte noch einmal versuchen.")
}

func nichtGefundenOderIntern(w http.ResponseWriter, err error) {
	if errors.Is(err, errNichtGefunden) {
		fehler(w, http.StatusNotFound, "Nicht gefunden – vielleicht wurde es gerade gelöscht.")
		return
	}
	intern(w, err)
}

func leseJSON(w http.ResponseWriter, r *http.Request, ziel any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 256<<10)
	if err := json.NewDecoder(r.Body).Decode(ziel); err != nil {
		fehler(w, http.StatusBadRequest, "Die Anfrage war nicht lesbar.")
		return false
	}
	return true
}

func idAusPfad(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		fehler(w, http.StatusBadRequest, "Ungültige Nummer.")
		return 0, false
	}
	return id, true
}
