import { chromium } from 'playwright-core';
// „Zurück“ überall gleich: Ein Bereich und „Anderer Ort“ haben einen eigenen Eintrag im Verlauf, damit
// die Zurück-Geste (Android, Safari) genau eine Ebene zurückführt; der Pfeil geht denselben Weg. In der
// App vom Home-Bildschirm am iPhone führt Wischen vom Rand zurück. Die Räume liefert hier der Test (20,
// also erst der Bereich, dann der Raum), und an den Server geht nichts, damit seine Daten unberührt bleiben.
const B = process.env.ADRESSE ?? 'http://127.0.0.1:8090';
const browser = await chromium.launch();
const fehler = [];
const pruefe = (bed, text) => { console.log((bed ? '✓ ' : '✗ ') + text); if (!bed) fehler.push('Prüfung: ' + text); };
const raeume = [...Array.from({ length: 10 }, (_, i) => `A1${String(i + 1).padStart(2, '0')}`), ...Array.from({ length: 8 }, (_, i) => `B20${i + 1}`), 'Aula', 'Mensa']
	.map((name, i) => ({ id: 9000 + i, name, bereich: '', code: `zurueck${i}` }));

async function kontext(name, { homescreen = false } = {}) {
	const k = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, isMobile: true, hasTouch: true, locale: 'de-DE' });
	await k.request.post(`${B}/api/anmelden`, { headers: { 'X-Schaden': '1' }, data: { email: name, passwort: 'x' } });
	await k.route('**/api/raeume', (r) => (r.request().method() === 'GET' ? r.fulfill({ json: raeume }) : r.continue()));
	// Abschicken antwortet der Test selbst: Eine echte Meldung brächte die Testdaten der anderen Prüfungen durcheinander.
	await k.route('**/api/meldungen', (r) => (r.request().method() === 'POST' ? r.fulfill({ json: { id: 99999, ort: 'Schulhof', kategorie: 'wasser', was: 'Wasserhahn tropft', status: 'offen', gemeldet_am: new Date().toISOString() } }) : r.continue()));
	// So sieht die App vom Home-Bildschirm am iPhone aus: navigator.standalone gibt es nur dort.
	if (homescreen) await k.addInitScript(() => Object.defineProperty(Navigator.prototype, 'standalone', { get: () => true }));
	const s = await k.newPage();
	s.on('pageerror', (e) => fehler.push(String(e)));
	s.on('console', (m) => m.type() === 'error' && !/401|410/.test(m.text()) && fehler.push(m.text()));
	return [k, s, await k.newCDPSession(s)];
}
const verlauf = (s) => s.evaluate(() => ({ laenge: history.length, unter: history.state?.unter, pfad: location.pathname }));
// Wie die Zurück-Geste von Android oder das Wischen in Safari: ein Schritt zurück im Verlauf.
const geste = (s) => s.evaluate(() => history.back());
const ruhe = (s) => s.waitForTimeout(300);
async function wische(cdp, von, bis, { y = 500, hoch = 0 } = {}) {
	await cdp.send('Input.dispatchTouchEvent', { type: 'touchStart', touchPoints: [{ x: von, y }] });
	for (let i = 1; i <= 8; i++) await cdp.send('Input.dispatchTouchEvent', { type: 'touchMove', touchPoints: [{ x: von + ((bis - von) * i) / 8, y: y - (hoch * i) / 8 }] });
	await cdp.send('Input.dispatchTouchEvent', { type: 'touchEnd', touchPoints: [] });
	await new Promise((r) => setTimeout(r, 400));
}
const bereichA = (s) => s.getByRole('button', { name: 'A: 10 Räume' });
const pfeilBereich = (s) => s.getByRole('button', { name: 'Zurück zu allen Bereichen' });
const pfeilOrt = (s) => s.getByRole('button', { name: 'Zurück zu den Räumen' });
const andererOrt = (s) => s.getByRole('button', { name: 'Anderer Ort: Schulhof, Flur …' });
const segment = (s, name) => s.getByRole('group', { name: 'Ansicht' }).getByRole('button', { name });

// 1. Kollegium: Bereich auf und wieder zu, per Pfeil und per Geste
let [k, s, cdp] = await kontext('s.schmidt');
await s.goto(B + '/');
await bereichA(s).waitFor();
const vorher = await verlauf(s);
await bereichA(s).click();
await pfeilBereich(s).waitFor();
let v = await verlauf(s);
pruefe(v.laenge === vorher.laenge + 1 && v.unter !== undefined, 'Ein Bereich legt einen Eintrag im Verlauf an');
await pfeilBereich(s).dblclick();
await bereichA(s).waitFor();
await ruhe(s);
v = await verlauf(s);
pruefe(v.unter === undefined && v.pfad === '/', 'Der Pfeil führt im Verlauf zurück zu den Bereichen, auch doppelt getippt nur einen Schritt');
await bereichA(s).click();
await pfeilBereich(s).waitFor();
pruefe((await verlauf(s)).laenge === vorher.laenge + 1, 'Wieder geöffnet: kein weiterer Eintrag, der alte wird ersetzt');
await wische(cdp, 4, 220);
pruefe(await pfeilBereich(s).isVisible(), 'Im Browser wischt die App nicht selbst (das macht Safari über den Verlauf)');
await geste(s);
await bereichA(s).waitFor();
pruefe((await verlauf(s)).unter === undefined, 'Die Zurück-Geste schließt den Bereich');

