package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"
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

// meldungKurz ist alles, was das Kollegium von den Meldungen anderer bekommt: genug für „Hier schon
// gemeldet“, damit niemand doppelt meldet. Notiz, Vermerk, Foto und Zeitpunkte bleiben beim Backend.
type meldungKurz struct {
	ID         int64     `json:"id"`
	RaumID     *int64    `json:"raumId"`
	Kategorie  string    `json:"kategorie"`
	Was        string    `json:"was"`
	Status     string    `json:"status"`
	GemeldetAm time.Time `json:"gemeldetAm"`
}

func (a *app) meldungen(w http.ResponseWriter, r *http.Request, b benutzer) {
	offen, err := a.sp.offeneMeldungen(r.Context())
	if err != nil {
		intern(w, err)
		return
	}
	if !b.darfSehen() {
		kurz := make([]meldungKurz, len(offen))
		for i, m := range offen {
			kurz[i] = meldungKurz{ID: m.ID, RaumID: m.RaumID, Kategorie: m.Kategorie, Was: m.Was,
				Status: m.Status, GemeldetAm: m.GemeldetAm}
		}
		schreibeJSON(w, http.StatusOK, map[string]any{"offen": kurz, "erledigt": []meldungKurz{}})
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
	// Gegen ein Skript, das dem Hausmeister tausende Mails schickt. Bearbeiter melden ohne Grenze,
	// etwa bei einem Rundgang durchs Haus.
	if !b.darfBearbeiten() {
		if rest := a.meldeBremse.versuch(b.Email, a.jetzt()); rest > 0 {
			fehler(w, http.StatusTooManyRequests,
				fmt.Sprintf("Sehr viele Meldungen in kurzer Zeit. Bitte in %d Minuten erneut versuchen.", int(rest.Minutes())+1))
			return
		}
	}
	r.Body = http.MaxBytesReader(w, r.Body, fotoHoechstGroesse+1<<20)
	if err := r.ParseMultipartForm(4 << 20); err != nil {
		fehler(w, http.StatusBadRequest, "Das Foto ist zu groß oder die Meldung unvollständig.")
		return
	}
	defer func() { _ = r.MultipartForm.RemoveAll() }()

	m, err := a.leseMeldung(r, b)
	if err == nil && a.schonDa(w, r, b, m.Kennung) {
		return
	}
	var foto []byte
	if err == nil {
		foto, err = leseFoto(r, &m)
	}
	if antworteFehler(w, err) {
		return
	}
	neu, err := a.sp.legeMeldungAn(r.Context(), m, foto)
	if errors.Is(err, errSchonDa) && a.schonDa(w, r, b, m.Kennung) {
		return
	}
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
	// Nicht im Browser behalten: Auf einem geteilten Gerät soll nach dem Abmelden kein Foto mehr im
	// Speicher liegen. Die Bilder sind klein (höchstens 1600 px), neu laden kostet wenig.
	w.Header().Set("Cache-Control", "private, no-store")
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
