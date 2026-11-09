import { chromium } from 'playwright-core';
// Ablauf im Handyformat: QR → Anmeldung → Raum erkannt, nur Formular fürs Kollegium, Rollen,
// Mail-Knopf (braucht Mailpit). Bildschirmfotos landen im Ordner aus argv[2], sonst im Temp-Ordner.
const ziel = process.argv[2] ?? (await import('node:os')).tmpdir();
const B = 'http://127.0.0.1:8090';
const browser = await chromium.launch();
const fehler = [];
const pruefe = (bed, text) => { console.log((bed ? '✓ ' : '✗ ') + text); if (!bed) fehler.push('Prüfung: ' + text); };
async function kontext(name, schema = 'light') {
	const k = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, isMobile: true, hasTouch: true, colorScheme: schema, locale: 'de-DE' });
	if (name) await k.request.post(`${B}/api/anmelden`, { headers: { 'X-Schaden': '1' }, data: { email: name, passwort: 'x' } });
	const s = await k.newPage();
	s.on('pageerror', (e) => fehler.push(String(e)));
	// Erwartete Absagen: 401 abgemeldet, 409 überholte Seite hinter dem Mail-Knopf, 410 alter Link.
	s.on('console', (m) => m.type() === 'error' && !/401|409|410/.test(m.text()) && fehler.push(m.text()));
	return [k, s];
}
const bild = async (s, d, ganz = false) => { await s.waitForTimeout(600); await s.screenshot({ path: `${ziel}/${d}.png`, fullPage: ganz }); };

// 1. QR-Code ohne Anmeldung → anmelden → Raum ist erkannt
let [k, s] = await kontext(null);
const raeume = await (await browser.newContext()).request.get(B + '/api/raeume').catch(() => null);
const admin = await browser.newContext(); await admin.request.post(`${B}/api/anmelden`, { headers: { 'X-Schaden': '1' }, data: { email: 'admin', passwort: 'x' } });
const liste = await (await admin.request.get(B + '/api/raeume')).json();
const r204 = liste.find((r) => r.name === '204');
await s.goto(`${B}/r/${r204.code}`);
await bild(s, '01-qr-anmelden');
pruefe(await s.getByRole('heading', { name: 'Schadenmeldung' }).isVisible(), 'QR ohne Anmeldung zeigt die Anmeldung');
await s.getByLabel('Name oder E-Mail').fill('k.neumann');
await s.getByLabel('Passwort des Schul-Postfachs').fill('geheim');
await s.getByRole('button', { name: 'Anmelden' }).click();
await s.getByRole('heading', { name: 'Schaden melden' }).waitFor();
pruefe(await s.getByRole('button', { name: /204.*Ändern/s }).isVisible(), 'Nach der Anmeldung ist Raum 204 aus dem QR-Code gewählt');
pruefe(s.url().endsWith('/r/' + r204.code), 'Adresse des QR-Codes bleibt stehen');
await bild(s, '02-qr-nach-anmeldung');
// Zweiter Scan auf demselben Handy: ohne Anmeldung direkt ins Formular
const r101 = liste.find((r) => r.name === '101');
await s.goto(`${B}/r/${r101.code}`);
await s.getByRole('heading', { name: 'Schaden melden' }).waitFor();
pruefe(await s.getByRole('button', { name: /101.*Ändern/s }).isVisible(), 'Zweiter Scan: direkt ins Formular, Raum 101 gewählt, keine Anmeldung');
// Kollegium sieht nichts anderes
await s.goto(B + '/uebersicht');
await s.getByRole('heading', { name: 'Schaden melden' }).waitFor();
pruefe(!(await s.getByText('Aufgaben').isVisible()) && !(await s.getByText('Übersicht').isVisible()), 'Kollegium: /uebersicht führt zurück aufs Formular');
await k.close();

