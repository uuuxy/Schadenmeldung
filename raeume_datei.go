package main

import (
	"bytes"
	"encoding/csv"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Räume aus einer Datei, Excel (xlsx) oder CSV. Der Server liest die Tabelle und antwortet mit Zeilen
// „Raum; Bereich“. Angelegt wird erst über raeumeAnlegen, nachdem jemand die Zeilen gesehen hat.
const raumDateiGrenze = 5 << 20

var (
	kopfRaum    = regexp.MustCompile(`(?i)^(raum|räume|raumnummer|raumname|raum-nr\.?|nr\.?|nummer|name|bezeichnung)$`)
	kopfBereich = regexp.MustCompile(`(?i)^(bereich|gebäude|gebaeude|haus|trakt|etage|stockwerk|geschoss|ort|standort)$`)
	// Trenner der Zeilen dürfen nicht im Namen stehen, sonst zerfiele ein Raum in zwei.
	raumSauber = strings.NewReplacer(";", ",", "\t", " ", "\r", " ", "\n", " ")
)

func (a *app) raeumeLesen(w http.ResponseWriter, r *http.Request, _ benutzer) {
	r.Body = http.MaxBytesReader(w, r.Body, raumDateiGrenze+1<<20)
	if err := r.ParseMultipartForm(1 << 20); err != nil {
		fehler(w, http.StatusBadRequest, "Die Datei ist zu groß, höchstens 5 MB.")
		return
	}
	defer func() { _ = r.MultipartForm.RemoveAll() }()
	zeilen, err := leseHochgeladeneTabelle(r)
	if antworteFehler(w, err) {
		return
	}
	text, anzahl := raumZeilen(zeilen)
	if anzahl == 0 {
		fehler(w, http.StatusBadRequest, "In der Datei stehen keine Räume. Erwartet: erste Spalte Raum, zweite Bereich.")
		return
	}
	schreibeJSON(w, http.StatusOK, map[string]any{"text": text, "anzahl": anzahl})
}

func leseHochgeladeneTabelle(r *http.Request) ([][]string, error) {
	datei, _, err := r.FormFile("datei")
	if err != nil {
		return nil, eingabeFehler("Bitte eine Excel- oder CSV-Datei wählen.")
	}
	defer func() { _ = datei.Close() }()
	roh, err := io.ReadAll(io.LimitReader(datei, raumDateiGrenze+1))
	if err != nil || len(roh) > raumDateiGrenze {
		return nil, eingabeFehler("Die Datei ist zu groß, höchstens 5 MB.")
	}
	// Excel erkennt man am Kopf des Zip, nicht an der Endung; alles andere ist Text.
	if bytes.HasPrefix(roh, []byte("PK\x03\x04")) {
		return leseXLSX(roh)
	}
	return leseCSV(roh)
}

// leseCSV verträgt, was Tabellenprogramme schreiben: Semikolon, Tab oder Komma als Trenner, UTF-8 mit
// oder ohne BOM und Windows-1252, wie Excel unter Windows CSV speichert.
func leseCSV(roh []byte) ([][]string, error) {
	roh = bytes.TrimPrefix(roh, []byte("\xef\xbb\xbf"))
	if !utf8.Valid(roh) {
		roh = []byte(ausWindows1252(roh))
	}
	erste, _, _ := bytes.Cut(roh, []byte("\n"))
	leser := csv.NewReader(bytes.NewReader(roh))
	leser.Comma = trenner(erste)
	leser.FieldsPerRecord = -1
	leser.LazyQuotes = true
	zeilen, err := leser.ReadAll()
	if err != nil {
		return nil, eingabeFehler("Die CSV-Datei ließ sich nicht lesen.")
	}
	return zeilen, nil
}

// trenner nimmt das häufigste Trennzeichen der ersten Zeile; ohne eines bleibt es beim Semikolon.
func trenner(zeile []byte) rune {
	beste, anzahl := ';', bytes.Count(zeile, []byte(";"))
	for _, t := range []rune{'\t', ','} {
		if n := bytes.Count(zeile, []byte(string(t))); n > anzahl {
			beste, anzahl = t, n
		}
	}
	return beste
}

// ausWindows1252: Umlaute und ß liegen dort wie in Latin-1; nur 0x80–0x9F weichen ab (€, „ “ …).
func ausWindows1252(roh []byte) string {
	var b strings.Builder
	for _, c := range roh {
		if c >= 0x80 && c <= 0x9f {
			b.WriteRune(cp1252[c-0x80])
			continue
		}
		b.WriteRune(rune(c))
	}
	return b.String()
}

var cp1252 = [32]rune{'€', '\uFFFD', '‚', 'ƒ', '„', '…', '†', '‡', 'ˆ', '‰', 'Š', '‹', 'Œ', '\uFFFD', 'Ž', '\uFFFD',
	'\uFFFD', '‘', '’', '“', '”', '•', '–', '—', '˜', '™', 'š', '›', 'œ', '\uFFFD', 'ž', 'Ÿ'}

// raumZeilen macht aus der Tabelle Zeilen „Raum; Bereich“. Eine Kopfzeile erkennt sie an ihren
// Wörtern und nimmt deren Spalten; sonst gilt: erste Spalte Raum, zweite Bereich.
func raumZeilen(tabelle [][]string) (string, int) {
	raum, bereich, start := spalten(tabelle)
	var b strings.Builder
	anzahl := 0
	for _, z := range tabelle[start:] {
		name := zelle(z, raum)
		if name == "" {
			continue
		}
		b.WriteString(raumSauber.Replace(name))
		if ber := zelle(z, bereich); ber != "" {
			b.WriteString("; " + raumSauber.Replace(ber))
		}
		b.WriteString("\n")
		anzahl++
	}
	return b.String(), anzahl
}

func zelle(z []string, i int) string {
	if i < 0 || i >= len(z) {
		return ""
	}
	return strings.TrimSpace(z[i])
}

// spalten findet Raum- und Bereichsspalte und ob die erste Zeile eine Kopfzeile ist (start = 1).
func spalten(tabelle [][]string) (raum, bereich, start int) {
	if len(tabelle) == 0 {
		return 0, 1, 0
	}
	raum, bereich = -1, -1
	for i, z := range tabelle[0] {
		t := strings.TrimSpace(z)
		if raum < 0 && kopfRaum.MatchString(t) {
			raum = i
		} else if bereich < 0 && kopfBereich.MatchString(t) {
			bereich = i
		}
	}
	switch {
	case raum < 0 && bereich < 0:
		return 0, 1, 0
	case raum < 0 && bereich == 0:
		return 1, 0, 1
	case raum < 0:
		return 0, bereich, 1
	}
	return raum, bereich, 1
}
