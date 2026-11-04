package main

import (
	"context"
	"time"
)

// Rückmeldung an die Meldenden: Eine Lehrkraft erfährt beim nächsten Öffnen, dass ihre Meldung
// erledigt ist (Erledigt.svelte). Wer sieht, dass Melden etwas bewirkt, meldet wieder.

// meldungErledigt ist eine eigene, erledigte Meldung: was, wo und wann; mehr braucht die Rückmeldung nicht.
type meldungErledigt struct {
	ID         int64     `json:"id"`
	Ort        string    `json:"ort"`
	Was        string    `json:"was"`
	ErledigtAm time.Time `json:"erledigtAm"`
}

// sichtFuerKollegium: von fremden Meldungen nur das Nötigste für „Hier schon gemeldet“, dazu die
// eigenen, die in den letzten 30 Tagen erledigt wurden.
func (a *app) sichtFuerKollegium(ctx context.Context, b benutzer, offen []meldung) (map[string]any, error) {
	kurz := make([]meldungKurz, len(offen))
	for i, m := range offen {
		kurz[i] = meldungKurz{ID: m.ID, RaumID: m.RaumID, Kategorie: m.Kategorie, Was: m.Was,
			Status: m.Status, GemeldetAm: m.GemeldetAm}
	}
	eigene, err := a.sp.abfrage(ctx, `WHERE m.gemeldet_von = $1 AND m.status = 'erledigt' AND m.erledigt_am >= $2
		ORDER BY m.erledigt_am DESC, m.id DESC LIMIT 5`, b.Email, a.jetzt().AddDate(0, 0, -30))
	if err != nil {
		return nil, err
	}
	fertig := make([]meldungErledigt, 0, len(eigene))
	for _, m := range eigene {
		if m.ErledigtAm != nil {
			fertig = append(fertig, meldungErledigt{ID: m.ID, Ort: m.Ort, Was: m.Was, ErledigtAm: *m.ErledigtAm})
		}
	}
	return map[string]any{"offen": kurz, "erledigt": []meldungKurz{}, "meineErledigt": fertig}, nil
}
