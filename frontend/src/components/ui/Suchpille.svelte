<script>
	import Icon from '../Icon.svelte';

	/**
	 * @component Suchpille — das Suchfeld als Werkzeug einer Seite, wie in der Bibliothek.
	 *
	 * Farben nach den M3-Token (search bar): Fläche surface-container-high, Symbol und Eingabe
	 * on-surface, Platzhalter on-surface-variant. Im Fokus helle Fläche mit Umriss in primary, kein
	 * Schatten; der zweite Pixel des Umrisses liegt innen. Chromes eigenes Löschkreuz ist
	 * weggeblendet, das eigene steht rechts. aktion ist ein Symbol rechts (M3 „trailing icon“), etwa der
	 * Scanner; sobald etwas eingetippt ist, macht es dem Löschkreuz Platz, wie in der Google-Suche.
	 */
	let { wert = $bindable(''), platzhalter, etikett, onkeydown, aktion } = $props();
</script>

<!-- label: Ein Tipp irgendwo auf die Pille, auch auf das Symbol, setzt den Cursor ins Feld. Der Fokus
     ist ein Ring statt eines Rands, so behält das Feld seine vollen 48 px. -->
<label
	class="group flex h-12 w-full cursor-text items-center rounded-full bg-surface-container-high pr-1 pl-4 ring-inset transition-colors duration-200 focus-within:bg-surface-container-lowest focus-within:ring-2 focus-within:ring-primary"
>
	<Icon name="search" groesse={22} class="text-on-surface transition-colors group-focus-within:text-primary" />
	<input
		type="search"
		inputmode="search"
		enterkeyhint="search"
		autocomplete="off"
		spellcheck="false"
		bind:value={wert}
		aria-label={etikett}
		placeholder={platzhalter}
		{onkeydown}
		class="h-full min-w-0 flex-1 bg-transparent px-3 text-body-large text-on-surface outline-none placeholder:text-on-surface-variant"
	/>
	<!-- ml-1: Die 48-px-Fläche eines Symbolknopfs ragt 4 px über ihn hinaus; so reicht sie nicht ins Feld. -->
	{#if wert}
		<button type="button" class="ml-1 icon-btn text-on-surface-variant" aria-label="Suche leeren" onclick={() => (wert = '')}>
			<Icon name="close" />
		</button>
	{:else if aktion}
		<span class="ml-1 flex">{@render aktion()}</span>
	{/if}
</label>
