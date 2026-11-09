<script>
	import { zustand, darfBearbeiten } from '../../lib/zustand.svelte.js';
	import { rueckmeldung } from '../../lib/haptik.js';
	import { zuletztGewaehlt, merkeRaum } from '../../lib/zuletzt.js';
	import { raumCodeAus } from '../../lib/weg.svelte.js';
	import Icon from '../Icon.svelte';
	import Button from '../ui/Button.svelte';
	import ZurueckZeile from '../ZurueckZeile.svelte';
	import Feld from '../ui/Feld.svelte';
	import FilterChips from '../ui/FilterChips.svelte';
	import RaumWahl from '../RaumWahl.svelte';
	import HierGemeldet from './HierGemeldet.svelte';
	import QrScanner from './QrScanner.svelte';

	/**
	 * Wo? Mit QR-Code ist der Raum schon gewählt. Sonst die Raumwahl (suchen oder Bereich, dann Raum),
	 * darüber „Zuletzt“ als Vorschläge für einen Tipp. Für alles ohne Raum gibt es „Anderer Ort“.
	 * Am Handy liest das Symbol rechts in der Suche den QR-Aufkleber auch in der App.
	 * Gewählt schrumpft die Frage auf eine Zeile mit „Ändern“.
	 */
	let { raum = $bindable(), ortFrei = $bindable(), freierOrt = $bindable(), codeUnbekannt = false, ongewaehlt } = $props();

	let suche = $state('');
	let zuletzt = $state(zuletztGewaehlt());
	const zuletztRaeume = $derived(zuletzt.map((id) => zustand.raeume.find((r) => r.id === id)).filter(Boolean));
	const nurFreitext = $derived(zustand.raeume.length === 0 || freierOrt);
	// Den Scanner gibt es, wo man mit dem Finger tippt und die Kamera erreichbar ist; am Rechner führt
	// die Suche schneller ans Ziel.
	const kamera = !!navigator.mediaDevices?.getUserMedia && matchMedia('(pointer: coarse)').matches;
	let scannen = $state(false);

	// Gelesen wird nur der Raum-Code aus /r/…, die Adresse davor nicht: Die App öffnet nie, was im Code
	// steht, sie wählt höchstens einen eigenen Raum.
	function codeGelesen(text) {
		let pfad = '';
		try {
			pfad = new URL(text).pathname;
		} catch {
			// kein Link, also kein Aufkleber von uns
		}
		const code = raumCodeAus(pfad);
		const r = code && zustand.raeume.find((x) => x.code === code);
		if (!r) return code ? 'Dieser QR-Code gehört zu keinem Raum mehr.' : 'Das ist kein QR-Aufkleber der Schadenmeldung.';
		scannen = false;
		waehleRaum(r);
	}

	function waehleRaum(r) {
		raum = r;
		freierOrt = false;
		merkeRaum(r.id);
		zuletzt = zuletztGewaehlt();
		rueckmeldung();
		ongewaehlt?.();
	}

	function ortAusSuche(text) {
		ortFrei = text;
		suche = '';
		freierOrt = true;
	}
</script>

{#if codeUnbekannt && !raum}
	<p class="mb-3 rounded-md bg-warning-container px-4 py-3 text-body-medium text-on-warning-container">
		Dieser QR-Code gehört zu keinem Raum mehr. Bitte den Raum auswählen.
	</p>
{/if}

{#if raum}
	<button class="flex w-full items-center gap-4 rounded-xl bg-secondary-container p-4 text-left text-on-secondary-container" onclick={() => (raum = null)}>
		<Icon name="location_on" gefuellt groesse={24} />
		<span class="min-w-0 flex-1">
			<span class="block truncate text-title-medium">{raum.name}</span>
			{#if raum.bereich}<span class="block truncate text-body-medium">{raum.bereich}</span>{/if}
		</span>
		<span class="shrink-0 text-label-large">Ändern</span>
	</button>
	<HierGemeldet raumId={raum.id} />
{:else if nurFreitext}
	<!-- Wie in einem Bereich: oben der Pfeil zurück zu den Räumen, kein Link unter dem Feld. -->
	{#if zustand.raeume.length}<ZurueckZeile titel="Anderer Ort" etikett="Zurück zu den Räumen" onzurueck={() => (freierOrt = false)} />{/if}
	<div class={zustand.raeume.length && 'mt-1'}>
		<Feld etikett="Wo ist der Schaden?" placeholder="z. B. Schulhof, Treppenhaus Nord" bind:wert={ortFrei} autocomplete="off" enterkeyhint="next" maxlength="80" />
	</div>
	{#if !zustand.raeume.length && darfBearbeiten()}
		<p class="mt-2 text-body-small text-on-surface-variant">Tipp: Unter Einstellungen (Kontomenü) die Raumliste anlegen – dann genügt hier ein Tipp.</p>
	{/if}
{:else if scannen}
	<ZurueckZeile titel="QR-Code scannen" etikett="Zurück zu den Räumen" onzurueck={() => (scannen = false)} />
	<div class="mt-1"><QrScanner onlesen={codeGelesen} /></div>
{:else}
	<RaumWahl raeume={zustand.raeume} bind:suche onwahl={waehleRaum} anderOrt={() => (freierOrt = true)} scannen={kamera ? () => (scannen = true) : null}>
		{#snippet vorschlaege()}
			{#if zuletztRaeume.length}
				<h3 class="mt-5 mb-2 text-label-large text-on-surface-variant">Zuletzt</h3>
				<FilterChips etikett="Zuletzt gewählte Räume" optionen={zuletztRaeume.map((r) => ({ wert: String(r.id), text: r.name, icon: 'history' }))} wert={null} onwahl={(id) => waehleRaum(zuletztRaeume.find((r) => String(r.id) === id))} />
			{/if}
		{/snippet}
		{#snippet ohneTreffer(text)}
			<Button variant="ghost" class="mt-2 -ml-4" onclick={() => ortAusSuche(text)}><Icon name="location_on" groesse={18} />„{text}“ als Ort nehmen</Button>
		{/snippet}
	</RaumWahl>
{/if}
