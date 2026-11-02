package main

import (
	"encoding/xml"
	"strconv"
	"strings"
)

// blattLeser liest ein Tabellenblatt: <row> mit <c r="B3" t="s"><v>…</v></c>. Texte stehen als
// Nummer in sharedStrings (t="s") oder eingebettet (<is><t>), Zahlen und Formelergebnisse in <v>.
type blattLeser struct {
	texte  []string
	zeilen [][]string
	zeile  []string
	spalte int
	typ    string
	wert   strings.Builder
	inWert bool
}

func (b *blattLeser) start(e xml.StartElement) {
	switch e.Name.Local {
	case "row":
		b.zeile = nil
	case "c":
		b.spalte = spalteAus(attribut(e, "r"), len(b.zeile))
		b.typ = attribut(e, "t")
		b.wert.Reset()
	case "v", "t":
		b.inWert = true
	}
}

func (b *blattLeser) ende(name string) {
	switch name {
	case "v", "t":
		b.inWert = false
	case "c":
		b.setze(b.zellwert())
	case "row":
		if len(b.zeilen) < xlsxZeilen {
			b.zeilen = append(b.zeilen, b.zeile)
		}
	}
}

func (b *blattLeser) text(t []byte) {
	if b.inWert {
		b.wert.Write(t)
	}
}

// zellwert löst geteilte Texte auf; alles andere steht schon als Text da.
func (b *blattLeser) zellwert() string {
	roh := b.wert.String()
	if b.typ != "s" {
		return roh
	}
	i, err := strconv.Atoi(strings.TrimSpace(roh))
	if err != nil || i < 0 || i >= len(b.texte) {
		return ""
	}
	return b.texte[i]
}

func (b *blattLeser) setze(wert string) {
	if b.spalte < 0 || b.spalte >= xlsxSpalten || wert == "" {
		return
	}
	for len(b.zeile) <= b.spalte {
		b.zeile = append(b.zeile, "")
	}
	b.zeile[b.spalte] = wert
}

// spalteAus macht aus „AB12“ die Spalte 27 (ab 0). Ohne Angabe gilt die nächste Spalte; mehr als
// drei Buchstaben kennt Excel nicht (letzte Spalte XFD).
func spalteAus(bezug string, sonst int) int {
	spalte, buchstaben := 0, 0
	for _, c := range bezug {
		if c < 'A' || c > 'Z' || buchstaben == 3 {
			break
		}
		spalte = spalte*26 + int(c-'A'+1)
		buchstaben++
	}
	if buchstaben == 0 {
		return sonst
	}
	return spalte - 1
}
