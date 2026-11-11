import { chromium } from 'playwright-core';
import { writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
// Drucken: auf der Seite einer Meldung ein Blatt, in der Liste mehrere über die Auswahl (M3 selection,
// wie in Gmail): langes Drücken oder ⋮ → Drucken, Tippen wählt dazu oder ab, Kreuz, Zurück
// und Drucken beenden. Braucht die Testdaten; gedruckt wird nur gelesen. Bild und PDF landen im Ordner
// aus argv[2], sonst im Temp-Ordner.
const B = process.env.ADRESSE ?? 'http://127.0.0.1:8090';
const ordner = process.argv[2] ?? tmpdir();
const browser = await chromium.launch();
const fehler = [];
const pruefe = (bed, text) => { console.log((bed ? '✓ ' : '✗ ') + text); if (!bed) fehler.push('Prüfung: ' + text); };

const k = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, isMobile: true, hasTouch: true, locale: 'de-DE', acceptDownloads: true });
await k.request.post(`${B}/api/anmelden`, { headers: { 'X-Schaden': '1' }, data: { email: 'hausmeister', passwort: 'x' } });
const s = await k.newPage();
s.on('pageerror', (e) => fehler.push(String(e)));
s.on('console', (m) => m.type() === 'error' && !/401|410/.test(m.text()) && fehler.push(m.text()));
const cdp = await k.newCDPSession(s);

const zeilen = () => s.getByRole('region', { name: 'Meldungen' }).locator('li > div > button:first-child');
const titel = () => s.locator('header h1').first().textContent();
const ruhe = () => s.waitForTimeout(400);
const seiten = (roh) => (roh.toString('latin1').match(/\/Type \/Page\b/g) ?? []).length;
async function drueckeLange(zeile) {
	const b = await zeile.boundingBox();
	const punkt = [{ x: b.x + b.width / 2, y: b.y + b.height / 2 }];
	await cdp.send('Input.dispatchTouchEvent', { type: 'touchStart', touchPoints: punkt });
	await s.waitForTimeout(700);
	await cdp.send('Input.dispatchTouchEvent', { type: 'touchEnd', touchPoints: [] });
	await ruhe();
}

// 1. Langes Drücken beginnt die Auswahl, ohne die Meldung zu öffnen
await s.goto(B + '/');
await zeilen().first().waitFor();
pruefe(await s.getByRole('button', { name: 'Weitere Optionen' }).isVisible(), 'Kopfleiste: ⋮ Weitere Optionen');
await drueckeLange(zeilen().first());
pruefe((await titel()) === '1 ausgewählt' && new URL(s.url()).pathname === '/', `Langes Drücken wählt aus und bleibt in der Liste (${await titel()})`);
pruefe((await zeilen().first().getAttribute('aria-pressed')) === 'true', 'Die Zeile ist gewählt (aria-pressed)');
pruefe(await s.evaluate(() => history.state?.unter !== undefined), 'Die Auswahl hat einen Eintrag im Verlauf');
pruefe((await s.getByRole('button', { name: /^Erledigt:/ }).count()) === 0, 'In der Auswahl fehlt das Häkchen „Erledigt“');

// 2. Ein Tipp wählt dazu; Drucken liefert ein PDF mit einer Seite je Meldung und beendet die Auswahl
await zeilen().nth(1).tap();
pruefe((await titel()) === '2 ausgewählt', 'Ein Tipp wählt eine zweite dazu');
pruefe((await s.locator('li.bg-secondary-container').count()) === 2, 'Gewählte Zeilen liegen auf secondary-container');
await s.screenshot({ path: `${ordner}/auswahl.png` });
const drucken = s.getByRole('link', { name: 'Drucken' });
const href = await drucken.getAttribute('href');
const antwort = await k.request.get(B + href);
const pdf = await antwort.body();
writeFileSync(`${ordner}/druck.pdf`, pdf);
pruefe(antwort.ok() && antwort.headers()['content-type'] === 'application/pdf' && seiten(pdf) === 2, `PDF mit 2 Seiten (${href})`);
const [download] = await Promise.all([s.waitForEvent('download'), drucken.click()]);
await ruhe();
pruefe(download.suggestedFilename().startsWith('Schadensmeldungen-'), `Download heißt ${download.suggestedFilename()}`);
pruefe(await s.getByRole('group', { name: 'Ansicht' }).isVisible(), 'Nach dem Drucken ist die Auswahl beendet');
pruefe(await s.evaluate(() => history.state?.unter === undefined), 'Der Eintrag der Auswahl ist abgeräumt');

