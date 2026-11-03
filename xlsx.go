package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"io"
	"strings"
)

// Excel-Dateien ohne fremde Bibliothek: Eine xlsx ist ein Zip mit XML. Für eine Raumliste genügen
// die erste Tabelle, geteilte und eingebettete Texte und Zahlen.
const (
	xlsxTeilGrenze = 20 << 20 // je Teil entpackt, damit eine Zip-Bombe den Speicher nicht füllt
	xlsxZeilen     = 5000
	xlsxSpalten    = 50
	xlsxTiefe      = 50 // Verschachtelung; eine echte Tabelle ist kaum zehn Ebenen tief
)

func leseXLSX(roh []byte) ([][]string, error) {
	z, err := zip.NewReader(bytes.NewReader(roh), int64(len(roh)))
	if err != nil {
		return nil, eingabeFehler("Die Excel-Datei ließ sich nicht öffnen.")
	}
	teile := map[string]*zip.File{}
	for _, f := range z.File {
		teile[f.Name] = f
	}
	texte := &textSammler{}
	if err := lese(teile["xl/sharedStrings.xml"], texte); err != nil {
		return nil, err
	}
	blatt := teile[ersteTabelle(teile)]
	if blatt == nil {
		return nil, eingabeFehler("In der Excel-Datei steht keine Tabelle.")
	}
	zellen := &blattLeser{texte: texte.texte}
	if err := lese(blatt, zellen); err != nil {
		return nil, err
	}
	return zellen.zeilen, nil
}

// xmlLeser bekommt die Teile eines XML-Dokuments der Reihe nach.
type xmlLeser interface {
	start(e xml.StartElement)
	ende(name string)
	text(b []byte)
}

// lese reicht einen Teil der Datei an einen xmlLeser weiter; ein fehlender Teil ist leer.
func lese(f *zip.File, l xmlLeser) error {
	if f == nil {
		return nil
	}
	rc, err := f.Open()
	if err != nil {
		return eingabeFehler("Die Excel-Datei ist beschädigt.")
	}
	defer func() { _ = rc.Close() }()
	d := xml.NewDecoder(io.LimitReader(rc, xlsxTeilGrenze))
	tiefe := 0
	for {
		t, err := d.Token()
		switch e := t.(type) {
		case xml.StartElement:
			// Tausendfach verschachtelt füllte schon eine winzige Datei den Speicher.
			if tiefe++; tiefe > xlsxTiefe {
				return eingabeFehler("Die Excel-Datei ist beschädigt.")
			}
			l.start(e)
		case xml.EndElement:
			tiefe--
			l.ende(e.Name.Local)
		case xml.CharData:
			l.text(e)
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return eingabeFehler("Die Excel-Datei ist beschädigt.")
		}
	}
}

// ersteTabelle folgt workbook.xml und dessen Verweisen zum ersten Blatt; sonst gilt sheet1.xml.
func ersteTabelle(teile map[string]*zip.File) string {
	verweise := &attributSammler{element: "sheet", attribut: "id"}
	_ = lese(teile["xl/workbook.xml"], verweise)
	ziele := &attributSammler{element: "Relationship", attribut: "Target", wennAttribut: "Id"}
	if len(verweise.werte) > 0 {
		ziele.wennWert = verweise.werte[0]
		_ = lese(teile["xl/_rels/workbook.xml.rels"], ziele)
	}
	switch {
	case len(ziele.werte) == 0:
		return "xl/worksheets/sheet1.xml"
	case strings.HasPrefix(ziele.werte[0], "/"):
		return strings.TrimPrefix(ziele.werte[0], "/")
	default:
		return "xl/" + ziele.werte[0]
	}
}

// attributSammler merkt sich ein Attribut jedes passenden Elements, auf Wunsch nur bei einem Wert.
type attributSammler struct {
	element, attribut      string
	wennAttribut, wennWert string
	werte                  []string
}

func (s *attributSammler) start(e xml.StartElement) {
	if e.Name.Local != s.element || (s.wennAttribut != "" && attribut(e, s.wennAttribut) != s.wennWert) {
		return
	}
	s.werte = append(s.werte, attribut(e, s.attribut))
}
func (s *attributSammler) ende(string) {}
func (s *attributSammler) text([]byte) {}

func attribut(e xml.StartElement, name string) string {
	for _, a := range e.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

// textSammler liest sharedStrings.xml: Jeder <si> ist ein Text, formatiert auch in mehreren <r><t>;
// die Lautschrift in <rPh> gehört nicht dazu.
type textSammler struct {
	texte       []string
	aktuell     strings.Builder
	inT, inLaut bool
}

func (s *textSammler) start(e xml.StartElement) {
	switch e.Name.Local {
	case "si":
		s.aktuell.Reset()
	case "t":
		s.inT = true
	case "rPh":
		s.inLaut = true
	}
}

func (s *textSammler) ende(name string) {
	switch name {
	case "si":
		s.texte = append(s.texte, s.aktuell.String())
	case "t":
		s.inT = false
	case "rPh":
		s.inLaut = false
	}
}

func (s *textSammler) text(b []byte) {
	if s.inT && !s.inLaut {
		s.aktuell.Write(b)
	}
}
