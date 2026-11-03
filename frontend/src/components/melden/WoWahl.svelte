<script>
	import { zustand, darfBearbeiten } from '../../lib/zustand.svelte.js';
	import { rueckmeldung } from '../../lib/haptik.js';
	import { zuletztGewaehlt, merkeRaum } from '../../lib/zuletzt.js';
	import Icon from '../Icon.svelte';
	import Button from '../ui/Button.svelte';
	import Feld from '../ui/Feld.svelte';
	import FilterChips from '../ui/FilterChips.svelte';
	import RaumWahl from '../RaumWahl.svelte';
	import HierGemeldet from './HierGemeldet.svelte';

	/**
	 * Wo? Mit QR-Code ist der Raum schon gewählt. Sonst die Raumwahl (suchen oder Bereich, dann Raum),
	 * darüber „Zuletzt“ als Vorschläge für einen Tipp. Für alles ohne Raum gibt es „Anderer Ort“.
	 * Gewählt schrumpft die Frage auf eine Zeile mit „Ändern“.
	 */
	let { raum = $bindable(), ortFrei = $bindable(), freierOrt = $bindable(), codeUnbekannt = false, ongewaehlt } = $props();

	let suche = $state('');
	let zuletzt = $state(zuletztGewaehlt());
	const zuletztRaeume = $derived(zuletzt.map((id) => zustand.raeume.find((r) => r.id === id)).filter(Boolean));
	const nurFreitext = $derived(zustand.raeume.length === 0 || freierOrt);

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
	<Feld etikett="Wo ist der Schaden?" placeholder="z. B. Schulhof, Treppenhaus Nord" bind:wert={ortFrei} autocomplete="off" enterkeyhint="next" maxlength="80" />
	{#if zustand.raeume.length}
		<Button variant="ghost" class="mt-2 -ml-4" onclick={() => (freierOrt = false)}><Icon name="meeting_room" groesse={18} />Doch einen Raum wählen</Button>
	{:else if darfBearbeiten()}
		<p class="mt-2 text-body-small text-on-surface-variant">Tipp: Unter Einstellungen (Kontomenü) die Raumliste anlegen – dann genügt hier ein Tipp.</p>
	{/if}
{:else}
	<RaumWahl raeume={zustand.raeume} bind:suche onwahl={waehleRaum}>
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
	<Button variant="ghost" class="mt-3 -ml-4" onclick={() => (freierOrt = true)}><Icon name="edit_note" groesse={18} />Anderer Ort (Schulhof, Flur …)</Button>
{/if}
