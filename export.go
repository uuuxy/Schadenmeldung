package main

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Export für alle mit Übersicht. Beide Dateien kommen ohne die Namen der Meldenden aus: Sie
// verlassen die App, etwa zum Schulträger, und für Rückfragen steht der Name in der App. Auch der
// Zeitpunkt von „In Arbeit“ fehlt, wie in der Oberfläche.

// exportExcel: alle gespeicherten Meldungen, offene und erledigte, in der Reihenfolge ihres Eingangs.
func (a *app) exportExcel(w http.ResponseWriter, r *http.Request, _ benutzer) {
	liste, err := a.sp.abfrage(r.Context(), "ORDER BY m.id")
	if err != nil {
		intern(w, err)
		return
	}
	roh, err := schreibeXLSX(meldungsTabelle(liste), a.konf.Zeitzone)
	if err != nil {
		intern(w, err)
		return
	}
	a.liefereDatei(w, roh, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "Schadensmeldungen", "xlsx", false)
}

func meldungsTabelle(liste []meldung) tabelle {
	t := tabelle{blatt: "Meldungen", spalten: []spalte{
		{"Nr.", 6}, {"Gemeldet am", 17}, {"Ort", 18}, {"Bereich", 24}, {"Art", 18}, {"Was", 26},
		{"Notiz", 40}, {"Dringend", 10}, {"Stand", 26}, {"Erledigt am", 17}, {"Vermerk", 40}, {"Foto", 7},
	}}
	for _, m := range liste {
		t.zeilen = append(t.zeilen, []any{
			m.ID, m.GemeldetAm, m.Ort, m.Bereich, artName(m.Kategorie), m.Was, m.Notiz,
			jaOderLeer(m.Dringend), statusNamen[m.Status], m.ErledigtAm, m.Antwort, jaOderLeer(m.Foto != ""),
		})
	}
	return t
}

func jaOderLeer(b bool) string {
	if b {
		return "ja"
	}
	return ""
}

// exportPDF: die Liste der Übersicht, wie sie gerade zu sehen ist (offen oder erledigt, auf Wunsch
// nur ein Bereich und eine Art), zum Drucken oder Weiterschicken. Eine unbekannte Art filtert nicht.
func (a *app) exportPDF(w http.ResponseWriter, r *http.Request, _ benutzer) {
	q := r.URL.Query()
	auswahl := pdfAuswahl{erledigt: q.Get("liste") == "erledigt", bereich: q.Get("bereich")}
	if artName(q.Get("art")) != "" {
		auswahl.art = q.Get("art")
	}
	liste, err := a.pdfListe(r.Context(), auswahl)
	if err != nil {
		intern(w, err)
		return
	}
	roh, err := a.schreibePDF(r.Context(), liste, auswahl)
	if err != nil {
		intern(w, err)
		return
	}
	name := "Schadensmeldungen-offen"
	if auswahl.erledigt {
		name = "Schadensmeldungen-erledigt"
	}
	a.liefereDatei(w, roh, "application/pdf", name, "pdf", q.Get("zeigen") == "1")
}

type pdfAuswahl struct {
	erledigt bool
	bereich  string
	art      string
}

// pdfListe ordnet nach Bereichen wie die Raumliste, darin nach Räumen (9 vor 10): So geht ein
// Rundgang mit dem Ausdruck Flur für Flur. Orte ohne Bereich stehen am Ende.
func (a *app) pdfListe(ctx context.Context, aus pdfAuswahl) ([]meldung, error) {
	var liste []meldung
	var err error
	if aus.erledigt {
		liste, err = a.sp.erledigteMeldungen(ctx, a.jetzt().AddDate(0, 0, -30))
	} else {
		liste, err = a.sp.offeneMeldungen(ctx)
	}
	if err != nil {
		return nil, err
	}
	raeume, err := a.sp.raeume(ctx)
	if err != nil {
		return nil, err
	}
	rang := map[string]int{}
	for _, r := range raeume {
		if _, da := rang[r.Bereich]; !da && r.Bereich != "" {
			rang[r.Bereich] = len(rang)
		}
	}
	gefiltert := []meldung{}
	for _, m := range liste {
		if (aus.bereich == "" || m.Bereich == aus.bereich) && (aus.art == "" || m.Kategorie == aus.art) {
			gefiltert = append(gefiltert, m)
		}
	}
	sort.SliceStable(gefiltert, func(i, j int) bool { return vorDem(gefiltert[i], gefiltert[j], rang) })
	return gefiltert, nil
}

func vorDem(x, y meldung, rang map[string]int) bool {
	if rx, ry := bereichsRang(x.Bereich, rang), bereichsRang(y.Bereich, rang); rx != ry {
		return rx < ry
	}
	if x.Ort != y.Ort {
		return natuerlichVor(x.Ort, y.Ort)
	}
	if x.Dringend != y.Dringend {
		return x.Dringend
	}
	return x.GemeldetAm.After(y.GemeldetAm)
}

// bereichsRang: bekannte Bereiche in der Reihenfolge der Raumliste, danach die übrigen, leer zuletzt.
func bereichsRang(b string, rang map[string]int) int {
	if r, da := rang[b]; da {
		return r
	}
	if b == "" {
		return 1 << 30
	}
	return 1<<30 - 1
}

// natuerlichVor ordnet Namen, wie Menschen zählen: „9“ vor „10“, „Bio 2“ vor „Bio 10“.
func natuerlichVor(x, y string) bool {
	x, y = strings.ToLower(x), strings.ToLower(y)
	for x != "" && y != "" {
		nx, restX := vorneZahl(x)
		ny, restY := vorneZahl(y)
		if nx >= 0 && ny >= 0 {
			if nx != ny {
				return nx < ny
			}
			x, y = restX, restY
			continue
		}
		rx, lx := utf8.DecodeRuneInString(x)
		ry, ly := utf8.DecodeRuneInString(y)
		if rx != ry {
			return rx < ry
		}
		x, y = x[lx:], y[ly:]
	}
	return len(x) < len(y)
}

// vorneZahl liest die Ziffern am Anfang; -1 heißt: keine (oder zu viele für eine Raumnummer).
func vorneZahl(s string) (int, string) {
	i := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i == 0 || i > 9 {
		return -1, s
	}
	n, _ := strconv.Atoi(s[:i])
	return n, s[i:]
}

// liefereDatei schickt eine Datei mit Datum im Namen, damit zwei Stände nebeneinander liegen können.
// zeigen öffnet sie im Browser (ein PDF am Rechner, gleich zum Drucken), sonst wird sie geladen.
// Nichts davon bleibt in einem Cache.
func (a *app) liefereDatei(w http.ResponseWriter, roh []byte, typ, name, endung string, zeigen bool) {
	art := "attachment"
	if zeigen {
		art = "inline"
	}
	h := w.Header()
	h.Set("Content-Type", typ)
	h.Set("Content-Disposition", art+`; filename="`+name+"-"+a.jetzt().In(a.konf.Zeitzone).Format("2006-01-02")+"."+endung+`"`)
	h.Set("Cache-Control", "no-store")
	h.Set("Content-Length", strconv.Itoa(len(roh)))
	_, _ = w.Write(roh)
}
