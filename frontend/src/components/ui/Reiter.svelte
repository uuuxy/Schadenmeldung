<script>
	import Icon from '../Icon.svelte';

	/**
	 * @component Reiter — Primary Tabs nach Material 3, wie ui/Reiter der Bibliothek: Beschriftung
	 * title-small, gewählt in primary mit 3-px-Indikator (oben gerundet) so breit wie der Inhalt, sonst
	 * on-surface-variant. Ein Symbol vor der Beschriftung ist gewählt gefüllt (M3 „inline icon“), so
	 * erkennt man auch ohne Lesen, wo man ist. Am Handy teilen sich die Reiter die Breite; ab sm stehen sie links und so breit wie ihr Text,
	 * wie in Googles Web-Apps, sonst lägen zwei Reiter am Rechner weit auseinander.
	 *
	 * @prop {{ id: string, label: string, icon?: string }[]} reiter
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
			class={['relative flex h-12 flex-1 basis-0 justify-center px-4 text-title-small transition-colors sm:flex-none sm:basis-auto', gewaehlt ? 'text-primary' : 'text-on-surface-variant']}
		>
			<span class="relative flex h-full items-center gap-2">
				{#if r.icon}<Icon name={r.icon} groesse={20} gefuellt={gewaehlt} />{/if}
				{r.label}
				{#if gewaehlt}<span class="absolute inset-x-0 bottom-0 h-0.75 rounded-t-[3px] bg-primary"></span>{/if}
			</span>
		</button>
	{/each}
</div>
