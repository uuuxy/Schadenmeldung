-- Mitteilungen aufs Handy (Web Push, push.go): je Gerät ein Abo mit der Adresse beim Push-Dienst des
-- Browsers und den Schlüsseln, mit denen der Server den Inhalt nur für dieses Gerät verschlüsselt.
-- gesehen: wann die App auf dem Gerät zuletzt offen war. Was danach gemeldet wurde, zählt die Zahl am
-- App-Symbol; ein Abo, das 180 Tage nicht offen war, löscht aufraeumen.go.
CREATE TABLE push_abos (
	endpunkt TEXT PRIMARY KEY CHECK (endpunkt LIKE 'https://%'),
	email    TEXT NOT NULL,
	p256dh   TEXT NOT NULL,
	auth     TEXT NOT NULL,
	gesehen  TIMESTAMPTZ NOT NULL
);
CREATE INDEX push_abos_email ON push_abos (email);

-- Das Schlüsselpaar des Servers gegenüber den Push-Diensten (VAPID); es entsteht beim ersten Gebrauch.
CREATE TABLE push_schluessel (
	id     INT PRIMARY KEY CHECK (id = 1),
	privat BYTEA NOT NULL
);
