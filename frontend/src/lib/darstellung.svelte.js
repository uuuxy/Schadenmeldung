// „Dunkles Design“ an oder aus, gewählt je Gerät – dieselbe Regel wie in der Bibliothek
// (stores/darstellung.svelte.js) und wie bei Google: Ohne Wahl gilt das Farbschema des Systems,
// eine Wahl gilt in diesem Browser auf diesem Gerät. public/darstellung.js setzt die Klasse schon
// vor dem ersten Bild nach derselben Regel.
import { MediaQuery } from 'svelte/reactivity';

const SCHLUESSEL = 'schaden.darstellung';

const leseWahl = (gespeichert) => (gespeichert === 'hell' || gespeichert === 'dunkel' ? gespeichert : 'system');

function gespeichert() {
	try {
		return localStorage.getItem(SCHLUESSEL);
	} catch {
		return null;
	}
}

let wahl = $state(leseWahl(gespeichert()));
const system = new MediaQuery('(prefers-color-scheme: dark)');

export const darstellung = {
	/** Steht die Seite dunkel? Ohne Wahl entscheidet das System. */
	get dunkel() {
		return wahl === 'dunkel' || (wahl === 'system' && system.current);
	},

	schalte(dunkel) {
		wahl = dunkel ? 'dunkel' : 'hell';
		try {
			localStorage.setItem(SCHLUESSEL, wahl);
		} catch {
			// Ohne Speicher (privates Fenster) gilt die Wahl bis zum Neuladen.
		}
	}
};

// Klasse an <html> und Farbe der Statusleiste aktuell halten: bei einer Wahl, einer Änderung im
// System und einer Wahl in einem anderen Fenster.
$effect.root(() => {
	$effect(() => {
		document.documentElement.classList.toggle('schema-dunkel', darstellung.dunkel);
		document.querySelector('meta[name="theme-color"]')?.setAttribute('content', darstellung.dunkel ? '#121316' : '#faf9fd');
	});
});
window.addEventListener('storage', (e) => {
	if (e.key === SCHLUESSEL) wahl = leseWahl(e.newValue);
});