// 2. Kollegium: Formular ohne QR, Melden, Bestätigung
[k, s] = await kontext('s.schmidt');
await s.goto(B + '/');
await bild(s, '03-melden-leer');
pruefe(await s.getByRole('heading', { name: 'Neubau · Naturwissenschaften' }).isVisible(), 'Bis 15 Räume: alle auf einen Blick, nach Bereichen');
await s.getByRole('button', { name: 'Chemie 2' }).click();
await s.getByRole('button', { name: 'Wasser & WC' }).click();
await s.getByRole('button', { name: 'Wasserhahn tropft' }).click();
await bild(s, '04-melden-ausgefuellt');
await bild(s, '04b-melden-ausgefuellt-ganz', true);
await s.getByRole('button', { name: /Melden · Chemie 2/ }).click();
await s.getByText('Gemeldet – danke!').waitFor();
await bild(s, '05-gemeldet');
await k.close();

// 3. Hausmeister hell und dunkel, Kontomenü mit Schaltern
[k, s] = await kontext('hausmeister');
await s.goto(B + '/'); await bild(s, '06-aufgaben-hell');
await s.getByRole('button', { name: 'Konto und Einstellungen' }).click(); await bild(s, '07-kontomenue-hausmeister');
pruefe(await s.getByRole('switch', { name: 'Dunkles Design' }).isVisible(), 'Kontomenü: Schalter „Dunkles Design“');
pruefe(await s.getByRole('switch', { name: 'Bei jeder neuen Meldung' }).getAttribute('aria-checked') === 'true', 'Hausmeister: Mail „Bei jeder neuen Meldung“ ist an');
await k.close();
[k, s] = await kontext('hausmeister', 'dark');
await s.goto(B + '/'); await bild(s, '08-aufgaben-dunkel'); await k.close();

// 4. Leitung darf alles, die Verwaltung (Sekretariat) sieht alles und ändert nichts
const haken = /^Erledigt: /;
[k, s] = await kontext('direktorin');
await s.goto(B + '/'); await s.getByRole('button', { name: haken }).first().waitFor(); await bild(s, '09-leitung');
pruefe((await s.getByRole('button', { name: haken }).count()) >= 6, 'Leitung: jede offene Meldung mit Häkchen „Erledigt“');
await k.close();
[k, s] = await kontext('sekretariat');
await s.goto(B + '/'); await s.getByRole('button', { name: /WC verstopft/ }).waitFor(); await bild(s, '09b-verwaltung');
pruefe((await s.getByRole('region', { name: 'Meldungen' }).getByRole('listitem').count()) >= 6 && (await s.getByRole('button', { name: haken }).count()) === 0, 'Verwaltung: alle Meldungen, ohne Häkchen');
await s.getByRole('button', { name: /^Bio 1/ }).click();
await s.getByRole('heading', { name: 'Licht flackert' }).waitFor();
pruefe(!(await s.getByRole('group', { name: 'Stand' }).isVisible()) && !(await s.getByLabel('Interner Vermerk').isVisible()) && !(await s.getByRole('button', { name: 'Erledigt' }).isVisible()), 'Verwaltung: Meldung nur lesend, ohne Stand, Vermerk-Feld und Erledigt');
await bild(s, '09c-verwaltung-meldung');
await s.getByRole('button', { name: 'Zurück' }).click();
await s.getByRole('button', { name: /Konto/ }).click();
pruefe(await s.getByRole('button', { name: 'QR-Aufkleber drucken' }).isVisible() && !(await s.getByRole('button', { name: 'Einstellungen', exact: true }).isVisible()), 'Verwaltung: QR-Aufkleber im Kontomenü, keine Einstellungen');
await k.close();

