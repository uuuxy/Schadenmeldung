package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/textproto"
	"strings"
	"time"
)

// Das Postfach, über das die App ihre Mails verschickt. Steht MAIL_ABSENDER in der .env, gilt das
// fest. Sonst trägt es jemand mit allen Rechten unter Einstellungen → Mail ein, am besten die Person,
// der es gehört: So muss sie ihr Passwort niemandem sagen. Die App prüft es mit einer Testmail,
// speichert es verschlüsselt und zeigt es nie wieder an.

// zugang ist die Anmeldung beim Mailserver samt Absender.
type zugang struct{ absender, benutzer, passwort string }

// postfachStand ist, was die Einstellungen zeigen; das Passwort gehört nie dazu.
type postfachStand struct {
	Mailserver  bool       `json:"mailserver"`
	Fest        bool       `json:"fest"`
	Speicherbar bool       `json:"speicherbar"`
	Adresse     string     `json:"adresse"`
	Eingetragen *time.Time `json:"eingetragen,omitempty"`
}

func (a *app) postfachFest() bool { return a.konf.MailAbsender != "" }

// mailAktiv: In der .env steht ein Mailserver, und es gibt ein Postfach, fest oder eingetragen.
func (a *app) mailAktiv(ctx context.Context) bool {
	if a.konf.SMTPHost == "" {
		return false
	}
	if a.postfachFest() {
		return true
	}
	_, _, _, err := a.sp.versandPostfach(ctx)
	return err == nil
}

// versandZugang liest die Anmeldung vor jedem Versand neu: Ein frisch eingetragenes Passwort gilt
// ab der nächsten Mail, ohne Neustart.
func (a *app) versandZugang(ctx context.Context) (zugang, error) {
	if a.postfachFest() {
		return zugang{a.konf.MailAbsender, a.konf.SMTPBenutzer, a.konf.SMTPPasswort}, nil
	}
	adresse, geheim, _, err := a.sp.versandPostfach(ctx)
	if err != nil {
		return zugang{}, fmt.Errorf("kein Postfach für den Versand: %w", err)
	}
	klar, err := entschluesseleGeheim(a.konf.Schluessel, geheim)
	if err != nil {
		return zugang{}, fmt.Errorf("Passwort des Postfachs nicht lesbar, wurde SCHLUESSEL geändert? %w", err)
	}
	return zugang{adresse, adresse, string(klar)}, nil
}

func (a *app) postfachLesen(w http.ResponseWriter, r *http.Request, _ benutzer) {
	st := postfachStand{Mailserver: a.konf.SMTPHost != "", Fest: a.postfachFest(), Speicherbar: len(a.konf.Schluessel) == schluesselLaenge}
	if st.Fest {
		st.Adresse = a.konf.MailAbsender
		schreibeJSON(w, http.StatusOK, st)
		return
	}
	adresse, _, wann, err := a.sp.versandPostfach(r.Context())
	switch {
	case err == nil:
		st.Adresse, st.Eingetragen = adresse, &wann
	case !errors.Is(err, errNichtGefunden):
		intern(w, err)
		return
	}
	schreibeJSON(w, http.StatusOK, st)
}

// postfachSetzen schickt mit dem neuen Zugang eine Testmail an die Person, die ihn einträgt, und
// speichert erst, wenn sie rausging. Die Bremse schützt die ganze App: Der Mailserver der Schule
// sperrt nach ein paar falschen Passwörtern auch die Anmeldung aller.
func (a *app) postfachSetzen(w http.ResponseWriter, r *http.Request, b benutzer) {
	var ein struct {
		Adresse  string `json:"adresse"`
		Passwort string `json:"passwort"`
	}
	if !leseJSON(w, r, &ein) {
		return
	}
	adresse := strings.ToLower(a.konf.vervollstaendige(strings.TrimSpace(ein.Adresse)))
	if status, text := a.postfachHindernis(adresse, ein.Passwort); text != "" {
		fehler(w, status, text)
		return
	}
	if rest := a.postfachBremse.versuch("postfach", a.jetzt()); rest > 0 {
		zuVieleVersuche(w, rest)
		return
	}
	if err := a.pruefeVersand(r.Context(), zugang{adresse, adresse, ein.Passwort}, b.Email); err != nil {
		slog.Warn("Postfach: Testmail gescheitert", "fehler", fuersLog(err))
		fehler(w, http.StatusBadRequest, postfachFehlerText(err))
		return
	}
	a.postfachBremse.erfolg("postfach")
	geheim, err := verschluesseleGeheim(a.konf.Schluessel, []byte(ein.Passwort))
	if err == nil {
		err = a.sp.setzeVersandPostfach(r.Context(), adresse, geheim, a.jetzt())
	}
	if err != nil {
		intern(w, err)
		return
	}
	a.postfachLesen(w, r, b)
}

