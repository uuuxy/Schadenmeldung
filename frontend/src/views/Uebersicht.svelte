<script>
	import { flip } from 'svelte/animate';
	import { fade } from 'svelte/transition';
	import { zustand, darfBearbeiten, auffrischen } from '../lib/zustand.svelte.js';
	import { gehe } from '../lib/weg.svelte.js';
	import Icon from '../components/Icon.svelte';
	import KopfLeiste from '../components/KopfLeiste.svelte';
	import MeldungKarte from '../components/MeldungKarte.svelte';
	import Ziehen from '../components/Ziehen.svelte';
	import FilterChips from '../components/ui/FilterChips.svelte';
	import Segmente from '../components/ui/Segmente.svelte';

	/** Das Backend: Aufgaben für Hausmeister und Verwaltung, Übersicht (nur lesen) für die Leitung. */
	let reiter = $state('offen');
	let bereich = $state(null);

	const quelle = $derived(reiter === 'offen' ? zustand.offen : zustand.erledigt);
	// Bereiche als Filter für den Rundgang durchs Haus, sobald es mehr als einen gibt.
	const bereiche = $derived([...new Set(quelle.map((m) => m.bereich).filter(Boolean))]);
	const liste = $derived(bereich ? quelle.filter((m) => m.bereich === bereich) : quelle);

	$effect(() => {
		if (bereich && !bereiche.includes(bereich)) bereich = null;
	});

	// Die Snackbar steht über dem Knopf „Melden“ (M3: Snackbars erscheinen über einem FAB).
	$effect(() => {
		document.body.style.setProperty('--snackbar-abstand', '4.5rem');
		return () => document.body.style.removeProperty('--snackbar-abstand');
	});
</script>

<KopfLeiste titel={darfBearbeiten() ? 'Aufgaben' : 'Übersicht'} />

<Ziehen aktualisieren={auffrischen}>
	<main class="mx-auto max-w-3xl px-4 pt-2 pb-28">
		<Segmente
			etikett="Meldungen"
			klasse="w-full"
			optionen={[
				{ wert: 'offen', text: `Offen · ${zustand.offen.length}` },
				{ wert: 'erledigt', text: 'Erledigt' }
			]}
			wert={reiter}
			onwahl={(w) => (reiter = w)}
		/>

		{#if bereiche.length > 1}
			<div class="mt-4">
				<FilterChips etikett="Bereich" optionen={bereiche.map((b) => ({ wert: b, text: b }))} wert={bereich} onwahl={(b) => (bereich = b)} />
			</div>
		{/if}

		{#if liste.length === 0}
			<div class="flex flex-col items-center px-6 py-20 text-center" in:fade={{ duration: 200 }}>
				<span class="mb-5 grid size-20 place-items-center rounded-2xl bg-success-container text-on-success-container">
					<Icon name="task_alt" groesse={40} />
				</span>
				<p class="text-title-large text-on-surface">{reiter === 'offen' ? 'Alles erledigt' : 'Noch nichts erledigt'}</p>
				<p class="mt-2 max-w-xs text-body-medium text-on-surface-variant">
					{reiter === 'offen' ? 'Gerade ist kein Schaden offen.' : 'Erledigte Meldungen der letzten 30 Tage stehen hier.'}
				</p>
			</div>
		{:else}
			<ul class="mt-4 flex flex-col gap-3">
				{#each liste as m (m.id)}
					<li animate:flip={{ duration: 300 }} out:fade={{ duration: 150 }}>
						<MeldungKarte meldung={m} />
					</li>
				{/each}
			</ul>
		{/if}
	</main>
</Ziehen>

<!-- M3 Extended FAB: die eine Hauptaktion dieser Seite, rechts unten in Daumenreichweite. -->
<button
	class="fixed right-4 bottom-[calc(env(safe-area-inset-bottom)+1rem)] z-20 flex h-14 items-center gap-3 rounded-lg bg-primary-container pr-5 pl-4 text-label-large text-on-primary-container shadow-3 transition-transform duration-200 active:scale-95"
	onclick={() => gehe('/melden')}
>
	<Icon name="add" />Melden
</button>
