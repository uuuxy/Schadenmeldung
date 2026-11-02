package main

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// raeume liefert die Bereiche in der Reihenfolge, in der die Verwaltung sie eingegeben hat (EG vor
// 1. OG, nicht nach dem Alphabet), darin die Räume nach Namen.
func (s *speicher) raeume(ctx context.Context) ([]raum, error) {
	zeilen, err := s.db.Query(ctx, `SELECT r.id, r.name, r.bereich, r.code FROM raeume r
		ORDER BY (SELECT min(x.id) FROM raeume x WHERE x.bereich = r.bereich), lower(r.name)`)
	if err != nil {
		return nil, err
	}
	liste, err := pgx.CollectRows(zeilen, pgx.RowToStructByPos[raum])
	if liste == nil {
		liste = []raum{}
	}
	return liste, err
}

func (s *speicher) raum(ctx context.Context, id int64) (raum, error) {
	var r raum
	err := s.db.QueryRow(ctx, "SELECT id, name, bereich, code FROM raeume WHERE id = $1", id).
		Scan(&r.ID, &r.Name, &r.Bereich, &r.Code)
	return r, nichtGefunden(err)
}

// legeRaeumeAn übernimmt neue Räume und überspringt Namen, die es schon gibt.
func (s *speicher) legeRaeumeAn(ctx context.Context, neue []raum) (int, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	angelegt := 0
	for _, r := range neue {
		tag, err := tx.Exec(ctx, "INSERT INTO raeume (name, bereich, code) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING",
			r.Name, r.Bereich, raumCode())
		if err != nil {
			return 0, err
		}
		angelegt += int(tag.RowsAffected())
	}
	return angelegt, tx.Commit(ctx)
}

func (s *speicher) loescheRaum(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, "DELETE FROM raeume WHERE id = $1", id)
	if err == nil && tag.RowsAffected() == 0 {
		return errNichtGefunden
	}
	return err
}

func raumCode() string {
	b := make([]byte, 5)
	_, _ = rand.Read(b)
	return strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b))
}

// aendereRaum schreibt Name und Bereich und zieht den Ort der Meldungen dieses Raums nach, in einer
// Transaktion. Gibt es den Namen schon, lehnt die Datenbank ab (raeume_name) und der Nutzer erfährt es.
func (s *speicher) aendereRaum(ctx context.Context, r raum) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, "UPDATE raeume SET name = $2, bereich = $3 WHERE id = $1", r.ID, r.Name, r.Bereich)
	var pgFehler *pgconn.PgError
	switch {
	case errors.As(err, &pgFehler) && pgFehler.Code == "23505":
		return eingabeFehler("Einen Raum mit diesem Namen gibt es schon.")
	case err != nil:
		return err
	case tag.RowsAffected() == 0:
		return errNichtGefunden
	}
	if _, err := tx.Exec(ctx, "UPDATE meldungen SET ort = $2 WHERE raum_id = $1", r.ID, r.Name); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
