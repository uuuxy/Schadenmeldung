-- Das Postfach, über das die App ihre Mails verschickt, wenn es nicht fest in der .env steht: unter
-- Einstellungen → Mail eingetragen. Höchstens eine Zeile; das Passwort nur verschlüsselt (AES-GCM,
-- der Schlüssel steht allein in der .env, also weder hier noch in den Sicherungen).
CREATE TABLE versand_postfach (
    eins        boolean PRIMARY KEY DEFAULT true CHECK (eins),
    adresse     text NOT NULL CHECK (adresse <> ''),
    passwort    bytea NOT NULL,
    eingetragen timestamptz NOT NULL
);
