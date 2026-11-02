package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	htmlvorlage "html/template"
	"log/slog"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"strings"
	textvorlage "text/template"
	"time"
)

// Benachrichtigung: Der Hausmeister bekommt neue Meldungen per Mail, ohne gehetzt zu werden.
// Dringendes kommt sofort, alles andere gesammelt zu festen Zeiten (Mo–Fr). Jede Meldung in der
// Mail trägt zwei Knöpfe – „Gesehen“ und „Erledigt“ –, die ohne Anmeldung wirken.

const linkGueltigkeit = 14 * 24 * time.Hour

// versender verschickt eine Mail; in den Tests steht hier eine Attrappe.
type versender interface {
	sende(ctx context.Context, an, betreff, text, html string) error
}

type smtpVersender struct{ k konfiguration }

func (v smtpVersender) sende(ctx context.Context, an, betreff, text, html string) error {
	absender, err := mail.ParseAddress(v.k.MailAbsender)
	if err != nil {
		return fmt.Errorf("MAIL_ABSENDER: %w", err)
	}
	adresse := net.JoinHostPort(v.k.SMTPHost, v.k.SMTPPort)
	tlsKonfig := &tls.Config{ServerName: v.k.SMTPHost, MinVersion: tls.VersionTLS12}
	verbinder := &net.Dialer{Timeout: 15 * time.Second}
	var verbindung net.Conn
	if v.k.SMTPPort == "465" {
		verbindung, err = (&tls.Dialer{NetDialer: verbinder, Config: tlsKonfig}).DialContext(ctx, "tcp", adresse)
	} else {
		verbindung, err = verbinder.DialContext(ctx, "tcp", adresse)
	}
	if err != nil {
		return err
	}
	_ = verbindung.SetDeadline(time.Now().Add(45 * time.Second))
	c, err := smtp.NewClient(verbindung, v.k.SMTPHost)
	if err != nil {
		_ = verbindung.Close()
		return err
	}
	defer func() { _ = c.Close() }()
	if v.k.SMTPPort != "465" {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(tlsKonfig); err != nil {
				return err
			}
		}
	}
	if v.k.SMTPBenutzer != "" {
		// PlainAuth verweigert sich ohne TLS (außer auf localhost): Das Passwort geht nie im Klartext.
		if err := c.Auth(smtp.PlainAuth("", v.k.SMTPBenutzer, v.k.SMTPPasswort, v.k.SMTPHost)); err != nil {
			return err
		}
	}
	if err := c.Mail(absender.Address); err != nil {
		return err
	}
	if err := c.Rcpt(an); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(baueNachricht(absender, an, betreff, text, html)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

// baueNachricht setzt eine Mail mit Text- und HTML-Teil zusammen (multipart/alternative, UTF-8).
func baueNachricht(von *mail.Address, an, betreff, text, html string) []byte {
	grenze := "schaden-" + zufallsHex(12)
	domain := von.Address[strings.LastIndex(von.Address, "@")+1:]
	var b bytes.Buffer
	kopf := func(name, wert string) { fmt.Fprintf(&b, "%s: %s\r\n", name, wert) }
	kopf("From", (&mail.Address{Name: "Schadenmeldung", Address: von.Address}).String())
	kopf("To", an)
	kopf("Subject", mime.QEncoding.Encode("utf-8", betreff))
	kopf("Date", time.Now().Format(time.RFC1123Z))
	kopf("Message-ID", fmt.Sprintf("<%s@%s>", zufallsHex(16), domain))
	kopf("MIME-Version", "1.0")
	kopf("Auto-Submitted", "auto-generated")
	kopf("Content-Type", `multipart/alternative; boundary="`+grenze+`"`)
	b.WriteString("\r\n")
	for _, teil := range []struct{ typ, inhalt string }{{"text/plain", text}, {"text/html", html}} {
		fmt.Fprintf(&b, "--%s\r\nContent-Type: %s; charset=utf-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n", grenze, teil.typ)
		qp := quotedprintable.NewWriter(&b)
		_, _ = qp.Write([]byte(teil.inhalt))
		_ = qp.Close()
		b.WriteString("\r\n")
	}
	fmt.Fprintf(&b, "--%s--\r\n", grenze)
	return b.Bytes()
}

// ── Wer bekommt was ──────────────────────────────────────────────────────────

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

// ── Mails bauen ──────────────────────────────────────────────────────────────

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

// ── Sammelmail ───────────────────────────────────────────────────────────────

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
	jetzt := a.jetzt().In(a.konf.Zeitzone)
	roh, err := a.sp.wert(ctx, "sammelmail_zuletzt")
	if errors.Is(err, errNichtGefunden) {
		if err := a.sp.setzeWert(ctx, "sammelmail_zuletzt", jetzt.Format(time.RFC3339)); err != nil {
			slog.Error("Sammelmail: Stand nicht gespeichert", "fehler", err)
		}
		return
	}
	if err != nil {
		slog.Error("Sammelmail: Stand nicht gelesen", "fehler", err)
		return
	}
	zuletzt, _ := time.Parse(time.RFC3339, roh)
	termin, ok := letzterTermin(jetzt, a.konf.SammelZeiten)
	if !ok || !zuletzt.Before(termin) {
		return
	}
	if err := a.sp.setzeWert(ctx, "sammelmail_zuletzt", termin.Format(time.RFC3339)); err != nil {
		slog.Error("Sammelmail: Stand nicht gespeichert", "fehler", err)
		return
	}
	liste, err := a.empfaengerFuer(ctx, false)
	if err != nil {
		slog.Error("Sammelmail: Empfänger nicht gelesen", "fehler", err)
		return
	}
	for _, e := range liste {
		neu, err := a.sp.ungesendeteFuer(ctx, e.Email, termin)
		if err != nil {
			slog.Error("Sammelmail: Meldungen nicht gelesen", "fehler", err)
			continue
		}
		if len(neu) == 0 {
			continue
		}
		if err := a.verschicke(ctx, e, neu, true); err != nil {
			slog.Error("Sammelmail nicht verschickt", "an", e.Email, "fehler", err)
		}
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

// ── Türen ────────────────────────────────────────────────────────────────────

func (a *app) benachrichtigungLesen(w http.ResponseWriter, r *http.Request, b benutzer) {
	e, err := a.sp.mailEinstellung(r.Context(), b.Email, vorgabeFuer(b.Rolle))
	if err != nil {
		intern(w, err)
		return
	}
	zeiten := make([]string, 0, len(a.konf.SammelZeiten))
	for _, z := range a.konf.SammelZeiten {
		zeiten = append(zeiten, fmt.Sprintf("%02d:%02d", z.Stunde, z.Minute))
	}
	schreibeJSON(w, http.StatusOK, map[string]any{
		"mailAktiv": a.konf.mailAktiv(), "dringendSofort": e.DringendSofort, "sammelmail": e.Sammelmail, "zeiten": zeiten,
	})
}

func (a *app) benachrichtigungSetzen(w http.ResponseWriter, r *http.Request, b benutzer) {
	var ein struct {
		DringendSofort *bool `json:"dringendSofort"`
		Sammelmail     *bool `json:"sammelmail"`
	}
	if !leseJSON(w, r, &ein) {
		return
	}
	e, err := a.sp.mailEinstellung(r.Context(), b.Email, vorgabeFuer(b.Rolle))
	if err != nil {
		intern(w, err)
		return
	}
	if ein.DringendSofort != nil {
		e.DringendSofort = *ein.DringendSofort
	}
	if ein.Sammelmail != nil {
		e.Sammelmail = *ein.Sammelmail
	}
	if err := a.sp.setzeMailEinstellung(r.Context(), b.Email, e); err != nil {
		intern(w, err)
		return
	}
	a.benachrichtigungLesen(w, r, b)
}

// aktionsLink liest den Link aus der Mail. Wer ihn hat, ist der Empfänger der Mail; seine Rolle
// zählt so, wie sie jetzt ist, nicht wie beim Verschicken.
func (a *app) aktionsLink(w http.ResponseWriter, r *http.Request) (meldung, benutzer, bool) {
	id, email, ablauf, err := a.sp.mailLink(r.Context(), tokenHash(r.PathValue("token")))
	if err != nil || a.jetzt().After(ablauf) {
		fehler(w, http.StatusGone, "Dieser Link ist abgelaufen. Die Meldung lässt sich in der App bearbeiten.")
		return meldung{}, benutzer{}, false
	}
	rolle, err := a.rolle(r.Context(), email)
	if err != nil {
		intern(w, err)
		return meldung{}, benutzer{}, false
	}
	b := benutzer{Email: email, Rolle: rolle}
	if !b.darfBearbeiten() {
		fehler(w, http.StatusForbidden, "Mit dieser Rolle lässt sich der Status nicht ändern.")
		return meldung{}, benutzer{}, false
	}
	m, err := a.sp.meldung(r.Context(), id)
	if err != nil {
		nichtGefundenOderIntern(w, err)
		return meldung{}, benutzer{}, false
	}
	return m, b, true
}

func (a *app) aktionLesen(w http.ResponseWriter, r *http.Request) {
	m, b, ok := a.aktionsLink(w, r)
	if !ok {
		return
	}
	schreibeJSON(w, http.StatusOK, a.fuerAktion(m, b, r.PathValue("token")))
}

func (a *app) aktionAusfuehren(w http.ResponseWriter, r *http.Request) {
	m, b, ok := a.aktionsLink(w, r)
	if !ok {
		return
	}
	var ein struct {
		Status string `json:"status"`
	}
	if !leseJSON(w, r, &ein) {
		return
	}
	if !stati[ein.Status] {
		fehler(w, http.StatusBadRequest, "Unbekannter Status.")
		return
	}
	neu, err := a.sp.setzeStatus(r.Context(), m.ID, &ein.Status, nil, a.jetzt())
	if err != nil {
		nichtGefundenOderIntern(w, err)
		return
	}
	a.verteiler.melde()
	schreibeJSON(w, http.StatusOK, a.fuerAktion(neu, b, r.PathValue("token")))
}

func (a *app) aktionFoto(w http.ResponseWriter, r *http.Request) {
	m, _, ok := a.aktionsLink(w, r)
	if !ok || m.Foto == "" {
		if ok {
			http.NotFound(w, r)
		}
		return
	}
	a.liefereFoto(w, r, m.Foto)
}

func (a *app) fuerAktion(m meldung, b benutzer, token string) meldung {
	m = fuerAnsicht(m, b)
	if m.Foto != "" {
		m.Foto = "/api/aktion/" + token + "/foto"
	}
	return m
}

// ── Speicher ─────────────────────────────────────────────────────────────────

func (s *speicher) mailEinstellung(ctx context.Context, email string, vorgabe mailEinstellung) (mailEinstellung, error) {
	e := vorgabe
	err := s.db.QueryRow(ctx, "SELECT dringend_sofort, sammelmail FROM benachrichtigungen WHERE email = $1", email).
		Scan(&e.DringendSofort, &e.Sammelmail)
	if errors.Is(nichtGefunden(err), errNichtGefunden) {
		return vorgabe, nil
	}
	return e, err
}

func (s *speicher) setzeMailEinstellung(ctx context.Context, email string, e mailEinstellung) error {
	_, err := s.db.Exec(ctx, `INSERT INTO benachrichtigungen (email, dringend_sofort, sammelmail) VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE SET dringend_sofort = excluded.dringend_sofort, sammelmail = excluded.sammelmail`,
		email, e.DringendSofort, e.Sammelmail)
	return err
}

func (s *speicher) legeMailLinkAn(ctx context.Context, hash string, meldungID int64, email string, ablauf time.Time) error {
	_, err := s.db.Exec(ctx, "INSERT INTO mail_links (token_hash, meldung_id, empfaenger, laeuft_ab) VALUES ($1, $2, $3, $4)",
		hash, meldungID, email, ablauf)
	return err
}

func (s *speicher) mailLink(ctx context.Context, hash string) (id int64, email string, ablauf time.Time, err error) {
	err = s.db.QueryRow(ctx, "SELECT meldung_id, empfaenger, laeuft_ab FROM mail_links WHERE token_hash = $1", hash).
		Scan(&id, &email, &ablauf)
	return id, email, ablauf, nichtGefunden(err)
}

func (s *speicher) loescheAbgelaufeneMailLinks(ctx context.Context, jetzt time.Time) error {
	_, err := s.db.Exec(ctx, "DELETE FROM mail_links WHERE laeuft_ab < $1", jetzt)
	return err
}

func (s *speicher) vermerkeVersand(ctx context.Context, meldungID int64, email string, zeit time.Time) error {
	_, err := s.db.Exec(ctx, `INSERT INTO mail_versand (meldung_id, empfaenger, gesendet_am) VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`, meldungID, email, zeit)
	return err
}

// ungesendeteFuer: offene Meldungen bis zum Termin, die an diese Person noch nicht gingen und die
// sie nicht selbst gemeldet hat.
func (s *speicher) ungesendeteFuer(ctx context.Context, email string, bis time.Time) ([]meldung, error) {
	return s.abfrage(ctx, `WHERE m.status = 'offen' AND m.gemeldet_am <= $1 AND m.gemeldet_von <> $2
		AND NOT EXISTS (SELECT 1 FROM mail_versand v WHERE v.meldung_id = m.id AND v.empfaenger = $2)
		ORDER BY m.dringend DESC, m.gemeldet_am`, bis, email)
}

func (s *speicher) anzahlOffen(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRow(ctx, "SELECT count(*) FROM meldungen WHERE status <> 'erledigt'").Scan(&n)
	return n, err
}

func (s *speicher) wert(ctx context.Context, schluessel string) (string, error) {
	var w string
	err := s.db.QueryRow(ctx, "SELECT wert FROM einstellungen WHERE schluessel = $1", schluessel).Scan(&w)
	return w, nichtGefunden(err)
}

func (s *speicher) setzeWert(ctx context.Context, schluessel, wert string) error {
	_, err := s.db.Exec(ctx, `INSERT INTO einstellungen (schluessel, wert) VALUES ($1, $2)
		ON CONFLICT (schluessel) DO UPDATE SET wert = excluded.wert`, schluessel, wert)
	return err
}

// ── Vorlagen ─────────────────────────────────────────────────────────────────
// Schlicht und hell wie die App (M3-Farben der Bibliothek), mit Stil am Element: Viele Mailprogramme
// verwerfen Stylesheets im Kopf.

var mailHTML = htmlvorlage.Must(htmlvorlage.New("mail").Parse(`<!doctype html>
<html lang="de"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><meta name="color-scheme" content="light"><title>{{.Betreff}}</title></head>
<body style="margin:0;padding:0;background:#faf9fd;color:#1a1c1e;font-family:Roboto,'Segoe UI',Helvetica,Arial,sans-serif;">
<div style="max-width:560px;margin:0 auto;padding:24px 16px;">
<p style="margin:0 0 4px;font-size:14px;line-height:20px;color:#42474e;">Schadenmeldung{{if .Schule}} · {{.Schule}}{{end}}</p>
<h1 style="margin:0 0 20px;font-size:24px;line-height:32px;font-weight:400;">{{.Ueberschrift}}</h1>
{{range .Meldungen}}<div style="background:#f4f3f7;border-radius:12px;padding:16px;margin:0 0 12px;">
{{if .Dringend}}<p style="margin:0 0 8px;"><span style="display:inline-block;background:#ffdad6;color:#93000a;border-radius:8px;padding:2px 8px;font-size:12px;line-height:16px;font-weight:500;">Dringend</span></p>{{end}}
<p style="margin:0;font-size:16px;line-height:24px;font-weight:500;">{{.Ort}}{{if .Bereich}}<span style="font-weight:400;color:#42474e;font-size:14px;"> · {{.Bereich}}</span>{{end}}</p>
<p style="margin:2px 0 0;font-size:16px;line-height:24px;">{{.Was}}</p>
{{if .Notiz}}<p style="margin:2px 0 0;font-size:14px;line-height:20px;color:#42474e;">{{.Notiz}}</p>{{end}}
<p style="margin:8px 0 0;font-size:12px;line-height:16px;color:#42474e;">{{.Wann}}{{if .Von}} · {{.Von}}{{end}}{{if .MitFoto}} · mit Foto{{end}}</p>
{{if $.MitKnoepfen}}<p style="margin:14px 0 0;">
<a href="{{.LinkGesehen}}" style="display:inline-block;margin:0 8px 0 0;padding:9px 20px;border:1px solid #c2c7cf;border-radius:999px;background:#ffffff;color:#42474e;font-size:14px;line-height:20px;font-weight:500;text-decoration:none;">Gesehen</a>
<a href="{{.LinkErledigt}}" style="display:inline-block;margin:0;padding:10px 21px;border-radius:999px;background:#0061a4;color:#ffffff;font-size:14px;line-height:20px;font-weight:500;text-decoration:none;">Erledigt</a>
</p>{{end}}
</div>
{{end}}<p style="margin:20px 0 0;font-size:14px;line-height:20px;"><a href="{{.LinkApp}}" style="color:#0061a4;font-weight:500;text-decoration:none;">Alle Aufgaben in der App</a><span style="color:#42474e;"> · {{.Offen}} offen</span></p>
<p style="margin:24px 0 0;font-size:12px;line-height:16px;color:#72777f;">{{if .MitKnoepfen}}„Gesehen“ zeigt dem Kollegium, dass Bescheid gesagt ist – ohne Zusage, wann. {{end}}Diese Mail kommt, weil die Adresse in der Schadenmeldung eingetragen ist. Abbestellen: in der App im Kontomenü.</p>
</div></body></html>`))

var mailText = textvorlage.Must(textvorlage.New("mail").Parse(`{{.Ueberschrift}}
{{range .Meldungen}}
{{if .Dringend}}[DRINGEND] {{end}}{{.Ort}}{{if .Bereich}} ({{.Bereich}}){{end}}
{{.Was}}{{if .Notiz}}
{{.Notiz}}{{end}}
{{.Wann}}{{if .Von}} · {{.Von}}{{end}}
{{if $.MitKnoepfen}}Gesehen:  {{.LinkGesehen}}
Erledigt: {{.LinkErledigt}}
{{end}}{{end}}
Alle Aufgaben ({{.Offen}} offen): {{.LinkApp}}

Abbestellen: in der App im Kontomenü.
`))
