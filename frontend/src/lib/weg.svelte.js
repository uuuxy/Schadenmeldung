import { flushSync } from 'svelte';

/** Ein kleiner Router über die History-Schnittstelle; der Server liefert für jeden Weg die index.html. */
export const weg = $state({ pfad: location.pathname });

export const wenigBewegung = matchMedia('(prefers-reduced-motion: reduce)');
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
	if (history.state?.zurueck) history.back();
	else gehe(ersatz, { ersetzen: true });
}

window.addEventListener('popstate', () => {
	weg.pfad = location.pathname;
	const y = history.state?.scrollY ?? 0;
	// Erst wenn die Liste wieder steht, lässt sich an ihre alte Stelle rollen.
	requestAnimationFrame(() => window.scrollTo({ top: y }));
});
