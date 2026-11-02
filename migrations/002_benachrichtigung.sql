-- Benachrichtigung des Hausmeisters per Mail, mit Knöpfen direkt in der Mail.

-- „In Arbeit“ wird zu „Gesehen“ (zur Kenntnis genommen): Ein Antippen genügt, damit alle wissen,
-- dass der Hausmeister Bescheid weiß – ohne Zusage, wann es erledigt ist.
ALTER TABLE meldungen DROP CONSTRAINT meldungen_status_check;
UPDATE meldungen SET status = 'gesehen' WHERE status = 'in_arbeit';
ALTER TABLE meldungen ADD CONSTRAINT meldungen_status_check
	CHECK (status IN ('offen', 'gesehen', 'bestellt', 'fachfirma', 'erledigt'));

-- Die Knöpfe einer Mail: Wer den Link hat, darf den Status dieser einen Meldung setzen, 14 Tage lang.
-- Nur der Hash des Links steht hier, wie bei den Sitzungen.
CREATE TABLE mail_links (
	token_hash TEXT PRIMARY KEY,
	meldung_id BIGINT NOT NULL REFERENCES meldungen (id) ON DELETE CASCADE,
	empfaenger TEXT NOT NULL,
	laeuft_ab  TIMESTAMPTZ NOT NULL
);

-- Welche Meldung ist an wen schon gegangen: Keine Meldung kommt doppelt, und was sofort nicht
-- rausging, nimmt die nächste Sammelmail mit.
CREATE TABLE mail_versand (
	meldung_id  BIGINT NOT NULL REFERENCES meldungen (id) ON DELETE CASCADE,
	empfaenger  TEXT NOT NULL,
	gesendet_am TIMESTAMPTZ NOT NULL,
	PRIMARY KEY (meldung_id, empfaenger)
);

-- Was der Hausmeister bekommen möchte; ohne Zeile gilt beides an.
CREATE TABLE benachrichtigungen (
	email           TEXT PRIMARY KEY,
	dringend_sofort BOOLEAN NOT NULL DEFAULT true,
	sammelmail      BOOLEAN NOT NULL DEFAULT true
);

-- Kleine Merkzettel des Servers, etwa wann die letzte Sammelmail lief.
CREATE TABLE einstellungen (
	schluessel TEXT PRIMARY KEY,
	wert       TEXT NOT NULL
);

-- Wer mehr sieht als das Meldeformular: Hausmeister (bearbeitet), Leitung (sieht alles, ändert nichts)
-- und Verwaltung (alles, auch Räume und Personen). Jedes andere Postfach der Schule ist Kollegium und
-- meldet nur. Die Adressen aus ADMINS gelten zusätzlich immer als Verwaltung, damit sich niemand
-- aussperren kann.
CREATE TABLE rollen (
	email        TEXT PRIMARY KEY,
	rolle        TEXT NOT NULL CHECK (rolle IN ('hausmeister', 'leitung', 'admin')),
	angelegt_am  TIMESTAMPTZ NOT NULL DEFAULT now(),
	angelegt_von TEXT NOT NULL DEFAULT ''
);
