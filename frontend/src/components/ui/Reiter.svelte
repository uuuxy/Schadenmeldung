<script>
	/**
	 * @component Reiter — Primary Tabs nach Material 3, wie ui/Reiter der Bibliothek: Beschriftung
	 * title-small, gewählt in primary mit 3-px-Indikator (oben gerundet), sonst on-surface-variant.
	 *
	 * @prop {{ id: string, label: string }[]} reiter
	 * @prop {string} aktiv
	 * @prop {(id: string) => void} onwahl
	 * @prop {string} etikett
	 */
	let { reiter, aktiv, onwahl, etikett } = $props();
</script>

<div role="tablist" aria-label={etikett} class="grid auto-cols-fr grid-flow-col border-b border-outline-variant">
	{#each reiter as r (r.id)}
		{@const gewaehlt = r.id === aktiv}
		<button
			type="button"
			role="tab"
			aria-selected={gewaehlt}
			onclick={() => onwahl(r.id)}
			class={['relative h-12 text-title-small transition-colors', gewaehlt ? 'text-primary' : 'text-on-surface-variant']}
		>
			{r.label}
			{#if gewaehlt}
				<span class="absolute bottom-0 left-1/2 h-0.75 w-12 -translate-x-1/2 rounded-t-[3px] bg-primary"></span>
			{/if}
		</button>
	{/each}
</div>
