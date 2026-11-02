package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/uudashr/gocognit"
)

// Höchstens 200 Zeilen je Datei und Komplexität 15 je Funktion, wie in der Bibliothek: Kleine
// Dateien sagen, wo etwas steht, und SonarQube meldet ab 16 (go:S3776, kognitive Komplexität).
const (
	zeilenGrenze       = 200
	komplexitaetGrenze = 15
)

// bestand ist eine Arbeitsliste, kein Freibrief: Eine Datei darf hier nur schrumpfen, und wer sie
// unter die Grenze bringt, trägt sie aus.
var bestand = map[string]int{}

// quellOrdner und quellEndungen: was wir selbst schreiben. symbole.js erzeugt npm run symbole.
var (
	quellOrdner   = []string{".", "migrations", "frontend/src", "frontend/public", "frontend/e2e", "frontend/scripts"}
	quellEndungen = map[string]bool{".go": true, ".sql": true, ".svelte": true, ".js": true, ".mjs": true, ".css": true, ".sh": true}
	erzeugt       = map[string]bool{"frontend/src/lib/symbole.js": true}
)

func TestKeineDateiUeber200Zeilen(t *testing.T) {
	for _, pfad := range quelldateien(t) {
		roh, err := os.ReadFile(pfad)
		if err != nil {
			t.Fatal(err)
		}
		zeilen := strings.Count(strings.TrimRight(string(roh), "\n"), "\n") + 1
		grenze, imBestand := bestand[pfad]
		switch {
		case !imBestand && zeilen > zeilenGrenze:
			t.Errorf("%s hat %d Zeilen, erlaubt sind %d: nach Aufgaben aufteilen", pfad, zeilen, zeilenGrenze)
		case imBestand && zeilen > grenze:
			t.Errorf("%s ist auf %d Zeilen gewachsen (Bestand %d): die Ausnahme ist kein Freibrief", pfad, zeilen, grenze)
		case imBestand && zeilen <= zeilenGrenze:
			t.Errorf("%s hat nur noch %d Zeilen: aus dem Bestand austragen", pfad, zeilen)
		}
	}
}

func TestKeineFunktionUeberKomplexitaet15(t *testing.T) {
	fset := token.NewFileSet()
	dateien, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, datei := range dateien {
		if strings.HasSuffix(datei, "_test.go") {
			continue
		}
		baum, err := parser.ParseFile(fset, datei, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range baum.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if k := gocognit.Complexity(fn); k > komplexitaetGrenze {
				t.Errorf("%s: %s hat Komplexität %d, erlaubt sind %d: zusammenhängende Prüfungen in einen Helfer ziehen",
					fset.Position(fn.Pos()), fn.Name.Name, k, komplexitaetGrenze)
			}
		}
	}
}

func quelldateien(t *testing.T) []string {
	t.Helper()
	var aus []string
	for _, ordner := range quellOrdner {
		err := filepath.WalkDir(ordner, func(pfad string, e fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if e.IsDir() && pfad != ordner && (ordner == "." || e.Name() == "node_modules" || e.Name() == "dist") {
				return filepath.SkipDir
			}
			if !e.IsDir() && quellEndungen[filepath.Ext(pfad)] && !erzeugt[filepath.ToSlash(pfad)] {
				aus = append(aus, filepath.ToSlash(pfad))
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return aus
}
