import { chromium } from 'playwright-core';
// Live im Handyformat: Die Übersicht zeigt eine neue Meldung und einen neuen Stand ohne Neuladen
// (/api/ereignisse), das Kollegium öffnet keinen Strom. Dazu Bildschirmfotos von Umschalter und
// Farben der Arten, hell und dunkel, in den Ordner aus argv[2] (sonst Temp-Ordner).
// Braucht den lokalen Server mit Testdaten aus e2e/testdaten.sh.
const ziel = process.argv[2] ?? (await import('node:os')).tmpdir();
const B = 'http://127.0.0.1:8090';
const browser = await chromium.launch();
const fehler = [];
const pruefe = (bed, text) => { console.log((bed ? '✓ ' : '✗ ') + text); if (!bed) fehler.push('Prüfung: ' + text); };
async function kontext(name, schema = 'light') {
	const k = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, isMobile: true, hasTouch: true, colorScheme: schema, locale: 'de-DE' });
	await k.request.post(`${B}/api/anmelden`, { headers: { 'X-Schaden': '1' }, data: { email: name, passwort: 'x' } });
	const s = await k.newPage();
	s.on('pageerror', (e) => fehler.push(String(e)));
	const stroeme = [];
	s.on('request', (r) => r.url().endsWith('/api/ereignisse') && stroeme.push(r.url()));
	return [k, s, stroeme];
}
const bild = async (s, d) => { await s.waitForTimeout(400); await s.screenshot({ path: `${ziel}/${d}.png` }); };

const [, hm, hmStroeme] = await kontext('hausmeister');
await hm.goto(B + '/');
await hm.getByRole('group', { name: 'Ansicht' }).waitFor();
await hm.waitForTimeout(500);
await hm.evaluate(() => (window.__ohneNeuladen = true));
const vorher = await hm.locator('article').count();
pruefe(hmStroeme.length === 1, 'Hausmeister öffnet einen Ereignisstrom');

// Eine Lehrkraft meldet von ihrem Handy aus; gemessen wird bis zur Karte beim Hausmeister.
const [lk, lehrerin, lehrerinStroeme] = await kontext('s.schmidt');
await lehrerin.goto(B + '/');
await lehrerin.getByRole('heading', { name: 'Schaden melden' }).waitFor();
const aula = (await (await lk.request.get(B + '/api/raeume')).json()).find((r) => r.name === 'Aula');
const start = Date.now();
const antwort = await lk.request.post(`${B}/api/meldungen`, { headers: { 'X-Schaden': '1' }, multipart: { raumId: String(aula.id), kategorie: 'licht', was: 'Lampe defekt' } });
const neu = await antwort.json();
await hm.locator('article', { hasText: 'Aula' }).filter({ hasText: 'Lampe defekt' }).waitFor({ timeout: 5000 });
const dauer = Date.now() - start;
pruefe((await hm.locator('article').count()) === vorher + 1, `Neue Meldung erscheint beim Hausmeister nach ${dauer} ms`);
pruefe(dauer < 2000, 'schneller als 2 Sekunden');
pruefe(await hm.evaluate(() => window.__ohneNeuladen === true), '… ohne die Seite neu zu laden');
pruefe(lehrerinStroeme.length === 0, 'Kollegium öffnet keinen Ereignisstrom');

// Die Leitung sieht den neuen Stand, sobald der Hausmeister „Gesehen“ setzt.
const [, leitung] = await kontext('direktorin');
await leitung.goto(B + '/');
const karte = leitung.locator('article', { hasText: 'Aula' }).filter({ hasText: 'Lampe defekt' });
await karte.waitFor();
const start2 = Date.now();
await hm.context().request.post(`${B}/api/meldungen/${neu.id}/status`, { headers: { 'X-Schaden': '1' }, data: { status: 'gesehen' } });
await karte.getByText('Gesehen').first().waitFor({ timeout: 5000 });
pruefe(true, `Leitung sieht „Gesehen“ nach ${Date.now() - start2} ms`);

// Umschalter: von der Übersicht zur Meldeebene und zurück.
await hm.getByRole('button', { name: 'Melden', exact: true }).click();
await hm.getByRole('button', { name: 'Licht & Strom' }).waitFor();
pruefe(new URL(hm.url()).pathname === '/melden', 'Umschalter „Melden“ öffnet das Formular');
await bild(hm, 'live-01-melden-umschalter');
await hm.getByRole('button', { name: 'Licht & Strom' }).click();
await hm.getByRole('button', { name: 'Licht & Strom' }).scrollIntoViewIfNeeded();
await bild(hm, 'live-02-melden-licht-gewaehlt');
await hm.getByRole('button', { name: 'Übersicht' }).click();
await hm.locator('article').first().waitFor();
pruefe(new URL(hm.url()).pathname === '/', 'Umschalter „Übersicht“ führt zurück');
await bild(hm, 'live-03-uebersicht');
await bild(lehrerin, 'live-04-kollegium-ohne-umschalter');

// Dasselbe dunkel.
const [, dunkel] = await kontext('hausmeister', 'dark');
await dunkel.goto(B + '/');
await dunkel.locator('article').first().waitFor();
await bild(dunkel, 'live-05-uebersicht-dunkel');
await dunkel.goto(B + '/melden');
await dunkel.getByRole('button', { name: 'Wasser & WC' }).click();
await dunkel.getByRole('button', { name: 'Wasser & WC' }).scrollIntoViewIfNeeded();
await bild(dunkel, 'live-06-melden-wasser-dunkel');

await browser.close();
if (fehler.length) {
	console.log('\nFEHLER:\n' + fehler.join('\n'));
	process.exit(1);
}
console.log('alles in Ordnung');
