import { chromium } from 'playwright-core';
import { encode } from 'uqr';
// QR-Aufkleber in der App lesen: Das Symbol rechts in „Raum suchen“ öffnet den Scanner, ein Aufkleber
// wählt den Raum, fremde und veraltete Codes nennen den Grund, und die Kamera geht danach wieder aus.
// Die Kamera stellt der Test selbst: eine Leinwand, auf der ein echter QR-Code steht, so läuft das
// Lesen (jsQR) wirklich durch. Räume liefert der Test, an den Server geht keine Meldung.
const B = process.env.ADRESSE ?? 'http://127.0.0.1:8090';
const browser = await chromium.launch();
const fehler = [];
const pruefe = (bed, text) => { console.log((bed ? '✓ ' : '✗ ') + text); if (!bed) fehler.push('Prüfung: ' + text); };
const raeume = [
	{ id: 9101, name: '204', bereich: 'Hauptgebäude · 2. OG', code: 'scan204' },
	{ id: 9102, name: '205', bereich: 'Hauptgebäude · 2. OG', code: 'scan205' },
	{ id: 9103, name: 'Aula', bereich: 'Hauptgebäude · EG', code: 'scanaula' }
];
const codes = {
	raum: encode(`${B}/r/scan204`, { border: 2, ecc: 'M' }).data,
	alt: encode(`${B}/r/gibtsnichtmehr`, { border: 2, ecc: 'M' }).data,
	fremd: encode('https://example.org/hallo', { border: 2, ecc: 'M' }).data
};

// Ersetzt die Kamera: getUserMedia liefert den Strom einer Leinwand mit dem QR-Code aus window.__kamera.inhalt.
function kameraAttrappe(codes) {
	window.__kamera = { inhalt: 'raum', verweigern: false, gestartet: 0, gestoppt: 0 };
	navigator.mediaDevices.getUserMedia = async () => {
		if (window.__kamera.verweigern) throw new DOMException('verweigert', 'NotAllowedError');
		window.__kamera.gestartet++;
		const z = 10;
		const leinwand = Object.assign(document.createElement('canvas'), { width: 640, height: 480 });
		const pinsel = leinwand.getContext('2d');
		const male = () => {
			const m = codes[window.__kamera.inhalt];
			const x0 = (640 - m.length * z) / 2, y0 = (480 - m.length * z) / 2;
			pinsel.fillStyle = '#fff';
			pinsel.fillRect(0, 0, 640, 480);
			pinsel.fillStyle = '#000';
			m.forEach((zeile, y) => zeile.forEach((an, x) => an && pinsel.fillRect(x0 + x * z, y0 + y * z, z, z)));
		};
		male();
		const uhr = setInterval(male, 100);
		const strom = leinwand.captureStream(10);
		for (const spur of strom.getTracks()) {
			const stop = spur.stop.bind(spur);
			spur.stop = () => { window.__kamera.gestoppt++; clearInterval(uhr); stop(); };
		}
		return strom;
	};
}

async function kontext(name, handy = true) {
	const k = await browser.newContext(handy
		? { viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, isMobile: true, hasTouch: true, locale: 'de-DE' }
		: { viewport: { width: 1280, height: 800 }, locale: 'de-DE' });
	await k.request.post(`${B}/api/anmelden`, { headers: { 'X-Schaden': '1' }, data: { email: name, passwort: 'x' } });
	await k.route('**/api/raeume', (r) => (r.request().method() === 'GET' ? r.fulfill({ json: raeume }) : r.continue()));
	await k.route('**/api/meldungen', (r) => (r.request().method() === 'POST' ? r.abort() : r.continue()));
	await k.addInitScript(kameraAttrappe, codes);
	const s = await k.newPage();
	s.on('pageerror', (e) => fehler.push(String(e)));
	s.on('console', (m) => m.type() === 'error' && !/401|410/.test(m.text()) && fehler.push(m.text()));
	return [k, s];
}
const kamera = (s) => s.evaluate(() => ({ ...window.__kamera }));
const symbol = (s) => s.getByRole('button', { name: 'QR-Code scannen' });
const pfeil = (s) => s.getByRole('button', { name: 'Zurück zu den Räumen' });
const gewaehlt = (s) => s.getByRole('button', { name: /Ändern/ });

