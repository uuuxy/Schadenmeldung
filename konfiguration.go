package main

import (
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
	// Die Zeitzonen stecken im Programm: Das schlanke Docker-Abbild bringt keine mit.
	_ "time/tzdata"
)

// konfiguration hält alles, was der Betrieb über Umgebungsvariablen einstellt.
type konfiguration struct {
	Port             string
	DatenbankURL     string
	Lokal            bool
	IMAPHost         string
	IMAPPort         string
	Domains          []string
	Aliase           []string // Kurzdomains derselben Postfächer, etwa kurz.example (siehe vervollstaendige)
	Admins           map[string]bool
	Schulname        string
	ProxyVertrauen   bool
	FotoTage         int
	AufbewahrungTage int

	// Mail an den Hausmeister; ohne SMTP_HOST ist sie aus.
	SMTPHost            string
	SMTPPort            string
	SMTPBenutzer        string
	SMTPPasswort        string
	MailAbsender        string
	OeffentlicheAdresse string
	SammelZeiten        []uhrzeit
	Zeitzone            *time.Location
}

// uhrzeit ist eine Tageszeit für die Sammelmail, etwa 07:00.
type uhrzeit struct{ Stunde, Minute int }

func ladeKonfiguration() (konfiguration, error) {
	k := konfiguration{
		Port:           wert("PORT", "8090"),
		DatenbankURL:   strings.TrimSpace(os.Getenv("DATABASE_URL")),
		Lokal:          istLokal(os.Getenv("APP_ENV")),
		IMAPHost:       strings.TrimSpace(os.Getenv("IMAP_HOST")),
		IMAPPort:       wert("IMAP_PORT", "993"),
		Domains:        liste(os.Getenv("ANMELDE_DOMAINS")),
		Aliase:         liste(os.Getenv("ALIAS_DOMAINS")),
		Schulname:      wert("SCHULNAME", ""),
		ProxyVertrauen: os.Getenv("PROXY_VERTRAUEN") == "1",
	}
	// ADMINS geht durch dieselbe Ergänzung wie die Anmeldung: Wer sich als Name oder mit der
	// Kurzdomain einträgt, wird sonst nie erkannt und sperrt sich aus.
	k.Admins = map[string]bool{}
	for _, adresse := range liste(os.Getenv("ADMINS")) {
		k.Admins[k.vervollstaendige(adresse)] = true
	}
	var err error
	if k.FotoTage, err = zahl("FOTO_TAGE", 30); err != nil {
		return k, err
	}
	if k.AufbewahrungTage, err = zahl("AUFBEWAHRUNG_TAGE", 365); err != nil {
		return k, err
	}
	k.SMTPHost = strings.TrimSpace(os.Getenv("SMTP_HOST"))
	k.SMTPPort = wert("SMTP_PORT", "587")
	k.SMTPBenutzer = strings.TrimSpace(os.Getenv("SMTP_BENUTZER"))
	k.SMTPPasswort = os.Getenv("SMTP_PASSWORT")
	k.MailAbsender = strings.TrimSpace(os.Getenv("MAIL_ABSENDER"))
	k.OeffentlicheAdresse = strings.TrimRight(strings.TrimSpace(os.Getenv("OEFFENTLICHE_ADRESSE")), "/")
	if k.SammelZeiten, err = leseUhrzeiten(wert("SAMMELMAIL_ZEITEN", "07:00,12:00")); err != nil {
		return k, err
	}
	if k.Zeitzone, err = time.LoadLocation(wert("ZEITZONE", "Europe/Berlin")); err != nil {
		return k, fmt.Errorf("ZEITZONE: %w", err)
	}
	return k, k.pruefe()
}