// 4b. Hausmeister: Meldung öffnen, Stand setzen (führt zurück zur Liste), in der Liste abhaken und rückgängig machen
[k, s] = await kontext('hausmeister');
await s.goto(B + '/');
await s.getByRole('button', { name: /^Lehrerzimmer/ }).click();
await bild(s, '09d-meldung-hausmeister');
await s.getByRole('group', { name: 'Stand' }).getByRole('button', { name: 'In Arbeit' }).click();
await s.getByText('In Arbeit: Lehrerzimmer').waitFor({ timeout: 5000 });
pruefe(new URL(s.url()).pathname === '/' && (await s.getByRole('button', { name: /^Lehrerzimmer/ }).getByText('In Arbeit').isVisible()), '„In Arbeit“ führt zurück zur Liste, die Zeile zeigt den Stand');
await s.getByRole('button', { name: 'Erledigt: Lehrerzimmer, Tisch wackelt' }).click();
await s.getByText('Erledigt: Lehrerzimmer').waitFor();
await s.getByRole('button', { name: /^Lehrerzimmer/ }).waitFor({ state: 'hidden' });
pruefe(true, 'Häkchen in der Liste: Meldung ist erledigt');
await bild(s, '09e-abgehakt');
await s.getByRole('button', { name: 'Rückgängig' }).click();
await s.getByRole('button', { name: /^Lehrerzimmer/ }).waitFor();
pruefe(true, 'Rückgängig holt die Meldung zurück');
await k.close();

// 5. Admin: Personen, Räume, Arten
[k, s] = await kontext('admin');
await s.goto(B + '/einstellungen'); await s.getByRole('tab', { name: 'Personen' }).click(); await bild(s, '10-personen');
await s.getByRole('tab', { name: 'Räume' }).click(); await bild(s, '11-raeume');
// Räume aus einer Datei: Die Zeilen landen zum Prüfen im Feld, angelegt wird erst mit „Hinzufügen“.
await s.getByRole('button', { name: 'Räume hinzufügen' }).click();
await s.locator('input[type=file]').setInputFiles({ name: 'raeume.csv', mimeType: 'text/csv', buffer: Buffer.from('Raum;Gebäude\nProbe 1;Erprobung\nProbe 2;Erprobung\n') });
await s.getByText('2 Räume gelesen').waitFor();
pruefe((await s.getByLabel('Räume', { exact: true }).inputValue()).includes('Probe 1; Erprobung'), 'Räume aus CSV: Zeilen stehen zum Prüfen im Feld');
await s.getByRole('button', { name: 'Hinzufügen', exact: true }).click();
await s.getByText('2 Räume angelegt').waitFor();
pruefe(await s.getByRole('button', { name: 'Probe 2' }).isVisible(), 'Räume aus CSV angelegt');
// Umbenennen behält den QR-Code.
await s.getByRole('button', { name: 'Probe 1' }).click();
await s.getByLabel('Name', { exact: true }).fill('Probe 1a');
await s.getByRole('button', { name: 'Speichern' }).click();
await s.getByText('Raum gespeichert').waitFor();
pruefe(await s.getByRole('button', { name: 'Probe 1a' }).isVisible(), 'Raum umbenannt');
// Arten: Medientechnik ist zu Beginn aus und lässt sich anschalten.
await s.getByRole('tab', { name: 'Arten' }).click(); await bild(s, '11b-arten');
const technik = s.getByRole('switch', { name: 'Medientechnik' });
pruefe((await technik.getAttribute('aria-checked')) === 'false', 'Arten: Medientechnik ist zu Beginn aus');
const [kk, lehrer] = await kontext('m.mustermann');
await lehrer.goto(B + '/melden'); await lehrer.getByRole('button', { name: 'Licht & Strom' }).waitFor();
pruefe(!(await lehrer.getByRole('button', { name: 'Medientechnik' }).isVisible()), 'Formular ohne Kachel Medientechnik');
await technik.click(); await s.getByText('steht wieder im Formular').waitFor();
await lehrer.reload(); await lehrer.getByRole('button', { name: 'Licht & Strom' }).waitFor();
pruefe(await lehrer.getByRole('button', { name: 'Medientechnik' }).isVisible(), 'Angeschaltet: Kachel Medientechnik ist da');
await technik.click(); await s.getByText('aus dem Formular genommen').waitFor();
await kk.close(); await k.close();

