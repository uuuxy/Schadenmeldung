package main

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/base32"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Die Migrationen stecken im Programm und laufen beim Start, wie in der Bibliothek.
//
//go:embed migrations/*.sql
var migrationen embed.FS

// speicher kapselt PostgreSQL.
type speicher struct {
	db *pgxpool.Pool
}

var errNichtGefunden = errors.New("nicht gefunden")

type raum struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Bereich string `json:"bereich"`
	Code    string `json:"code"`
}

type meldung struct {
	ID          int64      `json:"id"`
	RaumID      *int64     `json:"raumId"`
	Ort         string     `json:"ort"`
	Bereich     string     `json:"bereich"`
	Kategorie   string     `json:"kategorie"`
	Was         string     `json:"was"`
	Notiz       string     `json:"notiz"`
	Foto        string     `json:"foto"`
	Dringend    bool       `json:"dringend"`
	Status      string     `json:"status"`
	Antwort     string     `json:"antwort"`
	GemeldetVon string     `json:"gemeldetVon,omitempty"`
	GemeldetAm  time.Time  `json:"gemeldetAm"`
	GeaendertAm time.Time  `json:"geaendertAm"`
	ErledigtAm  *time.Time `json:"erledigtAm"`
}

func oeffneSpeicher(ctx context.Context, url string) (*speicher, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("DATABASE_URL ist nicht lesbar: %w", err)
	}
	return oeffneMitKonfiguration(ctx, cfg)
}

func oeffneMitKonfiguration(ctx context.Context, cfg *pgxpool.Config) (*speicher, error) {
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	// Startet alles zugleich (Docker Compose), ist die Datenbank manchmal ein paar Sekunden später bereit.
	for versuch := 1; ; versuch++ {
		if err = pool.Ping(ctx); err == nil {
			break
		}
		if versuch == 30 {
			pool.Close()
			return nil, fmt.Errorf("datenbank nicht erreichbar: %w", err)
		}
		select {
		case <-ctx.Done():
			pool.Close()
			return nil, ctx.Err()
		case <-time.After(time.Second):
		}
	}
	s := &speicher{db: pool}
	if err := s.migriere(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return s, nil
}

func (s *speicher) schliesse() { s.db.Close() }

// migriere spielt die nummerierten Dateien aus migrations/ ein, die noch fehlen, jede in eigener
// Transaktion. Die Sperre hält zwei gleichzeitig startende Server davon ab, doppelt zu migrieren.
func (s *speicher) migriere(ctx context.Context) error {
	if _, err := s.db.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_version (
		version INT PRIMARY KEY, angewendet TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
		return err
	}
	dateien, err := fs.Glob(migrationen, "migrations/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(dateien)
	for _, datei := range dateien {
		nummer, err := strconv.Atoi(strings.SplitN(path.Base(datei), "_", 2)[0])
		if err != nil {
			return fmt.Errorf("migration %s: der Name muss mit einer Nummer beginnen", datei)
		}
		if err := s.spieleEin(ctx, nummer, datei); err != nil {
			return fmt.Errorf("migration %s: %w", datei, err)
		}
	}
	return nil
}

func (s *speicher) spieleEin(ctx context.Context, nummer int, datei string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(7262)"); err != nil {
		return err
	}
	var vorhanden bool
	if err := tx.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM schema_version WHERE version = $1)", nummer).Scan(&vorhanden); err != nil {
		return err
	}
	if vorhanden {
		return nil
	}
	inhalt, err := migrationen.ReadFile(datei)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, string(inhalt)); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "INSERT INTO schema_version (version) VALUES ($1)", nummer); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func nichtGefunden(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return errNichtGefunden
	}
	return err
}

// ── Räume ────────────────────────────────────────────────────────────────────

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

