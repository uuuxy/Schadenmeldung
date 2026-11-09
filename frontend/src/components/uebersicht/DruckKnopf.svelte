<script module>
	// So viele druckt der Server auf einmal (druckHoechstens in druck.go); mehr macht die Datei zu groß.
	export const HOECHSTENS = 50;
</script>

<script>
	import Icon from '../Icon.svelte';

	/**
	 * Drucken als Symbol in der Kopfleiste: ein PDF mit einer Seite je Meldung samt Foto (druck.go),
	 * auf der Seite einer Meldung und während einer Auswahl in der Liste. Ein Link wie beim Export, so
	 * öffnet jedes Handy das PDF zum Drucken oder Teilen. Ohne Auswahl oder mit zu vielen ist er
	 * gesperrt (M3: 38 % Deckkraft).
	 */
	let { ids, onclick } = $props();
	const geht = $derived(ids.length > 0 && ids.length <= HOECHSTENS);
</script>

{#if geht}
	<a class="icon-btn text-on-surface-variant" href="/api/druck.pdf?ids={ids.join(',')}" download aria-label="Drucken" title="Drucken" {onclick}><Icon name="print" /></a>
{:else}
	<button class="icon-btn text-on-surface-variant opacity-38" disabled aria-label="Drucken" title={ids.length ? `Höchstens ${HOECHSTENS} auf einmal` : 'Erst Meldungen antippen'}><Icon name="print" /></button>
{/if}
