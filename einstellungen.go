package main

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"
)

// arten: die Arten von Schäden in der Reihenfolge der Kacheln, mit den Namen für Export und PDF.
// Dieselben Schlüssel und Namen stehen in frontend/src/lib/kategorien.js; namen_test.go hält beide
// gleich, damit eine umbenannte Kachel nicht in der Excel-Tabelle anders heißt.
var arten = []struct{ schluessel, name string }{
	{"licht", "Licht & Strom"},
	{"heizung", "Heizung & Lüftung"},
	{"wasser", "Wasser & WC"},
	{"fenster", "Fenster & Türen"},
	{"moebel", "Möbel & Tafel"},
	{"technik", "Medientechnik"},
	{"reinigung", "Reinigung"},
	{"sonstiges", "Sonstiges"},
}

// statusNamen: wie frontend/src/lib/status.js; die Schlüssel prüft zusätzlich der CHECK aus
// migrations/002.
var statusNamen = map[string]string{
	"offen":     "Offen",
	"in_arbeit": "In Arbeit",
	"bestellt":  "Material bestellt",
	"fachfirma": "Fachfirma / Kreis informiert",
	"erledigt":  "Erledigt",
}

// artName gibt den Namen einer Art; leer heißt unbekannt.
func artName(schluessel string) string {
	for _, a := range arten {
		if a.schluessel == schluessel {
			return a.name
		}
	}
	return ""
}

// einstellungen braucht keine Anmeldung: Die Anmeldeseite zeigt den Schulnamen, und das Formular
// soll die abgeschalteten Arten auch kennen, wenn es aus dem Funkloch startet.
func (a *app) einstellungen(w http.ResponseWriter, r *http.Request) {
	domain := ""
	if len(a.konf.Domains) == 1 {
		domain = a.konf.Domains[0]
	}
	aus, err := a.sp.artenAus(r.Context())
	if err != nil {
		intern(w, err)
		return
	}
	schreibeJSON(w, http.StatusOK, map[string]any{
		"schulname":        a.konf.Schulname,
		"domain":           domain,
		"fotoTage":         a.konf.FotoTage,
		"aufbewahrungTage": a.konf.AufbewahrungTage,
		"mailAktiv":        a.konf.mailAktiv(),
		"artenAus":         aus,
	})
}

// artSchalten blendet eine Art im Meldeformular aus oder wieder ein. Meldungen dieser Art nimmt der
// Server weiter an: Was im Funkloch schon im Ausgang lag, soll nicht verloren gehen.
func (a *app) artSchalten(w http.ResponseWriter, r *http.Request, _ benutzer) {
	var ein struct {
		Kategorie string `json:"kategorie"`
		Sichtbar  bool   `json:"sichtbar"`
	}
	if !leseJSON(w, r, &ein) {
		return
	}
	if artName(ein.Kategorie) == "" || ein.Kategorie == "sonstiges" {
		fehler(w, http.StatusBadRequest, "Diese Art lässt sich nicht abschalten.")
		return
	}
	if err := a.sp.schalteArt(r.Context(), ein.Kategorie, ein.Sichtbar); err != nil {
		intern(w, err)
		return
	}
	aus, err := a.sp.artenAus(r.Context())
	if err != nil {
		intern(w, err)
		return
	}
	schreibeJSON(w, http.StatusOK, map[string]any{"artenAus": aus})
}

func (s *speicher) artenAus(ctx context.Context) ([]string, error) {
	return alle(ctx, s.db, pgx.RowTo[string], "SELECT kategorie FROM arten_aus ORDER BY kategorie")
}

func (s *speicher) schalteArt(ctx context.Context, kategorie string, sichtbar bool) error {
	befehl := "INSERT INTO arten_aus (kategorie) VALUES ($1) ON CONFLICT DO NOTHING"
	if sichtbar {
		befehl = "DELETE FROM arten_aus WHERE kategorie = $1"
	}
	_, err := s.db.Exec(ctx, befehl, kategorie)
	return err
}
