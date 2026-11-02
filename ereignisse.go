package main

import (
	"io"
	"net/http"
	"sync"
	"time"
)

const (
	// pulsTakt hält den Strom offen: Ohne Lebenszeichen trennen Proxys und Mobilfunknetze stille
	// Verbindungen, und ein toter Strom fiele erst beim nächsten Schreiben auf.
	pulsTakt      = 25 * time.Second
	schreibFrist  = 10 * time.Second
	ereignisZeile = "event: meldungen\ndata: neu\n\n"
)

// verteiler sagt jedem offenen Ereignisstrom Bescheid, wenn sich Meldungen ändern. Er trägt keine
// Inhalte: Jede Oberfläche holt danach selbst, was ihre Rolle sehen darf.
type verteiler struct {
	mu     sync.Mutex
	hoerer map[chan struct{}]struct{}
	ende   chan struct{}
	zu     sync.Once
}

func neuerVerteiler() *verteiler {
	return &verteiler{hoerer: map[chan struct{}]struct{}{}, ende: make(chan struct{})}
}

// melde blockiert nie: Wartet bei einem Hörer schon ein Bescheid, genügt der eine, denn er holt
// ohnehin den ganzen Stand.
func (v *verteiler) melde() {
	v.mu.Lock()
	defer v.mu.Unlock()
	for h := range v.hoerer {
		select {
		case h <- struct{}{}:
		default:
		}
	}
}

func (v *verteiler) abonniere() (chan struct{}, func()) {
	h := make(chan struct{}, 1)
	v.mu.Lock()
	v.hoerer[h] = struct{}{}
	v.mu.Unlock()
	return h, func() {
		v.mu.Lock()
		delete(v.hoerer, h)
		v.mu.Unlock()
	}
}

// schliesse beendet alle Ströme. Shutdown wartet auf laufende Anfragen, und ein Strom endet sonst nie.
func (v *verteiler) schliesse() { v.zu.Do(func() { close(v.ende) }) }

// ereignisse hält einen Server-Sent-Events-Strom offen, damit die Übersicht neue Meldungen und
// jeden neuen Stand sofort zeigt statt erst beim nächsten Nachfragen.
func (a *app) ereignisse(w http.ResponseWriter, r *http.Request, _ benutzer) {
	rc := http.NewResponseController(w)
	// Die Fristen des Servers gelten für gewöhnliche Anfragen; dieser Strom bleibt offen. Jedes
	// Schreiben bekommt unten seine eigene Frist, damit ein hängendes Gerät nichts blockiert.
	_ = rc.SetReadDeadline(time.Time{})
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Accel-Buffering", "no")

	hoerer, abmelden := a.verteiler.abonniere()
	defer abmelden()
	puls := time.NewTicker(pulsTakt)
	defer puls.Stop()

	schreibe := func(text string) bool {
		_ = rc.SetWriteDeadline(time.Now().Add(schreibFrist))
		if _, err := io.WriteString(w, text); err != nil {
			return false
		}
		return rc.Flush() == nil
	}
	// retry: Nach einer Lücke versucht es der Browser nach 5 Sekunden wieder.
	if !schreibe("retry: 5000\n: verbunden\n\n") {
		return
	}
	for {
		select {
		case <-r.Context().Done():
			return
		case <-a.verteiler.ende:
			return
		case <-hoerer:
			if !schreibe(ereignisZeile) {
				return
			}
		case <-puls.C:
			if !schreibe(": puls\n\n") {
				return
			}
		}
	}
}
