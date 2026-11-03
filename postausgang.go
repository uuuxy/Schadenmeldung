package main

import (
	"context"
	"log/slog"
	"time"
)

// Postausgang: Eine neue Meldung merkt für jeden Empfänger eine Zeile in mail_versand vor, und ein
// einziger Takt verschickt sie gleich darauf. Was scheitert, etwa weil der Mailserver nicht antwortet,
// versucht er alle fünf Minuten erneut, einen Tag lang. Weil nur dieser Takt verschickt, kommt keine
// Mail doppelt.
const (
	nachholTakt  = 5 * time.Minute
	nachholFrist = 24 * time.Hour
)

// mailVormerken läuft nach dem Speichern einer Meldung. Wer selbst meldet, bekommt dazu keine Mail.
func (a *app) mailVormerken(ctx context.Context, m meldung) {
	if !a.konf.mailAktiv() {
		return
	}
	wer, err := a.mailEmpfaenger(ctx)
	if err != nil {
		slog.Error("Mail: Empfänger nicht gelesen", "fehler", err)
		return
	}
	for email := range wer {
		if email == m.GemeldetVon {
			continue
		}
		if err := a.sp.merkeMailVor(ctx, m.ID, email); err != nil {
			slog.Error("Mail nicht vorgemerkt", "id", m.ID, "fehler", err)
		}
	}
	select {
	case a.postWecker <- struct{}{}:
	default: // Ein Wecken steht schon aus; diese Runde nimmt die neue Meldung mit.
	}
}

// postausgangTakt verschickt, was aussteht: gleich nach jeder neuen Meldung, sonst alle fünf Minuten.
func (a *app) postausgangTakt(ctx context.Context) {
	takt := time.NewTicker(nachholTakt)
	defer takt.Stop()
	for {
		a.versendeAusstehende(ctx)
		select {
		case <-ctx.Done():
			return
		case <-takt.C:
		case <-a.postWecker:
		}
	}
}

func (a *app) versendeAusstehende(ctx context.Context) {
	if !a.konf.mailAktiv() {
		return
	}
	meldungen, err := a.sp.meldungenMitAusstehenderMail(ctx, a.jetzt().Add(-nachholFrist))
	if err != nil {
		slog.Error("Mail: Postausgang nicht gelesen", "fehler", err)
		return
	}
	if len(meldungen) == 0 {
		return
	}
	wer, err := a.mailEmpfaenger(ctx)
	if err != nil {
		slog.Error("Mail: Empfänger nicht gelesen", "fehler", err)
		return
	}
	for _, m := range meldungen {
		a.versendeMeldung(ctx, m, wer)
	}
}

// versendeMeldung schickt eine Meldung an alle, an die sie noch aussteht. Es zählt, wer jetzt Mail
// will: Wer sie inzwischen abgeschaltet hat oder nicht mehr im Backend ist, bekommt nichts mehr.
func (a *app) versendeMeldung(ctx context.Context, m meldung, wer map[string]empfaenger) {
	offen, err := a.sp.ausstehendeEmpfaenger(ctx, m.ID)
	if err != nil {
		slog.Error("Mail: Postausgang nicht gelesen", "id", m.ID, "fehler", err)
		return
	}
	for _, email := range offen {
		e, ok := wer[email]
		if !ok || ctx.Err() != nil {
			continue
		}
		if err := a.verschicke(ctx, e, m); err != nil {
			slog.Error("Mail nicht verschickt, neuer Versuch in fünf Minuten", "id", m.ID, "fehler", fuersLog(err))
		}
	}
}
