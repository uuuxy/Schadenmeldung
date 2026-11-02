<script>
	/**
	 * Räume zum Antippen, nach Stockwerken gegliedert, wenn die Gruppe welche hat. Kurze Namen
	 * („C204“) stehen am Handy zu viert in einer Reihe, längere zu dritt; getroffen wird jeder Raum
	 * auf 48 px Höhe. mitBereich nennt unter dem Namen den Bereich, für Suchtreffer aus allen Gruppen.
	 */
	let { raeume, stockwerke = [], mitBereich = false, onwahl } = $props();

	const kurz = $derived(raeume.every((r) => r.name.length <= 5));
	const abschnitte = $derived(stockwerke.length ? stockwerke : [{ titel: '', raeume }]);
</script>

<div class="flex flex-col gap-4">
	{#each abschnitte as s (s.titel)}
		<div>
			{#if s.titel}<h3 class="mb-2 text-label-large text-on-surface-variant">{s.titel}</h3>{/if}
			<div class={['grid gap-2', kurz ? 'grid-cols-[repeat(auto-fill,minmax(4.5rem,1fr))]' : 'grid-cols-[repeat(auto-fill,minmax(6.5rem,1fr))]']}>
				{#each s.raeume as r (r.id)}
					<button
						class="flex min-h-12 min-w-0 flex-col items-center justify-center rounded-md border border-outline-variant bg-surface-container-lowest px-2 py-2 text-center text-on-surface transition-transform duration-200 active:scale-[0.97]"
						onclick={() => onwahl(r)}
					>
						<span class="text-label-large [overflow-wrap:anywhere]">{r.name}</span>
						{#if mitBereich && r.bereich}<span class="mt-0.5 line-clamp-1 text-body-small text-on-surface-variant">{r.bereich}</span>{/if}
					</button>
				{/each}
			</div>
		</div>
	{/each}
</div>
