<script>
	import { zustand } from '../../lib/zustand.svelte.js';
	import { kategorieNach } from '../../lib/kategorien.js';
	import Icon from '../Icon.svelte';
	import Blatt from '../Blatt.svelte';
	import Button from '../ui/Button.svelte';

	/**
	 * Export für alle mit Übersicht. Das PDF zeigt die Liste, wie sie gerade steht (Reiter und
	 * Filter), zum Drucken oder Weiterschicken; die Excel-Tabelle alle gespeicherten Meldungen zum
	 * Sortieren und Filtern. Beide ohne Namen der Meldenden: Die Dateien verlassen die App.
	 */
	let { liste, bereich, art = null, anzahl } = $props();
	let offen = $state(false);

	const pdf = $derived(`/api/export.pdf?liste=${liste}${bereich ? `&bereich=${encodeURIComponent(bereich)}` : ''}${art ? `&art=${art}` : ''}`);
	const filter = $derived([bereich, art && kategorieNach(art).name].filter(Boolean).join(', '));
	const pdfText = $derived.by(() => {
		const menge = liste === 'offen' ? (anzahl === 1 ? '1 offene Meldung' : `${anzahl} offene Meldungen`) : anzahl === 1 ? '1 erledigte Meldung der letzten 30 Tage' : `${anzahl} erledigte Meldungen der letzten 30 Tage`;
		return `${menge}${filter ? `, nur ${filter}` : ''}, wie hier zu sehen – zum Drucken oder Weiterschicken.`;
	});
</script>

{#snippet eintrag(href, symbol, titel, text)}
	<li>
		<a {href} download class="schicht flex items-center gap-4 px-6 py-3 text-on-surface" onclick={() => setTimeout(() => (offen = false))}>
			<span class="grid size-10 shrink-0 place-items-center rounded-full bg-secondary-container text-on-secondary-container"><Icon name={symbol} /></span>
			<span class="min-w-0 flex-1">
				<span class="block text-body-large">{titel}</span>
				<span class="block text-body-medium text-on-surface-variant">{text}</span>
			</span>
			<Icon name="download" class="text-on-surface-variant" />
		</a>
	</li>
{/snippet}

<Button variant="ghost" class="-mr-3" onclick={() => (offen = true)}><Icon name="download" groesse={18} />Exportieren</Button>

<Blatt bind:offen titel="Exportieren">
	<ul class="pb-2">
		{@render eintrag(pdf, 'picture_as_pdf', 'PDF', pdfText)}
		{@render eintrag('/api/export.xlsx', 'table_view', 'Excel-Tabelle', 'Alle gespeicherten Meldungen, offene und erledigte – zum Sortieren und Filtern.')}
	</ul>
	<p class="px-6 pb-6 text-body-small text-on-surface-variant">
		Ohne Namen der Meldenden. Erledigte Meldungen bleiben {zustand.einstellungen.aufbewahrungTage} Tage gespeichert.
	</p>
</Blatt>