// 2. Raum im Bereich gewählt: Sein Eintrag geht mit, sonst täte „Zurück“ danach einmal nichts
await bereichA(s).click();
await s.getByRole('button', { name: 'A103', exact: true }).click();
await s.getByRole('button', { name: /A103.*Ändern/s }).waitFor();
await ruhe(s);
v = await verlauf(s);
pruefe(v.unter === undefined && v.laenge === vorher.laenge + 1, 'Raum gewählt: Der Eintrag des Bereichs ist abgeräumt');

// 3. „Anderer Ort“: eigener Eintrag; abgeschickt geht er mit, „Noch etwas melden“ legt ihn wieder an
await s.getByRole('button', { name: /A103.*Ändern/s }).click();
await andererOrt(s).click();
await pfeilOrt(s).waitFor();
pruefe((await verlauf(s)).unter !== undefined, '„Anderer Ort“ legt einen Eintrag im Verlauf an');
await geste(s);
await bereichA(s).waitFor();
pruefe(!(await pfeilOrt(s).isVisible()), 'Die Zurück-Geste führt von „Anderer Ort“ zu den Räumen');
await andererOrt(s).click();
await s.getByLabel('Wo ist der Schaden?').fill('Schulhof (Prüfung Zurück)');
await s.getByRole('button', { name: 'Wasser & WC' }).click();
await s.getByRole('button', { name: 'Wasserhahn tropft' }).click();
await s.getByRole('button', { name: /Melden · Schulhof/ }).click();
await s.getByText('Gemeldet – danke!').waitFor();
await ruhe(s);
pruefe((await verlauf(s)).unter === undefined, 'Abgeschickt: Der Eintrag von „Anderer Ort“ ist abgeräumt');
await s.getByRole('button', { name: 'Noch etwas melden' }).click();
await pfeilOrt(s).waitFor();
pruefe((await verlauf(s)).unter !== undefined, '„Noch etwas melden“: „Anderer Ort“ steht wieder da, mit Eintrag');
await pfeilOrt(s).click();
await bereichA(s).waitFor();
await k.close();

// 4. Hausmeister-Ebene: Wer aus einem Bereich zur Übersicht wechselt, kommt mit „Zurück“ ohne toten Schritt heim
[k, s, cdp] = await kontext('admin');
await s.goto(B + '/');
await segment(s, 'Melden').click();
await bereichA(s).click();
await pfeilBereich(s).waitFor();
await segment(s, 'Übersicht').click();
await s.waitForURL(B + '/');
await geste(s);
await bereichA(s).waitFor();
await ruhe(s);
v = await verlauf(s);
pruefe(v.pfad === '/melden' && v.unter === undefined, 'Zurück aus der Übersicht: Melden ohne offenen Bereich, kein Eintrag übrig');
await geste(s);
await s.waitForURL(B + '/');
pruefe(await segment(s, 'Übersicht').getAttribute('aria-pressed') === 'true', 'Noch einmal zurück: wieder in der Übersicht');

// 5. Der Pfeil der Kopfleiste verlässt die Einstellungen auch dann, wenn ein Bereich offen ist
await s.getByRole('button', { name: 'Konto und Einstellungen' }).click();
await s.getByRole('button', { name: 'Einstellungen', exact: true }).click();
await bereichA(s).click();
await pfeilBereich(s).waitFor();
await s.getByRole('button', { name: 'Zurück', exact: true }).click();
await s.waitForURL(B + '/');
await ruhe(s);
v = await verlauf(s);
pruefe(v.pfad === '/' && v.unter === undefined, 'Pfeil der Kopfleiste: aus den Einstellungen zurück zur Übersicht');
await k.close();

// 6. iPhone, App vom Home-Bildschirm: Wischen vom Rand schließt erst das Blatt, dann den Bereich, dann die Seite
[k, s, cdp] = await kontext('admin', { homescreen: true });
await s.goto(B + '/');
await s.getByRole('button', { name: 'Konto und Einstellungen' }).click();
await s.getByRole('button', { name: 'Einstellungen', exact: true }).click();
await bereichA(s).click();
await s.getByRole('button', { name: 'A105', exact: true }).click();
const blatt = s.getByRole('dialog', { name: 'Raum bearbeiten' });
await blatt.waitFor();
await wische(cdp, 4, 220);
pruefe(!(await blatt.isVisible()) && (await pfeilBereich(s).isVisible()), 'Wischen schließt zuerst das Blatt, der Bereich bleibt');
await wische(cdp, 4, 50);
pruefe(await pfeilBereich(s).isVisible(), 'Zu kurz gewischt: nichts passiert');
await wische(cdp, 120, 330);
pruefe(await pfeilBereich(s).isVisible(), 'Mitten auf der Seite gewischt: nichts passiert');
await wische(cdp, 4, 20, { hoch: 200 });
pruefe(await pfeilBereich(s).isVisible(), 'Am Rand gerollt statt gewischt: nichts passiert');
await wische(cdp, 4, 220);
pruefe(!(await pfeilBereich(s).isVisible()) && (await bereichA(s).isVisible()), 'Wischen vom Rand: zurück zu den Bereichen');
await wische(cdp, 4, 220);
await s.waitForURL(B + '/');
pruefe(await segment(s, 'Übersicht').getAttribute('aria-pressed') === 'true', 'Noch einmal gewischt: aus den Einstellungen zur Übersicht');
await wische(cdp, 4, 220);
pruefe(new URL(s.url()).pathname === '/', 'In der Übersicht gibt es kein Zurück: Wischen tut nichts');
await k.close();

await browser.close();
if (fehler.length) {
	console.log('\n' + fehler.join('\n'));
	process.exit(1);
}
