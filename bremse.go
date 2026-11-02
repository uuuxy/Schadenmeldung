package main

import (
	"sync"
	"time"
)

// bremse sperrt einen Schlüssel (Konto oder Adresse) für ein Zeitfenster, wenn darin zu viele
// Fehlversuche auflaufen. Sie bremst das Durchprobieren von Passwörtern über diese App.
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

// gesperrt liefert die verbleibende Sperrzeit, sonst 0.
func (b *bremse) gesperrt(schluessel string, jetzt time.Time) time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()
	if v := b.eintraege[schluessel]; v != nil && jetzt.Before(v.gesperrtBis) {
		return v.gesperrtBis.Sub(jetzt)
	}
	return 0
}

func (b *bremse) fehlversuch(schluessel string, jetzt time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
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
