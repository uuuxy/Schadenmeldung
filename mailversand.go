package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"
)

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
	c, err := v.verbinde(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()
	return uebergebe(c, absender.Address, an, baueNachricht(absender, an, betreff, text, html))
}

// verbinde öffnet die Verbindung zum Mailserver: auf Port 465 gleich über TLS, sonst per STARTTLS.
func (v smtpVersender) verbinde(ctx context.Context) (*smtp.Client, error) {
	adresse := net.JoinHostPort(v.k.SMTPHost, v.k.SMTPPort)
	tlsKonfig := &tls.Config{ServerName: v.k.SMTPHost, MinVersion: tls.VersionTLS12}
	verbinder := &net.Dialer{Timeout: 15 * time.Second}
	direktTLS := v.k.SMTPPort == "465"
	var verbindung net.Conn
	var err error
	if direktTLS {
		verbindung, err = (&tls.Dialer{NetDialer: verbinder, Config: tlsKonfig}).DialContext(ctx, "tcp", adresse)
	} else {
		verbindung, err = verbinder.DialContext(ctx, "tcp", adresse)
	}
	if err != nil {
		return nil, err
	}
	_ = verbindung.SetDeadline(time.Now().Add(45 * time.Second))
	c, err := smtp.NewClient(verbindung, v.k.SMTPHost)
	if err != nil {
		_ = verbindung.Close()
		return nil, err
	}
	if err := v.sichereUndMeldeAn(c, direktTLS, tlsKonfig); err != nil {
		_ = c.Close()
		return nil, err
	}
	return c, nil
}

// sichereUndMeldeAn schaltet ohne direktes TLS auf STARTTLS um und meldet sich an. PlainAuth
// verweigert sich ohne TLS (außer auf localhost): Das Passwort geht nie im Klartext.
func (v smtpVersender) sichereUndMeldeAn(c *smtp.Client, direktTLS bool, tlsKonfig *tls.Config) error {
	if ok, _ := c.Extension("STARTTLS"); ok && !direktTLS {
		if err := c.StartTLS(tlsKonfig); err != nil {
			return err
		}
	}
	if v.k.SMTPBenutzer == "" {
		return nil
	}
	return c.Auth(smtp.PlainAuth("", v.k.SMTPBenutzer, v.k.SMTPPasswort, v.k.SMTPHost))
}

// uebergebe reicht eine fertige Nachricht an einen Empfänger weiter.
func uebergebe(c *smtp.Client, von, an string, nachricht []byte) error {
	if err := c.Mail(von); err != nil {
		return err
	}
	if err := c.Rcpt(an); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(nachricht); err != nil {
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
