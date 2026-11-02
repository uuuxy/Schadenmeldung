// Tippt mit dem Touchscreen knapp NEBEN die sichtbaren Ränder (so wie ein Daumen danebenliegt) und
// prüft, ob trotzdem das Richtige passiert.
import { chromium } from 'playwright-core';
const B = 'http://127.0.0.1:8090';
const browser = await chromium.launch();
const fehler = [];
const pruefe = (bed, text) => { console.log((bed ? '✓ ' : '✗ ') + text); if (!bed) fehler.push(text); };
async function seite(name) {
	const k = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, isMobile: true, hasTouch: true, locale: 'de-DE' });
	if (name) await k.request.post(`${B}/api/anmelden`, { headers: { 'X-Schaden': '1' }, data: { email: name, passwort: 'x' } });
	return [k, await k.newPage()];
}
// Tippt dx/dy Pixel neben die Kante: unten (dy>0) bzw. oben (dy<0) außerhalb des sichtbaren Bauteils.
async function tippeNeben(s, ort, kante, abstand) {
	await ort.scrollIntoViewIfNeeded();
	const b = await ort.boundingBox();
	const y = kante === 'unten' ? b.y + b.height + abstand : b.y - abstand;
	await s.touchscreen.tap(b.x + b.width / 2, y);
	await s.waitForTimeout(350);
	return `${kante} +${abstand}px neben ${Math.round(b.height)}-px-Bauteil`;
}

let [k, s] = await seite('s.schmidt');
await s.goto(B + '/melden');
let wo = await tippeNeben(s, s.getByRole('button', { name: 'Sport', exact: true }), 'unten', 6);
pruefe(await s.getByRole('button', { name: 'Sport', exact: true }).getAttribute('aria-pressed') === 'true', `Bereichs-Chip „Sport“ (${wo})`);
wo = await tippeNeben(s, s.getByRole('button', { name: 'Turnhalle' }), 'unten', 0.5);
pruefe(await s.getByRole('button', { name: /Turnhalle.*Ändern/s }).isVisible(), `Raumknopf „Turnhalle“ (${wo})`);
await s.getByRole('button', { name: 'Licht & Strom' }).click();
await s.waitForTimeout(500);
wo = await tippeNeben(s, s.getByRole('button', { name: 'Licht flackert' }), 'oben', 6);
pruefe(await s.getByRole('button', { name: 'Licht flackert' }).getAttribute('aria-pressed') === 'true', `Fall-Chip „Licht flackert“ (${wo})`);
wo = await tippeNeben(s, s.getByRole('button', { name: 'Dringend / Gefahr' }), 'unten', 6);
pruefe(await s.getByRole('button', { name: 'Dringend / Gefahr' }).getAttribute('aria-pressed') === 'true', `Chip „Dringend / Gefahr“ (${wo})`);
await s.getByRole('button', { name: 'Dringend / Gefahr' }).click();
await s.goto(B + '/melden');
const pille = s.getByRole('searchbox', { name: 'Raum suchen' });
const pb = await pille.boundingBox();
await s.touchscreen.tap(pb.x - 20, pb.y + pb.height / 2);
await s.waitForTimeout(200);
pruefe(await pille.evaluate((el) => document.activeElement === el), 'Suchpille: Tipp auf die Lupe setzt den Cursor ins Feld');
await k.close();

[k, s] = await seite('hausmeister');
await s.goto(B + '/');
const karte = s.locator('article').first();
const ortVorher = await karte.locator('h3').textContent();
wo = await tippeNeben(s, karte.getByRole('button', { name: 'Gesehen' }), 'oben', 3);
pruefe((await karte.getByRole('button', { name: 'Gesehen' }).getAttribute('aria-pressed')) === 'true', `Hausmeister „Gesehen“ (${wo})`);
wo = await tippeNeben(s, karte.getByRole('button', { name: 'Erledigt' }), 'unten', 3);
await s.waitForTimeout(400);
pruefe((await s.locator('article').first().locator('h3').textContent()) !== ortVorher || (await s.getByText(/Erledigt: /).isVisible()), `Hausmeister „Erledigt“ (${wo})`);
wo = await tippeNeben(s, s.getByRole('group', { name: 'Ansicht' }).getByRole('button', { name: 'Melden' }), 'unten', 3);
await s.waitForTimeout(300);
pruefe(new URL(s.url()).pathname === '/melden', `Umschalter „Melden“ in der Kopfleiste (${wo})`);
await k.close();

[k, s] = await seite(null);
await s.goto(B + '/');
wo = await tippeNeben(s, s.getByRole('button', { name: 'Datenschutz' }), 'unten', 3);
pruefe(s.url().endsWith('/datenschutz'), `Anmeldung: „Datenschutz“ (${wo})`);
await k.close();

await browser.close();
console.log(fehler.length ? `\n${fehler.length} Fehler` : '\nalles in Ordnung');
