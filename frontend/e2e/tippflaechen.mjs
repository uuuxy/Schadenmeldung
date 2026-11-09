// Misst jede Seite in sieben Bildschirmgrößen, vom kleinen Android bis zum Tablet quer: jede Tippfläche
// (sichtbar und tatsächliche Trefferfläche inkl. ::before), Überlappungen zwischen Nachbarn, ob die
// Seite seitlich über den Rand ragt und ob Eingabefelder unter 16 px Schrift haben (dann zoomt das iPhone).
import { chromium } from 'playwright-core';
const B = 'http://127.0.0.1:8090';
const browser = await chromium.launch();
// 320: kleinstes verbreitetes Android, 360: übliches Android, 390: iPhone 13–16, 430: großes iPhone,
// 844 × 390: Handy quer, 768 und 1024: Tablet hoch und quer.
const groessen = [[320, 640], [360, 780], [390, 844], [430, 932], [844, 390], [768, 1024], [1024, 768]];
async function kontext(name, width, height) {
	const k = await browser.newContext({ viewport: { width, height }, deviceScaleFactor: 2, isMobile: true, hasTouch: true, locale: 'de-DE' });
	if (name) await k.request.post(`${B}/api/anmelden`, { headers: { 'X-Schaden': '1' }, data: { email: name, passwort: 'x' } });
	return k;
}
const messen = () => {
	const sel = 'button, a[href], input:not([type=hidden]):not([type=file]), label:has(input[type=file]), textarea';
	const liste = [];
	// Ist ein Blatt, Dialog oder Menü offen, zählt nur, was darin liegt: Der Rest ist nicht antippbar.
	const oben = document.querySelector('dialog[open], :popover-open') ?? document;
	// Unter einer festen oder mitlaufenden Leiste (Kopfleiste, Melden, Erledigt) rollt der Inhalt durch;
	// das ist keine Überlappung.
	const fest = (el) => { for (let e = el; e; e = e.parentElement) if (['fixed', 'sticky'].includes(getComputedStyle(e).position)) return true; return false; };
	for (const el of oben.querySelectorAll(sel)) {
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
		liste.push({ name, fest: fest(el), feld: el.matches('input, textarea'), sichtbar: `${Math.round(r.width)}×${Math.round(r.height)}`, w: Math.round(w), h: Math.round(h), box: { top, left, bottom: top + h, right: left + w } });
	}
	const ueber = [];
	const innen = (x, y) => y.box.left >= x.box.left - 1 && y.box.right <= x.box.right + 1 && y.box.top >= x.box.top - 1 && y.box.bottom <= x.box.bottom + 1;
	for (let i = 0; i < liste.length; i++) for (let j = i + 1; j < liste.length; j++) {
		if (liste[i].fest !== liste[j].fest) continue;
		// Ein Symbolknopf im Feld (Passwort zeigen) liegt mit Absicht darin, wie bei M3.
		if ((liste[i].feld && innen(liste[i], liste[j])) || (liste[j].feld && innen(liste[j], liste[i]))) continue;
		const a = liste[i].box, b = liste[j].box;
		const dx = Math.min(a.right, b.right) - Math.max(a.left, b.left), dy = Math.min(a.bottom, b.bottom) - Math.max(a.top, b.top);
		if (dx > 1 && dy > 1) ueber.push(`${liste[i].name} ↔ ${liste[j].name} (${Math.round(dx)}×${Math.round(dy)} px)`);
	}
	const felder = [...document.querySelectorAll('input:not([type=hidden]):not([type=file]):not([type=checkbox]):not([type=radio]), textarea, select')];
	const kleineSchrift = felder.filter((f) => f.getBoundingClientRect().width > 0 && parseFloat(getComputedStyle(f).fontSize) < 16)
		.map((f) => f.getAttribute('aria-label') || f.placeholder || f.name || f.type);
	const ueberstand = document.documentElement.scrollWidth - window.innerWidth;
	return { liste: liste.map(({ box, fest, feld, ...x }) => x), ueber, ueberstand, kleineSchrift };
};
const seiten = [
	['Anmelden', null, async (s) => s.goto(B + '/')],
	['Melden (leer)', 's.schmidt', async (s) => s.goto(B + '/melden')],
	['Melden (Art gewählt)', 's.schmidt', async (s) => { await s.goto(B + '/melden'); await s.getByRole('button', { name: '101' }).click(); await s.getByRole('button', { name: 'Licht & Strom' }).click(); await s.waitForTimeout(500); }],
	['Aufgaben', 'hausmeister', async (s) => s.goto(B + '/')],
	['Kontomenü', 'hausmeister', async (s) => { await s.goto(B + '/'); await s.getByRole('button', { name: 'Konto und Einstellungen' }).click(); await s.waitForTimeout(400); }],
	['Meldung', 'hausmeister', async (s) => { await s.goto(B + '/'); await s.getByRole('button', { name: /^Bio 1/ }).click(); await s.waitForTimeout(400); }],
	['Blatt Export', 'hausmeister', async (s) => { await s.goto(B + '/'); await s.getByRole('button', { name: 'Exportieren' }).click(); await s.waitForTimeout(500); }],
	['Blatt Bereich', 'hausmeister', async (s) => { await s.goto(B + '/'); await s.getByRole('button', { name: 'Bereich' }).click(); await s.waitForTimeout(500); }],
	['Räume', 'admin', async (s) => { await s.goto(B + '/einstellungen'); await s.waitForTimeout(400); }],
	['Arten', 'admin', async (s) => { await s.goto(B + '/einstellungen'); await s.getByRole('tab', { name: 'Arten' }).click(); await s.waitForTimeout(400); }],
	['Personen', 'admin', async (s) => { await s.goto(B + '/einstellungen'); await s.getByRole('tab', { name: 'Personen' }).click(); await s.waitForTimeout(400); }],
	['Mail', 'admin', async (s) => { await s.goto(B + '/einstellungen'); await s.getByRole('tab', { name: 'Mail' }).click(); await s.waitForTimeout(400); }]
];
let gesamt = 0, zuKlein = 0, auffaellig = 0;
for (const [breite, hoehe] of groessen) {
	let n = 0, klein = 0;
	const funde = [];
	for (const [titel, name, hin] of seiten) {
		const k = await kontext(name, breite, hoehe); const s = await k.newPage();
		s.setDefaultTimeout(5000);
		try {
			await hin(s); await s.waitForTimeout(500);
			const m = await s.evaluate(messen);
			const zk = m.liste.filter((e) => e.h < 48 || e.w < 48);
			n += m.liste.length; klein += zk.length;
			for (const e of zk) funde.push(`${titel}: ${e.name} sichtbar ${e.sichtbar}, Treffer ${e.w}×${e.h}`);
			if (m.ueber.length) funde.push(`${titel}: überlappend ${m.ueber.slice(0, 3).join('; ')}`);
			if (m.ueberstand > 1) funde.push(`${titel}: ragt ${m.ueberstand} px seitlich über den Rand`);
			for (const f of m.kleineSchrift) funde.push(`${titel}: Feld „${f}“ unter 16 px Schrift`);
		} catch (e) {
			funde.push(`${titel}: nicht erreichbar (${e.message.split('\n')[0].slice(0, 80)})`);
		}
		await k.close();
	}
	gesamt += n; zuKlein += klein; auffaellig += funde.length;
	console.log(`${breite}×${hoehe}: ${n} Tippflächen, ${klein} unter 48 px, ${funde.length} Auffälligkeiten`);
	for (const f of funde) console.log(`   ${f}`);
}
console.log(`\nGesamt: ${gesamt} Tippflächen in ${groessen.length} Größen, davon ${zuKlein} unter 48 px; ${auffaellig} Auffälligkeiten`);
await browser.close();
