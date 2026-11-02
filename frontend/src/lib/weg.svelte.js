import { flushSync } from 'svelte';

/** Ein kleiner Router über die History-Schnittstelle; der Server liefert für jeden Weg die index.html. */
export const weg = $state({ pfad: location.pathname });

const wenigBewegung = matchMedia('(prefers-reduced-motion: reduce)');

export function gehe(pfad, { ersetzen = false } = {}) {
	if (pfad === weg.pfad) return;
	const wechsel = () => {
		history[ersetzen ? 'replaceState' : 'pushState']({}, '', pfad);
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

window.addEventListener('popstate', () => {
	weg.pfad = location.pathname;
});
