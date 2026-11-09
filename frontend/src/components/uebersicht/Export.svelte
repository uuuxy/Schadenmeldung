<script>
	import { zustand } from '../../lib/zustand.svelte.js';
	import { kategorieNach } from '../../lib/kategorien.js';
	import Icon from '../Icon.svelte';
	import Blatt from '../Blatt.svelte';

	/**
	 * Drucken und Export für alle mit Übersicht. „Meldungen auswählen“ beginnt die Auswahl in der Liste
	 * (onauswaehlen), gedruckt wird dann je Meldung eine Seite mit Foto. Das PDF der Liste zeigt sie,
	 * wie sie gerade steht (Reiter und Filter), die Excel-Tabelle alle gespeicherten Meldungen zum
	 * Sortieren und Filtern. Alles ohne Namen der Meldenden: Die Dateien verlassen die App.
	 * Als Symbol in der Kopfleiste (M3: Aktionen für den ganzen Bildschirm stehen in der App Bar), der
	 * Drucker, weil Drucken das Häufigste ist.
	 */
	let { liste, bereich, art = null, anzahl, onauswaehlen } = $props();
	let offen = $state(false);

	const pdf = $derived(`/api/export.pdf?liste=${liste}${bereich ? `&bereich=${encodeURIComponent(bereich)}` : ''}${art ? `&art=${art}` : ''}`);
	const filter = $derived([bereich, art && kategorieNach(art).name].filter(Boolean).join(', '));
	const pdfText = $derived.by(() => {
		const menge = liste === 'offen' ? (anzahl === 1 ? '1 offene Meldung' : `${anzahl} offene Meldungen`) : anzahl === 1 ? '1 erledigte Meldung der letzten 30 Tage' : `${anzahl} erledigte Meldungen der letzten 30 Tage`;
		return `${menge}${filter ? `, nur ${filter}` : ''}, wie hier zu sehen – zum Drucken oder Weiterschicken.`;
	});
</script>

{#snippet inhalt(symbol, titel, text, hinten)}
	<span class="grid size-10 shrink-0 place-items-center rounded-full bg-secondary-container text-on-secondary-container"><Icon name={symbol} /></span>
	<span class="min-w-0 flex-1">
		<span class="block text-body-large">{titel}</span>
		<span class="block text-body-medium text-on-surface-variant">{text}</span>
	</span>
	<Icon name={hinten} class="text-on-surface-variant" />
{/snippet}

{#snippet eintrag(href, symbol, titel, text)}
	<li>
		<a {href} download class="schicht flex items-center gap-4 px-6 py-3 text-on-surface" onclick={() => setTimeout(() => (offen = false))}>
			{@render inhalt(symbol, titel, text, 'download')}
		</a>
	</li>
{/snippet}

<button class="icon-btn text-on-surface-variant" aria-label="Drucken und exportieren" title="Drucken und exportieren" onclick={() => (offen = true)}><Icon name="print" /></button>

<Blatt bind:offen titel="Drucken und exportieren">
	<ul class="pb-2">
		{#if anzahl}
			<li>
				<button class="schicht flex w-full items-center gap-4 px-6 py-3 text-left text-on-surface" onclick={() => ((offen = false), onauswaehlen())}>
					{@render inhalt('print', 'Meldungen auswählen', 'Je Meldung eine Seite mit Foto – danach in der Liste antippen, was mit soll.', 'chevron_right')}
				</button>
			</li>
		{/if}
		{@render eintrag(pdf, 'picture_as_pdf', 'Liste als PDF', pdfText)}
		{@render eintrag('/api/export.xlsx', 'table_view', 'Excel-Tabelle', 'Alle gespeicherten Meldungen, offene und erledigte – zum Sortieren und Filtern.')}
	</ul>
	<p class="px-6 pb-6 text-body-small text-on-surface-variant">
		Ohne Namen der Meldenden. Erledigte Meldungen bleiben {zustand.einstellungen.aufbewahrungTage} Tage gespeichert.
	</p>
</Blatt>
