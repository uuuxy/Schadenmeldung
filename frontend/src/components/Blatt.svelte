<script>
	/** M3 Bottom Sheet auf Basis von <dialog>: Fokus, Esc und Zurücktreten des Hintergrunds kommen vom Browser. */
	let { offen = $bindable(false), titel = '', label = '', children } = $props();
	let dialog;

	$effect(() => {
		if (offen && !dialog.open) dialog.showModal();
		if (!offen && dialog.open) dialog.close();
	});

	// Ein Tipp auf den abgedunkelten Hintergrund trifft den Dialog selbst und schließt das Blatt.
	function tippDaneben(e) {
		if (e.target === dialog) offen = false;
	}
</script>

<dialog bind:this={dialog} class="blatt" aria-label={label || titel} onclose={() => (offen = false)} onclick={tippDaneben}>
	<div class="flex justify-center pt-4 pb-1 sm:hidden" aria-hidden="true">
		<span class="h-1 w-8 rounded-full bg-on-surface-variant/40"></span>
	</div>
	{#if titel}
		<h2 class="px-6 pt-3 pb-4 text-title-large text-on-surface sm:pt-6">{titel}</h2>
	{/if}
	{@render children?.()}
</dialog>
