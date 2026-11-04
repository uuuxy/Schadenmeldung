package main

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

const meldungSpalten = `m.id, m.raum_id, m.ort, coalesce(r.bereich, ''), m.kategorie, m.was, m.notiz,
	coalesce(m.foto, ''), m.dringend, m.status, m.antwort, m.gemeldet_von, m.gemeldet_am, m.geaendert_am,
	m.erledigt_am
	FROM meldungen m LEFT JOIN raeume r ON r.id = m.raum_id`

func scanneMeldung(z pgx.Row) (meldung, error) {
	var m meldung
	err := z.Scan(&m.ID, &m.RaumID, &m.Ort, &m.Bereich, &m.Kategorie, &m.Was, &m.Notiz, &m.Foto, &m.Dringend,
		&m.Status, &m.Antwort, &m.GemeldetVon, &m.GemeldetAm, &m.GeaendertAm, &m.ErledigtAm)
	return m, err
}

func (s *speicher) abfrage(ctx context.Context, rest string, args ...any) ([]meldung, error) {
	return alle(ctx, s.db, func(z pgx.CollectableRow) (meldung, error) { return scanneMeldung(z) },
		"SELECT "+meldungSpalten+" "+rest, args...)
}

// offeneMeldungen zeigt Dringendes zuerst, sonst das Neueste oben.
func (s *speicher) offeneMeldungen(ctx context.Context) ([]meldung, error) {
	return s.abfrage(ctx, "WHERE m.status <> 'erledigt' ORDER BY m.dringend DESC, m.gemeldet_am DESC, m.id DESC")
}

func (s *speicher) erledigteMeldungen(ctx context.Context, seit time.Time) ([]meldung, error) {
	return s.abfrage(ctx, "WHERE m.status = 'erledigt' AND m.erledigt_am >= $1 ORDER BY m.erledigt_am DESC, m.id DESC LIMIT 300", seit)
}

func (s *speicher) meldung(ctx context.Context, id int64) (meldung, error) {
	m, err := scanneMeldung(s.db.QueryRow(ctx, "SELECT "+meldungSpalten+" WHERE m.id = $1", id))
	return m, nichtGefunden(err)
}

// errSchonDa: Dieselbe Kennung kam gleichzeitig zweimal (zwei Tabs); die andere Anfrage hat gespeichert.
var errSchonDa = errors.New("meldung mit dieser Kennung gibt es schon")

// meldungMitKennung findet eine Meldung, die das Handy dieser Person schon einmal geschickt hat.
func (s *speicher) meldungMitKennung(ctx context.Context, email, kennung string) (meldung, error) {
	m, err := scanneMeldung(s.db.QueryRow(ctx, "SELECT "+meldungSpalten+" WHERE m.gemeldet_von = $1 AND m.kennung = $2", email, kennung))
	return m, nichtGefunden(err)
}

// legeMeldungAn speichert Meldung und Foto in einer Transaktion: Es gibt kein Foto ohne Meldung.
func (s *speicher) legeMeldungAn(ctx context.Context, m meldung, foto []byte) (meldung, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return meldung{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var fotoName *string
	if foto != nil {
		if _, err := tx.Exec(ctx, "INSERT INTO fotos (name, daten, angelegt) VALUES ($1, $2, $3)", m.Foto, foto, m.GemeldetAm); err != nil {
			return meldung{}, err
		}
		fotoName = &m.Foto
	}
	var id int64
	// Dieselbe Kennung gleichzeitig zweimal: nichts anlegen, auch das Foto nicht (Rollback).
	err = tx.QueryRow(ctx, `INSERT INTO meldungen
		(raum_id, ort, kategorie, was, notiz, foto, dringend, gemeldet_von, gemeldet_am, geaendert_am, kennung)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9, NULLIF($10, ''))
		ON CONFLICT (gemeldet_von, kennung) DO NOTHING RETURNING id`,
		m.RaumID, m.Ort, m.Kategorie, m.Was, m.Notiz, fotoName, m.Dringend, m.GemeldetVon, m.GemeldetAm, m.Kennung).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return meldung{}, errSchonDa
	}
	if err != nil {
		return meldung{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return meldung{}, err
	}
	return s.meldung(ctx, id)
}

// setzeStatus ändert Status und Antwort; nil lässt den bisherigen Wert stehen. Der Zeitpunkt der
// Erledigung bleibt beim ersten Abhaken stehen und fällt weg, sobald die Meldung wieder offen ist.
// Erledigt wer die eigene Meldung selbst, braucht es keine Rückmeldung per Mail.
func (s *speicher) setzeStatus(ctx context.Context, id int64, status, antwort *string, wer string, jetzt time.Time) (meldung, error) {
	tag, err := s.db.Exec(ctx, `UPDATE meldungen SET
			status = coalesce($2, status),
			antwort = coalesce($3, antwort),
			erledigt_am = CASE WHEN coalesce($2, status) = 'erledigt' THEN coalesce(erledigt_am, $4) END,
			rueckmeldung_am = CASE WHEN $2 = 'erledigt' AND gemeldet_von = $5
				THEN coalesce(rueckmeldung_am, $4) ELSE rueckmeldung_am END,
			geaendert_am = $4
		WHERE id = $1`, id, status, antwort, jetzt, wer)
	if err != nil {
		return meldung{}, err
	}
	if tag.RowsAffected() == 0 {
		return meldung{}, errNichtGefunden
	}
	return s.meldung(ctx, id)
}

// loescheMeldung entfernt die Meldung samt Foto.
func (s *speicher) loescheMeldung(ctx context.Context, id int64) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var foto *string
	if err := tx.QueryRow(ctx, "DELETE FROM meldungen WHERE id = $1 RETURNING foto", id).Scan(&foto); err != nil {
		return nichtGefunden(err)
	}
	if foto != nil {
		if _, err := tx.Exec(ctx, "DELETE FROM fotos WHERE name = $1", *foto); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *speicher) foto(ctx context.Context, name string) ([]byte, error) {
	var daten []byte
	err := s.db.QueryRow(ctx, "SELECT daten FROM fotos WHERE name = $1", name).Scan(&daten)
	return daten, nichtGefunden(err)
}
