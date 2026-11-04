package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"
)

// Benachrichtigung: Jede neue Meldung geht sofort als eigene Mail an alle aus dem Backend, die Mail
// eingeschaltet haben; der Hausmeister hat sie von sich aus an. Die Mail trägt drei Knöpfe –
// „In Arbeit“, „Fachfirma“ und „Erledigt“ –, die ohne Anmeldung wirken; mehr wären in einer Zeile zu
// viel (M3), „Material bestellt“ steht auf der Seite dahinter. Wie sie hinausgeht: postausgang.go.

const linkGueltigkeit = 14 * 24 * time.Hour

// mailVorgabe: Ohne eigene Wahl bekommt der Hausmeister Mail, alle anderen nicht; sie schalten sie
// sich bei Bedarf selbst an.
func mailVorgabe(rolle string) bool { return rolle == rolleHausmeister }

type empfaenger struct {
	Email          string
	DarfBearbeiten bool
}

// mailEmpfaenger sind alle aus dem Backend, die jetzt Mail wollen, nach Adresse.
func (a *app) mailEmpfaenger(ctx context.Context) (map[string]empfaenger, error) {
	personen, err := a.backendPersonen(ctx)
	if err != nil {
		return nil, err
	}
	wer := map[string]empfaenger{}
	for _, p := range personen {
		an, err := a.sp.mailAn(ctx, p.Email, mailVorgabe(p.Rolle))
		if err != nil {
			return nil, err
		}
		if an {
			b := benutzer{Email: p.Email, Rolle: p.Rolle}
			wer[p.Email] = empfaenger{Email: p.Email, DarfBearbeiten: b.darfBearbeiten()}
		}
	}
	return wer, nil
}

type mailMeldung struct {
	Ort, Bereich, Was, Notiz, Wann, Von       string
	Dringend, MitFoto                         bool
	LinkInArbeit, LinkFachfirma, LinkErledigt string
}

type mailDaten struct {
	Betreff, Ueberschrift, Schule, LinkApp string
	Offen                                  int
	MitKnoepfen                            bool
	Meldung                                mailMeldung
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

// baueMail legt für die Meldung einen eigenen Link an. Wer nur sehen darf (Verwaltung), bekommt die
// Mail ohne Knöpfe.
func (a *app) baueMail(ctx context.Context, e empfaenger, m meldung) (mailDaten, error) {
	d := mailDaten{Schule: a.konf.Schulname, LinkApp: a.konf.OeffentlicheAdresse + "/", MitKnoepfen: e.DarfBearbeiten,
		Betreff: fmt.Sprintf("Neue Schadensmeldung: %s – %s", m.Was, m.Ort), Ueberschrift: "Neue Meldung",
		Meldung: mailMeldung{Ort: m.Ort, Bereich: m.Bereich, Was: m.Was, Notiz: m.Notiz, Wann: a.wannText(m.GemeldetAm),
			Von: strings.SplitN(m.GemeldetVon, "@", 2)[0], Dringend: m.Dringend, MitFoto: m.Foto != ""}}
	if m.Dringend {
		d.Betreff, d.Ueberschrift = fmt.Sprintf("Dringend: %s – %s", m.Was, m.Ort), "Dringende Meldung"
	}
	if e.DarfBearbeiten {
		token, hash := neuesToken()
		if err := a.sp.legeMailLinkAn(ctx, hash, m.ID, e.Email, a.jetzt().Add(linkGueltigkeit)); err != nil {
			return d, err
		}
		basis := a.konf.OeffentlicheAdresse + "/aktion/" + token + "?tun="
		d.Meldung.LinkInArbeit, d.Meldung.LinkFachfirma, d.Meldung.LinkErledigt = basis+"in_arbeit", basis+"fachfirma", basis+"erledigt"
	}
	offen, err := a.sp.anzahlOffen(ctx)
	d.Offen = offen
	return d, err
}

// verschicke schickt eine Meldung an eine Person und hakt sie im Postausgang ab.
func (a *app) verschicke(ctx context.Context, e empfaenger, m meldung) error {
	d, err := a.baueMail(ctx, e, m)
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
	err = a.versender.sende(ctx, e.Email, d.Betreff, text.String(), html.String())
	a.mailVersuch(err)
	if err != nil {
		return err
	}
	return a.sp.vermerkeVersand(ctx, m.ID, e.Email, a.jetzt())
}
