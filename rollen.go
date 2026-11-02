package main

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strings"
)

// Rollen: Wer mehr sieht als das Meldeformular. Ohne Eintrag ist jedes Postfach der Schule Kollegium
// und meldet nur. Die Verwaltung vergibt die Rollen in der App; die Adressen aus ADMINS sind immer
// Verwaltung, damit sich niemand aussperren kann.
const (
	rolleKollegium   = "kollegium"
	rolleHausmeister = "hausmeister"
	rolleLeitung     = "leitung"
	rolleAdmin       = "admin"
)

var vergebbareRollen = map[string]bool{rolleHausmeister: true, rolleLeitung: true, rolleAdmin: true}

type rollenEintrag struct {
	Email string `json:"email"`
	Rolle string `json:"rolle"`
	// Fest: aus ADMINS, in der App nicht änderbar.
	Fest bool `json:"fest"`
}

func (a *app) rolle(ctx context.Context, email string) (string, error) {
	if a.konf.Admins[email] {
		return rolleAdmin, nil
	}
	r, err := a.sp.rolle(ctx, email)
	if errors.Is(err, errNichtGefunden) {
		return rolleKollegium, nil
	}
	return r, err
}

// backendPersonen sind alle mit einer Rolle über dem Kollegium, samt der festen Verwaltung.
func (a *app) backendPersonen(ctx context.Context) ([]rollenEintrag, error) {
	eintraege, err := a.sp.rollen(ctx)
	if err != nil {
		return nil, err
	}
	for email := range a.konf.Admins {
		eintraege = append(eintraege, rollenEintrag{Email: email, Rolle: rolleAdmin, Fest: true})
	}
	// Feste Einträge verdrängen einen gleichlautenden aus der Datenbank.
	nachAdresse := map[string]rollenEintrag{}
	for _, e := range eintraege {
		if alt, da := nachAdresse[e.Email]; !da || e.Fest || !alt.Fest {
			nachAdresse[e.Email] = e
		}
	}
	liste := make([]rollenEintrag, 0, len(nachAdresse))
	for _, e := range nachAdresse {
		liste = append(liste, e)
	}
	sort.Slice(liste, func(i, j int) bool {
		if liste[i].Rolle != liste[j].Rolle {
			return rollenRang(liste[i].Rolle) < rollenRang(liste[j].Rolle)
		}
		return liste[i].Email < liste[j].Email
	})
	return liste, nil
}

func rollenRang(r string) int {
	return map[string]int{rolleHausmeister: 0, rolleLeitung: 1, rolleAdmin: 2}[r]
}

// ── Türen ────────────────────────────────────────────────────────────────────

func (a *app) rollenListe(w http.ResponseWriter, r *http.Request, _ benutzer) {
	liste, err := a.backendPersonen(r.Context())
	if err != nil {
		intern(w, err)
		return
	}
	schreibeJSON(w, http.StatusOK, liste)
}

func (a *app) rolleSetzen(w http.ResponseWriter, r *http.Request, b benutzer) {
	var ein struct {
		Email string `json:"email"`
		Rolle string `json:"rolle"`
	}
	if !leseJSON(w, r, &ein) {
		return
	}
	email := strings.ToLower(a.konf.vervollstaendige(ein.Email))
	switch {
	case !strings.Contains(email, "@") || len(email) > 200:
		fehler(w, http.StatusBadRequest, "Bitte eine E-Mail-Adresse der Schule eingeben.")
		return
	case !a.konf.domainErlaubt(email):
		fehler(w, http.StatusBadRequest, "Nur Postfächer der Schule können eine Rolle bekommen.")
		return
	case !vergebbareRollen[ein.Rolle]:
		fehler(w, http.StatusBadRequest, "Unbekannte Rolle.")
		return
	case a.konf.Admins[email]:
		fehler(w, http.StatusBadRequest, "Diese Adresse ist fest als Verwaltung eingetragen (ADMINS).")
		return
	case email == b.Email:
		// Sonst nimmt sich jemand aus Versehen die eigenen Rechte.
		fehler(w, http.StatusBadRequest, "Die eigene Rolle ändert eine andere Person der Verwaltung.")
		return
	}
	if err := a.sp.setzeRolle(r.Context(), email, ein.Rolle, b.Email); err != nil {
		intern(w, err)
		return
	}
	a.rollenListe(w, r, b)
}

func (a *app) rolleLoeschen(w http.ResponseWriter, r *http.Request, b benutzer) {
	email := strings.ToLower(r.PathValue("email"))
	switch {
	case a.konf.Admins[email]:
		fehler(w, http.StatusBadRequest, "Diese Adresse ist fest als Verwaltung eingetragen (ADMINS).")
		return
	case email == b.Email:
		fehler(w, http.StatusBadRequest, "Die eigene Rolle ändert eine andere Person der Verwaltung.")
		return
	}
	if err := a.sp.loescheRolle(r.Context(), email); err != nil {
		nichtGefundenOderIntern(w, err)
		return
	}
	a.rollenListe(w, r, b)
}

// ── Speicher ─────────────────────────────────────────────────────────────────

func (s *speicher) rolle(ctx context.Context, email string) (string, error) {
	var r string
	err := s.db.QueryRow(ctx, "SELECT rolle FROM rollen WHERE email = $1", email).Scan(&r)
	return r, nichtGefunden(err)
}

func (s *speicher) rollen(ctx context.Context) ([]rollenEintrag, error) {
	zeilen, err := s.db.Query(ctx, "SELECT email, rolle FROM rollen")
	if err != nil {
		return nil, err
	}
	defer zeilen.Close()
	var liste []rollenEintrag
	for zeilen.Next() {
		var e rollenEintrag
		if err := zeilen.Scan(&e.Email, &e.Rolle); err != nil {
			return nil, err
		}
		liste = append(liste, e)
	}
	return liste, zeilen.Err()
}

func (s *speicher) setzeRolle(ctx context.Context, email, rolle, von string) error {
	_, err := s.db.Exec(ctx, `INSERT INTO rollen (email, rolle, angelegt_von) VALUES ($1, $2, $3)
		ON CONFLICT (email) DO UPDATE SET rolle = excluded.rolle, angelegt_von = excluded.angelegt_von`, email, rolle, von)
	return err
}

func (s *speicher) loescheRolle(ctx context.Context, email string) error {
	tag, err := s.db.Exec(ctx, "DELETE FROM rollen WHERE email = $1", email)
	if err == nil && tag.RowsAffected() == 0 {
		return errNichtGefunden
	}
	return err
}
