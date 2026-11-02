package main

import (
	"net/http"
	"strings"
	"unicode/utf8"
)

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

// raumAendern benennt einen Raum um oder setzt ihn in einen anderen Bereich. Sein QR-Code bleibt,
// der Aufkleber an der Tür gilt also weiter; seine Meldungen tragen danach den neuen Namen.
func (a *app) raumAendern(w http.ResponseWriter, r *http.Request, _ benutzer) {
	id, ok := idAusPfad(w, r)
	if !ok {
		return
	}
	var ein raum
	if !leseJSON(w, r, &ein) {
		return
	}
	ein.ID, ein.Name, ein.Bereich = id, strings.TrimSpace(ein.Name), strings.TrimSpace(ein.Bereich)
	if ein.Name == "" || utf8.RuneCountInString(ein.Name) > 60 || utf8.RuneCountInString(ein.Bereich) > 80 {
		fehler(w, http.StatusBadRequest, "Bitte einen Namen eingeben (höchstens 60 Zeichen, Bereich 80).")
		return
	}
	if antworteFehler(w, a.sp.aendereRaum(r.Context(), ein)) {
		return
	}
	a.verteiler.melde()
	liste, err := a.sp.raeume(r.Context())
	if err != nil {
		intern(w, err)
		return
	}
	schreibeJSON(w, http.StatusOK, liste)
}
