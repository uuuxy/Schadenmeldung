package main

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-imap/commands"
)

// pruefeIMAP meldet sich mit den Zugangsdaten am Mailserver der Schule an, über TLS ab der ersten
// Zeile. Das Passwort geht nur dorthin und wird nirgends gespeichert.
func (a *app) pruefeIMAP(ctx context.Context, email, passwort string) error {
	if a.konf.IMAPHost == "mock" {
		// konfiguration.pruefe lässt mock nur mit APP_ENV=local zu.
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, imapFrist)
	defer cancel()

	adresse := net.JoinHostPort(a.konf.IMAPHost, a.konf.IMAPPort)
	verbinder := &tls.Dialer{Config: &tls.Config{ServerName: a.konf.IMAPHost, MinVersion: tls.VersionTLS12}}
	verbindung, err := verbinder.DialContext(ctx, "tcp", adresse)
	if err != nil {
		slog.Error("Mailserver nicht erreichbar", "adresse", adresse, "fehler", err)
		return errMailserverWeg
	}
	// Endet die Frist, schließt diese Goroutine die Verbindung; das löst jedes hängende Lesen.
	fertig := make(chan struct{})
	defer close(fertig)
	go func() {
		select {
		case <-ctx.Done():
			_ = verbindung.Close()
		case <-fertig:
		}
	}()

	c, err := client.New(verbindung)
	if err != nil {
		_ = verbindung.Close()
		slog.Error("Mailserver ohne Begrüßung", "fehler", err)
		return errMailserverWeg
	}
	c.Timeout = imapBefehlFrist
	defer func() { _ = c.Logout() }()

	// Execute statt Login: Nur so lässt sich „Server lehnt ab“ (NO, BAD) von „Server ist weg“ trennen.
	status, err := c.Execute(&commands.Login{Username: email, Password: passwort}, nil)
	switch {
	case err != nil:
		slog.Warn("Anmeldung am Mailserver abgebrochen", "fehler", err)
		return errMailserverWeg
	case status == nil:
		return errMailserverWeg
	case status.Type == imap.StatusRespOk:
		c.SetState(imap.AuthenticatedState, nil)
		return nil
	default:
		return errAbgelehnt
	}
}
