import { chromium } from 'playwright-core';
// Dringend auf der Seite einer Meldung im Handyformat: Jede Person mit der Rolle Hausmeister, auch eine
// zweite, setzt und nimmt es bei jeder Meldung, die Übersicht sortiert gleich um. Ohne Netz wartet es im
// Ausgang und geht raus, sobald wieder Netz da ist. Die Verwaltung sieht nur das Abzeichen. Bilder landen
// im Ordner aus argv[2]. Braucht den lokalen Server mit Testdaten aus e2e/testdaten.sh.
const B = process.env.ADRESSE ?? 'http://127.0.0.1:8090';
const H = { 'X-Schaden': '1' };
const browser = await chromium.launch();
const fehler = [];
const pruefe = (bed, text) => { console.log((bed ? '✓ ' : '✗ ') + text); if (!bed) fehler.push('Prüfung: ' + text); };
async function angemeldet(email) {
	const k = await browser.newContext({ viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, isMobile: true, hasTouch: true, locale: 'de-DE' });
	await k.request.post(`${B}/api/anmelden`, { headers: H, data: { email, passwort: 'x' } });
	return k;
}
const bild = (s, name) => process.argv[2] && s.screenshot({ path: `${process.argv[2]}/dringend-${name}.png` });

// Eine zweite Hausmeisterin und eine eigene Meldung aus dem Kollegium, damit der Test nicht an den
// übrigen Testdaten hängt; geprüft wird am Server über eine Admin-Sitzung.
const admin = await angemeldet('admin@schule.de');
await admin.request.post(`${B}/api/rollen`, { headers: H, data: { email: 'hausmeisterin', rolle: 'hausmeister' } });
const lehrer = await angemeldet('m.mustermann');
const raum = (await (await lehrer.request.get(`${B}/api/raeume`)).json())[0];
for (const alt of (await (await admin.request.get(`${B}/api/meldungen`)).json()).offen.filter((m) => m.was === 'Dringend-Probe')) {
	await admin.request.delete(`${B}/api/meldungen/${alt.id}`, { headers: H });
}
const meldung = await (await lehrer.request.post(`${B}/api/meldungen`, { headers: H, multipart: { raumId: String(raum.id), kategorie: 'licht', was: 'Dringend-Probe' } })).json();
const amServer = async () => (await (await admin.request.get(`${B}/api/meldungen`)).json()).offen.find((m) => m.id === meldung.id)?.dringend;

const k = await angemeldet('hausmeisterin');
const s = await k.newPage();
s.on('pageerror', (e) => fehler.push(String(e)));
const chip = s.getByRole('button', { name: 'Dringend', exact: true });
const an = async () => (await chip.getAttribute('aria-pressed')) === 'true';
const ersteZeile = () => s.getByRole('region', { name: 'Meldungen' }).getByRole('listitem').first().textContent();

// 1. Die zweite Hausmeisterin setzt Dringend bei einer fremden Meldung; die Übersicht stellt sie nach oben.
await s.goto(`${B}/meldung/${meldung.id}`);
await chip.waitFor();
pruefe(!(await an()), 'Zweite Hausmeisterin: oben der Chip „Dringend“ wie beim Melden, nicht gewählt');
await bild(s, 'aus');
await chip.click();
await s.waitForTimeout(500);
pruefe((await an()) && (await amServer()) === true, 'Ein Tipp auf „Dringend“ setzt es, auch beim Server');
pruefe(await chip.evaluate((e) => e.classList.contains('bg-error-container')), 'Gesetzt ist der Chip rot (error-container)');
await bild(s, 'an');
await s.getByRole('button', { name: 'Zurück' }).click();
await s.getByRole('heading', { name: 'Dringend' }).waitFor();
pruefe((await ersteZeile()).includes('Dringend-Probe'), 'Übersicht: die Meldung steht oben unter „Dringend“');

// 2. Wegnehmen: zurück unter „Weitere“.
await s.getByRole('button', { name: `${raum.name} Dringend-Probe` }).click();
await chip.click();
await s.waitForTimeout(500);
pruefe(!(await an()) && (await amServer()) === false, 'Ein zweiter Tipp nimmt Dringend wieder weg, auch beim Server');

// 3. Ohne Netz: wartet im Ausgang und geht raus, sobald wieder Netz da ist.
await k.setOffline(true);
await chip.click();
await s.getByText('Dringend · wird nachgereicht').waitFor({ timeout: 5000 });
pruefe((await an()) && (await amServer()) === false, 'Ohne Netz: Chip gesetzt, der Server hat noch nichts');
await k.setOffline(false);
let da = false;
for (let i = 0; i < 20 && !da; i++) da = (await s.waitForTimeout(500), (await amServer()) === true);
pruefe(da, 'Wieder Netz: Dringend ist nachgereicht');

// 4. Die Verwaltung sieht Dringend, kann es aber nicht antippen.
const v = await angemeldet('sekretariat');
const vs = await v.newPage();
await vs.goto(`${B}/meldung/${meldung.id}`);
await vs.getByRole('heading', { name: 'Dringend-Probe' }).waitFor();
pruefe((await vs.getByRole('button', { name: 'Dringend', exact: true }).count()) === 0 && (await vs.locator('main').getByText('Dringend', { exact: true }).isVisible()), 'Verwaltung: nur das Abzeichen „Dringend“, nichts zum Antippen');

await admin.request.delete(`${B}/api/meldungen/${meldung.id}`, { headers: H });
await admin.request.delete(`${B}/api/rollen/hausmeisterin@schule.de`, { headers: H });
await browser.close();
if (fehler.length) {
	console.log(fehler.join('\n'));
	process.exit(1);
}
console.log('alles in Ordnung');
