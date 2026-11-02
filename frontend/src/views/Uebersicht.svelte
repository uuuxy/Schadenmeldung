<script>
	import { flip } from 'svelte/animate';
	import { fade } from 'svelte/transition';
	import { zustand, darfBearbeiten, auffrischen } from '../lib/zustand.svelte.js';
	import Icon from '../components/Icon.svelte';
	import KopfLeiste from '../components/KopfLeiste.svelte';
	import MeldungKarte from '../components/MeldungKarte.svelte';
	import Ziehen from '../components/Ziehen.svelte';
	import FilterChips from '../components/ui/FilterChips.svelte';
	import Reiter from '../components/ui/Reiter.svelte';

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
</script>

<KopfLeiste titel={darfBearbeiten() ? 'Aufgaben' : 'Übersicht'} ebene="uebersicht" />

<Ziehen aktualisieren={auffrischen}>
	<main class="mx-auto max-w-3xl px-4 pb-12">
		<!-- Primary Tabs direkt unter der Kopfleiste (M3): zwei Datensätze derselben Ansicht. Sie sehen
		     anders aus als der Umschalter der Ebene darüber, damit die Rangfolge auf einen Blick klar ist. -->
		<div class="-mx-4">
			<Reiter
				etikett="Meldungen"
				reiter={[
					{ id: 'offen', label: `Offen · ${zustand.offen.length}` },
					{ id: 'erledigt', label: 'Erledigt' }
				]}
				aktiv={reiter}
				onwahl={(w) => (reiter = w)}
			/>
		</div>

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
