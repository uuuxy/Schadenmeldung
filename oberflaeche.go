package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strconv"
	"strings"
)

// oberflaeche liefert die gebaute Svelte-App. Pfade ohne Datei (/r/…, /uebersicht) bekommen die
// index.html, die ihren Weg selbst liest; die gehashten Dateien unter /assets/ dürfen ewig im Cache liegen.
func (a *app) oberflaeche() http.Handler {
	dateien := http.FileServerFS(a.statisch)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { a.liefereOberflaeche(w, r, dateien) })
}

func (a *app) liefereOberflaeche(w http.ResponseWriter, r *http.Request, dateien http.Handler) {
	pfad := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if pfad != "" && pfad != "index.html" {
		if a.istDatei(pfad) {
			w.Header().Set("Cache-Control", cacheFuer(pfad))
			dateien.ServeHTTP(w, r)
			return
		}
		if path.Ext(pfad) != "" {
			http.NotFound(w, r)
			return
		}
	}
	a.liefereIndex(w)
}

func (a *app) istDatei(pfad string) bool {
	info, err := fs.Stat(a.statisch, pfad)
	return err == nil && !info.IsDir()
}

// cacheFuer: Die gehashten Dateien unter /assets/ dürfen ewig im Cache liegen, alles andere prüft der
// Browser bei jedem Laden, damit eine neue Fassung sofort ankommt.
func cacheFuer(pfad string) string {
	if strings.HasPrefix(pfad, "assets/") {
		return "public, max-age=31536000, immutable"
	}
	return "no-cache"
}

func (a *app) liefereIndex(w http.ResponseWriter) {
	index, err := fs.ReadFile(a.statisch, "index.html")
	if err != nil {
		http.Error(w, "Die Oberfläche ist nicht gebaut: im Ordner frontend »npm run build« ausführen.",
			http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	_, _ = w.Write(index)
}

func schreibeJSON(w http.ResponseWriter, status int, wert any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(wert); err != nil {
		slog.Warn("Antwort nicht geschrieben", "fehler", err)
	}
}

func fehler(w http.ResponseWriter, status int, text string) {
	schreibeJSON(w, status, map[string]string{"fehler": text})
}

func intern(w http.ResponseWriter, err error) {
	slog.Error("Interner Fehler", "fehler", err)
	fehler(w, http.StatusInternalServerError, "Da ist etwas schiefgegangen. Bitte noch einmal versuchen.")
}

func nichtGefundenOderIntern(w http.ResponseWriter, err error) {
	if errors.Is(err, errNichtGefunden) {
		fehler(w, http.StatusNotFound, "Nicht gefunden – vielleicht wurde es gerade gelöscht.")
		return
	}
	intern(w, err)
}

func leseJSON(w http.ResponseWriter, r *http.Request, ziel any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 256<<10)
	if err := json.NewDecoder(r.Body).Decode(ziel); err != nil {
		fehler(w, http.StatusBadRequest, "Die Anfrage war nicht lesbar.")
		return false
	}
	return true
}

func idAusPfad(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		fehler(w, http.StatusBadRequest, "Ungültige Nummer.")
		return 0, false
	}
	return id, true
}
