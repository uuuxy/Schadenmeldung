package main

import (
	"context"
	"log/slog"
	"time"
)

// raeumeAuf setzt die Löschfristen um: Fotos verschwinden FOTO_TAGE nach der Erledigung, die Meldung
// selbst nach AUFBEWAHRUNG_TAGE. Abgelaufene Sitzungen und Push-Abos gehen gleich mit.
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
	if err := a.raeumePushAbosAuf(ctx, jetzt); err != nil {
		slog.Error("Aufräumen: Push-Abos nicht gelöscht", "fehler", err)
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

// loescheAlteFotos entfernt die Fotos erledigter Meldungen; der Verweis in der Meldung fällt über den
// Fremdschlüssel (ON DELETE SET NULL) von selbst weg. Fotos ohne Meldung gehen gleich mit.
func (s *speicher) loescheAlteFotos(ctx context.Context, grenze, jetzt time.Time) (int64, error) {
	tag, err := s.db.Exec(ctx, `DELETE FROM fotos f WHERE f.name IN (
			SELECT foto FROM meldungen WHERE status = 'erledigt' AND erledigt_am < $1 AND foto IS NOT NULL)
		OR (NOT EXISTS (SELECT 1 FROM meldungen m WHERE m.foto = f.name) AND f.angelegt < $2)`,
		grenze, jetzt.Add(-24*time.Hour))
	return tag.RowsAffected(), err
}

// loescheAlteMeldungen entfernt erledigte Meldungen nach der Aufbewahrungsfrist samt Foto.
func (s *speicher) loescheAlteMeldungen(ctx context.Context, grenze time.Time) (int64, error) {
	var anzahl int64
	err := s.db.QueryRow(ctx, `WITH weg AS (
			DELETE FROM meldungen WHERE status = 'erledigt' AND erledigt_am < $1 RETURNING foto),
		bilder AS (DELETE FROM fotos WHERE name IN (SELECT foto FROM weg WHERE foto IS NOT NULL))
		SELECT count(*) FROM weg`, grenze).Scan(&anzahl)
	return anzahl, err
}
