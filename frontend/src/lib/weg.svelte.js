import { flushSync } from 'svelte';

/** Ein kleiner Router über die History-Schnittstelle; der Server liefert für jeden Weg die index.html. */
export const weg = $state({ pfad: location.pathname });

export const wenigBewegung = matchMedia('(prefers-reduced-motion: reduce)');

// Der Raum-Code aus einem Pfad wie /r/abc: für den Link im QR-Aufkleber und für den Scanner in der App,
// damit beide denselben Code lesen. Ein kaputtes %-Zeichen ergibt keinen Code statt eines Fehlers.
export function raumCodeAus(pfad) {
	if (!pfad.startsWith('/r/')) return '';
	try {
		return decodeURIComponent(pfad.slice(3));
	} catch {
		return '';
	}
}
// Die Stelle in der Liste stellt die App beim Zurückgehen selbst wieder her (siehe popstate).
history.scrollRestoration = 'manual';

export function gehe(pfad, { ersetzen = false } = {}) {
	if (pfad === weg.pfad) return;
	history.replaceState({ ...history.state, scrollY: window.scrollY }, '');
	const wechsel = () => {
		// zurueck: Dieser Eintrag kam aus der App, „Zurück“ darf also einen Schritt in der History gehen.
		history[ersetzen ? 'replaceState' : 'pushState']({ zurueck: !ersetzen || !!history.state?.zurueck }, '', pfad);
		weg.pfad = pfad;
		window.scrollTo({ top: 0 });
	};
	// Seitenwechsel mit sanftem Überblenden (M3 „fade through“), wo der Browser View Transitions kann.
	if (document.startViewTransition && !wenigBewegung.matches) {
		document.startViewTransition(() => flushSync(wechsel));
	} else {
		wechsel();
	}
}

/** Zurück, wie der Pfeil in der Kopfleiste: in der History, wenn die Seite aus der App kam, sonst nach ersatz. */
export function zurueck(ersatz = '/') {
	// Über offene Unteransichten hinweg, denn der Pfeil der Kopfleiste verlässt die ganze Seite.
	if (history.state?.zurueck) history.go(-1 - (history.state.tiefe ?? 0));
	else gehe(ersatz, { ersetzen: true });
}

/*
 * Unteransichten (ein Bereich, „Anderer Ort“) haben einen eigenen Eintrag im Verlauf, solange sie offen
 * sind. So führen die Zurück-Geste von Android und das Wischen in Safari genau eine Ebene zurück, wie
 * ihr Pfeil. Zu einem Eintrag ohne offene Unteransicht (nach dem Neuladen, nach einem Seitenwechsel)
 * geht die App gleich weiter zurück, sonst täte „Zurück“ einmal sichtbar nichts.
 */
const offen = new Set();
// Ein stiller Schritt räumt einen Eintrag ab, ohne die Seite zu bewegen.
let still = false;

function verwaist() {
	const unter = history.state?.unter;
	return unter !== undefined && !offen.has(unter);
}

function raeumeAuf() {
	// Erst wenn Svelte fertig ist, steht fest, ob eine neue Unteransicht den Eintrag übernommen hat.
	setTimeout(() => {
		if (!verwaist()) return;
		still = true;
		history.back();
	});
}

/**
 * Für die Zeile mit dem Pfeil (ZurueckZeile): legt den Eintrag an; schliessen läuft, sobald er
 * verlassen wird. zurueck() ist der Pfeil, ende() räumt den Eintrag ab, wenn die Unteransicht anders
 * verschwindet (Raum gewählt, Meldung abgeschickt).
 */
export function unteransicht(schliessen) {
	const id = Math.random();
	let tiefe = history.state?.tiefe;
	if (verwaist()) {
		// Eine eben verschwundene Unteransicht gibt ihren Eintrag weiter, statt dass ein zweiter entsteht.
		history.replaceState({ ...history.state, unter: id }, '');
	} else {
		tiefe = (tiefe ?? 0) + 1;
		history.replaceState({ ...history.state, scrollY: window.scrollY }, '');
		history.pushState({ ...history.state, unter: id, tiefe }, '');
	}
	offen.add(id);
	const verlassen = () => (history.state?.tiefe ?? 0) < tiefe && schliessen();
	window.addEventListener('popstate', verlassen);
	// Ein zweiter Tipp, bevor der Browser zurückgeblättert hat, ginge sonst noch einen Schritt weiter.
	let unterwegs = false;
	return {
		zurueck() {
			if (history.state?.unter !== id) schliessen();
			else if (!unterwegs) history.back();
			unterwegs = true;
		},
		ende() {
			window.removeEventListener('popstate', verlassen);
			offen.delete(id);
			raeumeAuf();
		}
	};
}

window.addEventListener('popstate', () => {
	weg.pfad = location.pathname;
	if (verwaist()) return history.back();
	if (still) return (still = false);
	const y = history.state?.scrollY ?? 0;
	// Erst wenn die Liste wieder steht, lässt sich an ihre alte Stelle rollen.
	requestAnimationFrame(() => window.scrollTo({ top: y }));
});
raeumeAuf();
