<script>
	import Icon from '../Icon.svelte';
	import Blatt from '../Blatt.svelte';
	import { pdfLink } from './DruckKnopf.svelte';

	/**
	 * Das Überlaufmenü der Übersicht (M3: weitere Aktionen hinter ⋮, am Handy als Blatt). „Drucken“
	 * beginnt die Auswahl in der Liste, „Liste als PDF“ lädt die Liste, wie sie gerade steht (Reiter und
	 * Filter), die Excel-Tabelle alle gespeicherten Meldungen. Einträge nach M3 nur mit Symbol und
	 * kurzem Namen; was gerade nicht geht, ist gesperrt statt versteckt. Die Dateien verlassen die App
	 * und tragen deshalb keine Namen der Meldenden.
	 */
	let { liste, bereich, art = null, anzahl, onauswaehlen } = $props();
	let offen = $state(false);

	const pdf = $derived(`/api/export.pdf?liste=${liste}${bereich ? `&bereich=${encodeURIComponent(bereich)}` : ''}${art ? `&art=${art}` : ''}`);
	const zeile = 'schicht flex h-14 w-full items-center gap-4 px-6 text-left text-body-large text-on-surface';
	const zu = () => setTimeout(() => (offen = false));
</script>

<button class="icon-btn text-on-surface-variant" aria-label="Weitere Optionen" title="Weitere Optionen" onclick={() => (offen = true)}><Icon name="more_vert" /></button>

<Blatt bind:offen label="Weitere Optionen">
	<ul class="pb-4 sm:pt-4">
		<li>
			<button class={[zeile, 'disabled:opacity-38']} disabled={!anzahl} onclick={() => ((offen = false), onauswaehlen())}>
				<Icon name="print" class="text-on-surface-variant" />Drucken
			</button>
		</li>
		<li><a class={zeile} {...pdfLink(pdf)} onclick={zu}><Icon name="picture_as_pdf" class="text-on-surface-variant" />Liste als PDF</a></li>
		<li><a class={zeile} href="/api/export.xlsx" download onclick={zu}><Icon name="table_view" class="text-on-surface-variant" />Excel-Tabelle</a></li>
	</ul>
</Blatt>
