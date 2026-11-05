<script>
	import Icon from './Icon.svelte';

	/**
	 * Vom linken Rand wischen führt zurück, wie in jeder iPhone-App. Safari bringt das über den Verlauf
	 * mit, die App vom Home-Bildschirm nicht, deshalb baut sie es dort selbst ein (App.svelte). Die Geste
	 * tippt den innersten Zurück-Pfeil an oder schließt ein offenes Blatt, wie die Zurück-Geste von
	 * Android: So führt „zurück“ überall gleich weit. Wo es keinen Pfeil gibt, passiert nichts.
	 */
	// Der Streifen, in dem die Geste beginnt, ist etwas breiter als der Seitenrand (16 px).
	const RAND = 24;
	const SCHWELLE = 80;
	let zug = $state(0);
	let hoehe = $state(0);
	let start = $state(null);

	// Ein offenes Blatt oder eine Rückfrage zuerst, sonst der innerste Pfeil: Er steht im Dokument zuletzt.
	const ziel = () => [...document.querySelectorAll('dialog[open]')].at(-1) ?? [...document.querySelectorAll('[data-zurueck]')].at(-1);

	$effect(() => {
		let seitlich = false;
		const bewegung = (e) => {
			const dx = e.touches[0].clientX - start.x;
			const dy = e.touches[0].clientY - start.y;
			if (!seitlich) {
				// Erst wenn klar ist, dass seitlich gewischt wird, gehört die Bewegung der Geste; sonst rollt die Seite.
				if (Math.abs(dy) > 10 && Math.abs(dy) > dx) return aufhoeren();
				if (dx < 10) return;
				seitlich = true;
			}
			e.preventDefault();
			zug = Math.min(160, Math.max(0, dx));
			hoehe = e.touches[0].clientY;
		};
		const aufhoeren = () => {
			window.removeEventListener('touchmove', bewegung);
			window.removeEventListener('popstate', aufhoeren);
			start = null;
			seitlich = false;
			zug = 0;
		};
		const beginn = (e) => {
			const t = e.touches[0];
			if (e.touches.length !== 1 || t.clientX > RAND || !ziel()) return;
			start = { x: t.clientX, y: t.clientY };
			hoehe = t.clientY;
			window.addEventListener('touchmove', bewegung, { passive: false });
			// Blättert das System selbst zurück, solange der Finger liegt, macht die Geste keinen zweiten Schritt.
			window.addEventListener('popstate', aufhoeren);
		};
		const ende = () => {
			if (!start) return;
			// Wer vor dem Loslassen wieder nach links zieht, bricht ab.
			const z = zug >= SCHWELLE && ziel();
			aufhoeren();
			if (z instanceof HTMLDialogElement) z.close();
			else if (z) z.click();
		};
		window.addEventListener('touchstart', beginn, { passive: true });
		window.addEventListener('touchend', ende);
		window.addEventListener('touchcancel', aufhoeren);
		return () => {
			aufhoeren();
			window.removeEventListener('touchstart', beginn);
			window.removeEventListener('touchend', ende);
			window.removeEventListener('touchcancel', aufhoeren);
		};
	});
</script>

<!-- Der Pfeil kommt mit dem Finger vom Rand; ab der Schwelle wird er voll, dann reicht Loslassen. -->
<div
	class={[
		'pointer-events-none fixed left-0 z-30 grid size-10 place-items-center rounded-full shadow-2',
		zug >= SCHWELLE ? 'bg-primary text-on-primary' : 'bg-surface-container-high text-primary',
		!start && 'transition-[translate,opacity] duration-300 ease-standard'
	]}
	style:top="{hoehe - 20}px"
	style:translate="{Math.min(zug, SCHWELLE) / 2 - 32}px 0"
	style:opacity={Math.min(1, zug / 40)}
	aria-hidden="true"
>
	<Icon name="arrow_back" groesse={22} />
</div>
