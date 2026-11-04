package main

import (
	"context"
	"embed"
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
	// Kennung vom Handy gegen doppelt geschickte Meldungen; geht nie hinaus.
	Kennung string `json:"-"`
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

// alle fragt ab und sammelt die Zeilen; nie nil, damit die Oberfläche [] statt null bekommt.
func alle[T any](ctx context.Context, db *pgxpool.Pool, lies pgx.RowToFunc[T], sql string, args ...any) ([]T, error) {
	zeilen, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	liste, err := pgx.CollectRows(zeilen, lies)
	if liste == nil {
		liste = []T{}
	}
	return liste, err
}
