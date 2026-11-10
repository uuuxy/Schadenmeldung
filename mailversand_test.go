package main

import (
	"context"
	"encoding/base64"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/textproto"
	"strings"
	"testing"
)

// smtpAttrappe ist ein winziger Mailserver für die Tests: Er nimmt eine Mail an und merkt sich, was
// ankam. Mit ablehnen weist er die Anmeldung ab wie ein echter Server bei falschem Passwort (535).
type smtpAttrappe struct {
	host, port         string
	anmeldung, von, an string
	daten              []byte
	fertig             chan struct{}
	ablehnen           bool
}

func starteSMTPAttrappe(t *testing.T, ablehnen bool) *smtpAttrappe {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	a := &smtpAttrappe{fertig: make(chan struct{}), ablehnen: ablehnen}
	a.host, a.port, _ = net.SplitHostPort(l.Addr().String())
	t.Cleanup(func() { _ = l.Close() })
	go func() {
		defer close(a.fertig)
		c, err := l.Accept()
		if err != nil {
			return
		}
		defer func() { _ = c.Close() }()
		a.bediene(textproto.NewConn(c))
	}()
	return a
}

func (a *smtpAttrappe) bediene(tp *textproto.Conn) {
	_ = tp.PrintfLine("220 attrappe ESMTP")
	for {
		zeile, err := tp.ReadLine()
		if err != nil {
			return
		}
		switch strings.ToUpper(strings.Fields(zeile + " x")[0]) {
		case "EHLO", "HELO":
			_ = tp.PrintfLine("250-attrappe")
			_ = tp.PrintfLine("250 AUTH PLAIN")
		case "AUTH":
			felder := strings.Fields(zeile)
			roh, _ := base64.StdEncoding.DecodeString(felder[len(felder)-1])
			a.anmeldung = string(roh)
			if a.ablehnen {
				_ = tp.PrintfLine("535 5.7.8 authentication failed")
				continue
			}
			_ = tp.PrintfLine("235 angemeldet")
		case "MAIL":
			a.von = zeile
			_ = tp.PrintfLine("250 ok")
		case "RCPT":
			a.an = zeile
			_ = tp.PrintfLine("250 ok")
		case "DATA":
			_ = tp.PrintfLine("354 los")
			a.daten, _ = tp.ReadDotBytes()
			_ = tp.PrintfLine("250 angenommen")
		case "QUIT":
			_ = tp.PrintfLine("221 tschüss")
			return
		default:
			_ = tp.PrintfLine("502 unbekannt")
		}
	}
}

func (a *smtpAttrappe) versender(lokal bool, z zugang) smtpVersender {
	return smtpVersender{
		k:      konfiguration{SMTPHost: a.host, SMTPPort: a.port, Lokal: lokal},
		zugang: func(context.Context) (zugang, error) { return z, nil },
	}
}

var postfachDesHausmeisters = zugang{absender: "hm@schule.de", benutzer: "hm@schule.de", passwort: "Sommer2026!"}

func TestMailGehtMitAnmeldungRaus(t *testing.T) {
	a := starteSMTPAttrappe(t, false)
	betreff := "Neue Schadensmeldung: Tür klemmt – Raum 204"
	err := a.versender(true, postfachDesHausmeisters).sende(context.Background(), "lehrer@schule.de", betreff, "Gemeldet: Tür klemmt.", "<p>Tür klemmt</p>")
	if err != nil {
		t.Fatal(err)
	}
	<-a.fertig
	if a.anmeldung != "\x00hm@schule.de\x00Sommer2026!" || !strings.Contains(a.von, "<hm@schule.de>") || !strings.Contains(a.an, "<lehrer@schule.de>") {
		t.Fatalf("Anmeldung %q, von %q, an %q", a.anmeldung, a.von, a.an)
	}
	m, err := mail.ReadMessage(strings.NewReader(string(a.daten)))
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := new(mime.WordDecoder).DecodeHeader(m.Header.Get("Subject")); b != betreff {
		t.Fatalf("Betreff kam als %q an", b)
	}
	if von, _ := m.Header.AddressList("From"); len(von) != 1 || von[0].Name != "Schadensmeldung" || von[0].Address != "hm@schule.de" {
		t.Fatalf("Absender: %v", von)
	}
	teile := mailTeile(t, m)
	if teile["text/plain"] != "Gemeldet: Tür klemmt." || teile["text/html"] != "<p>Tür klemmt</p>" {
		t.Fatalf("Teile: %q", teile)
	}
}

// mailTeile liest die Teile einer multipart/alternative-Mail, nach Typ, quoted-printable entschlüsselt.
func mailTeile(t *testing.T, m *mail.Message) map[string]string {
	t.Helper()
	_, param, err := mime.ParseMediaType(m.Header.Get("Content-Type"))
	if err != nil {
		t.Fatal(err)
	}
	teile := map[string]string{}
	r := multipart.NewReader(m.Body, param["boundary"])
	for {
		teil, err := r.NextRawPart()
		if errors.Is(err, io.EOF) {
			return teile
		}
		if err != nil {
			t.Fatal(err)
		}
		typ, _, _ := mime.ParseMediaType(teil.Header.Get("Content-Type"))
		inhalt, _ := io.ReadAll(quotedprintable.NewReader(teil))
		teile[typ] = strings.TrimRight(string(inhalt), "\r\n")
	}
}

func TestFalschesPasswortMeldetDerMailserver(t *testing.T) {
	a := starteSMTPAttrappe(t, true)
	err := a.versender(true, postfachDesHausmeisters).sende(context.Background(), "lehrer@schule.de", "Test", "x", "x")
	<-a.fertig
	var antwort *textproto.Error
	if !errors.As(err, &antwort) || antwort.Code != 535 || !strings.Contains(postfachFehlerText(err), "Passwort") {
		t.Fatalf("Fehler %v, Hinweis %q", err, postfachFehlerText(err))
	}
	if a.daten != nil {
		t.Fatal("ohne Anmeldung darf keine Mail rausgehen")
	}
}

func TestPasswortNieOhneVerschluesselung(t *testing.T) {
	a := starteSMTPAttrappe(t, false)
	// Außerhalb der Entwicklung muss der Mailserver STARTTLS anbieten, sonst ginge das Passwort im Klartext.
	err := a.versender(false, postfachDesHausmeisters).sende(context.Background(), "lehrer@schule.de", "Test", "x", "x")
	<-a.fertig
	if err == nil || !strings.Contains(err.Error(), "STARTTLS") || a.anmeldung != "" {
		t.Fatalf("ohne STARTTLS: Fehler %v, Anmeldung %q", err, a.anmeldung)
	}
}
