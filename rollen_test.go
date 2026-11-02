package main

import (
	"net/http"
	"strconv"
	"testing"
)

func TestRaeumeNurDurchVerwaltung(t *testing.T) {
	u := neueTestUmgebung(t)
	admin, lehrer := u.nutzer("admin"), u.nutzer("lehrer")
	text := "204; Hauptgebäude · 2. OG\nTurnhalle\tSport\n\n204; doppelt\n"
	if status, _ := lehrer.json("POST", "/api/raeume", map[string]string{"text": text}); status != http.StatusForbidden {
		t.Fatalf("Kollegium legt Räume an: Status %d", status)
	}
	status, antwort := admin.json("POST", "/api/raeume", map[string]string{"text": text})
	if status != 200 || antwort["neu"].(float64) != 2 {
		t.Fatalf("Anlegen: %d, %v", status, antwort)
	}
	raeume := eintraege(antwort, "raeume")
	if raeume[0]["name"] != "204" || raeume[0]["bereich"] != "Hauptgebäude · 2. OG" || len(raeume[0]["code"].(string)) != 8 {
		t.Fatalf("Räume: %v", raeume)
	}

	// Eine Meldung behält ihren Ort, auch wenn der Raum später gelöscht wird.
	raumID := int(raeume[0]["id"].(float64))
	if status, m := lehrer.melde(map[string]string{"raumId": strconv.Itoa(raumID), "kategorie": "licht", "was": "Lampe defekt"}, nil); status != 201 || m["ort"] != "204" || m["bereich"] != "Hauptgebäude · 2. OG" {
		t.Fatalf("Meldung mit Raum: %d, %v", status, m)
	}
	if status, _ := admin.json("DELETE", "/api/raeume/"+strconv.Itoa(raumID), nil); status != http.StatusNoContent {
		t.Fatalf("Löschen: Status %d", status)
	}
	_, alle := lehrer.json("GET", "/api/meldungen", nil)
	if offen := eintraege(alle, "offen"); len(offen) != 1 || offen[0]["ort"] != "204" || offen[0]["raumId"] != nil {
		t.Fatalf("nach Löschen des Raums: %v", offen)
	}
}

func TestRollenInDerApp(t *testing.T) {
	u := neueTestUmgebung(t)
	admin, lehrer := u.nutzer("admin"), u.nutzer("lehrer")
	lehrer.melde(map[string]string{"ort": "Aula", "kategorie": "licht", "was": "Lampe defekt"}, nil)

	if status, _ := lehrer.json("POST", "/api/rollen", map[string]string{"email": "lehrer", "rolle": "admin"}); status != http.StatusForbidden {
		t.Fatalf("Kollegium vergibt Rollen: Status %d", status)
	}
	status, antwort := admin.json("POST", "/api/rollen", map[string]string{"email": "direktorin", "rolle": "leitung"})
	if status != 200 || len(eintraege(antwort, "liste")) != 3 {
		t.Fatalf("Rolle vergeben: %d, %v", status, antwort)
	}

	leitung := u.nutzer("direktorin")
	_, ich := leitung.json("GET", "/api/ich", nil)
	_, alle := leitung.json("GET", "/api/meldungen", nil)
	offen := eintraege(alle, "offen")
	if ich["rolle"] != "leitung" || len(offen) != 1 || offen[0]["gemeldetVon"] != "lehrer@schule.de" {
		t.Fatalf("Leitung sieht: %v, %v", ich, offen)
	}
	id := strconv.Itoa(int(offen[0]["id"].(float64)))
	if status, _ := leitung.json("POST", "/api/meldungen/"+id+"/status", map[string]string{"status": "erledigt"}); status != http.StatusForbidden {
		t.Fatalf("Leitung ändert Status: %d, erwartet 403 (nur Einsicht)", status)
	}

	if status, _ := admin.json("POST", "/api/rollen", map[string]string{"email": "admin", "rolle": "leitung"}); status != http.StatusBadRequest {
		t.Fatalf("feste Verwaltung umgestuft: Status %d", status)
	}
	admin.json("POST", "/api/rollen", map[string]string{"email": "zweit", "rolle": "admin"})
	if status, _ := u.nutzer("zweit").json("DELETE", "/api/rollen/zweit@schule.de", nil); status != http.StatusBadRequest {
		t.Fatalf("eigene Rolle gelöscht: Status %d, erwartet 400", status)
	}
	if status, _ := admin.json("DELETE", "/api/rollen/direktorin@schule.de", nil); status != 200 {
		t.Fatalf("Rolle löschen: %d", status)
	}
	if _, ich := leitung.json("GET", "/api/ich", nil); ich["rolle"] != "kollegium" {
		t.Fatalf("nach dem Löschen: %v, erwartet kollegium", ich)
	}
}
