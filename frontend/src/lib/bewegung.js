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
const raeumlichLangsam = kurve(0.39, 1.29, 0.35, 0.98); // slow spatial, 650 ms
const bremsend = kurve(0.05, 0.7, 0.1, 1); // emphasized decelerate (M3), wie --ease-emphasized-decelerate
const dauer = (ms) => (wenigBewegung.matches ? 0 : ms);

/**
 * Eine Frage klappt auf und steigt dabei federnd an ihren Platz, so weit wie bei M3 „shared axis“ (30 px).
 * Die Feder sitzt im Steigen, dort sieht man sie; die Höhe wächst ohne Überschwingen, sonst zuckte, was
 * darunter steht. schnell für kleine Teile wie die Zeile mit „Ändern“.
 */
export function aufklappen(node, { schnell = false } = {}) {
	const flaeche = slide(node).css;
	const feder = schnell ? raeumlichSchnell : raeumlich;
	return {
		duration: dauer(schnell ? 350 : 500),
		// translate vorn: slide endet ohne Semikolon, dahinter ginge es verloren.
		css: (t) => `translate: 0 ${(1 - feder(t)) * 30}px;${flaeche(bremsend(t), 1 - bremsend(t))}`
	};
}

/** Zuklappen federt nicht: Was geht, soll nur kurz Platz machen. */
export const zuklappen = (node) => slide(node, { duration: dauer(200) });

/*
 * Der eine Moment am Ende einer Meldung (M3: Expressive für die großen Momente, „slow“ für Bewegung über
 * den ganzen Bildschirm). Er beginnt erst, wenn die Seite eingeblendet ist, sonst ginge er darin unter:
 * Die Form dreht sich herein, dann springt das Häkchen auf.
 */
export const eindrehen = () => ({
	delay: dauer(150),
	duration: dauer(650),
	easing: raeumlichLangsam,
	css: (t) => `scale: ${0.3 + 0.7 * t}; rotate: ${(t - 1) * 120}deg; opacity: ${Math.min(1, 3 * t)}`
});
export const aufspringen = () => ({ delay: dauer(400), duration: dauer(350), easing: raeumlichSchnell, css: (t) => `scale: ${t}` });
