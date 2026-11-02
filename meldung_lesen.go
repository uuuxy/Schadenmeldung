package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// eingabeFehler ist eine Eingabe, die der Server ablehnt. Der Text geht so an die Oberfläche; alle
// anderen Fehler werden zu 404 oder 500 (antworteFehler).
type eingabeFehler string

func (e eingabeFehler) Error() string { return string(e) }

// antworteFehler schreibt die Antwort zu err und sagt, ob es eine gab. So bleibt in jeder Tür eine
// einzige Abzweigung statt einer je Prüfung.
func antworteFehler(w http.ResponseWriter, err error) bool {
	var eingabe eingabeFehler
	switch {
	case err == nil:
		return false
	case errors.As(err, &eingabe):
		fehler(w, http.StatusBadRequest, string(eingabe))
	default:
		nichtGefundenOderIntern(w, err)
	}
	return true
}

// leseMeldung baut die Meldung aus dem Formular und prüft Art, Beschreibung und Ort.
func (a *app) leseMeldung(r *http.Request, b benutzer) (meldung, error) {
	m := meldung{
		Kategorie:   r.FormValue("kategorie"),
		Was:         strings.TrimSpace(r.FormValue("was")),
		Notiz:       strings.TrimSpace(r.FormValue("notiz")),
		Dringend:    r.FormValue("dringend") == "1",
		GemeldetVon: b.Email,
		GemeldetAm:  zeitpunktAusAusgang(r.FormValue("erstellt"), a.jetzt()),
	}
	if !kategorien[m.Kategorie] {
		return m, eingabeFehler("Bitte eine Art des Schadens wählen.")
	}
	if m.Was == "" || utf8.RuneCountInString(m.Was) > 160 || utf8.RuneCountInString(m.Notiz) > 600 {
		return m, eingabeFehler("Bitte kurz beschreiben, was kaputt ist (höchstens 160 Zeichen, Notiz 600).")
	}
	return m, a.setzeOrt(r, &m)
}

// zeitpunktAusAusgang: Im Funkloch gespeicherte Meldungen bringen ihren Zeitpunkt mit. Er gilt, wenn
// er plausibel ist (nicht in der Zukunft, höchstens zwei Tage alt); sonst zählt die Uhr des Servers.
func zeitpunktAusAusgang(roh string, jetzt time.Time) time.Time {
	t, err := time.Parse(time.RFC3339, roh)
	if err != nil || !t.Before(jetzt.Add(5*time.Minute)) || !t.After(jetzt.Add(-48*time.Hour)) {
		return jetzt
	}
	return t
}

// setzeOrt nimmt den gewählten Raum oder, ohne Raum, den frei eingegebenen Ort.
func (a *app) setzeOrt(r *http.Request, m *meldung) error {
	roh := r.FormValue("raumId")
	if roh == "" {
		m.Ort = strings.TrimSpace(r.FormValue("ort"))
		if m.Ort == "" || utf8.RuneCountInString(m.Ort) > 80 {
			return eingabeFehler("Bitte angeben, wo der Schaden ist.")
		}
		return nil
	}
	id, err := strconv.ParseInt(roh, 10, 64)
	if err != nil {
		return eingabeFehler("Unbekannter Raum.")
	}
	rm, err := a.sp.raum(r.Context(), id)
	if err != nil {
		return err
	}
	m.RaumID, m.Ort = &rm.ID, rm.Name
	return nil
}

// leseFoto verkleinert ein mitgeschicktes Foto und gibt ihm einen Namen; ohne Foto bleibt alles leer.
func leseFoto(r *http.Request, m *meldung) ([]byte, error) {
	datei, _, err := r.FormFile("foto")
	if err != nil {
		return nil, nil
	}
	defer func() { _ = datei.Close() }()
	foto, err := bereiteFotoAuf(datei)
	if err != nil {
		return nil, eingabeFehler("Das Foto ließ sich nicht lesen. Bitte ein JPG- oder PNG-Foto nehmen.")
	}
	m.Foto = zufallsHex(16) + ".jpg"
	return foto, nil
}
