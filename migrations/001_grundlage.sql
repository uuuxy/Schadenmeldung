-- Grundlage: Räume, Fotos, Meldungen, Sitzungen.

CREATE TABLE raeume (
	id      BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	name    TEXT NOT NULL,
	bereich TEXT NOT NULL DEFAULT '',
	-- Zufälliger Code für den QR-Aufkleber; bleibt beim Umbenennen gleich.
	code    TEXT NOT NULL UNIQUE
);
-- „Aula“ und „aula“ sind derselbe Raum.
CREATE UNIQUE INDEX raeume_name ON raeume (lower(name));

-- Fotos liegen in der Datenbank: Eine Sicherung mit pg_dump nimmt alles mit, und der Container der
-- App hat keine eigenen Daten. Sie sind verkleinert (höchstens 1600 px) und ohne Metadaten.
CREATE TABLE fotos (
	name     TEXT PRIMARY KEY,
	daten    BYTEA NOT NULL,
	angelegt TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE meldungen (
	id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	-- Der Ort steht als Text in der Meldung, damit sie lesbar bleibt, wenn der Raum gelöscht wird.
	raum_id      BIGINT REFERENCES raeume (id) ON DELETE SET NULL,
	ort          TEXT NOT NULL,
	kategorie    TEXT NOT NULL,
	was          TEXT NOT NULL,
	notiz        TEXT NOT NULL DEFAULT '',
	foto         TEXT REFERENCES fotos (name) ON DELETE SET NULL,
	dringend     BOOLEAN NOT NULL DEFAULT false,
	status       TEXT NOT NULL DEFAULT 'offen'
		CHECK (status IN ('offen', 'in_arbeit', 'bestellt', 'fachfirma', 'erledigt')),
	antwort      TEXT NOT NULL DEFAULT '',
	gemeldet_von TEXT NOT NULL,
	gemeldet_am  TIMESTAMPTZ NOT NULL,
	geaendert_am TIMESTAMPTZ NOT NULL,
	erledigt_am  TIMESTAMPTZ,
	CHECK ((status = 'erledigt') = (erledigt_am IS NOT NULL))
);
CREATE INDEX meldungen_offen ON meldungen (gemeldet_am DESC) WHERE status <> 'erledigt';
CREATE INDEX meldungen_erledigt ON meldungen (erledigt_am DESC) WHERE status = 'erledigt';

-- Nur der Hash des Sitzungsschlüssels: Wer die Datenbank liest, kann sich damit nicht anmelden.
CREATE TABLE sitzungen (
	token_hash TEXT PRIMARY KEY,
	email      TEXT NOT NULL,
	laeuft_ab  TIMESTAMPTZ NOT NULL
);
