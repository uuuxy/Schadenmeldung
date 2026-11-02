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
	s.on('console', (m) => m.type() === 'error' && !/401|410/.test(m.text()) && fehler.push(m.text()));
	return [k, s];
}
const bild = async (s, d, ganz = false) => { await s.waitForTimeout(600); await s.screenshot({ path: `${ziel}/${d}.png`, fullPage: ganz }); };

// 1. QR-Code ohne Anmeldung → anmelden → Raum ist erkannt
let [k, s] = await kontext(null);
const raeume = await (await browser.newContext()).request.get(B + '/api/raeume').catch(() => null);
const admin = await browser.newContext(); await admin.request.post(`${B}/api/anmelden`, { headers: { 'X-Schaden': '1' }, data: { email: 'verwaltung', passwort: 'x' } });
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
await s.getByRole('button', { name: 'Neubau · Naturwissenschaften' }).click();
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
pruefe(await s.getByRole('switch', { name: 'Dringendes sofort' }).getAttribute('aria-checked') === 'true', 'Hausmeister: „Dringendes sofort“ ist an');
await k.close();
[k, s] = await kontext('hausmeister', 'dark');
await s.goto(B + '/'); await bild(s, '08-aufgaben-dunkel'); await k.close();

// 4. Leitung: sieht alles, ändert nichts
[k, s] = await kontext('direktorin');
await s.goto(B + '/'); await bild(s, '09-leitung');
pruefe(await s.getByRole('heading', { name: 'Übersicht' }).isVisible() && (await s.locator('article').count()) > 0 && (await s.locator('article').getByRole('button', { name: /Erledigt|Antwort|Gesehen|Bestellt|Fachfirma|löschen/ }).count()) === 0 && (await s.locator('article').count()) >= 6, 'Leitung: Übersicht, Karten ohne Bearbeiten-Knöpfe');
await k.close();

// 5. Verwaltung: Personen
[k, s] = await kontext('verwaltung');
await s.goto(B + '/verwaltung'); await s.getByRole('tab', { name: 'Personen' }).click(); await bild(s, '10-personen');
await s.getByRole('tab', { name: 'Räume' }).click(); await bild(s, '11-raeume');
await k.close();

// 6. Die Mail an den Hausmeister und ihr Knopf „Erledigt“
const post = await (await fetch('http://127.0.0.1:8025/api/v1/messages')).json();
const nachricht = post.messages.find((m) => m.Subject.startsWith('Dringend'));
const voll = await (await fetch(`http://127.0.0.1:8025/api/v1/message/${nachricht.ID}`)).json();
[k, s] = await kontext(null);
await s.setContent(voll.HTML); await bild(s, '12-mail');
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
await k.close();

await browser.close();
console.log(fehler.length ? 'FEHLER:\n' + fehler.join('\n') : 'alles in Ordnung');
