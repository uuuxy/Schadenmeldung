package main

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// versandPostfach liest das unter Einstellungen → Mail eingetragene Postfach, das Passwort noch
// verschlüsselt; ohne Eintrag errNichtGefunden.
func (s *speicher) versandPostfach(ctx context.Context) (adresse string, passwort []byte, eingetragen time.Time, err error) {
	err = s.db.QueryRow(ctx, "SELECT adresse, passwort, eingetragen FROM versand_postfach").Scan(&adresse, &passwort, &eingetragen)
	if errors.Is(err, pgx.ErrNoRows) {
		err = errNichtGefunden
	}
	return adresse, passwort, eingetragen, err
}

// setzeVersandPostfach ersetzt den Eintrag; mehr als einen lässt die Tabelle nicht zu.
func (s *speicher) setzeVersandPostfach(ctx context.Context, adresse string, passwort []byte, jetzt time.Time) error {
	_, err := s.db.Exec(ctx, `INSERT INTO versand_postfach (adresse, passwort, eingetragen) VALUES ($1, $2, $3)
		ON CONFLICT (eins) DO UPDATE SET adresse = excluded.adresse, passwort = excluded.passwort, eingetragen = excluded.eingetragen`,
		adresse, passwort, jetzt)
	return err
}

func (s *speicher) loescheVersandPostfach(ctx context.Context) error {
	_, err := s.db.Exec(ctx, "DELETE FROM versand_postfach")
	return err
}