// postfachHindernis nennt, was dem Eintragen im Weg steht, samt Status; leer, wenn nichts.
func (a *app) postfachHindernis(adresse, passwort string) (int, string) {
	switch {
	case a.konf.SMTPHost == "":
		return http.StatusConflict, "Auf dem Server ist kein Mailserver eingetragen (SMTP_HOST in der .env)."
	case a.postfachFest():
		return http.StatusConflict, "Das Postfach steht fest in der .env des Servers."
	case len(a.konf.Schluessel) != schluesselLaenge:
		return http.StatusConflict, "Zum sicheren Speichern fehlt SCHLUESSEL in der .env; ./update.sh legt ihn an."
	case !gueltigeAdresse(adresse) || !a.konf.domainErlaubt(adresse):
		return http.StatusBadRequest, "Bitte ein Postfach der Schule eingeben."
	case passwort == "" || len(passwort) > 512:
		return http.StatusBadRequest, "Bitte das Passwort des Postfachs eingeben."
	}
	return 0, ""
}

// postfachFehlerText trennt ein falsches Passwort (SMTP 535) von allem anderen.
func postfachFehlerText(err error) string {
	var antwort *textproto.Error
	if errors.As(err, &antwort) && antwort.Code == 535 {
		return "Der Mailserver hat das Passwort nicht angenommen. Bitte genau prüfen: Nach mehreren Fehlversuchen sperrt er die App für eine Weile."
	}
	return "Die Testmail ging nicht raus. Stimmt die Adresse, und ist der Mailserver erreichbar?"
}

func (a *app) postfachLoeschen(w http.ResponseWriter, r *http.Request, b benutzer) {
	if err := a.sp.loescheVersandPostfach(r.Context()); err != nil {
		intern(w, err)
		return
	}
	a.postfachLesen(w, r, b)
}

// pruefeVersandSMTP schickt mit dem neuen Zugang eine Testmail; kommt sie raus, stimmen Adresse,
// Passwort und Absender.
func (a *app) pruefeVersandSMTP(ctx context.Context, z zugang, an string) error {
	v := smtpVersender{k: a.konf, zugang: func(context.Context) (zugang, error) { return z, nil }}
	return v.sende(ctx, an, "", "Schadensmeldung: Postfach eingerichtet", testmailText, testmailHTML)
}

const testmailText = `Postfach eingerichtet

Über dieses Postfach verschickt die Schadensmeldung ab jetzt ihre Mails: neue Meldungen an den Hausmeister und die Nachricht an die Meldenden, sobald etwas erledigt ist.
`

const testmailHTML = `<!doctype html><html lang="de"><body style="margin:0;padding:0;background:#faf9fd;color:#1a1c1e;font-family:Roboto,'Segoe UI',Helvetica,Arial,sans-serif;">
<div style="max-width:560px;margin:0 auto;padding:24px 16px;">
<p style="margin:0 0 4px;font-size:14px;line-height:20px;color:#42474e;">Schadensmeldung</p>
<h1 style="margin:0 0 12px;font-size:24px;line-height:32px;font-weight:400;">Postfach eingerichtet</h1>
<p style="margin:0;font-size:16px;line-height:24px;">Über dieses Postfach verschickt die Schadensmeldung ab jetzt ihre Mails: neue Meldungen an den Hausmeister und die Nachricht an die Meldenden, sobald etwas erledigt ist.</p>
</div></body></html>`
