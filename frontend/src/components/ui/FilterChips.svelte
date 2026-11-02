<script>
	import Icon from '../Icon.svelte';

	/**
	 * @component FilterChips — die Filter-Chips aus Material 3, einfach wählbar, wie in der Bibliothek.
	 *
	 * Die Chips brechen um, statt seitlich aus dem Bild zu laufen: So steht jeder Chip auf der Kante
	 * des Inhalts, und keiner ist halb verdeckt.
	 *
	 * Maße aus den M3-Token (filter chip): 32 px hoch, Ecke 8 px, Beschriftung label-large; getroffen
	 * wird auf 48 px (.treffer). Ungewählt:
	 * Umriss outline, Text on-surface-variant. Gewählt: Fläche secondary-container ohne Umriss, vorn
	 * ein 18-px-Häkchen; der Rand bleibt durchsichtig stehen, damit die Höhe gleich bleibt.
	 *
	 * @prop {{ wert: string, text: string, icon?: string }[]} optionen
	 * @prop {string | null} wert - der gewählte Wert; null heißt keiner.
	 * @prop {(wert: string | null) => void} onwahl - bekommt null, wenn der Chip zurückgenommen wird.
	 * @prop {string} etikett - Name der Gruppe für Screenreader.
	 * @prop {boolean} [abwaehlbar=true]
	 */
	let { optionen, wert, onwahl, etikett, abwaehlbar = true } = $props();
</script>

<div role="group" aria-label={etikett} class="flex flex-wrap gap-x-2 gap-y-4">
	{#each optionen as o (o.wert)}
		{@const gewaehlt = o.wert === wert}
		<button
			type="button"
			aria-pressed={gewaehlt}
			onclick={() => onwahl(gewaehlt && abwaehlbar ? null : o.wert)}
			class={[
				'treffer flex h-8 items-center gap-2 rounded-sm border text-label-large transition-colors duration-200',
				gewaehlt ? 'border-transparent bg-secondary-container pr-4 pl-2 text-on-secondary-container' : 'border-outline px-4 text-on-surface-variant',
				!gewaehlt && o.icon && 'pl-2'
			]}
		>
			{#if gewaehlt}
				<Icon name="check" groesse={18} />
			{:else if o.icon}
				<Icon name={o.icon} groesse={18} class="text-primary" />
			{/if}
			{o.text}
		</button>
	{/each}
</div>
