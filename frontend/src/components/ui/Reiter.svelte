<script>
	/**
	 * @component Reiter — Primary Tabs nach Material 3, wie ui/Reiter der Bibliothek: Beschriftung
	 * title-small, gewählt in primary mit 3-px-Indikator (oben gerundet), sonst on-surface-variant.
	 * Am Handy teilen sich die Reiter die Breite; ab sm stehen sie links und so breit wie ihr Text,
	 * wie in Googles Web-Apps, sonst lägen zwei Reiter am Rechner weit auseinander.
	 *
	 * @prop {{ id: string, label: string }[]} reiter
	 * @prop {string} aktiv
	 * @prop {(id: string) => void} onwahl
	 * @prop {string} etikett
	 */
	let { reiter, aktiv, onwahl, etikett } = $props();
</script>

<div role="tablist" aria-label={etikett} class="flex border-b border-outline-variant">
	{#each reiter as r (r.id)}
		{@const gewaehlt = r.id === aktiv}
		<button
			type="button"
			role="tab"
			aria-selected={gewaehlt}
			onclick={() => onwahl(r.id)}
			class={['relative h-12 flex-1 basis-0 px-4 text-title-small transition-colors sm:flex-none sm:basis-auto', gewaehlt ? 'text-primary' : 'text-on-surface-variant']}
		>
			{r.label}
			{#if gewaehlt}
				<span class="absolute bottom-0 left-1/2 h-0.75 w-12 -translate-x-1/2 rounded-t-[3px] bg-primary"></span>
			{/if}
		</button>
	{/each}
</div>
