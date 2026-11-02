<script>
	import { renderSVG } from 'uqr';
	import { zustand, darfBearbeiten } from '../lib/zustand.svelte.js';
	import { zurueck } from '../lib/weg.svelte.js';
	import { formatNach, aufBlaetter, gemerktesFormat, merkeFormat } from '../lib/etiketten.js';
	import Icon from '../components/Icon.svelte';
	import KopfLeiste from '../components/KopfLeiste.svelte';
	import Button from '../components/ui/Button.svelte';
	import FormatKacheln from '../components/qr/FormatKacheln.svelte';
	import Druckblatt from '../components/qr/Druckblatt.svelte';

	/**
	 * QR-Aufkleber in vier Formaten: zum Ausschneiden, auf Etikettenbögen groß und klein, oder ein
	 * Raum je Blatt. Die Vorschau zeigt die echten Blätter verkleinert; gedruckt wird in
	 * Originalgröße ohne Seitenrand, sonst träfen die Aufkleber die Etiketten nicht.
	 */
	// Der Code führt auf die Adresse, unter der die App gerade läuft.
	const basis = location.origin;
	// Dieselbe Reihenfolge wie in der Raumliste (vom Server): so liegen die Aufkleber eines Flurs beisammen.
	const raeume = $derived(zustand.raeume);
	const qr = (r) => renderSVG(`${basis}/r/${r.code}`, { border: 1, ecc: 'M' });

	let formatKey = $state(gemerktesFormat());
	const format = $derived(formatNach(formatKey));
	const blaetter = $derived(aufBlaetter(raeume, format));
	// Am Bildschirm genügen zwei Blätter als Vorschau; gedruckt werden alle.
	const VORSCHAU = 2;

	// Ein Blatt ist 210 mm breit (96 px je Zoll); am Bildschirm wird es auf die freie Breite verkleinert,
	// am Rechner auf höchstens 80 %, damit ein Blatt nicht den ganzen Bildschirm füllt.
	let breite = $state(0);
	const zoom = $derived(breite ? Math.min(0.8, breite / ((210 * 96) / 25.4)) : 0.45);
	// Die Maße des Formats als CSS-Variablen für die Blätter (components/qr/Druckblatt.svelte).
	const mass = $derived(
		[
			`--zoom: ${zoom}`,
			`--spalten: ${format.spalten}`,
			`--breite: ${format.breite}mm`,
			`--hoehe: ${format.hoehe}mm`,
			`--oben: ${format.oben}mm`,
			`--links: ${format.links}mm`,
			`--innen: ${format.anordnung === 'seite' ? format.hoehe * 0.13 : Math.min(format.breite, format.hoehe) * 0.06}mm`,
			`--qr: ${format.qr}mm`,
			`--marke: ${format.schrift.marke}pt`,
			`--name: ${format.schrift.name}pt`,
			`--bereich: ${format.schrift.bereich}pt`,
			`--tipp: ${format.schrift.tipp}pt`
		].join('; ')
	);

	function waehle(key) {
		formatKey = key;
		merkeFormat(key);
	}
	const menge = (n, eins, mehr) => (n === 1 ? `1 ${eins}` : `${n} ${mehr}`);
</script>

<div class="contents print:hidden">
	<KopfLeiste titel="QR-Aufkleber" konto={false} zurueck={() => zurueck(darfBearbeiten() ? '/einstellungen' : '/')}>
		{#snippet aktionen()}
			<Button class="mr-2" disabled={raeume.length === 0} onclick={() => window.print()}><Icon name="print" groesse={18} />Drucken</Button>
		{/snippet}
	</KopfLeiste>

	<section class="mx-auto max-w-5xl px-4">
		<h2 class="mb-3 text-title-medium text-on-surface">Format</h2>
		<FormatKacheln wert={format.key} onwahl={waehle} />
		<p class="mt-4 text-body-medium text-on-surface">{format.hinweis}</p>
		<p class="mt-1 text-body-medium text-on-surface-variant">
			{menge(raeume.length, 'Aufkleber', 'Aufkleber')} auf {menge(blaetter.length, 'Blatt', 'Blättern')}. Beim Drucken „Tatsächliche Größe“ (100 %) wählen,
			sonst verrutschen die Aufkleber.
		</p>
	</section>
</div>

<div class="mx-auto max-w-5xl px-4 pt-6 pb-10 print:max-w-none print:p-0">
	{#if raeume.length === 0}
		<p class="text-body-large text-on-surface-variant print:hidden">Noch keine Räume angelegt (Einstellungen → Räume).</p>
	{/if}
	<div class="flex flex-col items-center gap-6 print:block" bind:clientWidth={breite} style={mass}>
		{#each blaetter as blatt, b (b)}
			<Druckblatt raeume={blatt} {format} {qr} nurDruck={b >= VORSCHAU} />
		{/each}
	</div>
	{#if blaetter.length > VORSCHAU}
		<p class="mt-4 text-center text-body-medium text-on-surface-variant print:hidden">… und {menge(blaetter.length - VORSCHAU, 'weiteres Blatt', 'weitere Blätter')}, gedruckt werden alle.</p>
	{/if}
</div>