// pruefe lehnt Einstellungen ab, die die Anmeldung unbrauchbar oder zur offenen Tür machen.
func (k konfiguration) pruefe() error {
	if k.DatenbankURL == "" {
		return errors.New("DATABASE_URL fehlt, z. B. postgres://schaden:geheim@db:5432/schaden?sslmode=disable")
	}
	if k.IMAPHost == "" {
		return errors.New("IMAP_HOST fehlt: ohne Mailserver ist keine Anmeldung möglich (lokal: IMAP_HOST=mock mit APP_ENV=local)")
	}
	if k.IMAPHost == "mock" && !k.Lokal {
		return errors.New("IMAP_HOST=mock nimmt jedes Passwort an und gilt nur zusammen mit APP_ENV=local")
	}
	if len(k.Aliase) > 0 && len(k.Domains) != 1 {
		return errors.New("ALIAS_DOMAINS braucht genau eine Domain in ANMELDE_DOMAINS: auf sie werden die Kurzadressen umgeschrieben")
	}
	if k.SMTPHost != "" && (k.MailAbsender == "" || k.OeffentlicheAdresse == "") {
		return errors.New("mit SMTP_HOST braucht es auch MAIL_ABSENDER und OEFFENTLICHE_ADRESSE (für die Knöpfe in der Mail)")
	}
	return nil
}

func (k konfiguration) mailAktiv() bool { return k.SMTPHost != "" }

// leseUhrzeiten liest „07:00,12:00“. Eine leere Angabe schaltet die Sammelmail ab.
func leseUhrzeiten(roh string) ([]uhrzeit, error) {
	var zeiten []uhrzeit
	for _, teil := range strings.Split(roh, ",") {
		teil = strings.TrimSpace(teil)
		if teil == "" || teil == "aus" {
			continue
		}
		t, err := time.Parse("15:04", teil)
		if err != nil {
			return nil, fmt.Errorf("SAMMELMAIL_ZEITEN: %q ist keine Uhrzeit wie 07:00", teil)
		}
		zeiten = append(zeiten, uhrzeit{t.Hour(), t.Minute()})
	}
	return zeiten, nil
}

// vervollstaendige macht aus der Eingabe die volle Adresse, die der Mailserver als Benutzernamen
// kennt: Bei genau einer Schuldomain genügt der Name vor dem @, und eine Kurzdomain aus
// ALIAS_DOMAINS wird ersetzt. Das spart Tipparbeit, und die gewohnte kurze Adresse führt trotzdem hinein.
func (k konfiguration) vervollstaendige(email string) string {
	email = strings.TrimSpace(email)
	if email == "" || len(k.Domains) != 1 {
		return email
	}
	at := strings.LastIndex(email, "@")
	switch {
	case at < 0:
		return email + "@" + k.Domains[0]
	case at > 0 && slices.Contains(k.Aliase, strings.ToLower(email[at+1:])):
		return email[:at+1] + k.Domains[0]
	}
	return email
}

// domainErlaubt prüft die Maildomain. Ohne ANMELDE_DOMAINS entscheidet allein der Mailserver.
func (k konfiguration) domainErlaubt(email string) bool {
	if len(k.Domains) == 0 {
		return true
	}
	at := strings.LastIndex(email, "@")
	if at < 0 {
		return false
	}
	domain := email[at+1:]
	for _, d := range k.Domains {
		if domain == d {
			return true
		}
	}
	return false
}

func istLokal(appEnv string) bool {
	switch strings.ToLower(strings.TrimSpace(appEnv)) {
	case "local", "development", "test":
		return true
	default:
		return false
	}
}

func wert(name, vorgabe string) string {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		return v
	}
	return vorgabe
}

func zahl(name string, vorgabe int) (int, error) {
	v := strings.TrimSpace(os.Getenv(name))
	if v == "" {
		return vorgabe, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return 0, fmt.Errorf("%s muss eine ganze Zahl ab 1 sein, ist aber %q", name, v)
	}
	return n, nil
}

// liste zerlegt eine durch Komma getrennte Angabe und schreibt alles klein, weil Mailadressen und
// Domains ohne Rücksicht auf Groß- und Kleinschreibung verglichen werden.
func liste(roh string) []string {
	var aus []string
	for _, teil := range strings.Split(roh, ",") {
		if t := strings.ToLower(strings.TrimSpace(teil)); t != "" {
			aus = append(aus, t)
		}
	}
	return aus
}
