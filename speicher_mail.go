package main

import (
	"context"
	"errors"
	"time"
)

func (s *speicher) mailEinstellung(ctx context.Context, email string, vorgabe mailEinstellung) (mailEinstellung, error) {
	e := vorgabe
	err := s.db.QueryRow(ctx, "SELECT dringend_sofort, sammelmail FROM benachrichtigungen WHERE email = $1", email).
		Scan(&e.DringendSofort, &e.Sammelmail)
	if errors.Is(nichtGefunden(err), errNichtGefunden) {
		return vorgabe, nil
	}
	return e, err
}

func (s *speicher) setzeMailEinstellung(ctx context.Context, email string, e mailEinstellung) error {
	_, err := s.db.Exec(ctx, `INSERT INTO benachrichtigungen (email, dringend_sofort, sammelmail) VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE SET dringend_sofort = excluded.dringend_sofort, sammelmail = excluded.sammelmail`,
		email, e.DringendSofort, e.Sammelmail)
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

func (s *speicher) vermerkeVersand(ctx context.Context, meldungID int64, email string, zeit time.Time) error {
	_, err := s.db.Exec(ctx, `INSERT INTO mail_versand (meldung_id, empfaenger, gesendet_am) VALUES ($1, $2, $3)
		ON CONFLICT DO NOTHING`, meldungID, email, zeit)
	return err
}

// ungesendeteFuer: offene Meldungen bis zum Termin, die an diese Person noch nicht gingen und die
// sie nicht selbst gemeldet hat.
func (s *speicher) ungesendeteFuer(ctx context.Context, email string, bis time.Time) ([]meldung, error) {
	return s.abfrage(ctx, `WHERE m.status = 'offen' AND m.gemeldet_am <= $1 AND m.gemeldet_von <> $2
		AND NOT EXISTS (SELECT 1 FROM mail_versand v WHERE v.meldung_id = m.id AND v.empfaenger = $2)
		ORDER BY m.dringend DESC, m.gemeldet_am`, bis, email)
}

func (s *speicher) anzahlOffen(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRow(ctx, "SELECT count(*) FROM meldungen WHERE status <> 'erledigt'").Scan(&n)
	return n, err
}

func (s *speicher) wert(ctx context.Context, schluessel string) (string, error) {
	var w string
	err := s.db.QueryRow(ctx, "SELECT wert FROM einstellungen WHERE schluessel = $1", schluessel).Scan(&w)
	return w, nichtGefunden(err)
}

func (s *speicher) setzeWert(ctx context.Context, schluessel, wert string) error {
	_, err := s.db.Exec(ctx, `INSERT INTO einstellungen (schluessel, wert) VALUES ($1, $2)
		ON CONFLICT (schluessel) DO UPDATE SET wert = excluded.wert`, schluessel, wert)
	return err
}
