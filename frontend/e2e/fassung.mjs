import { chromium } from 'playwright-core';
// Neue Fassung: Wechselt die App in den Hintergrund oder zurück und nennt der Server ein anderes
// Startskript, lädt sie sich selbst neu, aber nicht, solange im Formular etwas getippt oder angetippt ist.
// Braucht den lokalen Server mit Testdaten aus e2e/testdaten.sh.
const B = process.env.ADRESSE ?? 'http://127.0.0.1:8090';
const browser = await chromium.launch();
const fehler = [];
const pruefe = (bed, text) => { console.log((bed ? '✓ ' : '✗ ') + text); if (!bed) fehler.push('Prüfung: ' + text); };

// Ohne Service Worker, damit die Nachfrage hier ankommt und umgebogen werden kann.
const k = await browser.newContext({ viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true, locale: 'de-DE', serviceWorkers: 'block' });
await k.request.post(`${B}/api/anmelden`, { headers: { 'X-Schaden': '1' }, data: { email: 's.schmidt', passwort: 'x' } });
const s = await k.newPage();
s.on('pageerror', (e) => fehler.push(String(e)));

// Sobald es eine neue Fassung „gibt“, nennt die Antwort auf die Nachfrage ein anderes Startskript;
// das Neuladen selbst bekommt die echte Seite.
let neueFassung = false;
await k.route(`${B}/`, async (route) => {
	if (!neueFassung || route.request().isNavigationRequest()) return route.continue();
	const antwort = await route.fetch();
	await route.fulfill({ response: antwort, body: (await antwort.text()).replace(/index-[\w-]+\.js/, 'index-neu.js') });
});

// Wie auf dem Handy: Die App geht in den Hintergrund (oder kommt zurück). Gemessen wird, ob sie sich
// binnen zwei Sekunden neu lädt.
async function wechsel(zustand = 'hidden') {
	const geladen = s.waitForEvent('load', { timeout: 2000 }).then(() => true, () => false);
	await s.evaluate((z) => {
		Object.defineProperty(document, 'visibilityState', { value: z, configurable: true });
		document.dispatchEvent(new Event('visibilitychange'));
	}, zustand);
	return geladen;
}
const frisch = async () => {
	await s.goto(B + '/');
	await s.getByRole('heading', { name: 'Schaden melden' }).waitFor();
};

await frisch();
pruefe(!(await wechsel()), 'Ohne neue Fassung bleibt die Seite stehen');

neueFassung = true;
await frisch();
const suche = s.getByRole('searchbox', { name: 'Raum suchen' });
await suche.fill('20');
pruefe(!(await wechsel()), 'Getippter Text hält die neue Fassung auf');
await suche.fill('');
pruefe(await wechsel(), 'Ist das Feld wieder leer, lädt sich die App neu');

await frisch();
await s.getByRole('button', { name: /Licht & Strom/ }).click();
pruefe(!(await wechsel()), 'Eine angetippte Kachel hält die neue Fassung auf');

await frisch();
pruefe(await wechsel('visible'), 'Zurück aus dem Hintergrund lädt sich die App neu');
await s.getByRole('heading', { name: 'Schaden melden' }).waitFor();

await browser.close();
if (fehler.length) {
	console.log('\nFEHLER:\n' + fehler.join('\n'));
	process.exit(1);
}
console.log('alles in Ordnung');
