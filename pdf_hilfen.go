package main

import (
	"fmt"
	"strings"
	"time"
)

// gofpdf schreibt die Standardschriften in cp1252; was diese Kodierung nicht kennt, druckt sein
// Übersetzer als Punkt. Buchstaben aus den Sprachen einer hessischen Schule werden vorher ersetzt,
// wie in der Bibliothek (pkg/pdfzeichen): „Ayşe“ wird zu „Ayse“ statt „Ay.e“.
var pdfErsatz = strings.NewReplacer(
	"ş", "s", "Ş", "S", "ğ", "g", "Ğ", "G", "ı", "i", "İ", "I",
	"ć", "c", "Ć", "C", "č", "c", "Č", "C", "ł", "l", "Ł", "L",
	"ń", "n", "Ń", "N", "ś", "s", "Ś", "S", "ż", "z", "Ż", "Z", "ź", "z", "Ź", "Z",
	"ą", "a", "Ą", "A", "ę", "e", "Ę", "E", "ő", "o", "Ő", "O", "ű", "u", "Ű", "U",
	"ș", "s", "Ș", "S", "ț", "t", "Ț", "T", "ă", "a", "Ă", "A",
	"ř", "r", "Ř", "R", "ě", "e", "Ě", "E", "ů", "u", "Ů", "U",
	"š", "s", "Š", "S", "ž", "z", "Ž", "Z", "đ", "d", "Đ", "D",
)

func pdfUebersetzer(cp1252 func(string) string) func(string) string {
	return func(text string) string { return cp1252(pdfErsatz.Replace(text)) }
}

var (
	wochentage = [...]string{"Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag"}
	monate     = [...]string{"Januar", "Februar", "März", "April", "Mai", "Juni", "Juli", "August", "September",
		"Oktober", "November", "Dezember"}
)

// langesDatum: „Montag, 2. November 2026, 14:30 Uhr“ – Go kennt keine deutschen Namen.
func langesDatum(t time.Time) string {
	return fmt.Sprintf("%s, %d. %s %d, %s Uhr", wochentage[t.Weekday()], t.Day(), monate[t.Month()-1], t.Year(), t.Format("15:04"))
}

// kurzesDatum: „2.11.2026, 09:41“.
func kurzesDatum(t time.Time) string {
	return t.Format("2.1.2006, 15:04")
}
