package main

import (
	"context"
	"log/slog"
	"time"
)

// raeumeAuf setzt die Löschfristen um: Fotos verschwinden FOTO_TAGE nach der Erledigung, die Meldung
// selbst nach AUFBEWAHRUNG_TAGE. Abgelaufene Sitzungen gehen gleich mit.
func (a *app) raeumeAuf(ctx context.Context) {
	jetzt := a.jetzt()
	meldungen, err := a.sp.loescheAlteMeldungen(ctx, jetzt.AddDate(0, 0, -a.konf.AufbewahrungTage))
	if err != nil {
		slog.Error("Aufräumen: Meldungen nicht gelöscht", "fehler", err)
	}
	if meldungen > 0 {
		a.verteiler.melde()
	}
	fotos, err := a.sp.loescheAlteFotos(ctx, jetzt.AddDate(0, 0, -a.konf.FotoTage), jetzt)
	if err != nil {
		slog.Error("Aufräumen: Fotos nicht gelöscht", "fehler", err)
	}
	if err := a.sp.loescheAbgelaufeneSitzungen(ctx, jetzt); err != nil {
		slog.Error("Aufräumen: Sitzungen nicht gelöscht", "fehler", err)
	}
	if err := a.sp.loescheAbgelaufeneMailLinks(ctx, jetzt); err != nil {
		slog.Error("Aufräumen: Mail-Links nicht gelöscht", "fehler", err)
	}
	if meldungen+fotos > 0 {
		slog.Info("Aufgeräumt", "meldungen", meldungen, "fotos", fotos)
	}
}

func (a *app) raeumeRegelmaessigAuf(ctx context.Context) {
	a.raeumeAuf(ctx)
	takt := time.NewTicker(time.Hour)
	defer takt.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-takt.C:
			a.raeumeAuf(ctx)
		}
	}
}
