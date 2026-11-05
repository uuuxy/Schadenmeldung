<script>
	import Icon from '../Icon.svelte';

	/**
	 * Eine Kachel zum Antippen für Bereiche, Räume und „Anderer Ort“: überall gleich hoch, gleich rund
	 * und auf derselben Fläche, damit ein Tipp in einen Bereich nur den Inhalt wechselt. Oben groß (Name,
	 * Buchstabe oder Symbol), darunter klein. kurz stellt den Inhalt wie eine Taste in die Mitte und
	 * zeigt das Symbol statt des Namens oben; sonst steht der Name links wie ein Titel.
	 */
	let { oben, unten = '', symbol = '', kurz = false, etikett = undefined, onclick } = $props();
</script>

<button
	class={[
		'flex min-h-16 min-w-0 flex-col justify-center rounded-xl bg-surface-container-low py-2 text-on-surface transition-transform duration-200 active:scale-[0.97]',
		kurz ? 'items-center px-2 text-center' : 'items-start px-3 text-left'
	]}
	aria-label={etikett}
	{onclick}
>
	{#if kurz && symbol}
		<Icon name={symbol} groesse={24} class="my-0.5" />
		<span class="line-clamp-1 max-w-full text-body-small text-on-surface-variant">{oben}</span>
	{:else}
		<span class={[kurz ? 'text-title-large' : 'text-title-medium', 'max-w-full truncate']}>{oben}</span>
		{#if unten}<span class="line-clamp-1 max-w-full text-body-small text-on-surface-variant">{unten}</span>{/if}
	{/if}
</button>
