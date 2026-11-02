// Misst jede Tippfläche im Handyformat: sichtbare Größe und tatsächliche Trefferfläche (inkl.
// unsichtbarer Erweiterung über ::before), dazu Überlappungen zwischen Nachbarn.
import { chromium } from 'playwright-core';
const B = 'http://127.0.0.1:8090';
const browser = await chromium.launch();
async function kontext(name) {
	const k = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, isMobile: true, hasTouch: true, locale: 'de-DE' });
	if (name) await k.request.post(`${B}/api/anmelden`, { headers: { 'X-Schaden': '1' }, data: { email: name, passwort: 'x' } });
	return k;
}
const messen = () => {
	const sel = 'button, a[href], input:not([type=hidden]):not([type=file]), label:has(input[type=file]), textarea';
	const liste = [];
	for (const el of document.querySelectorAll(sel)) {
		const r = el.getBoundingClientRect();
		if (r.width === 0 || r.height === 0 || getComputedStyle(el).visibility === 'hidden') continue;
		let h = r.height, w = r.width, top = r.top, left = r.left;
		const vor = getComputedStyle(el, '::before');
		if (vor.content !== 'none' && vor.position === 'absolute') {
			const vh = parseFloat(vor.height) || 0, vw = parseFloat(vor.width) || 0;
			if (vh > h) { top = r.top + (r.height - vh) / 2; h = vh; }
			if (vw > w) { left = r.left + (r.width - vw) / 2; w = vw; }
		}
		const name = (el.getAttribute('aria-label') || el.textContent || el.placeholder || el.type || '').trim().replace(/\s+/g, ' ').slice(0, 34);
		liste.push({ name, sichtbar: `${Math.round(r.width)}×${Math.round(r.height)}`, w: Math.round(w), h: Math.round(h), box: { top, left, bottom: top + h, right: left + w } });
	}
	const ueber = [];
	for (let i = 0; i < liste.length; i++) for (let j = i + 1; j < liste.length; j++) {
		const a = liste[i].box, b = liste[j].box;
		const dx = Math.min(a.right, b.right) - Math.max(a.left, b.left), dy = Math.min(a.bottom, b.bottom) - Math.max(a.top, b.top);
		if (dx > 1 && dy > 1) ueber.push(`${liste[i].name} ↔ ${liste[j].name} (${Math.round(dx)}×${Math.round(dy)} px)`);
	}
	return { liste: liste.map(({ box, ...x }) => x), ueber };
};
const seiten = [
	['Anmelden', null, async (s) => s.goto(B + '/')],
	['Melden (leer)', 's.schmidt', async (s) => s.goto(B + '/melden')],
	['Melden (Art gewählt)', 's.schmidt', async (s) => { await s.goto(B + '/melden'); await s.getByRole('button', { name: '101' }).click(); await s.getByRole('button', { name: 'Licht & Strom' }).click(); await s.waitForTimeout(500); }],
	['Aufgaben', 'hausmeister', async (s) => s.goto(B + '/')],
	['Kontomenü', 'hausmeister', async (s) => { await s.goto(B + '/'); await s.getByRole('button', { name: 'Konto und Einstellungen' }).click(); await s.waitForTimeout(400); }],
	['Personen', 'verwaltung', async (s) => { await s.goto(B + '/verwaltung'); await s.getByRole('tab', { name: 'Personen' }).click(); await s.waitForTimeout(400); }]
];
let gesamt = 0, zuKlein = 0;
for (const [titel, name, hin] of seiten) {
	const k = await kontext(name); const s = await k.newPage();
	await hin(s); await s.waitForTimeout(500);
	if (titel === 'Kontomenü') { /* nur das Menü messen */ }
	const { liste, ueber } = await s.evaluate(messen);
	const klein = liste.filter((e) => e.h < 48 || e.w < 48);
	gesamt += liste.length; zuKlein += klein.length;
	console.log(`\n${titel}: ${liste.length} Tippflächen, ${klein.length} unter 48 px`);
	for (const e of klein) console.log(`  ${e.h < 24 || e.w < 24 ? '✗✗' : '✗ '} ${e.name.padEnd(34)} sichtbar ${e.sichtbar}, Treffer ${e.w}×${e.h}`);
	if (ueber.length) console.log(`  Überlappende Trefferflächen: ${ueber.slice(0, 6).join('; ')}${ueber.length > 6 ? ' …' : ''}`);
	await k.close();
}
console.log(`\nGesamt: ${gesamt} Tippflächen, davon ${zuKlein} unter 48 px`);
await browser.close();
