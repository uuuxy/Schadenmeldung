package main

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

// letzterTermin ist die jüngste Sammelzeit von heute, die schon vorbei ist. Am Wochenende gibt es
// keine; was dann kommt, steht Montag früh in der ersten Mail.
func letzterTermin(jetzt time.Time, zeiten []uhrzeit) (time.Time, bool) {
	if tag := jetzt.Weekday(); tag == time.Saturday || tag == time.Sunday {
		return time.Time{}, false
	}
	var termin time.Time
	for _, z := range zeiten {
		t := time.Date(jetzt.Year(), jetzt.Month(), jetzt.Day(), z.Stunde, z.Minute, 0, 0, jetzt.Location())
		if !t.After(jetzt) && t.After(termin) {
			termin = t
		}
	}
	return termin, !termin.IsZero()
}

// pruefeSammelmail läuft jede Minute. Ein verpasster Termin (Server war aus) wird nachgeholt, aber
// nur einmal; beim allerersten Start zählt erst der nächste Termin.
func (a *app) pruefeSammelmail(ctx context.Context) {
	if !a.konf.mailAktiv() || len(a.konf.SammelZeiten) == 0 {
		return
	}
	termin, faellig := a.faelligerTermin(ctx)
	if !faellig {
		return
	}
	liste, err := a.empfaengerFuer(ctx, false)
	if err != nil {
		slog.Error("Sammelmail: Empfänger nicht gelesen", "fehler", err)
		return
	}
	for _, e := range liste {
		a.sammelmailAn(ctx, e, termin)
	}
}

// faelligerTermin sagt, ob seit der letzten Sammelmail ein Termin verstrichen ist, und merkt ihn sich
// vor dem Versand: Lieber fällt eine Sammelmail einmal aus, als dass sie doppelt kommt.
func (a *app) faelligerTermin(ctx context.Context) (time.Time, bool) {
	jetzt := a.jetzt().In(a.konf.Zeitzone)
	roh, err := a.sp.wert(ctx, "sammelmail_zuletzt")
	if errors.Is(err, errNichtGefunden) {
		// Beim ersten Start gilt jetzt als letzter Versand, sonst käme sofort alles Alte.
		a.merkeSammelmail(ctx, jetzt)
		return time.Time{}, false
	}
	if err != nil {
		slog.Error("Sammelmail: Stand nicht gelesen", "fehler", err)
		return time.Time{}, false
	}
	zuletzt, _ := time.Parse(time.RFC3339, roh)
	termin, ok := letzterTermin(jetzt, a.konf.SammelZeiten)
	if !ok || !zuletzt.Before(termin) {
		return time.Time{}, false
	}
	return termin, a.merkeSammelmail(ctx, termin)
}

func (a *app) merkeSammelmail(ctx context.Context, t time.Time) bool {
	if err := a.sp.setzeWert(ctx, "sammelmail_zuletzt", t.Format(time.RFC3339)); err != nil {
		slog.Error("Sammelmail: Stand nicht gespeichert", "fehler", err)
		return false
	}
	return true
}

// sammelmailAn schickt einer Person, was bis zum Termin neu ist. Die Adresse bleibt aus dem Log.
func (a *app) sammelmailAn(ctx context.Context, e empfaenger, termin time.Time) {
	neu, err := a.sp.ungesendeteFuer(ctx, e.Email, termin)
	if err != nil {
		slog.Error("Sammelmail: Meldungen nicht gelesen", "fehler", err)
		return
	}
	if len(neu) == 0 {
		return
	}
	if err := a.verschicke(ctx, e, neu, true); err != nil {
		slog.Error("Sammelmail nicht verschickt", "fehler", err)
	}
}

func (a *app) sammelmailTakt(ctx context.Context) {
	takt := time.NewTicker(time.Minute)
	defer takt.Stop()
	for {
		a.pruefeSammelmail(ctx)
		select {
		case <-ctx.Done():
			return
		case <-takt.C:
		}
	}
}
