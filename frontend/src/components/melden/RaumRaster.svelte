<script>
	import Kachel from './Kachel.svelte';

	/**
	 * Räume zum Antippen, nach Stockwerken gegliedert, wenn die Gruppe welche hat. Die Kacheln sehen aus
	 * wie die Bereiche davor, damit sich beim Tipp auf einen Bereich nur der Inhalt ändert: kurze Namen
	 * („C204“) wie Tasten zu dritt, ab sm zu viert, längere zu zweit. mitBereich nennt unter dem Namen
	 * den Bereich, für Suchtreffer aus allen Gruppen.
	 */
	let { raeume, stockwerke = [], mitBereich = false, kurz = undefined, onwahl } = $props();

	// Bei wenigen Räumen gibt RaumWahl kurz für alle Gruppen vor, damit jede gleich breite Kacheln hat.
	const tasten = $derived(kurz ?? raeume.every((r) => r.name.length <= 5));
	const abschnitte = $derived(stockwerke.length ? stockwerke : [{ titel: '', raeume }]);
</script>

<div class="flex flex-col gap-4">
	{#each abschnitte as s (s.titel)}
		<div>
			{#if s.titel}<h3 class="mb-2 text-label-large text-on-surface-variant">{s.titel}</h3>{/if}
			<div class={['grid gap-3', tasten ? 'grid-cols-3 sm:grid-cols-4' : 'grid-cols-2']}>
				{#each s.raeume as r (r.id)}
					<Kachel kurz={tasten} oben={r.name} unten={mitBereich ? (r.bereich ?? '') : ''} onclick={() => onwahl(r)} />
				{/each}
			</div>
		</div>
	{/each}
</div>
