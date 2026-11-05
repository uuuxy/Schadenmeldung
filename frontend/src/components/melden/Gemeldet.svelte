<script>
	import { fade } from 'svelte/transition';
	import { eindrehen, aufspringen } from '../../lib/bewegung.js';
	import { zustand, darfSehen } from '../../lib/zustand.svelte.js';
	import { gehe } from '../../lib/weg.svelte.js';
	import Icon from '../Icon.svelte';
	import Button from '../ui/Button.svelte';

	/**
	 * Bestätigung statt Formular: Die Meldung ist sicher angekommen oder sicher im Ausgang. Der Text
	 * verspricht nur, was stimmt: eine Mail nur, wenn der Server Mails verschickt.
	 */
	let { fertig, onweiter } = $props();
	// M3 Expressive: Das Häkchen steht in einer Form der neuen Formenbibliothek, dem „Cookie“ mit neun
	// Wellen. Fürs Web gibt es die Formen nicht fertig, also ein Kreis, dessen Rand sanft schwingt.
	const keks = Array.from({ length: 180 }, (_, i) => {
		const w = (i / 180) * 2 * Math.PI;
		const r = 1 + 0.07 * Math.cos(9 * w);
		return `${(r * Math.sin(w)).toFixed(3)},${(-r * Math.cos(w)).toFixed(3)}`;
	}).join(' ');
	// Kommt die Meldung aus dem Ausgang beim Server an, springt die Bestätigung um: „Gerade ist kein Netz“
	// stimmte dann nicht mehr. Eine abgelehnte zählt nicht, das sagt die Snackbar.
	const gemeldet = $derived(fertig.art === 'gemeldet' || zustand.nachgereicht.includes(fertig.ausgangId));
	const text = $derived.by(() => {
		if (!gemeldet) return 'Gerade ist kein Netz. Die Meldung geht automatisch raus, sobald wieder Verbindung besteht.';
		// Gleich dazu, wie der Bescheid kommt, wenn sie erledigt ist (rueckmeldung.go).
		if (zustand.einstellungen.mailAktiv) return 'Sie ist beim Hausmeister angekommen. Ist sie erledigt, kommt Bescheid per Mail.';
		return 'Sie ist beim Hausmeister angekommen. Ist sie erledigt, steht das beim nächsten Öffnen hier.';
	});
</script>

<main class="mx-auto flex max-w-xl flex-col items-center px-6 pt-16 pb-12 text-center" in:fade={{ duration: 200 }}>
	{#if gemeldet}
		<!-- global: Das läuft auch, wenn die Bestätigung gerade erst erscheint. -->
		<span class="relative mb-6 grid size-24 place-items-center text-on-success-container">
			<svg class="absolute inset-0 size-full fill-success-container" viewBox="-1.07 -1.07 2.14 2.14" aria-hidden="true" in:eindrehen|global><polygon points={keks} /></svg>
			<span class="relative" in:aufspringen|global><Icon name="check" groesse={48} /></span>
		</span>
	{:else}
		<span class="mb-6 grid size-24 place-items-center rounded-2xl bg-surface-container-highest text-on-surface-variant">
			<Icon name="schedule_send" groesse={48} />
		</span>
	{/if}
	<h2 class="text-headline-small text-on-surface">{gemeldet ? 'Gemeldet – danke!' : 'Gespeichert'}</h2>
	<p class="mt-2 text-body-large text-on-surface-variant">{fertig.was} · {fertig.ort}</p>
	<p class="mt-4 max-w-sm text-balance text-body-medium text-on-surface-variant">{text}</p>
	<div class="mt-10 flex flex-col gap-3 self-stretch">
		<Button size="lg" onclick={onweiter}><Icon name="add" />Noch etwas melden</Button>
		{#if darfSehen()}
			<Button variant="ghost" onclick={() => gehe('/')}>Zur Übersicht</Button>
		{/if}
	</div>
</main>
