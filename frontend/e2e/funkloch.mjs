import { chromium } from 'playwright-core';
// Funkloch beim Hausmeister im Handyformat: Ein Stand, der ohne Netz gesetzt wird, bleibt stehen, auch
// nach dem Neuladen, steht als „wartet auf Netz“ in der Liste und geht raus, sobald wieder Netz da ist.
// „Rückgängig“ nimmt eine wartende Änderung zurück, ohne dass etwas beim Server ankommt.
// Braucht den lokalen Server mit Testdaten aus e2e/testdaten.sh.
const B = process.env.ADRESSE ?? 'http://127.0.0.1:8090';
const H = { 'X-Schaden': '1' };
const browser = await chromium.launch();
const fehler = [];
const pruefe = (bed, text) => { console.log((bed ? '✓ ' : '✗ ') + text); if (!bed) fehler.push('Prüfung: ' + text); };
async function angemeldet(email, optionen = {}) {
	const k = await browser.newContext(optionen);
	await k.request.post(`${B}/api/anmelden`, { headers: H, data: { email, passwort: 'x' } });
	return k;
}

// Zwei eigene Meldungen, damit der Test nicht an den übrigen Testdaten hängt; geprüft wird am Server
// über eine zweite Sitzung, die vom Funkloch des Handys nichts merkt.
const lehrer = await angemeldet('m.mustermann');
const raum = (await (await lehrer.request.get(`${B}/api/raeume`)).json())[0];
const melde = async (was) => (await lehrer.request.post(`${B}/api/meldungen`, { headers: H, multipart: { raumId: String(raum.id), kategorie: 'licht', was } })).json();
const keller = await melde('Funkloch Keller');
const flur = await melde('Funkloch Flur');
const pruefer = await angemeldet('hausmeister');
const standAmServer = async (id) => {
	const d = await (await pruefer.request.get(`${B}/api/meldungen`)).json();
	return [...d.offen, ...d.erledigt].find((m) => m.id === id)?.status;
};

const k = await angemeldet('hausmeister', { viewport: { width: 390, height: 844 }, deviceScaleFactor: 2, isMobile: true, hasTouch: true, locale: 'de-DE' });
const s = await k.newPage();
s.on('pageerror', (e) => fehler.push(String(e)));
await s.goto(B + '/');
await s.getByRole('group', { name: 'Ansicht' }).waitFor();
// Ohne Service Worker startete die App im Funkloch gar nicht erst.
await s.evaluate(() => navigator.serviceWorker.ready);
const zeile = (was) => s.getByRole('region', { name: 'Meldungen' }).getByRole('listitem').filter({ hasText: was });
await zeile('Funkloch Keller').waitFor();

// 1. Ohne Netz abhaken: Der Stand bleibt stehen und wartet im Ausgang.
await k.setOffline(true);
await s.getByRole('button', { name: `Erledigt: ${raum.name}, Funkloch Keller` }).click();
await s.getByText(`Erledigt: ${raum.name} · wird nachgereicht`).waitFor({ timeout: 5000 });
// Die Zeile gleitet noch hinaus, während die Snackbar schon steht.
const weg = await zeile('Funkloch Keller').waitFor({ state: 'detached', timeout: 3000 }).then(() => true, () => false);
pruefe(weg, 'Ohne Netz abgehakt: nicht mehr unter „Offen“, die Snackbar sagt „wird nachgereicht“');

// 2. Neu laden ohne Netz: Der Stand bleibt, die Zeile sagt „wartet auf Netz“, beim Server ist nichts.
await s.reload();
await s.getByRole('tab', { name: /^Erledigt/ }).click();
await zeile('Funkloch Keller').waitFor({ timeout: 5000 });
pruefe((await zeile('Funkloch Keller').getByText('wartet auf Netz').count()) === 1, 'Nach dem Neuladen ohne Netz: unter „Erledigt“ mit „wartet auf Netz“');
if (process.argv[2]) await s.screenshot({ path: `${process.argv[2]}/funkloch-wartet.png` });
pruefe((await standAmServer(keller.id)) === 'offen', 'Der Server hat noch nichts bekommen');

// 3. „Rückgängig“ ohne Netz: Die wartende Änderung fällt weg.
await s.getByRole('tab', { name: /^Offen/ }).click();
await zeile('Funkloch Flur').getByRole('button').first().click();
await s.getByRole('group', { name: 'Stand' }).getByRole('button', { name: 'In Arbeit' }).click();
await s.getByText(`In Arbeit: ${raum.name} · wird nachgereicht`).waitFor({ timeout: 5000 });
pruefe((await zeile('Funkloch Flur').getByText('wartet auf Netz').count()) === 1, '„In Arbeit“ ohne Netz: zurück in der Liste, „wartet auf Netz“');
await s.getByRole('button', { name: 'Rückgängig' }).click();
await s.waitForTimeout(300);
pruefe((await zeile('Funkloch Flur').getByText(/wartet auf Netz|In Arbeit/).count()) === 0, '„Rückgängig“ ohne Netz nimmt die wartende Änderung zurück');

// 4. Wieder Netz: Was wartet, geht von selbst raus.
await k.setOffline(false);
await s.getByText('Änderung nachgereicht').waitFor({ timeout: 10000 });
pruefe((await standAmServer(keller.id)) === 'erledigt', 'Wieder Netz: „Erledigt“ ist beim Server angekommen');
pruefe((await standAmServer(flur.id)) === 'offen', 'Die zurückgenommene Änderung kam nie beim Server an');

for (const m of [keller, flur]) await pruefer.request.delete(`${B}/api/meldungen/${m.id}`, { headers: H });
await browser.close();
if (fehler.length) {
	console.log(fehler.join('\n'));
	process.exit(1);
}
console.log('alles in Ordnung');
