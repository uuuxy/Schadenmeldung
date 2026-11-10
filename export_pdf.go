package main

import (
	"bytes"
	"context"
	"fmt"

	"github.com/phpdave11/gofpdf"
)

// Das PDF der Übersicht: A4 hoch, nach Bereichen gegliedert, je Meldung ein Block mit kleinem Foto.
// Maße in Millimetern; die Farben sind die Rollen der App (on-surface, on-surface-variant, primary,
// error, outline-variant), damit der Ausdruck aussieht wie die App.
const (
	pdfRand      = 15.0
	pdfBreite    = 210 - 2*pdfRand
	pdfUnten     = 297 - pdfRand - 6 // darunter steht die Fußzeile
	pdfFotoKante = 30.0
)

type rgb [3]int

var (
	pdfText    = rgb{26, 28, 30}
	pdfLeise   = rgb{66, 71, 78}
	pdfPrimaer = rgb{0, 97, 164}
	pdfFehler  = rgb{186, 26, 26}
	pdfLinie   = rgb{194, 199, 207}
)

type pdfSatz struct {
	p  *gofpdf.Fpdf
	tr func(string) string
	// bereich: unter welchem Bereich gerade gesetzt wird; nach einem Seitenumbruch steht er erneut oben.
	bereich *string
}

// absatz ist eine Zeile oder ein umbrechender Text eines Eintrags in einer Schrift.
type absatz struct {
	text, stil     string
	groesse, hoehe float64
	farbe          rgb
}

func (a *app) schreibePDF(ctx context.Context, liste []meldung, aus pdfAuswahl) ([]byte, error) {
	titel := "Offene Schadensmeldungen"
	if aus.erledigt {
		titel = "Erledigte Schadensmeldungen der letzten 30 Tage"
	}
	s := a.neuesPDF(titel)
	s.p.AddPage()
	s.kopf(titel, a.pdfUnterzeile(len(liste), aus))
	for _, m := range liste {
		s.eintrag(m, a.pdfFoto(ctx, m.Foto, 400), pdfAbsaetze(m, aus.erledigt, a.konf.Zeitzone))
	}
	if len(liste) == 0 {
		s.setze(pdfBreite, absatz{"Keine Meldungen.", "", 11, 6, pdfLeise})
	}
	return s.fertig()
}

// neuesPDF: A4 hoch mit Rändern, Titel und Fußzeile, wie Liste und Einzelblatt sie teilen.
func (a *app) neuesPDF(titel string) *pdfSatz {
	p := gofpdf.New("P", "mm", "A4", "")
	p.SetMargins(pdfRand, pdfRand, pdfRand)
	p.SetAutoPageBreak(false, 0)
	p.AliasNbPages("{seiten}")
	s := &pdfSatz{p: p, tr: pdfUebersetzer(p.UnicodeTranslatorFromDescriptor(""))}
	p.SetTitle(titel, true)
	p.SetCreator("Schadensmeldung", true)
	p.SetFooterFunc(func() { s.fuss(a.konf.Schulname) })
	return s
}

func (s *pdfSatz) fertig() ([]byte, error) {
	var puffer bytes.Buffer
	err := s.p.Output(&puffer)
	return puffer.Bytes(), err
}

// setze schreibt Absätze in der gegebenen Breite untereinander, jeden in seiner Schrift.
func (s *pdfSatz) setze(breite float64, absaetze ...absatz) {
	for _, z := range absaetze {
		s.schrift(z.stil, z.groesse, z.farbe)
		s.p.SetX(pdfRand)
		s.p.MultiCell(breite, z.hoehe, s.tr(z.text), "", "L", false)
	}
}

// eintrag setzt eine Meldung. Passt sie nicht mehr auf die Seite, beginnt eine neue; ein Eintrag
// wird nie geteilt.
func (s *pdfSatz) eintrag(m meldung, foto *pdfBild, absaetze []absatz) {
	breite := pdfBreite
	if foto != nil {
		breite -= pdfFotoKante + 4
	}
	hoehe := s.hoehe(absaetze, breite)
	if foto != nil {
		hoehe = max(hoehe, foto.hoeheMM())
	}
	neuerBereich := s.bereich == nil || *s.bereich != m.Bereich
	bedarf := hoehe + 6
	if neuerBereich {
		bedarf += 11
	}
	switch {
	case s.p.GetY()+bedarf > pdfUnten:
		s.p.AddPage()
		s.bereichsKopf(m.Bereich, !neuerBereich)
	case neuerBereich:
		s.bereichsKopf(m.Bereich, false)
	}
	oben := s.p.GetY() + 3
	if m.Dringend && m.Status != "erledigt" {
		s.schrift("B", 8.5, pdfFehler)
		s.p.SetXY(pdfRand, oben)
		s.p.CellFormat(breite, 5.5, "DRINGEND", "", 0, "R", false, 0, "")
	}
	s.p.SetXY(pdfRand, oben)
	s.setze(breite, absaetze...)
	if foto != nil {
		m := foto.massstab()
		s.zeichneFoto(foto, pdfRand+pdfBreite-float64(foto.breite)*m, oben, m)
	}
	s.p.SetY(oben + hoehe + 3)
	s.linie()
}

func (s *pdfSatz) hoehe(absaetze []absatz, breite float64) float64 {
	summe := 0.0
	for _, z := range absaetze {
		s.schrift(z.stil, z.groesse, z.farbe)
		summe += float64(len(s.p.SplitLines([]byte(s.tr(z.text)), breite))) * z.hoehe
	}
	return summe
}

func (s *pdfSatz) kopf(titel, unterzeile string) {
	s.schrift("B", 16, pdfText)
	s.p.MultiCell(pdfBreite, 8, s.tr(titel), "", "L", false)
	s.schrift("", 9, pdfLeise)
	s.p.MultiCell(pdfBreite, 5, s.tr(unterzeile), "", "L", false)
	s.p.Ln(3)
}

func (s *pdfSatz) bereichsKopf(bereich string, fortsetzung bool) {
	name := bereich
	if name == "" {
		name = "Weitere Orte"
	}
	if fortsetzung {
		name += " (Fortsetzung)"
	}
	s.p.Ln(3)
	s.schrift("B", 11, pdfPrimaer)
	s.p.CellFormat(pdfBreite, 7, s.tr(name), "", 1, "L", false, 0, "")
	s.linie()
	s.bereich = &bereich
}

func (s *pdfSatz) fuss(schulname string) {
	s.p.SetXY(pdfRand, 297-pdfRand+3)
	s.schrift("", 8, pdfLeise)
	links := "Schadensmeldung"
	if schulname != "" {
		links += " · " + schulname
	}
	s.p.CellFormat(pdfBreite/2, 4, s.tr(links), "", 0, "L", false, 0, "")
	s.p.CellFormat(pdfBreite/2, 4, fmt.Sprintf("Seite %d von {seiten}", s.p.PageNo()), "", 0, "R", false, 0, "")
}

func (s *pdfSatz) schrift(stil string, groesse float64, farbe rgb) {
	s.p.SetFont("Helvetica", stil, groesse)
	s.p.SetTextColor(farbe[0], farbe[1], farbe[2])
}

func (s *pdfSatz) linie() {
	s.p.SetDrawColor(pdfLinie[0], pdfLinie[1], pdfLinie[2])
	s.p.SetLineWidth(0.2)
	s.p.Line(pdfRand, s.p.GetY(), pdfRand+pdfBreite, s.p.GetY())
}
