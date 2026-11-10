package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Einzelne Meldungen zum Ausdrucken, jede auf einer eigenen A4-Seite mit großem Foto: Der Hausmeister
// nimmt die Blätter mit oder gibt sie einer Fachfirma. Wie der Export ohne Namen der Meldenden, denn
// das Papier verlässt die App. Mehr als druckHoechstens auf einmal machte die Datei zu groß für Handy
// und Mail; für alle zusammen gibt es das PDF der Liste.
const (
	druckHoechstens = 50
	druckFotoKante  = 1200
	druckSpalte     = 32.0 // Breite der Namen in den Angaben, in Millimetern
)

func (a *app) druckPDF(w http.ResponseWriter, r *http.Request, _ benutzer) {
	ids, ok := druckIDs(r.URL.Query().Get("ids"))
	if !ok {
		fehler(w, http.StatusBadRequest, fmt.Sprintf("Bitte 1 bis %d Meldungen wählen.", druckHoechstens))
		return
	}
	liste, err := a.sp.meldungenMitIDs(r.Context(), ids)
	if err != nil {
		intern(w, err)
		return
	}
	if len(liste) == 0 {
		fehler(w, http.StatusNotFound, "Diese Meldungen gibt es nicht mehr.")
		return
	}
	roh, err := a.schreibeDruck(r.Context(), liste)
	if err != nil {
		intern(w, err)
		return
	}
	name := "Schadensmeldungen"
	if len(liste) == 1 {
		name = fmt.Sprintf("Schadensmeldung-%d", liste[0].ID)
	}
	a.liefereDatei(w, roh, "application/pdf", name, "pdf", r.URL.Query().Get("zeigen") == "1")
}

// druckIDs liest „3,1,2“ in der Reihenfolge der Liste; doppelte zählen einmal.
func druckIDs(text string) ([]int64, bool) {
	var ids []int64
	gesehen := map[int64]bool{}
	for teil := range strings.SplitSeq(text, ",") {
		id, err := strconv.ParseInt(strings.TrimSpace(teil), 10, 64)
		if err != nil || id <= 0 {
			return nil, false
		}
		if !gesehen[id] {
			gesehen[id] = true
			ids = append(ids, id)
		}
	}
	return ids, len(ids) > 0 && len(ids) <= druckHoechstens
}

func (a *app) schreibeDruck(ctx context.Context, liste []meldung) ([]byte, error) {
	titel := "Schadensmeldungen"
	if len(liste) == 1 {
		titel = "Schadensmeldung: " + liste[0].Ort
	}
	s := a.neuesPDF(titel)
	gedruckt := "gedruckt " + langesDatum(a.jetzt().In(a.konf.Zeitzone))
	for _, m := range liste {
		s.p.AddPage()
		s.druckSeite(m, gedruckt, druckAngaben(m, a.konf.Zeitzone), a.pdfFoto(ctx, m.Foto, druckFotoKante))
	}
	return s.fertig()
}

// druckAngaben: dieselben Zeilen wie auf der Seite der Meldung in der App.
func druckAngaben(m meldung, zone *time.Location) [][2]string {
	angaben := [][2]string{{"Art", artName(m.Kategorie)}, {"Stand", statusNamen[m.Status]}, {"Gemeldet", langesDatum(m.GemeldetAm.In(zone))}}
	if m.Status == "erledigt" && m.ErledigtAm != nil {
		angaben = append(angaben, [2]string{"Erledigt", langesDatum(m.ErledigtAm.In(zone))})
	}
	return angaben
}

// druckSeite: oben klein Nummer und Druckdatum („Stand“ heißt hier der Bearbeitungsstand), groß der Ort mit Bereich, dann was kaputt ist, die Angaben,
// Beschreibung und Vermerk; das Foto bekommt, was von der Seite bleibt.
func (s *pdfSatz) druckSeite(m meldung, gedruckt string, angaben [][2]string, foto *pdfBild) {
	s.setze(pdfBreite, absatz{fmt.Sprintf("Schadensmeldung Nr. %d · %s", m.ID, gedruckt), "", 9, 5, pdfLeise})
	s.p.Ln(5)
	s.setze(pdfBreite, absatz{m.Ort, "B", 22, 10, pdfText})
	if m.Bereich != "" {
		s.setze(pdfBreite, absatz{m.Bereich, "", 12, 6, pdfLeise})
	}
	s.p.Ln(6)
	if m.Dringend && m.Status != "erledigt" {
		s.setze(pdfBreite, absatz{"DRINGEND", "B", 10, 6, pdfFehler})
	}
	s.setze(pdfBreite, absatz{m.Was, "B", 15, 7.5, pdfText})
	s.p.Ln(4)
	for _, zeile := range angaben {
		s.angabe(zeile[0], zeile[1])
	}
	for _, block := range [][2]string{{"Beschreibung", m.Notiz}, {"Vermerk", m.Antwort}} {
		if block[1] != "" {
			s.p.Ln(4)
			s.setze(pdfBreite, absatz{block[0], "", 9.5, 5, pdfLeise}, absatz{block[1], "", 11, 6, pdfText})
		}
	}
	if foto != nil {
		s.p.Ln(6)
		s.grossesFoto(foto)
	}
}

// angabe: der Name leise in einer schmalen Spalte, der Wert daneben.
func (s *pdfSatz) angabe(name, wert string) {
	y := s.p.GetY()
	s.schrift("", 10, pdfLeise)
	s.p.SetXY(pdfRand, y)
	s.p.CellFormat(druckSpalte, 6, s.tr(name), "", 0, "L", false, 0, "")
	s.schrift("", 11, pdfText)
	s.p.SetXY(pdfRand+druckSpalte, y)
	s.p.MultiCell(pdfBreite-druckSpalte, 6, s.tr(wert), "", "L", false)
}

// grossesFoto setzt das Foto so groß, wie die Seite über der Fußzeile Platz lässt, ohne es zu
// verzerren; bleibt fast nichts, fehlt es lieber, als in die Fußzeile zu ragen.
func (s *pdfSatz) grossesFoto(b *pdfBild) {
	oben := s.p.GetY()
	frei := pdfUnten - oben
	if frei < 20 {
		return
	}
	s.zeichneFoto(b, pdfRand, oben, min(pdfBreite/float64(b.breite), frei/float64(b.hoehe)))
}
