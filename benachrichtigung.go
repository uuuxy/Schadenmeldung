package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// Benachrichtigung: Der Hausmeister bekommt neue Meldungen per Mail, ohne gehetzt zu werden.
// Dringendes kommt sofort, alles andere gesammelt zu festen Zeiten (Mo–Fr). Jede Meldung in der
// Mail trägt zwei Knöpfe – „Gesehen“ und „Erledigt“ –, die ohne Anmeldung wirken.

const linkGueltigkeit = 14 * 24 * time.Hour

type mailEinstellung struct {
	DringendSofort bool `json:"dringendSofort"`
	Sammelmail     bool `json:"sammelmail"`
}

// Ohne eigene Wahl bekommt der Hausmeister beides, Leitung und Verwaltung nichts: Sie schalten es
// sich bei Bedarf selbst an.
func vorgabeFuer(rolle string) mailEinstellung {
	an := rolle == rolleHausmeister
	return mailEinstellung{DringendSofort: an, Sammelmail: an}
}

type empfaenger struct {
	Email          string
	DarfBearbeiten bool
}

func (a *app) empfaengerFuer(ctx context.Context, sofort bool) ([]empfaenger, error) {
	personen, err := a.backendPersonen(ctx)
	if err != nil {
		return nil, err
	}
	var liste []empfaenger
	for _, p := range personen {
		e, err := a.sp.mailEinstellung(ctx, p.Email, vorgabeFuer(p.Rolle))
		if err != nil {
			return nil, err
		}
		if (sofort && e.DringendSofort) || (!sofort && e.Sammelmail) {
			b := benutzer{Email: p.Email, Rolle: p.Rolle}
			liste = append(liste, empfaenger{Email: p.Email, DarfBearbeiten: b.darfBearbeiten()})
		}
	}
	return liste, nil
}

type mailMeldung struct {
	Ort, Bereich, Was, Notiz, Wann, Von string
	Dringend, MitFoto                   bool
	LinkGesehen, LinkErledigt           string
}

type mailDaten struct {
	Betreff, Ueberschrift, Schule, LinkApp string
	Offen                                  int
	MitKnoepfen                            bool
	Meldungen                              []mailMeldung
}

func (a *app) wannText(t time.Time) string {
	t = t.In(a.konf.Zeitzone)
	jetzt := a.jetzt().In(a.konf.Zeitzone)
	if t.YearDay() == jetzt.YearDay() && t.Year() == jetzt.Year() {
		return "heute, " + t.Format("15:04")
	}
	tage := [...]string{"So.", "Mo.", "Di.", "Mi.", "Do.", "Fr.", "Sa."}
	return fmt.Sprintf("%s %s, %s", tage[t.Weekday()], t.Format("02.01."), t.Format("15:04"))
}

// baueMail legt für jede Meldung einen eigenen Link an. Wer nur sehen darf (Leitung), bekommt die
// Mail ohne Knöpfe.
func (a *app) baueMail(ctx context.Context, e empfaenger, liste []meldung, sammel bool) (mailDaten, error) {
	d := mailDaten{Schule: a.konf.Schulname, LinkApp: a.konf.OeffentlicheAdresse + "/", MitKnoepfen: e.DarfBearbeiten}
	for _, m := range liste {
		mm := mailMeldung{Ort: m.Ort, Bereich: m.Bereich, Was: m.Was, Notiz: m.Notiz, Wann: a.wannText(m.GemeldetAm),
			Von: strings.SplitN(m.GemeldetVon, "@", 2)[0], Dringend: m.Dringend, MitFoto: m.Foto != ""}
		if e.DarfBearbeiten {
			token, hash := neuesToken()
			if err := a.sp.legeMailLinkAn(ctx, hash, m.ID, e.Email, a.jetzt().Add(linkGueltigkeit)); err != nil {
				return d, err
			}
			basis := a.konf.OeffentlicheAdresse + "/aktion/" + token
			mm.LinkGesehen, mm.LinkErledigt = basis+"?tun=gesehen", basis+"?tun=erledigt"
		}
		d.Meldungen = append(d.Meldungen, mm)
	}
	n := len(liste)
	switch {
	case !sammel:
		d.Betreff = fmt.Sprintf("Dringend: %s – %s", liste[0].Was, liste[0].Ort)
		d.Ueberschrift = "Dringende Meldung"
	case n == 1:
		d.Betreff = fmt.Sprintf("Neue Schadensmeldung: %s – %s", liste[0].Was, liste[0].Ort)
		d.Ueberschrift = "Eine neue Meldung"
	default:
		d.Betreff = fmt.Sprintf("%d neue Schadensmeldungen", n)
		d.Ueberschrift = fmt.Sprintf("%d neue Meldungen", n)
	}
	offen, err := a.sp.anzahlOffen(ctx)
	d.Offen = offen
	return d, err
}

func (a *app) verschicke(ctx context.Context, e empfaenger, liste []meldung, sammel bool) error {
	d, err := a.baueMail(ctx, e, liste, sammel)
	if err != nil {
		return err
	}
	var text, html bytes.Buffer
	if err := mailText.Execute(&text, d); err != nil {
		return err
	}
	if err := mailHTML.Execute(&html, d); err != nil {
		return err
	}
	if err := a.versender.sende(ctx, e.Email, d.Betreff, text.String(), html.String()); err != nil {
		return err
	}
	for _, m := range liste {
		if err := a.sp.vermerkeVersand(ctx, m.ID, e.Email, a.jetzt()); err != nil {
			return err
		}
	}
	return nil
}

// benachrichtigeSofort schickt eine dringende Meldung gleich los. Geht es schief, nimmt die nächste
// Sammelmail sie mit (mail_versand fehlt dann).
func (a *app) benachrichtigeSofort(m meldung) {
	if !a.konf.mailAktiv() || !m.Dringend {
		return
	}
	a.hintergrund.Add(1)
	go func() {
		defer a.hintergrund.Done()
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		liste, err := a.empfaengerFuer(ctx, true)
		if err != nil {
			slog.Error("Mail: Empfänger nicht gelesen", "fehler", err)
			return
		}
		for _, e := range liste {
			if e.Email == m.GemeldetVon {
				continue
			}
			if err := a.verschicke(ctx, e, []meldung{m}, false); err != nil {
				slog.Error("Mail: dringende Meldung nicht verschickt", "id", m.ID, "fehler", err)
			}
		}
	}()
}
