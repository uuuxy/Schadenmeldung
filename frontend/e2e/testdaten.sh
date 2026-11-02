#!/bin/bash
# Testdaten für die lokale App (Mock-Login, Domain schule.de, ADMINS=verwaltung@schule.de):
# Rollen, Räume, Meldungen, Stati. Erwartet eine frische Datenbank und Mailpit.
HIER=$(cd "$(dirname "$0")" && pwd); SP=$(mktemp -d); B=http://127.0.0.1:8090; H='X-Schaden: 1'; J='Content-Type: application/json'
for i in $(seq 1 40); do curl -sf $B/healthz >/dev/null && break; sleep 0.3; done
login() { curl -s -c $SP/$1.txt -H "$H" -H "$J" -d "{\"email\":\"$2\",\"passwort\":\"x\"}" $B/api/anmelden; echo; }
login admin verwaltung
curl -s -b $SP/admin.txt -H "$H" -H "$J" -d '{"email":"hausmeister","rolle":"hausmeister"}' $B/api/rollen >/dev/null
curl -s -b $SP/admin.txt -H "$H" -H "$J" -d '{"email":"direktorin","rolle":"leitung"}' $B/api/rollen | head -c 200; echo
login lehrer m.mustermann; login lehrerin s.schmidt; login hm hausmeister; login leitung direktorin
curl -s -b $SP/admin.txt -H "$H" -H "$J" -d '{"text":"101; Hauptgebäude · EG\n102; Hauptgebäude · EG\nAula; Hauptgebäude · EG\nLehrerzimmer; Hauptgebäude · 1. OG\nWC Mädchen 1. OG; Hauptgebäude · 1. OG\n204; Hauptgebäude · 2. OG\n205; Hauptgebäude · 2. OG\nBio 1; Neubau · Naturwissenschaften\nChemie 2; Neubau · Naturwissenschaften\nTurnhalle; Sport"}' $B/api/raeume | head -c 40; echo
id() { curl -s -b $SP/admin.txt $B/api/raeume | python3 -c "import json,sys; print([r['id'] for r in json.load(sys.stdin) if r['name']==sys.argv[1]][0])" "$1"; }
m() { curl -s -b $SP/$1.txt -H "$H" "${@:2}" $B/api/meldungen | head -c 60; echo; }
m lehrer -F raumId=$(id 204) -F kategorie=technik -F 'was=Beamer / Display aus' -F 'notiz=Bleibt schwarz, auch nach Neustart.'
m lehrerin -F raumId=$(id 'WC Mädchen 1. OG') -F kategorie=wasser -F 'was=WC verstopft' -F dringend=1
m lehrer -F raumId=$(id 'Bio 1') -F kategorie=licht -F 'was=Licht flackert' -F foto=@$HIER/foto.jpg
m lehrerin -F raumId=$(id Turnhalle) -F kategorie=fenster -F 'was=Schloss defekt' -F 'notiz=Geräteraum lässt sich nicht abschließen.'
m lehrer -F raumId=$(id Lehrerzimmer) -F kategorie=moebel -F 'was=Tisch wackelt'
m lehrerin -F raumId=$(id 101) -F kategorie=heizung -F 'was=Zu kalt'
st() { curl -s -b $SP/hm.txt -H "$H" -H "$J" -d "$2" $B/api/meldungen/$1/status | head -c 50; echo; }
st 3 '{"status":"bestellt","antwort":"Neue Röhre ist bestellt, kommt Freitag."}'
st 4 '{"status":"fachfirma"}'
st 1 '{"status":"gesehen"}'
st 6 '{"status":"erledigt","antwort":"Ventil war zu – geht wieder."}'