// ── Meldungen ────────────────────────────────────────────────────────────────

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
	zeilen, err := s.db.Query(ctx, "SELECT "+meldungSpalten+" "+rest, args...)
	if err != nil {
		return nil, err
	}
	liste, err := pgx.CollectRows(zeilen, func(z pgx.CollectableRow) (meldung, error) { return scanneMeldung(z) })
	if liste == nil {
		liste = []meldung{}
	}
	return liste, err
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
	err = tx.QueryRow(ctx, `INSERT INTO meldungen
		(raum_id, ort, kategorie, was, notiz, foto, dringend, gemeldet_von, gemeldet_am, geaendert_am)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9) RETURNING id`,
		m.RaumID, m.Ort, m.Kategorie, m.Was, m.Notiz, fotoName, m.Dringend, m.GemeldetVon, m.GemeldetAm).Scan(&id)
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
func (s *speicher) setzeStatus(ctx context.Context, id int64, status, antwort *string, jetzt time.Time) (meldung, error) {
	tag, err := s.db.Exec(ctx, `UPDATE meldungen SET
			status = coalesce($2, status),
			antwort = coalesce($3, antwort),
			erledigt_am = CASE WHEN coalesce($2, status) = 'erledigt' THEN coalesce(erledigt_am, $4) END,
			geaendert_am = $4
		WHERE id = $1`, id, status, antwort, jetzt)
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

// ── Aufräumen ────────────────────────────────────────────────────────────────

// loescheAlteFotos entfernt die Fotos erledigter Meldungen; der Verweis in der Meldung fällt über den
// Fremdschlüssel (ON DELETE SET NULL) von selbst weg. Fotos ohne Meldung gehen gleich mit.
func (s *speicher) loescheAlteFotos(ctx context.Context, grenze, jetzt time.Time) (int64, error) {
	tag, err := s.db.Exec(ctx, `DELETE FROM fotos f WHERE f.name IN (
			SELECT foto FROM meldungen WHERE status = 'erledigt' AND erledigt_am < $1 AND foto IS NOT NULL)
		OR (NOT EXISTS (SELECT 1 FROM meldungen m WHERE m.foto = f.name) AND f.angelegt < $2)`,
		grenze, jetzt.Add(-24*time.Hour))
	return tag.RowsAffected(), err
}

// loescheAlteMeldungen entfernt erledigte Meldungen nach der Aufbewahrungsfrist samt Foto.
func (s *speicher) loescheAlteMeldungen(ctx context.Context, grenze time.Time) (int64, error) {
	var anzahl int64
	err := s.db.QueryRow(ctx, `WITH weg AS (
			DELETE FROM meldungen WHERE status = 'erledigt' AND erledigt_am < $1 RETURNING foto),
		bilder AS (DELETE FROM fotos WHERE name IN (SELECT foto FROM weg WHERE foto IS NOT NULL))
		SELECT count(*) FROM weg`, grenze).Scan(&anzahl)
	return anzahl, err
}

// ── Sitzungen ────────────────────────────────────────────────────────────────

func (s *speicher) legeSitzungAn(ctx context.Context, hash, email string, ablauf time.Time) error {
	_, err := s.db.Exec(ctx, "INSERT INTO sitzungen (token_hash, email, laeuft_ab) VALUES ($1, $2, $3)", hash, email, ablauf)
	return err
}

func (s *speicher) sitzung(ctx context.Context, hash string) (email string, ablauf time.Time, err error) {
	err = s.db.QueryRow(ctx, "SELECT email, laeuft_ab FROM sitzungen WHERE token_hash = $1", hash).Scan(&email, &ablauf)
	return email, ablauf, nichtGefunden(err)
}

func (s *speicher) verlaengereSitzung(ctx context.Context, hash string, ablauf time.Time) error {
	_, err := s.db.Exec(ctx, "UPDATE sitzungen SET laeuft_ab = $2 WHERE token_hash = $1", hash, ablauf)
	return err
}

func (s *speicher) loescheSitzung(ctx context.Context, hash string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM sitzungen WHERE token_hash = $1", hash)
	return err
}

func (s *speicher) loescheAbgelaufeneSitzungen(ctx context.Context, jetzt time.Time) error {
	_, err := s.db.Exec(ctx, "DELETE FROM sitzungen WHERE laeuft_ab < $1", jetzt)
	return err
}
