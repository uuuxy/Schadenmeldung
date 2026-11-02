<script>
	import { flip } from 'svelte/animate';
	import { fade, slide } from 'svelte/transition';
	import { zustand, darfBearbeiten, auffrischen } from '../lib/zustand.svelte.js';
	import Icon from '../components/Icon.svelte';
	import KopfLeiste from '../components/KopfLeiste.svelte';
	import Ziehen from '../components/Ziehen.svelte';
	import Reiter from '../components/ui/Reiter.svelte';
	import MeldungZeile from '../components/uebersicht/MeldungZeile.svelte';
	import BereichWahl from '../components/uebersicht/BereichWahl.svelte';
	import Export from '../components/uebersicht/Export.svelte';

	/**
	 * Die Übersicht für Hausmeister, Leitung, Verwaltung und Admins: eine schlanke Liste, Dringendes
	 * zuerst. Ein Tipp öffnet die Meldung mit allem Weiteren (M3 list-detail: am Handy immer nur eine
	 * der beiden Ansichten). Reiter und Bereich merkt sich die Sitzung, damit „Zurück“ dort landet.
	 */
	const merker = (() => {
		try {
			return JSON.parse(sessionStorage.getItem('schaden.uebersicht') ?? '{}');
		} catch {
			return {};
		}
	})();
	let reiter = $state(merker.reiter ?? 'offen');
	let bereich = $state(merker.bereich ?? null);

	const quelle = $derived(reiter === 'offen' ? zustand.offen : zustand.erledigt);
	// Bereiche in der Reihenfolge der Raumliste (EG vor 1. OG), nur die mit Meldungen.
	const bereiche = $derived.by(() => {
		const mitMeldung = new Set(quelle.map((m) => m.bereich).filter(Boolean));
		return [...new Set(zustand.raeume.map((r) => r.bereich))].filter((b) => mitMeldung.delete(b)).concat([...mitMeldung]);
	});
	const liste = $derived(bereich ? quelle.filter((m) => m.bereich === bereich) : quelle);

	$effect(() => {
		if (bereich && !bereiche.includes(bereich)) bereich = null;
	});
	$effect(() => {
		try {
			sessionStorage.setItem('schaden.uebersicht', JSON.stringify({ reiter, bereich }));
		} catch {
			/* Ohne Speicher beginnt die Übersicht eben bei „Offen“. */
		}
	});
</script>

<KopfLeiste titel={darfBearbeiten() ? 'Aufgaben' : 'Übersicht'} ebene="uebersicht" />

<Ziehen aktualisieren={auffrischen}>
	<main class="mx-auto max-w-3xl px-4 pb-12">
		<!-- Primary Tabs direkt unter der Kopfleiste (M3): zwei Datensätze derselben Ansicht. -->
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

		<div class="flex min-h-14 items-center gap-2">
			{#if bereiche.length > 1}<BereichWahl {bereiche} bind:bereich />{/if}
			<span class="flex-1"></span>
			<Export liste={reiter} {bereich} anzahl={liste.length} />
		</div>

		{#if liste.length === 0}
			<div class="flex flex-col items-center px-6 py-16 text-center" in:fade={{ duration: 200 }}>
				<span class="mb-5 grid size-20 place-items-center rounded-2xl bg-success-container text-on-success-container">
					<Icon name="task_alt" groesse={40} />
				</span>
				<p class="text-title-large text-on-surface">{reiter === 'offen' ? 'Alles erledigt' : 'Noch nichts erledigt'}</p>
				<p class="mt-2 max-w-xs text-body-medium text-on-surface-variant">
					{reiter === 'offen' ? 'Gerade ist kein Schaden offen.' : 'Erledigte Meldungen der letzten 30 Tage stehen hier.'}
				</p>
			</div>
		{:else}
			<!-- Segmentierte Liste (M3 Expressive): gefüllte Einträge mit 2 px Abstand, außen 16 px rund,
			     innen 4 px. -->
			<ul class="flex flex-col gap-0.5" aria-label="Meldungen">
				{#each liste as m (m.id)}
					<li class="rounded-xs bg-surface-container-low first:rounded-t-lg last:rounded-b-lg" animate:flip={{ duration: 300 }} out:slide={{ duration: 200 }}>
						<MeldungZeile meldung={m} />
					</li>
				{/each}
			</ul>
		{/if}
	</main>
</Ziehen>
