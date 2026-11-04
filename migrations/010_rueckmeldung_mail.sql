-- Rückmeldung per Mail: Wer einen Schaden meldet, bekommt eine Mail, sobald die Meldung erledigt ist,
-- höchstens eine je Meldung. rueckmeldung_am hält fest, wann sie hinausging, oder wann die meldende
-- Person die Meldung selbst erledigt hat und darum keine braucht.
ALTER TABLE meldungen ADD COLUMN rueckmeldung_am TIMESTAMPTZ;

-- Was schon erledigt ist, bekommt beim Einspielen keine Mail mehr.
UPDATE meldungen SET rueckmeldung_am = erledigt_am WHERE status = 'erledigt';
