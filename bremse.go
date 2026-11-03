package main

import (
	"sync"
	"time"
)

// bremse sperrt einen Schlüssel (Konto, Adresse oder meldende Person) für ein Zeitfenster, wenn
// darin zu viele Versuche auflaufen. Sie bremst das Durchprobieren von Passwörtern und Fluten von
// Meldungen.
type bremse struct {
	mu        sync.Mutex
	eintraege map[string]*versuche
	grenze    int
	fenster   time.Duration
}

type versuche struct {
	anzahl      int
	seit        time.Time
	gesperrtBis time.Time
}

func neueBremse(grenze int, fenster time.Duration) *bremse {
	return &bremse{eintraege: map[string]*versuche{}, grenze: grenze, fenster: fenster}
}

// versuch prüft und verbucht in einem Schritt und liefert die verbleibende Sperrzeit, sonst 0. Ein
// Versuch zählt sofort und wird zurückgenommen, wenn er gelingt: Würde erst nach der Antwort des
// Mailservers gezählt, kämen viele gleichzeitige Versuche alle an der Bremse vorbei.
func (b *bremse) versuch(schluessel string, jetzt time.Time) time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()
	if v := b.eintraege[schluessel]; v != nil && jetzt.Before(v.gesperrtBis) {
		return v.gesperrtBis.Sub(jetzt)
	}
	if len(b.eintraege) > 10000 {
		b.vergiss(jetzt)
	}
	v := b.eintraege[schluessel]
	if v == nil || jetzt.Sub(v.seit) > b.fenster {
		v = &versuche{seit: jetzt}
		b.eintraege[schluessel] = v
	}
	v.anzahl++
	if v.anzahl >= b.grenze {
		v.gesperrtBis = jetzt.Add(b.fenster)
		v.anzahl = 0
		v.seit = jetzt
	}
	return 0
}

// zuruecknehmen streicht einen mit versuch verbuchten Versuch wieder.
func (b *bremse) zuruecknehmen(schluessel string, jetzt time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	v := b.eintraege[schluessel]
	switch {
	case v == nil:
	case v.anzahl > 0:
		v.anzahl--
	case jetzt.Before(v.gesperrtBis):
		// Dieser Versuch hatte die Sperre ausgelöst; sie fällt mit ihm weg.
		v.gesperrtBis = time.Time{}
		v.anzahl = b.grenze - 1
	}
}

func (b *bremse) erfolg(schluessel string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.eintraege, schluessel)
}

func (b *bremse) vergiss(jetzt time.Time) {
	for k, v := range b.eintraege {
		if jetzt.Sub(v.seit) > b.fenster && jetzt.After(v.gesperrtBis) {
			delete(b.eintraege, k)
		}
	}
}
