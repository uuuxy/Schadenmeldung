package main

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// mailAn sagt, ob die Person Mail bekommt; ohne eigene Wahl gilt die Vorgabe ihrer Rolle.
func (s *speicher) mailAn(ctx context.Context, email string, vorgabe bool) (bool, error) {
	an := vorgabe
	err := s.db.QueryRow(ctx, "SELECT mail FROM benachrichtigungen WHERE email = $1", email).Scan(&an)
	if errors.Is(nichtGefunden(err), errNichtGefunden) {
		return vorgabe, nil
	}
	return an, err
}

func (s *speicher) setzeMailAn(ctx context.Context, email string, an bool) error {
	_, err := s.db.Exec(ctx, `INSERT INTO benachrichtigungen (email, mail) VALUES ($1, $2)
		ON CONFLICT (email) DO UPDATE SET mail = excluded.mail`, email, an)
	return err
}

func (s *speicher) legeMailLinkAn(ctx context.Context, hash string, meldungID int64, email string, ablauf time.Time) error {
	_, err := s.db.Exec(ctx, "INSERT INTO mail_links (token_hash, meldung_id, empfaenger, laeuft_ab) VALUES ($1, $2, $3, $4)",
		hash, meldungID, email, ablauf)
	return err
}

func (s *speicher) mailLink(ctx context.Context, hash string) (id int64, email string, ablauf time.Time, err error) {
	err = s.db.QueryRow(ctx, "SELECT meldung_id, empfaenger, laeuft_ab FROM mail_links WHERE token_hash = $1", hash).
		Scan(&id, &email, &ablauf)
	return id, email, ablauf, nichtGefunden(err)
}

func (s *speicher) loescheAbgelaufeneMailLinks(ctx context.Context, jetzt time.Time) error {
	_, err := s.db.Exec(ctx, "DELETE FROM mail_links WHERE laeuft_ab < $1", jetzt)
	return err
}

// merkeMailVor legt die Zeile im Postausgang an; gesendet_am bleibt leer, bis die Mail hinaus ist.
func (s *speicher) merkeMailVor(ctx context.Context, meldungID int64, email string) error {
	_, err := s.db.Exec(ctx, "INSERT INTO mail_versand (meldung_id, empfaenger) VALUES ($1, $2) ON CONFLICT DO NOTHING",
		meldungID, email)
	return err
}

func (s *speicher) vermerkeVersand(ctx context.Context, meldungID int64, email string, zeit time.Time) error {
	_, err := s.db.Exec(ctx, "UPDATE mail_versand SET gesendet_am = $3 WHERE meldung_id = $1 AND empfaenger = $2",
		meldungID, email, zeit)
	return err
}

// meldungenMitAusstehenderMail: offene Meldungen ab seit, für die noch eine Mail aussteht. Was
// inzwischen bearbeitet ist, braucht keine Mail mehr.
func (s *speicher) meldungenMitAusstehenderMail(ctx context.Context, seit time.Time) ([]meldung, error) {
	return s.abfrage(ctx, `WHERE m.status = 'offen' AND m.gemeldet_am >= $1
		AND EXISTS (SELECT 1 FROM mail_versand v WHERE v.meldung_id = m.id AND v.gesendet_am IS NULL)
		ORDER BY m.gemeldet_am, m.id`, seit)
}

// ausstehendeEmpfaenger: an wen die Mail zu dieser Meldung noch hinaus muss.
func (s *speicher) ausstehendeEmpfaenger(ctx context.Context, meldungID int64) ([]string, error) {
	zeilen, err := s.db.Query(ctx, `SELECT empfaenger FROM mail_versand
		WHERE meldung_id = $1 AND gesendet_am IS NULL ORDER BY empfaenger`, meldungID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(zeilen, pgx.RowTo[string])
}

func (s *speicher) anzahlOffen(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRow(ctx, "SELECT count(*) FROM meldungen WHERE status <> 'erledigt'").Scan(&n)
	return n, err
}
