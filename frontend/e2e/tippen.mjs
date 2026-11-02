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
const haken = s.getByRole('button', { name: /^Erledigt: / }).first();
const name = await haken.getAttribute('aria-label');
wo = await tippeNeben(s, haken, 'unten', 3);
await s.waitForTimeout(400);
pruefe(!(await s.getByRole('button', { name }).isVisible()) && (await s.getByText(/Erledigt: /).isVisible()), `Häkchen „Erledigt“ in der Liste (${wo})`);
await s.getByRole('button', { name: 'Rückgängig' }).click();
await s.getByRole('button', { name: /^Bio 1/ }).click();
// Erst nach dem Überblenden (View Transition, 0,3 s) nimmt die neue Seite Tipps an.
await s.waitForTimeout(500);
const gesehen = s.getByRole('group', { name: 'Stand' }).getByRole('button', { name: 'Gesehen' });
wo = await tippeNeben(s, gesehen, 'oben', 3);
await s.waitForTimeout(500);
pruefe((await gesehen.getAttribute('aria-pressed')) === 'true', `Stand „Gesehen“ auf der Meldung (${wo})`);
await s.getByRole('button', { name: 'Zurück' }).click();
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