// 1. Handy: Aufkleber lesen wählt den Raum
let [k, s] = await kontext('s.schmidt');
await s.goto(B + '/');
await symbol(s).waitFor();
pruefe(await symbol(s).evaluate((el) => el.classList.contains('icon-btn')), 'Symbol ist ein Symbolknopf (48-px-Fläche, misst pruefen:tippflaechen)');
await symbol(s).tap();
await s.getByText('QR-Code scannen', { exact: true }).waitFor();
pruefe(await s.evaluate(() => !!history.state?.unter), 'Scanner hat einen Eintrag im Verlauf');
await gewaehlt(s).waitFor({ timeout: 8000 });
pruefe((await gewaehlt(s).textContent()).includes('204'), 'Aufkleber wählt den Raum 204');
await s.waitForTimeout(300);
let kam = await kamera(s);
pruefe(kam.gestoppt >= kam.gestartet && kam.gestartet >= 1, `Kamera danach aus (${kam.gestartet} gestartet, ${kam.gestoppt} gestoppt)`);
// Wie ein im Bereich gewählter Raum: Der Eintrag des Scanners geht mit, sonst täte „Zurück“ einmal nichts.
pruefe(await s.evaluate(() => history.state?.unter === undefined), 'Der Eintrag des Scanners ist abgeräumt');

// 2. Veralteter und fremder Code: Grund nennen, weiter lesen; Zurück schließt und hält die Kamera an
await s.goto(B + '/');
await s.evaluate(() => (window.__kamera.inhalt = 'alt'));
await symbol(s).tap();
await s.getByText('Dieser QR-Code gehört zu keinem Raum mehr.').waitFor({ timeout: 8000 });
pruefe(true, 'Veralteter Code: „gehört zu keinem Raum mehr“');
await s.evaluate(() => (window.__kamera.inhalt = 'fremd'));
await s.getByText('Das ist kein QR-Aufkleber der Schadenmeldung.').waitFor({ timeout: 8000 });
pruefe(true, 'Fremder Code: „kein QR-Aufkleber der Schadenmeldung“');
pruefe(await s.getByText('QR-Code scannen', { exact: true }).isVisible(), 'Scanner bleibt offen und liest weiter');
await s.evaluate(() => history.back());
await symbol(s).waitFor();
kam = await kamera(s);
pruefe(kam.gestoppt >= kam.gestartet, `Zurück-Geste schließt den Scanner, Kamera aus (${kam.gestartet}/${kam.gestoppt})`);
await symbol(s).tap();
await pfeil(s).click();
await symbol(s).waitFor();
pruefe(true, 'Pfeil führt zurück zu den Räumen');

// 3. Kamera nicht erlaubt: verständlicher Grund statt eines Fehlers
await s.evaluate(() => (window.__kamera.verweigern = true));
await symbol(s).tap();
await s.getByRole('alert').filter({ hasText: 'Die Kamera ist nicht erlaubt' }).waitFor();
pruefe(true, 'Verweigerte Kamera: Hinweis auf die Einstellungen');
await pfeil(s).click();

// 4. Beim Tippen weicht das Symbol dem Löschkreuz
await s.getByRole('searchbox', { name: 'Raum suchen' }).fill('20');
pruefe(await s.getByRole('button', { name: 'Suche leeren' }).isVisible() && !(await symbol(s).isVisible()), 'Mit Text: Löschkreuz statt Scanner');
await k.close();

// 5. Am Rechner gibt es kein Scanner-Symbol
[k, s] = await kontext('s.schmidt', false);
await s.goto(B + '/');
await s.getByRole('searchbox', { name: 'Raum suchen' }).waitFor();
pruefe(!(await symbol(s).isVisible()), 'Rechner: kein Scanner-Symbol');
await k.close();

await browser.close();
if (fehler.length) {
	console.log('\n' + fehler.join('\n'));
	process.exit(1);
}
console.log('alles in Ordnung');
