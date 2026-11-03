package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"
)

var fotoName = regexp.MustCompile(`^[0-9a-f]{32}\.jpg$`)

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

	m, err := a.leseMeldung(r, b)
	var foto []byte
	if err == nil {
		foto, err = leseFoto(r, &m)
	}
	if antworteFehler(w, err) {
		return
	}
	neu, err := a.sp.legeMeldungAn(r.Context(), m, foto)
	if err != nil {
		intern(w, err)
		return
	}
	slog.Info("Neue Meldung", "id", neu.ID, "kategorie", neu.Kategorie, "dringend", neu.Dringend)
	a.verteiler.melde()
	// Die Meldung steht; reißt die Verbindung jetzt ab (Funkloch), soll die Mail trotzdem hinaus.
	a.mailVormerken(context.WithoutCancel(r.Context()), neu)
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
	if ein.Status != nil && statusNamen[*ein.Status] == "" {
		fehler(w, http.StatusBadRequest, "Unbekannter Status.")
		return
	}
	if ein.Antwort != nil {
		t := strings.TrimSpace(*ein.Antwort)
		if utf8.RuneCountInString(t) > 300 {
			fehler(w, http.StatusBadRequest, "Der Vermerk darf höchstens 300 Zeichen lang sein.")
			return
		}
		ein.Antwort = &t
	}
	neu, err := a.sp.setzeStatus(r.Context(), id, ein.Status, ein.Antwort, a.jetzt())
	if err != nil {
		nichtGefundenOderIntern(w, err)
		return
	}
	a.verteiler.melde()
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
	a.verteiler.melde()
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
