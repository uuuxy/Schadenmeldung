package main

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"time"

	"github.com/phpdave11/gofpdf"
)

// Was im PDF steht: die Zeile unter dem Titel, die Absätze einer Meldung und ihr Foto.

func (a *app) pdfUnterzeile(anzahl int, aus pdfAuswahl) string {
	text := "Stand: " + langesDatum(a.jetzt().In(a.konf.Zeitzone))
	if a.konf.Schulname != "" {
		text = a.konf.Schulname + " · " + text
	}
	if aus.bereich != "" {
		text += " · nur " + aus.bereich
	}
	if anzahl == 1 {
		return text + " · 1 Meldung"
	}
	return fmt.Sprintf("%s · %d Meldungen", text, anzahl)
}

// pdfAbsaetze: Ort fett, darunter was kaputt ist, die Notiz, eine leise Zeile mit Art, Stand und
// Datum und zuletzt der interne Vermerk.
func pdfAbsaetze(m meldung, erledigt bool, zone *time.Location) []absatz {
	stand := statusNamen[m.Status] + " · gemeldet " + kurzesDatum(m.GemeldetAm.In(zone))
	if erledigt && m.ErledigtAm != nil {
		stand = "erledigt " + kurzesDatum(m.ErledigtAm.In(zone)) + " · gemeldet " + kurzesDatum(m.GemeldetAm.In(zone))
	}
	liste := []absatz{
		{m.Ort, "B", 11, 5.5, pdfText},
		{m.Was, "", 10.5, 5, pdfText},
	}
	if m.Notiz != "" {
		liste = append(liste, absatz{m.Notiz, "", 9.5, 4.5, pdfLeise})
	}
	liste = append(liste, absatz{artName(m.Kategorie) + " · " + stand, "", 8.5, 4.5, pdfLeise})
	if m.Antwort != "" {
		liste = append(liste, absatz{"Vermerk: " + m.Antwort, "I", 9.5, 4.5, pdfText})
	}
	return liste
}

// pdfBild ist ein verkleinertes Foto für den Druck.
type pdfBild struct {
	name          string
	daten         []byte
	breite, hoehe int // Pixel
}

func (b *pdfBild) massstab() float64 {
	return min(pdfFotoKante/float64(b.breite), pdfFotoKante/float64(b.hoehe))
}

func (b *pdfBild) hoeheMM() float64 { return float64(b.hoehe) * b.massstab() }

func (s *pdfSatz) zeichneFoto(b *pdfBild, oben float64) {
	opt := gofpdf.ImageOptions{ImageType: "JPG"}
	s.p.RegisterImageOptionsReader(b.name, opt, bytes.NewReader(b.daten))
	m := b.massstab()
	s.p.ImageOptions(b.name, pdfRand+pdfBreite-float64(b.breite)*m, oben, float64(b.breite)*m, b.hoeheMM(), false, opt, 0, "")
}

// pdfFoto holt das Foto einer Meldung klein: 400 Pixel genügen für 3 cm Druck und halten das PDF
// klein genug für eine Mail. Fehlt es oder lässt es sich nicht lesen, bleibt der Platz leer.
func (a *app) pdfFoto(ctx context.Context, name string) *pdfBild {
	if name == "" {
		return nil
	}
	roh, err := a.sp.foto(ctx, name)
	if err != nil {
		return nil
	}
	bild, err := jpeg.Decode(bytes.NewReader(roh))
	if err != nil {
		return nil
	}
	klein := aufWeiss(bild, 400)
	var aus bytes.Buffer
	if err := jpeg.Encode(&aus, klein, &jpeg.Options{Quality: 75}); err != nil {
		return nil
	}
	g := klein.Bounds()
	return &pdfBild{name: name, daten: aus.Bytes(), breite: g.Dx(), hoehe: g.Dy()}
}
