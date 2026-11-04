// Neue Fassung ohne Zutun: Handys halten die App oft tagelang im Hintergrund offen, dann liefe nach
// einem Update die alte Fassung weiter. Bei jedem Wechsel in den Hintergrund und zurück fragt die App
// deshalb kurz nach und lädt sich selbst neu, am liebsten unbemerkt im Hintergrund, aber nie mitten
// in einer Eingabe.

const halter = new Set();
const getippt = new Set();
let neu = false;
let zuletztBedient = Date.now();
let dateiwahl = 0;

/**
 * Meldet an, was ein Neuladen verwerfen würde, etwa eine angetippte Kachel; getippten Text sieht die
 * App selbst. Gibt die Abmeldung zurück und passt so in ein $effect: `$effect(() => haltFest(() => x))`.
 * @param {() => boolean} hatEingabe
 */
export function haltFest(hatEingabe) {
	halter.add(hatEingabe);
	return () => halter.delete(hatEingabe);
}

function etwasGetippt() {
	for (const feld of getippt) {
		if (!feld.isConnected) getippt.delete(feld);
		else if (feld.value.trim()) return true;
	}
	return false;
}

// Wer ein Foto macht oder eine Datei wählt, verlässt die App nur kurz und will seine Seite zurück.
const beschaeftigt = () => Date.now() - dateiwahl < 120_000 || etwasGetippt() || [...halter].some((hat) => hat());

// Die Fassung steckt im gehashten Namen des Startskripts. Die Frage geht am Speicher des Service
// Workers vorbei (sw.js), sonst läge dort eine index.html, deren Dateien im Funkloch fehlen.
const startskript = (dok) => dok.querySelector('script[type="module"][src]')?.getAttribute('src');

async function pruefe() {
	if (neu) return;
	try {
		const antwort = await fetch('/', { cache: 'no-store' });
		if (!antwort.ok) return;
		const dort = startskript(new DOMParser().parseFromString(await antwort.text(), 'text/html'));
		neu = !!dort && dort !== startskript(document);
	} catch {
		// Funkloch: Es bleibt bei der laufenden Fassung.
	}
}

// Wer während der Nachfrage schon wieder tippt, wird nicht unterbrochen: Dann kommt die neue Fassung
// beim nächsten Wechsel.
async function ladeNeuWennFrei() {
	const seit = Date.now();
	await pruefe();
	if (neu && zuletztBedient < seit && !beschaeftigt()) location.reload();
}

document.addEventListener('visibilitychange', ladeNeuWennFrei);
for (const art of ['pointerdown', 'keydown']) {
	document.addEventListener(art, () => (zuletztBedient = Date.now()), { capture: true, passive: true });
}
document.addEventListener(
	'input',
	(e) => {
		const feld = /** @type {HTMLInputElement} */ (e.target);
		if (feld.matches?.('textarea, input:not([type=checkbox], [type=radio], [type=file])')) getippt.add(feld);
	},
	true
);
document.addEventListener(
	'click',
	(e) => {
		if (e.target instanceof HTMLInputElement && e.target.type === 'file') dateiwahl = Date.now();
	},
	true
);

// Bleibt die App ständig sichtbar, etwa die Übersicht auf einem Bildschirm im Büro, kommt die neue
// Fassung, sobald fünf Minuten lang niemand etwas tut.
setInterval(() => Date.now() - zuletztBedient > 300_000 && ladeNeuWennFrei(), 300_000);
