package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Eine Excel-Tabelle schreiben, ohne fremde Bibliothek wie beim Lesen (xlsx.go): ein Zip mit sechs
// XML-Teilen. Texte stehen eingebettet (inlineStr), Zeitpunkte als echte Excel-Daten, damit sich
// die Spalten sortieren und filtern lassen. Die Kopfzeile ist fest und trägt den Autofilter.

type spalte struct {
	kopf   string
	breite float64
}

// tabelle: Zellen sind string, int64, time.Time oder *time.Time; leere Werte bleiben leere Zellen.
type tabelle struct {
	blatt   string
	spalten []spalte
	zeilen  [][]any
}

const xlsxKopf = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n"

const xlsxTypen = xlsxKopf + `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` +
	`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>` +
	`<Default Extension="xml" ContentType="application/xml"/>` +
	`<Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>` +
	`<Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>` +
	`<Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>` +
	`</Types>`

const xlsxRels = xlsxKopf + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
	`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>` +
	`</Relationships>`

const xlsxMappeRels = xlsxKopf + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
	`<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>` +
	`<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>` +
	`</Relationships>`

// Stile: 0 Vorgabe, 1 Kopf fett, 2 Zeitpunkt (Tag und Uhrzeit), 3 Text oben bündig mit Umbruch,
// damit Notizen über mehrere Zeilen lesbar bleiben.
const xlsxStile = xlsxKopf + `<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">` +
	`<numFmts count="1"><numFmt numFmtId="164" formatCode="dd.mm.yyyy hh:mm"/></numFmts>` +
	`<fonts count="2"><font><sz val="11"/><name val="Calibri"/></font><font><b/><sz val="11"/><name val="Calibri"/></font></fonts>` +
	`<fills count="2"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill></fills>` +
	`<borders count="1"><border><left/><right/><top/><bottom/><diagonal/></border></borders>` +
	`<cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>` +
	`<cellXfs count="4"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/>` +
	`<xf numFmtId="0" fontId="1" fillId="0" borderId="0" xfId="0" applyFont="1"/>` +
	`<xf numFmtId="164" fontId="0" fillId="0" borderId="0" xfId="0" applyNumberFormat="1" applyAlignment="1"><alignment vertical="top"/></xf>` +
	`<xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0" applyAlignment="1"><alignment vertical="top" wrapText="1"/></xf>` +
	`</cellXfs><cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles></styleSheet>`

func schreibeXLSX(t tabelle, zone *time.Location) ([]byte, error) {
	var puffer bytes.Buffer
	z := zip.NewWriter(&puffer)
	teile := []struct{ name, inhalt string }{
		{"[Content_Types].xml", xlsxTypen},
		{"_rels/.rels", xlsxRels},
		{"xl/workbook.xml", mappeXML(t)},
		{"xl/_rels/workbook.xml.rels", xlsxMappeRels},
		{"xl/styles.xml", xlsxStile},
		{"xl/worksheets/sheet1.xml", blattXML(t, zone)},
	}
	for _, teil := range teile {
		w, err := z.Create(teil.name)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write([]byte(teil.inhalt)); err != nil {
			return nil, err
		}
	}
	if err := z.Close(); err != nil {
		return nil, err
	}
	return puffer.Bytes(), nil
}

// mappeXML nennt das Blatt und den Bereich des Autofilters, wie Excel ihn selbst ablegt.
func mappeXML(t tabelle) string {
	var b strings.Builder
	b.WriteString(xlsxKopf + `<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" ` +
		`xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><sheets><sheet name="`)
	_ = xml.EscapeText(&b, []byte(t.blatt))
	b.WriteString(`" sheetId="1" r:id="rId1"/></sheets><definedNames>` +
		`<definedName name="_xlnm._FilterDatabase" localSheetId="0" hidden="1">'`)
	_ = xml.EscapeText(&b, []byte(strings.ReplaceAll(t.blatt, "'", "''")))
	fmt.Fprintf(&b, `'!$A$1:$%s$%d</definedName></definedNames></workbook>`, spaltenName(len(t.spalten)-1), len(t.zeilen)+1)
	return b.String()
}

func blattXML(t tabelle, zone *time.Location) string {
	var b strings.Builder
	b.WriteString(xlsxKopf + `<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">` +
		`<sheetViews><sheetView workbookViewId="0"><pane ySplit="1" topLeftCell="A2" activePane="bottomLeft" state="frozen"/>` +
		`</sheetView></sheetViews><sheetFormatPr defaultRowHeight="15"/><cols>`)
	for i, s := range t.spalten {
		fmt.Fprintf(&b, `<col min="%d" max="%d" width="%g" customWidth="1"/>`, i+1, i+1, s.breite)
	}
	b.WriteString(`</cols><sheetData><row r="1">`)
	for i, s := range t.spalten {
		fmt.Fprintf(&b, `<c r="%s1" t="inlineStr" s="1"><is><t>`, spaltenName(i))
		_ = xml.EscapeText(&b, []byte(s.kopf))
		b.WriteString(`</t></is></c>`)
	}
	b.WriteString(`</row>`)
	for z, zeile := range t.zeilen {
		fmt.Fprintf(&b, `<row r="%d">`, z+2)
		for i, wert := range zeile {
			zellXML(&b, spaltenName(i)+strconv.Itoa(z+2), wert, zone)
		}
		b.WriteString(`</row>`)
	}
	fmt.Fprintf(&b, `</sheetData><autoFilter ref="A1:%s%d"/></worksheet>`, spaltenName(len(t.spalten)-1), len(t.zeilen)+1)
	return b.String()
}

func zellXML(b *strings.Builder, bezug string, wert any, zone *time.Location) {
	switch v := wert.(type) {
	case string:
		if v == "" {
			return
		}
		fmt.Fprintf(b, `<c r="%s" t="inlineStr" s="3"><is><t xml:space="preserve">`, bezug)
		_ = xml.EscapeText(b, []byte(v))
		b.WriteString(`</t></is></c>`)
	case int64:
		fmt.Fprintf(b, `<c r="%s" s="3"><v>%d</v></c>`, bezug, v)
	case time.Time:
		fmt.Fprintf(b, `<c r="%s" s="2"><v>%s</v></c>`, bezug, strconv.FormatFloat(excelDatum(v, zone), 'f', -1, 64))
	case *time.Time:
		if v != nil {
			zellXML(b, bezug, *v, zone)
		}
	}
}

// excelDatum: Excel zählt Tage seit dem 30.12.1899 und kennt keine Zeitzone; es zählt die Uhrzeit,
// die an der Wand der Schule steht.
func excelDatum(t time.Time, zone *time.Location) float64 {
	w := t.In(zone)
	wand := time.Date(w.Year(), w.Month(), w.Day(), w.Hour(), w.Minute(), w.Second(), 0, time.UTC)
	return wand.Sub(time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)).Hours() / 24
}

// spaltenName macht aus 0 „A“, aus 26 „AA“.
func spaltenName(i int) string {
	name := ""
	for i++; i > 0; i = (i - 1) / 26 {
		name = string(rune('A'+(i-1)%26)) + name
	}
	return name
}
