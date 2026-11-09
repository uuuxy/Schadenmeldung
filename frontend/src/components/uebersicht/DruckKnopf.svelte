<script module>
	// So viele druckt der Server auf einmal (druckHoechstens in druck.go); mehr macht die Datei zu groß.
	export const HOECHSTENS = 50;

	// Am Rechner öffnet ein PDF gleich im Browser, gedruckt ist es dann mit einem Klick; am Handy wird es
	// wie bisher geladen, denn ein neuer Tab aus der App vom Home-Bildschirm kennt am iPhone die Anmeldung nicht.
	const imBrowser = matchMedia('(pointer: fine)').matches && navigator.standalone !== true;
	export const pdfLink = (adresse) => (imBrowser ? { href: `${adresse}&zeigen=1`, target: '_blank', rel: 'noopener' } : { href: adresse, download: '' });
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
	<a class="icon-btn text-on-surface-variant" {...pdfLink(`/api/druck.pdf?ids=${ids.join(',')}`)} aria-label="Drucken" title="Drucken" {onclick}><Icon name="print" /></a>
{:else}
	<button class="icon-btn text-on-surface-variant opacity-38" disabled aria-label="Drucken" title={ids.length ? `Höchstens ${HOECHSTENS} auf einmal` : 'Drucken'}><Icon name="print" /></button>
{/if}
