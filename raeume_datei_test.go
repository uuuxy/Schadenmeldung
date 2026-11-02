package main

import (
	"archive/zip"
	"bytes"
	"mime/multipart"
	"net/http"
	"testing"
)

// Was Excel unter Windows als CSV speichert: Semikolon, Kopfzeile, Umlaute in Windows-1252.
func TestCSVAusExcelUnterWindows(t *testing.T) {
	zeilen, err := leseCSV([]byte("Raum;Geb\xe4ude\r\n101;Hauptgeb\xe4ude\r\nB\xfchne;Aula-Trakt\r\n;leer\r\n"))
	if err != nil {
		t.Fatal(err)
	}
	if text, n := raumZeilen(zeilen); n != 2 || text != "101; Hauptgebäude\nBühne; Aula-Trakt\n" {
		t.Errorf("%d Räume: %q", n, text)
	}
	zeilen, _ = leseCSV([]byte("\xef\xbb\xbf204,2. OG\nTurnhalle\n"))
	if text, n := raumZeilen(zeilen); n != 2 || text != "204; 2. OG\nTurnhalle\n" {
		t.Errorf("Komma ohne Kopfzeile: %d, %q", n, text)
	}
}

// Eine xlsx wie von Excel: geteilte Texte (einer formatiert in Abschnitten, einer mit Lautschrift),
// eine Zahl, ein eingebetteter Text, eine Zeilenlücke. Das Blatt heißt nicht sheet1, damit der
// Verweis aus der Arbeitsmappe zählt.
func TestXLSXWieExcel(t *testing.T) {
	zeilen, err := leseXLSX(beispielXLSX(t))
	if err != nil {
		t.Fatal(err)
	}
	if text, n := raumZeilen(zeilen); n != 3 || text != "101; Hauptgebäude\nAula\nBio 1; Neubau\n" {
		t.Errorf("%d Räume: %q", n, text)
	}
	if _, err := leseXLSX([]byte("PK\x03\x04kaputt")); err == nil {
		t.Error("kaputte Datei ohne Fehler gelesen")
	}
}

func TestSpalteAus(t *testing.T) {
	for bezug, erwartet := range map[string]int{"A1": 0, "Z9": 25, "AA3": 26, "XFD1": 16383, "": 7} {
		if s := spalteAus(bezug, 7); s != erwartet {
			t.Errorf("spalteAus(%q) = %d, erwartet %d", bezug, s, erwartet)
		}
	}
}

func TestRaeumeAusDateiUeberDieTuer(t *testing.T) {
	u := neueTestUmgebung(t)
	hm, lehrer := u.nutzer("hm"), u.nutzer("lehrer")
	if status, _ := lehrer.anfrage("POST", "/api/raeume/lesen", datei(t, beispielXLSX(t)), typVon); status != http.StatusForbidden {
		t.Errorf("Kollegium liest eine Raumdatei: Status %d", status)
	}
	status, antwort := hm.anfrage("POST", "/api/raeume/lesen", datei(t, beispielXLSX(t)), typVon)
	if status != 200 || antwort["anzahl"] != float64(3) || antwort["text"] != "101; Hauptgebäude\nAula\nBio 1; Neubau\n" {
		t.Errorf("Hausmeister: %d, %v", status, antwort)
	}
	status, antwort = hm.anfrage("POST", "/api/raeume/lesen", datei(t, []byte("PK\x03\x04kaputt")), typVon)
	if status != http.StatusBadRequest || antwort["fehler"] != "Die Excel-Datei ließ sich nicht öffnen." {
		t.Errorf("kaputte Datei: %d, %v", status, antwort)
	}
}

// typVon trägt den Inhaltstyp des letzten Formulars aus datei, mit Grenze.
var typVon string

func datei(t *testing.T, inhalt []byte) *bytes.Buffer {
	t.Helper()
	var puffer bytes.Buffer
	mw := multipart.NewWriter(&puffer)
	fw, err := mw.CreateFormFile("datei", "raeume.xlsx")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = fw.Write(inhalt)
	_ = mw.Close()
	typVon = mw.FormDataContentType()
	return &puffer
}

func beispielXLSX(t *testing.T) []byte {
	t.Helper()
	const haupt = `xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"`
	teile := map[string]string{
		"xl/workbook.xml": `<workbook ` + haupt + ` xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">` +
			`<sheets><sheet name="Räume" sheetId="1" r:id="rId3"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` +
			`<Relationship Id="rId1" Target="styles.xml"/><Relationship Id="rId3" Target="worksheets/sheet3.xml"/></Relationships>`,
		"xl/sharedStrings.xml": `<sst ` + haupt + `><si><t>Raumnummer</t></si><si><t>Gebäude</t></si>` +
			`<si><r><rPr><b/></rPr><t>Haupt</t></r><r><t>gebäude</t></r></si>` +
			`<si><t>Neubau</t><rPh sb="0" eb="1"><t>X</t></rPh></si></sst>`,
		"xl/worksheets/sheet3.xml": `<worksheet ` + haupt + `><sheetData>` +
			`<row r="1"><c r="A1" t="s"><v>0</v></c><c r="B1" t="s"><v>1</v></c></row>` +
			`<row r="2"><c r="A2"><v>101</v></c><c r="B2" t="s"><v>2</v></c></row>` +
			`<row r="3"><c r="A3" t="inlineStr"><is><t>Aula</t></is></c></row>` +
			`<row r="5"><c r="A5" t="str"><f>"Bio "&amp;1</f><v>Bio 1</v></c><c r="B5" t="s"><v>3</v></c></row>` +
			`<row r="6"><c r="B6" t="s"><v>3</v></c></row></sheetData></worksheet>`,
		"xl/worksheets/sheet1.xml": `<worksheet ` + haupt + `><sheetData><row><c t="inlineStr"><is><t>Falsches Blatt</t></is></c></row></sheetData></worksheet>`,
	}
	var puffer bytes.Buffer
	zw := zip.NewWriter(&puffer)
	for name, inhalt := range teile {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte(inhalt))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return puffer.Bytes()
}
