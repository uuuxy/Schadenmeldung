package main

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// Speicher für die Mitteilungen aufs Handy (push.go): die Abos je Gerät und das Schlüsselpaar des Servers.

type pushAbo struct {
	Endpunkt, Email, P256dh, Auth string
}

// pushZiel ist ein Abo samt der Zahl der Meldungen, die seit dem letzten Öffnen der App kamen.
type pushZiel struct {
	pushAbo
	Neu int
}

// geraeteJePerson: Mehr braucht niemand; bei einem weiteren Gerät fällt das am längsten nicht geöffnete weg.
const geraeteJePerson = 10

func (s *speicher) merkePushAbo(ctx context.Context, abo pushAbo, gesehen time.Time) error {
	_, err := s.db.Exec(ctx, `INSERT INTO push_abos (endpunkt, email, p256dh, auth, gesehen) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (endpunkt) DO UPDATE SET email = $2, p256dh = $3, auth = $4, gesehen = $5`,
		abo.Endpunkt, abo.Email, abo.P256dh, abo.Auth, gesehen)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, `DELETE FROM push_abos WHERE email = $1 AND endpunkt NOT IN
		(SELECT endpunkt FROM push_abos WHERE email = $1 ORDER BY gesehen DESC LIMIT $2)`, abo.Email, geraeteJePerson)
	return err
}

// pushZiele: die Abos dieser Personen und je Gerät die Zahl der neuen Meldungen seit dem letzten Öffnen;
// eigene zählen nicht. Die gerade gemeldete zählt immer mit, auch wenn sie aus dem Funkloch kommt und die
// Zeit trägt, zu der sie getippt wurde: Sonst käme ihre Mitteilung ohne Zahl.
func (s *speicher) pushZiele(ctx context.Context, emails []string, meldungID int64) ([]pushZiel, error) {
	return alle(ctx, s.db, func(row pgx.CollectableRow) (pushZiel, error) {
		var z pushZiel
		err := row.Scan(&z.Endpunkt, &z.Email, &z.P256dh, &z.Auth, &z.Neu)
		return z, err
	}, `SELECT a.endpunkt, a.email, a.p256dh, a.auth,
			(SELECT count(*) FROM meldungen m WHERE (m.gemeldet_am > a.gesehen OR m.id = $2) AND m.gemeldet_von <> a.email)
		FROM push_abos a WHERE a.email = ANY($1)`, emails, meldungID)
}

// loeschePushAbo löscht ein Abo; mit email nur, wenn es dieser Person gehört.
func (s *speicher) loeschePushAbo(ctx context.Context, endpunkt, email string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM push_abos WHERE endpunkt = $1 AND ($2 = '' OR email = $2)", endpunkt, email)
	return err
}

func (s *speicher) loeschePushAbosVon(ctx context.Context, email string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM push_abos WHERE email = $1", email)
	return err
}

// loescheAltePushAbos: was seit grenze nicht offen war oder wem die Rolle fehlt.
func (s *speicher) loescheAltePushAbos(ctx context.Context, grenze time.Time, backend []string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM push_abos WHERE gesehen < $1 OR NOT (email = ANY($2))", grenze, backend)
	return err
}

// pushSchluessel liest das Schlüsselpaar des Servers oder legt es mit neu an. Starten zwei Server
// gleichzeitig, gilt für beide der zuerst gespeicherte Schlüssel.
func (s *speicher) pushSchluessel(ctx context.Context, neu func() ([]byte, error)) ([]byte, error) {
	var der []byte
	err := s.db.QueryRow(ctx, "SELECT privat FROM push_schluessel WHERE id = 1").Scan(&der)
	if !errors.Is(err, pgx.ErrNoRows) {
		return der, err
	}
	if der, err = neu(); err != nil {
		return nil, err
	}
	if _, err := s.db.Exec(ctx, "INSERT INTO push_schluessel (id, privat) VALUES (1, $1) ON CONFLICT DO NOTHING", der); err != nil {
		return nil, err
	}
	err = s.db.QueryRow(ctx, "SELECT privat FROM push_schluessel WHERE id = 1").Scan(&der)
	return der, err
}
