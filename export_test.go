package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"
)

// holeDatei lädt eine Exportdatei wie der Browser: ein GET mit dem Cookie der Sitzung.
func (n *testNutzer) holeDatei(pfad string) (int, http.Header, []byte) {
	n.u.t.Helper()
	resp, err := n.client.Get(n.u.srv.URL + pfad)
	if err != nil {
		n.u.t.Fatal(err)
	}
	defer resp.Body.Close()
	roh, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, resp.Header, roh
}

func TestExportExcelAlleMeldungenOhneNamen(t *testing.T) {
	u := neueTestUmgebung(t)
	lehrer, hm := u.nutzer("lehrer"), u.nutzer("hm")
	lehrer.melde(map[string]string{"ort": "Schulhof", "kategorie": "licht", "was": "Lampe defekt",
		"notiz": "Zeile 1\nZeile 2 & <mehr>", "dringend": "1"}, nil)
	_, fertig := lehrer.melde(map[string]string{"ort": "Aula", "kategorie": "wasser", "was": "Hahn tropft"}, nil)
	hm.json("POST", fmt.Sprintf("/api/meldungen/%v/status", fertig["id"]), map[string]string{"status": "erledigt", "antwort": "Dichtung neu"})

	status, kopf, roh := u.nutzer("admin").holeDatei("/api/export.xlsx")
	if status != http.StatusOK || !strings.Contains(kopf.Get("Content-Disposition"), `filename="Schadenmeldungen-2026-11-02.xlsx"`) {
		t.Fatalf("Excel: Status %d, %v", status, kopf)
	}
	zeilen, err := leseXLSX(roh)
	if err != nil || len(zeilen) != 3 {
		t.Fatalf("Excel nicht lesbar oder falsche Zeilenzahl: %v, %v", err, zeilen)
	}
	if k := strings.Join(zeilen[0], "|"); k != "Nr.|Gemeldet am|Ort|Bereich|Art|Was|Notiz|Dringend|Stand|Erledigt am|Vermerk|Foto" {
		t.Errorf("Kopfzeile: %s", k)
	}
	// 2. November 2026, 9 Uhr in Berlin ist für Excel der Tag 46328 und drei Achtel.
	offen := zeilen[1]
	if offen[1] != "46328.375" || offen[4] != "Licht & Strom" || offen[6] != "Zeile 1\nZeile 2 & <mehr>" || offen[7] != "ja" || offen[8] != "Offen" {
		t.Errorf("offene Meldung: %q", offen)
	}
	if erledigt := zeilen[2]; erledigt[8] != "Erledigt" || erledigt[9] == "" || erledigt[10] != "Dichtung neu" {
		t.Errorf("erledigte Meldung: %q", erledigt)
	}
	if alles := fmt.Sprint(zeilen); strings.Contains(alles, "lehrer") {
		t.Errorf("Der Name der meldenden Person steht in der Tabelle: %s", alles)
	}
}

func TestExportPDFNachListeMitFoto(t *testing.T) {
	u := neueTestUmgebung(t)
	lehrer, hm := u.nutzer("lehrer"), u.nutzer("hm")
	lehrer.melde(map[string]string{"ort": "Schulhof", "kategorie": "licht", "was": "Lampe defekt", "notiz": "Ayşe sagt: flackert"}, fotoMitExif(t, 800, 600))
	_, fertig := lehrer.melde(map[string]string{"ort": "Aula", "kategorie": "wasser", "was": "Hahn tropft"}, nil)
	hm.json("POST", fmt.Sprintf("/api/meldungen/%v/status", fertig["id"]), map[string]string{"status": "erledigt"})

	status, kopf, roh := hm.holeDatei("/api/export.pdf")
	if status != http.StatusOK || !bytes.HasPrefix(roh, []byte("%PDF-")) || !strings.Contains(kopf.Get("Content-Disposition"), "Schadenmeldungen-offen-2026-11-02.pdf") {
		t.Fatalf("PDF offen: Status %d, %v", status, kopf)
	}
	text := pdfInhalt(t, roh)
	for _, soll := range []string{"(Offene Schadenmeldungen)", "(Lampe defekt)", "(Ayse sagt: flackert)", "(Weitere Orte)"} {
		if !strings.Contains(text, soll) {
			t.Errorf("PDF offen ohne %s", soll)
		}
	}
	if strings.Contains(text, "Hahn tropft") || strings.Contains(text, "lehrer") || !bytes.Contains(roh, []byte("/Subtype /Image")) {
		t.Errorf("PDF offen: erledigte Meldung, Name oder fehlendes Foto")
	}
	_, kopf, roh = hm.holeDatei("/api/export.pdf?liste=erledigt&bereich=Neubau")
	if !strings.Contains(kopf.Get("Content-Disposition"), "Schadenmeldungen-erledigt") || !strings.Contains(pdfInhalt(t, roh), "(Keine Meldungen.)") {
		t.Errorf("PDF erledigt, Bereich ohne Meldungen: %v", kopf)
	}
	if status, _, _ := lehrer.holeDatei("/api/export.pdf"); status != http.StatusForbidden {
		t.Errorf("Kollegium lädt das PDF: Status %d", status)
	}
}

// pdfInhalt entpackt die Seiteninhalte, damit der Test den gesetzten Text lesen kann.
func pdfInhalt(t *testing.T, roh []byte) string {
	t.Helper()
	var aus strings.Builder
	for _, s := range regexp.MustCompile(`(?s)stream\r?\n(.*?)\r?\nendstream`).FindAllSubmatch(roh, -1) {
		if z, err := zlib.NewReader(bytes.NewReader(s[1])); err == nil {
			entpackt, _ := io.ReadAll(z)
			aus.Write(entpackt)
		}
	}
	return aus.String()
}

func TestNatuerlicheReihenfolge(t *testing.T) {
	for _, f := range [][2]string{{"9", "10"}, {"Bio 2", "Bio 10"}, {"101", "Aula"}, {"aula", "Bio 1"}, {"204", "204a"}} {
		if !natuerlichVor(f[0], f[1]) || natuerlichVor(f[1], f[0]) {
			t.Errorf("%q gehört vor %q", f[0], f[1])
		}
	}
}
