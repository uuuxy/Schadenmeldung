package main

import (
	"net/http"
)

func (a *app) benachrichtigungLesen(w http.ResponseWriter, r *http.Request, b benutzer) {
	an, err := a.sp.mailAn(r.Context(), b.Email, mailVorgabe(b.Rolle))
	if err != nil {
		intern(w, err)
		return
	}
	schreibeJSON(w, http.StatusOK, map[string]any{"mailAktiv": a.mailAktiv(r.Context()), "mail": an})
}

// benachrichtigungSetzen: Ohne Angabe bleibt alles, wie es ist; so stört ein alter, noch
// zwischengespeicherter Stand der Oberfläche nicht.
func (a *app) benachrichtigungSetzen(w http.ResponseWriter, r *http.Request, b benutzer) {
	var ein struct {
		Mail *bool `json:"mail"`
	}
	if !leseJSON(w, r, &ein) {
		return
	}
	if ein.Mail != nil {
		if err := a.sp.setzeMailAn(r.Context(), b.Email, *ein.Mail); err != nil {
			intern(w, err)
			return
		}
	}
	a.benachrichtigungLesen(w, r, b)
}

// aktionsLink liest den Link aus der Mail. Wer ihn hat, ist der Empfänger der Mail; seine Rolle
// zählt so, wie sie jetzt ist, nicht wie beim Verschicken.
func (a *app) aktionsLink(w http.ResponseWriter, r *http.Request) (meldung, benutzer, bool) {
	id, email, ablauf, err := a.sp.mailLink(r.Context(), tokenHash(r.PathValue("token")))
	if err != nil || a.jetzt().After(ablauf) {
		fehler(w, http.StatusGone, "Dieser Link ist abgelaufen. Die Meldung lässt sich in der App bearbeiten.")
		return meldung{}, benutzer{}, false
	}
	rolle, err := a.rolle(r.Context(), email)
	if err != nil {
		intern(w, err)
		return meldung{}, benutzer{}, false
	}
	b := benutzer{Email: email, Rolle: rolle}
	if !b.darfBearbeiten() {
		fehler(w, http.StatusForbidden, "Mit dieser Rolle lässt sich der Status nicht ändern.")
		return meldung{}, benutzer{}, false
	}
	m, err := a.sp.meldung(r.Context(), id)
	if err != nil {
		nichtGefundenOderIntern(w, err)
		return meldung{}, benutzer{}, false
	}
	return m, b, true
}

func (a *app) aktionLesen(w http.ResponseWriter, r *http.Request) {
	m, b, ok := a.aktionsLink(w, r)
	if !ok {
		return
	}
	schreibeJSON(w, http.StatusOK, a.fuerAktion(m, b, r.PathValue("token")))
}

func (a *app) aktionAusfuehren(w http.ResponseWriter, r *http.Request) {
	m, b, ok := a.aktionsLink(w, r)
	if !ok {
		return
	}
	var ein struct {
		Status string `json:"status"`
		Vorher string `json:"vorher"` // der Stand, den die Seite gezeigt hat
	}
	if !leseJSON(w, r, &ein) {
		return
	}
	// Nur was die Seite anbietet: Zurück auf „offen“ geht nur in der App, nicht mit dem Link aus einer
	// vielleicht weitergeleiteten Mail.
	if statusNamen[ein.Status] == "" || ein.Status == "offen" {
		fehler(w, http.StatusBadRequest, "Unbekannter Status.")
		return
	}
	// Zwei Hausmeister bekommen dieselbe Meldung. Hat einer erledigt, während die Seite des anderen noch
	// den alten Stand zeigte, öffnet dessen Tipp die Meldung nicht wieder; die Seite zeigt dann „Schon
	// erledigt“, und wer wirklich ändern will, tut es von dort bewusst. Ohne Vorher (alte Seite) wie bisher.
	if m.Status == "erledigt" && ein.Vorher != "" && ein.Vorher != "erledigt" && ein.Status != "erledigt" {
		fehler(w, http.StatusConflict, "Inzwischen hat jemand die Meldung erledigt.")
		return
	}
	neu, err := a.sp.setzeStatus(r.Context(), m.ID, aenderung{Status: &ein.Status}, b.Email, a.jetzt())
	if err != nil {
		nichtGefundenOderIntern(w, err)
		return
	}
	a.verteiler.melde()
	schreibeJSON(w, http.StatusOK, a.fuerAktion(neu, b, r.PathValue("token")))
}

func (a *app) aktionFoto(w http.ResponseWriter, r *http.Request) {
	m, _, ok := a.aktionsLink(w, r)
	if !ok || m.Foto == "" {
		if ok {
			http.NotFound(w, r)
		}
		return
	}
	a.liefereFoto(w, r, m.Foto)
}

func (a *app) fuerAktion(m meldung, b benutzer, token string) meldung {
	m = fuerAnsicht(m, b)
	// Wer den Link hat, braucht weder die meldende Person noch den internen Vermerk: Mails werden
	// weitergeleitet, etwa an eine Fachfirma.
	m.GemeldetVon, m.Antwort = "", ""
	if m.Foto != "" {
		m.Foto = "/api/aktion/" + token + "/foto"
	}
	return m
}
