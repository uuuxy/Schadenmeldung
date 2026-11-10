<script>
	/**
	 * Die Hauptaktion unten in Daumenreichweite, auf einer angedockten Leiste (M3 docked toolbar mit
	 * eingebetteter Hauptaktion). Am Rechner steht sie mittig unter dem Inhalt, so breit wie die Seiten;
	 * der Knopf darin nimmt w-full.
	 */
	let { children } = $props();
</script>

<div class="fixed inset-x-0 bottom-0 z-20 bg-surface-container px-4 pt-3 pb-[calc(env(safe-area-inset-bottom)+0.75rem)]">
	<div class="mx-auto flex max-w-3xl">{@render children()}</div>
</div>
