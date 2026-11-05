import { slide } from 'svelte/transition';
import { wenigBewegung } from './weg.svelte.js';

/** Kurve wie CSS cubic-bezier(x1, y1, x2, y2) als Funktion, denn Svelte-Übergänge nehmen nur Funktionen. */
function kurve(x1, y1, x2, y2) {
	const bezier = (a, b, t) => 3 * a * t * (1 - t) ** 2 + 3 * b * t * t * (1 - t) + t ** 3;
	return (x) => {
		// x wächst mit t (x1 und x2 liegen zwischen 0 und 1), also findet Halbieren das passende t.
		let [unten, oben] = [0, 1];
		for (let i = 0; i < 20; i++) {
			const t = (unten + oben) / 2;
			if (bezier(x1, x2, t) < x) unten = t;
			else oben = t;
		}
		return bezier(y1, y2, (unten + oben) / 2);
	};
}

// M3 Expressive fürs Web: Google rechnet seine Federn in Kurven um (m3.material.io, Motion specs,
// „Convert springs to curves“). Räumliche Bewegung schwingt leicht über das Ziel hinaus.
const raeumlich = kurve(0.38, 1.21, 0.22, 1); // default spatial, 500 ms
const raeumlichSchnell = kurve(0.42, 1.67, 0.21, 0.9); // fast spatial, 350 ms (wie --ease-spring)
const dauer = (ms) => (wenigBewegung.matches ? 0 : ms);

/** Eine Frage klappt federnd auf; schnell für kleine Teile wie die Zeile mit „Ändern“. */
export const aufklappen = (node, { schnell = false } = {}) =>
	slide(node, { duration: dauer(schnell ? 350 : 500), easing: schnell ? raeumlichSchnell : raeumlich });

/** Zuklappen federt nicht: Was geht, soll nur kurz Platz machen. */
export const zuklappen = (node) => slide(node, { duration: dauer(200) });

/** Für den einen Moment am Ende einer Meldung: Das Häkchen federt herein, die Form dahinter dreht sich ein. */
export const hereinfedern = () => ({ duration: dauer(500), easing: raeumlich, css: (t) => `scale: ${0.4 + 0.6 * t}; opacity: ${Math.min(1, 2 * t)}` });
export const eindrehen = () => ({ duration: dauer(500), easing: raeumlich, css: (t) => `rotate: ${(t - 1) * 90}deg` });