// 3. Wer die letzte abwählt, beendet die Auswahl (M3)
await drueckeLange(zeilen().first());
await zeilen().first().tap();
await ruhe();
pruefe(await s.getByRole('group', { name: 'Ansicht' }).isVisible(), 'Letzte abgewählt: Auswahl beendet');

// 4. Über das Menü: ⋮ → Drucken, Alle, keine; Zurück beendet
await s.getByRole('button', { name: 'Weitere Optionen' }).click();
await ruhe();
await s.screenshot({ path: `${ordner}/menue.png` });
await s.getByRole('dialog').getByRole('button', { name: 'Drucken' }).click();
await ruhe();
pruefe((await titel()) === 'Meldungen auswählen' && (await s.locator('header').getByText('Zum Auswählen antippen').isVisible()) && (await s.getByRole('button', { name: 'Drucken' }).isDisabled()), 'Aus dem Menü: „Meldungen auswählen“, darunter „Zum Auswählen antippen“, Drucken gesperrt');
const anzahl = await zeilen().count();
await s.getByRole('button', { name: 'Alle auswählen' }).click();
pruefe((await titel()) === `${anzahl} ausgewählt`, `Alle auswählen: ${anzahl} ausgewählt`);
await s.getByRole('button', { name: 'Alle abwählen' }).click();
pruefe((await titel()) === 'Meldungen auswählen', 'Alle abwählen: keine, die Auswahl bleibt offen');
await s.evaluate(() => history.back());
await ruhe();
pruefe(await s.getByRole('group', { name: 'Ansicht' }).isVisible(), 'Zurück-Geste beendet die Auswahl');

// 5. Kurzer Tipp öffnet die Meldung; dort druckt das Symbol genau sie
await zeilen().first().tap();
await s.waitForURL(/\/meldung\/\d+$/);
const id = s.url().split('/').pop();
const einzeln = s.getByRole('link', { name: 'Drucken' });
pruefe((await einzeln.getAttribute('href')) === `/api/druck.pdf?ids=${id}`, 'Seite der Meldung: Drucken druckt sie');
pruefe(seiten(await (await k.request.get(`${B}/api/druck.pdf?ids=${id}`)).body()) === 1, 'Ein Blatt mit einer Seite');

// 6. Am Rechner öffnet das PDF gleich im Browser, zum Drucken mit einem Klick; am Handy wird es geladen
pruefe((await einzeln.getAttribute('download')) !== null && (await einzeln.getAttribute('target')) === null, 'Handy: Drucken lädt das PDF wie bisher');
const rechner = await browser.newContext({ viewport: { width: 1280, height: 800 }, locale: 'de-DE' });
await rechner.request.post(`${B}/api/anmelden`, { headers: { 'X-Schaden': '1' }, data: { email: 'hausmeister', passwort: 'x' } });
const r = await rechner.newPage();
await r.goto(`${B}/meldung/${id}`);
const link = r.getByRole('link', { name: 'Drucken' });
await link.waitFor();
const adresse = await link.getAttribute('href');
pruefe((await link.getAttribute('target')) === '_blank' && adresse.endsWith('&zeigen=1') && (await link.getAttribute('download')) === null, 'Rechner: Drucken öffnet das PDF im Browser');
pruefe((await rechner.request.get(B + adresse)).headers()['content-disposition'].startsWith('inline;'), 'Rechner: Der Server schickt es zum Ansehen');

await browser.close();
if (fehler.length) {
	console.log('\n' + fehler.join('\n'));
	process.exit(1);
}
console.log('alles in Ordnung');
