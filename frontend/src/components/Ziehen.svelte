<script>
	import Icon from './Icon.svelte';
	import Ladekreis from './Ladekreis.svelte';

	/**
	 * Ziehen zum Aktualisieren, wie in jeder Listen-App: ganz oben nach unten ziehen, loslassen. Nur
	 * am Handy (Touch); am Rechner frischt die App ohnehin alle 30 Sekunden auf.
	 */
	let { aktualisieren, children } = $props();

	const SCHWELLE = 64;
	let zug = $state(0);
	let laeuft = $state(false);
	let start = $state(null);

	$effect(() => {
		const beginn = (e) => {
			start = window.scrollY <= 0 && !laeuft && e.touches.length === 1 ? e.touches[0].clientY : null;
		};
		const bewegung = (e) => {
			if (start === null) return;
			// Halber Weg mit Deckel: Der Widerstand fühlt sich an wie ein Gummiband.
			zug = Math.min(112, Math.max(0, (e.touches[0].clientY - start) * 0.5));
		};
		const ende = async () => {
			if (start === null) return;
			start = null;
			if (zug < SCHWELLE) {
				zug = 0;
				return;
			}
			laeuft = true;
			zug = 56;
			try {
				await aktualisieren();
			} finally {
				laeuft = false;
				zug = 0;
			}
		};
		window.addEventListener('touchstart', beginn, { passive: true });
		window.addEventListener('touchmove', bewegung, { passive: true });
		window.addEventListener('touchend', ende);
		window.addEventListener('touchcancel', ende);
		return () => {
			window.removeEventListener('touchstart', beginn);
			window.removeEventListener('touchmove', bewegung);
			window.removeEventListener('touchend', ende);
			window.removeEventListener('touchcancel', ende);
		};
	});
</script>

<div
	class={['pointer-events-none fixed left-1/2 z-30 grid size-10 place-items-center rounded-full bg-surface-container-high text-primary shadow-2', start === null && 'transition-[top,opacity] duration-300 ease-standard']}
	style:top="calc(env(safe-area-inset-top) + 4rem + {zug - 48}px)"
	style:opacity={Math.min(1, zug / 40)}
	style:translate="-50% 0"
	aria-hidden={!laeuft}
>
	{#if laeuft}
		<Ladekreis groesse={22} />
	{:else}
		<span style:rotate="{zug >= SCHWELLE ? 180 : 0}deg" class="transition-[rotate] duration-200">
			<Icon name="arrow_downward" groesse={22} />
		</span>
	{/if}
</div>

{@render children?.()}
