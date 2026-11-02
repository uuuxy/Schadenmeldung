<script>
	import { rueckfrage, antworte } from '../lib/rueckfrage.svelte.js';

	/**
	 * Der Dialog nach M3 (basic dialog): Fläche surface-container-high, Ecke 28 px, Überschrift
	 * headline-small, Text body-medium, zwei Textknöpfe rechts. Die riskante Aktion steht in error,
	 * „Abbrechen“ bekommt den Fokus, damit ein versehentliches Enter nichts löscht.
	 */
	let dialog;

	$effect(() => {
		if (rueckfrage.offen && !dialog.open) dialog.showModal();
		if (!rueckfrage.offen && dialog.open) dialog.close();
	});
</script>

<dialog bind:this={dialog} class="rueckfrage" aria-labelledby="rueckfrage-titel" onclose={() => rueckfrage.offen && antworte(false)}>
	<h2 id="rueckfrage-titel" class="text-headline-small text-on-surface">{rueckfrage.titel}</h2>
	{#if rueckfrage.text}<p class="mt-4 text-body-medium text-on-surface-variant">{rueckfrage.text}</p>{/if}
	<div class="mt-6 flex justify-end gap-2">
		<!-- svelte-ignore a11y_autofocus -->
		<button autofocus class="treffer h-10 rounded-full px-3 text-label-large text-primary" onclick={() => antworte(false)}>Abbrechen</button>
		<button class="treffer h-10 rounded-full px-3 text-label-large text-error" onclick={() => antworte(true)}>{rueckfrage.ja}</button>
	</div>
</dialog>