// 6. Die Mail an den Hausmeister und ihr Knopf „Erledigt“
const post = await (await fetch('http://127.0.0.1:8025/api/v1/messages')).json();
const nachricht = post.messages.find((m) => m.Subject.startsWith('Dringend'));
const voll = await (await fetch(`http://127.0.0.1:8025/api/v1/message/${nachricht.ID}`)).json();
[k, s] = await kontext(null);
await s.setContent(voll.HTML); await bild(s, '12-mail');
const knoepfe = await Promise.all(['In Arbeit', 'Fachfirma', 'Erledigt'].map((n) => s.getByRole('link', { name: n, exact: true }).isVisible()));
pruefe(knoepfe.every(Boolean), 'Mail mit den Knöpfen In Arbeit, Fachfirma und Erledigt');
const token = voll.Text.match(/\/aktion\/([A-Za-z0-9_-]+)\?tun=erledigt/)[1];
await s.goto(`${B}/aktion/${token}?tun=erledigt`);
await s.getByRole('button', { name: 'Erledigt' }).waitFor();
await bild(s, '13-mail-knopf');
const vorher = await (await admin.request.get(B + '/api/meldungen')).json();
pruefe(vorher.offen.some((m) => m.was === 'WC verstopft'), 'Beim bloßen Öffnen des Links passiert nichts (Schutz vor Link-Scannern)');
await s.getByRole('button', { name: 'Erledigt' }).click();
await s.getByText('„Erledigt“ ist gesetzt').waitFor();
await bild(s, '14-mail-erledigt');
const nachher = await (await admin.request.get(B + '/api/meldungen')).json();
pruefe(nachher.erledigt.some((m) => m.was === 'WC verstopft'), 'Knopf aus der Mail hat die Meldung erledigt – ohne Anmeldung');

// 7. Später derselbe Link mit einem anderen Knopf, etwa beim zweiten Hausmeister: „Schon erledigt“
// statt eines Knopfes, der die Meldung aus Versehen wieder öffnet; ändern erst nach „Stand ändern“.
await s.goto(`${B}/aktion/${token}?tun=in_arbeit`);
await s.getByText('Schon erledigt').waitFor();
pruefe(!(await s.getByRole('button', { name: 'In Arbeit' }).isVisible()), 'Schon erledigt: kein Knopf „In Arbeit“ auf den ersten Blick');
await bild(s, '15-mail-schon-erledigt');
await s.getByRole('button', { name: 'Stand ändern' }).click();
pruefe(await s.getByRole('button', { name: 'In Arbeit' }).isVisible(), 'Nach „Stand ändern“ stehen die Knöpfe da');
await s.getByRole('button', { name: 'In Arbeit' }).click();
await s.getByText('„In Arbeit“ ist gesetzt').waitFor();
// Die Seite zeigt noch „In Arbeit“, da erledigt der andere in der App: Der nächste Tipp öffnet nicht wieder.
await s.getByRole('button', { name: 'Doch anders' }).click();
const wc = (await (await admin.request.get(B + '/api/meldungen')).json()).offen.find((m) => m.was === 'WC verstopft');
await admin.request.post(`${B}/api/meldungen/${wc.id}/status`, { headers: { 'X-Schaden': '1' }, data: { status: 'erledigt' } });
await s.getByRole('button', { name: 'Fachfirma' }).click();
await s.getByText('Schon erledigt').waitFor();
const zuletzt = await (await admin.request.get(B + '/api/meldungen')).json();
pruefe(zuletzt.erledigt.some((m) => m.was === 'WC verstopft'), 'Überholte Seite öffnet die erledigte Meldung nicht wieder');
await k.close();

// 8. „Wieder öffnen“ führt wie „Erledigt“ zurück in die Liste; bliebe die Seite stehen, lüde der große
// Knopf „Erledigt“ zum Wieder-Abhaken ein.
const as = await admin.newPage();
await as.goto(B + '/uebersicht'); await as.getByRole('tab', { name: 'Erledigt' }).click();
await as.getByRole('button', { name: /WC verstopft/ }).first().click();
await as.getByRole('button', { name: 'Wieder öffnen' }).click();
await as.getByText('Wieder offen: WC Mädchen 1. OG').waitFor();
pruefe(!new URL(as.url()).pathname.startsWith('/meldung/'), '„Wieder öffnen“ führt zurück in die Liste, mit „Rückgängig“');

await browser.close();
console.log(fehler.length ? 'FEHLER:\n' + fehler.join('\n') : 'alles in Ordnung');
