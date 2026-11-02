<script>
	import Icon from '../Icon.svelte';

	/**
	 * @component Segmente — der Segmented Button aus Material 3 (Einfachauswahl), wie in der Bibliothek.
	 *
	 * Zwei bis fünf Ansichten desselben Inhalts, von denen immer eine gilt („Offen | Erledigt“). Das
	 * gewählte Segment liegt auf secondary-container. Der Platz des Häkchens bleibt in jedem Segment
	 * reserviert, sonst rücken beim Umschalten die Beschriftungen, und der Umschalter zappelt.
	 *
	 * @prop {{ wert: string, text: string }[]} optionen
	 * @prop {string} wert
	 * @prop {(wert: string) => void} onwahl
	 * @prop {string} etikett
	 * @prop {string} [klasse]
	 */
	let { optionen, wert, onwahl, etikett, klasse = '' } = $props();
</script>

<!-- auto-cols-fr: M3 gibt allen Segmenten dieselbe Breite. Ohne overflow-hidden, damit die
     48-px-Trefferfläche (.treffer) über den Rand greift; die äußeren Segmente runden sich selbst. -->
<div role="group" aria-label={etikett} class="inline-grid h-10 auto-cols-fr grid-flow-col rounded-full border border-outline {klasse}">
	{#each optionen as o, i (o.wert)}
		{@const gewaehlt = o.wert === wert}
		<button
			type="button"
			aria-pressed={gewaehlt}
			onclick={() => onwahl(o.wert)}
			class={[
				'treffer flex min-w-0 items-center justify-center gap-2 px-3 text-label-large transition-colors duration-200 first:rounded-l-full last:rounded-r-full',
				i > 0 && 'border-l border-outline',
				gewaehlt ? 'bg-secondary-container text-on-secondary-container' : 'text-on-surface'
			]}
		>
			<Icon name="check" groesse={18} class={gewaehlt ? '' : 'invisible'} />
			<span class="truncate">{o.text}</span>
		</button>
	{/each}
</div>
