package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

// Rückmeldung an die Meldenden: Wer einen Schaden meldet, bekommt eine Mail, sobald die Meldung
// erledigt ist. Verschickt der Server keine Mail, erfährt es die Lehrkraft beim nächsten Öffnen
// (Erledigt.svelte). Wer sieht, dass Melden etwas bewirkt, meldet wieder.

// rueckmeldungWartezeit: So lange lässt sich ein versehentliches „Erledigt“ mit „Rückgängig“
// zurücknehmen, ohne dass eine Mail hinausgeht.
const rueckmeldungWartezeit = 10 * time.Minute

// meldungErledigt ist eine eigene, erledigte Meldung: was, wo und wann; mehr braucht die Rückmeldung nicht.
type meldungErledigt struct {
	ID         int64     `json:"id"`
	Ort        string    `json:"ort"`
	Was        string    `json:"was"`
	ErledigtAm time.Time `json:"erledigtAm"`
}

// sichtFuerKollegium: von fremden Meldungen nur das Nötigste für „Hier schon gemeldet“, dazu die
// eigenen, die in den letzten 30 Tagen erledigt wurden. Mit Mail bleibt diese Liste leer, sonst käme
// der Bescheid doppelt.
func (a *app) sichtFuerKollegium(ctx context.Context, b benutzer, offen []meldung) (map[string]any, error) {
	kurz := make([]meldungKurz, len(offen))
	for i, m := range offen {
		kurz[i] = meldungKurz{ID: m.ID, RaumID: m.RaumID, Kategorie: m.Kategorie, Was: m.Was,
			Status: m.Status, GemeldetAm: m.GemeldetAm}
	}
	fertig := []meldungErledigt{}
	if !a.mailAktiv(ctx) {
		eigene, err := a.sp.abfrage(ctx, `WHERE m.gemeldet_von = $1 AND m.status = 'erledigt' AND m.erledigt_am >= $2
			ORDER BY m.erledigt_am DESC, m.id DESC LIMIT 5`, b.Email, a.jetzt().AddDate(0, 0, -30))
		if err != nil {
			return nil, err
		}
		for _, m := range eigene {
			if m.ErledigtAm != nil {
				fertig = append(fertig, meldungErledigt{ID: m.ID, Ort: m.Ort, Was: m.Was, ErledigtAm: *m.ErledigtAm})
			}
		}
	}
	return map[string]any{"offen": kurz, "erledigt": []meldungKurz{}, "meineErledigt": fertig}, nil
}

// versendeRueckmeldungen schreibt allen, deren Meldung seit der Wartezeit erledigt ist. Scheitert der
// Versand, versucht es der nächste Takt, einen Tag lang wie bei den übrigen Mails.
func (a *app) versendeRueckmeldungen(ctx context.Context) {
	jetzt := a.jetzt()
	faellig, err := a.sp.faelligeRueckmeldungen(ctx, jetzt.Add(-nachholFrist), jetzt.Add(-rueckmeldungWartezeit))
	if err != nil {
		slog.Error("Mail: Rückmeldungen nicht gelesen", "fehler", err)
		return
	}
	for _, m := range faellig {
		if ctx.Err() != nil {
			return
		}
		if err := a.schreibeMeldendem(ctx, m); err != nil {
			slog.Error("Rückmeldung nicht verschickt, neuer Versuch in fünf Minuten", "id", m.ID, "fehler", fuersLog(err))
		}
	}
}

// schreibeMeldendem: was und wo, ohne den internen Vermerk und ohne Namen. Unsichtbar für die
// Übersicht, darum ohne verteiler.melde().
func (a *app) schreibeMeldendem(ctx context.Context, m meldung) error {
	d := mailDaten{Rueckmeldung: true, Schule: a.konf.Schulname, LinkApp: a.konf.OeffentlicheAdresse + "/",
		Betreff: fmt.Sprintf("Erledigt: %s – %s", m.Was, m.Ort), Ueberschrift: "Erledigt – danke fürs Melden!",
		Meldung: mailMeldung{Ort: m.Ort, Bereich: m.Bereich, Was: m.Was, Notiz: m.Notiz,
			Wann: "gemeldet " + a.wannText(m.GemeldetAm)}}
	if err := a.sendeMail(ctx, m.GemeldetVon, d); err != nil {
		return err
	}
	return a.sp.vermerkeRueckmeldung(ctx, m.ID, a.jetzt())
}
