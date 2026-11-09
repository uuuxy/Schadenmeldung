<script>
	import { slide } from 'svelte/transition';
	import Icon from './Icon.svelte';
	import KontoKnopf from './KontoKnopf.svelte';
	import Segmente from './ui/Segmente.svelte';
	import { zustand, darfSehen } from '../lib/zustand.svelte.js';
	import { gehe } from '../lib/weg.svelte.js';

	/**
	 * M3 App Bar (small, seit M3 Expressive nicht mehr „Top“). Beim Scrollen hebt sie sich mit
	 * surface-container vom Inhalt ab. Google rät zu höchstens zwei Aktionen; hier ist es das Konto.
	 *
	 * ebene ('uebersicht' | 'melden'): Wer die Übersicht sieht, wechselt hier zwischen ihr und dem
	 * Meldeformular. M3 nimmt für zwei Ansichten einen Segmented Button, keinen Switch („Avoid using
	 * switches to toggle between opposing options“); in der Kopfleiste bleibt er beim Scrollen da.
	 * Während einer Auswahl steht vorn das Kreuz statt des Pfeils (zurueckSymbol), wie in Gmail.
	 */
	let { titel, zurueck = null, zurueckSymbol = 'arrow_back', zurueckEtikett = 'Zurück', konto = true, aktionen, ebene = null } = $props();
	const ebenen = [
		{ wert: 'uebersicht', text: 'Übersicht' },
		{ wert: 'melden', text: 'Melden' }
	];
	let gescrollt = $state(false);

	$effect(() => {
		const pruefe = () => (gescrollt = window.scrollY > 2);
		pruefe();
		window.addEventListener('scroll', pruefe, { passive: true });
		return () => window.removeEventListener('scroll', pruefe);
	});

	const uhrzeit = new Intl.DateTimeFormat('de-DE', { hour: '2-digit', minute: '2-digit' });
	const anzahl = (n) => (n === 1 ? '1 Meldung wartet' : `${n} Meldungen warten`);
	// Funkloch und Ausgang sichtbar machen: Niemand soll glauben, eine Meldung sei verloren.
	const hinweis = $derived.by(() => {
		const n = zustand.ausgang.length;
		if (zustand.offline && n) return `Kein Netz · ${anzahl(n)} und geht automatisch raus.`;
		if (zustand.offline) return `Kein Netz – du siehst den Stand ${zustand.standVon ? `von ${uhrzeit.format(new Date(zustand.standVon))}` : 'von eben'}.`;
		if (n) return `${anzahl(n)} auf das Nachreichen …`;
		return '';
	});
</script>

<header class={['sticky top-0 z-20 pt-[env(safe-area-inset-top)] transition-colors duration-300', gescrollt ? 'bg-surface-container' : 'bg-surface']}>
	<div class="mx-auto flex h-16 max-w-3xl items-center gap-1 px-1">
		{#if zurueck}
			<button class="icon-btn text-on-surface" aria-label={zurueckEtikett} data-zurueck onclick={zurueck}><Icon name={zurueckSymbol} /></button>
		{:else}
			<span class="w-3"></span>
		{/if}
		{#if ebene && darfSehen()}
			<h1 class="sr-only">{titel}</h1>
			<div class="min-w-0 flex-1 px-1">
				<Segmente etikett="Ansicht" klasse="w-full max-w-xs" optionen={ebenen} wert={ebene} onwahl={(w) => w !== ebene && gehe(w === 'melden' ? '/melden' : '/')} />
			</div>
		{:else}
			<h1 class="min-w-0 flex-1 truncate px-1 text-title-large text-on-surface">{titel}</h1>
		{/if}
		{@render aktionen?.()}
		{#if konto}<KontoKnopf />{/if}
	</div>
	{#if hinweis}
		<div class="mx-auto max-w-3xl px-4 pb-2" transition:slide={{ duration: 200 }}>
			<p class="flex items-center gap-3 rounded-lg bg-surface-container-highest px-4 py-2.5 text-body-medium text-on-surface" role="status">
				<Icon name={zustand.offline ? 'cloud_off' : 'schedule_send'} groesse={20} class="text-on-surface-variant" />
				{hinweis}
			</p>
		</div>
	{/if}
</header>
