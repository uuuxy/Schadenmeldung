package main

import (
	"fmt"
	"net/http"
)

func (a *app) benachrichtigungLesen(w http.ResponseWriter, r *http.Request, b benutzer) {
	e, err := a.sp.mailEinstellung(r.Context(), b.Email, vorgabeFuer(b.Rolle))
	if err != nil {
		intern(w, err)
		return
	}
	zeiten := make([]string, 0, len(a.konf.SammelZeiten))
	for _, z := range a.konf.SammelZeiten {
		zeiten = append(zeiten, fmt.Sprintf("%02d:%02d", z.Stunde, z.Minute))
	}
	schreibeJSON(w, http.StatusOK, map[string]any{
		"mailAktiv": a.konf.mailAktiv(), "dringendSofort": e.DringendSofort, "sammelmail": e.Sammelmail, "zeiten": zeiten,
	})
}

func (a *app) benachrichtigungSetzen(w http.ResponseWriter, r *http.Request, b benutzer) {
	var ein struct {
		DringendSofort *bool `json:"dringendSofort"`
		Sammelmail     *bool `json:"sammelmail"`
	}
	if !leseJSON(w, r, &ein) {
		return
	}
	e, err := a.sp.mailEinstellung(r.Context(), b.Email, vorgabeFuer(b.Rolle))
	if err != nil {
		intern(w, err)
		return
	}
	if ein.DringendSofort != nil {
		e.DringendSofort = *ein.DringendSofort
	}
	if ein.Sammelmail != nil {
		e.Sammelmail = *ein.Sammelmail
	}
	if err := a.sp.setzeMailEinstellung(r.Context(), b.Email, e); err != nil {
		intern(w, err)
		return
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
	}
	if !leseJSON(w, r, &ein) {
		return
	}
	if statusNamen[ein.Status] == "" {
		fehler(w, http.StatusBadRequest, "Unbekannter Status.")
		return
	}
	neu, err := a.sp.setzeStatus(r.Context(), m.ID, &ein.Status, nil, a.jetzt())
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
	if m.Foto != "" {
		m.Foto = "/api/aktion/" + token + "/foto"
	}
	return m
}
