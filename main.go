// Schadenmeldung: Das Kollegium meldet Schäden im Schulgebäude, der Hausmeister hakt sie am Handy ab.
package main

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Die gebaute Oberfläche steckt im Programm; auf dem Server liegt damit nur eine Datei.
//
//go:embed all:frontend/dist
var gebauteOberflaeche embed.FS

func main() {
	// Für den Gesundheitscheck von Docker: Das Abbild hat keine Shell und kein curl.
	if len(os.Args) == 2 && os.Args[1] == "gesund" {
		antwort, err := http.Get("http://127.0.0.1:" + wert("PORT", "8090") + "/healthz")
		if err != nil || antwort.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		return
	}
	if err := starte(); err != nil {
		slog.Error("Start abgebrochen", "fehler", err)
		os.Exit(1)
	}
}

func starte() error {
	konf, err := ladeKonfiguration()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sp, err := oeffneSpeicher(ctx, konf.DatenbankURL)
	if err != nil {
		return err
	}
	defer sp.schliesse()

	statisch, err := fs.Sub(gebauteOberflaeche, "frontend/dist")
	if err != nil {
		return err
	}
	a := neueApp(konf, sp, statisch)
	go a.raeumeRegelmaessigAuf(ctx)
	go a.sammelmailTakt(ctx)

	srv := &http.Server{
		Addr:              ":" + konf.Port,
		Handler:           a.routen(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       2 * time.Minute,
		WriteTimeout:      2 * time.Minute,
		IdleTimeout:       2 * time.Minute,
	}
	srv.RegisterOnShutdown(a.verteiler.schliesse)
	go func() {
		<-ctx.Done()
		ende, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(ende)
		a.hintergrund.Wait()
	}()

	slog.Info("Schadenmeldung läuft", "port", konf.Port, "lokal", konf.Lokal)
	if konf.IMAPHost == "mock" {
		slog.Warn("IMAP_HOST=mock: Jedes Passwort wird angenommen. Nur für die Entwicklung!")
	}
	if len(konf.Admins) == 0 {
		slog.Warn("ADMINS ist leer: Niemand kann Räume anlegen, Rollen vergeben oder QR-Codes drucken.")
	}
	if !konf.mailAktiv() {
		slog.Warn("SMTP_HOST ist leer: Der Hausmeister bekommt keine Mails.")
	}
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
