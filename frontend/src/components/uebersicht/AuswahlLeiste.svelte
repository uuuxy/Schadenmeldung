<script>
	import Icon from '../Icon.svelte';
	import KopfLeiste from '../KopfLeiste.svelte';
	import DruckKnopf, { HOECHSTENS } from './DruckKnopf.svelte';

	/**
	 * Die Kopfleiste während einer Auswahl, wie in Gmail, Fotos und Dateien von Google: vorn das Kreuz,
	 * als Titel die Zahl der gewählten, rechts Drucken und Alle (alle sichtbaren oder keine). Drucken
	 * beendet die Auswahl, denn eine Aktion schließt sie ab (M3 selection).
	 */
	let { gewaehlt, alle, onbeenden, ongedruckt, onalle } = $props();
	const n = $derived(gewaehlt.length);
	const titel = $derived(n ? `${n} ausgewählt${n > HOECHSTENS ? ` · höchstens ${HOECHSTENS}` : ''}` : 'Meldungen antippen');
</script>

<KopfLeiste {titel} zurueck={onbeenden} zurueckSymbol="close" zurueckEtikett="Auswahl beenden" konto={false}>
	{#snippet aktionen()}
		<DruckKnopf ids={gewaehlt} onclick={() => setTimeout(ongedruckt)} />
		<button class="icon-btn text-on-surface-variant" aria-label={alle ? 'Alle abwählen' : 'Alle auswählen'} title={alle ? 'Alle abwählen' : 'Alle auswählen'} onclick={onalle}>
			<Icon name={alle ? 'deselect' : 'select_all'} />
		</button>
	{/snippet}
</KopfLeiste>
