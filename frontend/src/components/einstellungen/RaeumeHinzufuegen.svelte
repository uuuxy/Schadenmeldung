<script>
	import { api } from '../../lib/api.js';
	import { zustand } from '../../lib/zustand.svelte.js';
	import { zeige } from '../../lib/snackbar.svelte.js';
	import { haltFest } from '../../lib/fassung.js';
	import Icon from '../Icon.svelte';
	import Button from '../ui/Button.svelte';
	import Feld from '../ui/Feld.svelte';

	/**
	 * Räume anlegen: Zeilen tippen, Tabellenspalten einfügen oder eine Excel- oder CSV-Datei wählen.
	 * Eine Datei liest der Server; die Räume landen erst als Zeilen im Feld, damit man vor dem
	 * Anlegen sieht, was kommt.
	 */
	let { fertig } = $props();
	let text = $state('');
	let laeuft = $state(false);
	let dateiWahl = $state();
	// Eingelesene Räume stehen im Feld, ohne dass jemand getippt hat: Eine neue Fassung wartet darauf.
	$effect(() => haltFest(() => laeuft || text.trim() !== ''));

	async function hinzufuegen() {
		if (!text.trim() || laeuft) return;
		laeuft = true;
		try {
			const antwort = await api('/api/raeume', { methode: 'POST', json: { text } });
			zustand.raeume = antwort.raeume;
			zeige(antwort.neu === 1 ? '1 Raum angelegt' : `${antwort.neu} Räume angelegt`);
			text = '';
			fertig?.();
		} catch (e) {
			zeige(e.message);
		} finally {
			laeuft = false;
		}
	}

	async function dateiGewaehlt(e) {
		const datei = e.currentTarget.files?.[0];
		e.currentTarget.value = '';
		if (!datei) return;
		const formular = new FormData();
		formular.set('datei', datei);
		laeuft = true;
		try {
			const antwort = await api('/api/raeume/lesen', { methode: 'POST', formular });
			text = antwort.text;
			zeige(`${antwort.anzahl === 1 ? '1 Raum' : `${antwort.anzahl} Räume`} gelesen – bitte prüfen und hinzufügen`);
		} catch (err) {
			zeige(err.message);
		} finally {
			laeuft = false;
		}
	}
</script>

<div class="flex flex-col gap-4">
	<p class="text-body-medium text-on-surface-variant">
		Eine Zeile je Raum, der Bereich nach einem Semikolon. Spalten aus einer Tabelle lassen sich direkt einfügen, oder eine Excel- oder CSV-Datei wählen: erste Spalte Raum, zweite Bereich.
	</p>
	<Feld etikett="Räume" placeholder={'204; Hauptgebäude 2. OG\nTurnhalle; Sport'} bind:wert={text} mehrzeilig zeilen={5} />
	<input bind:this={dateiWahl} type="file" accept=".xlsx,.csv,.txt,text/csv,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" class="hidden" onchange={dateiGewaehlt} />
	<div class="flex flex-wrap justify-end gap-2">
		<Button variant="ghost" disabled={laeuft} onclick={() => dateiWahl.click()}><Icon name="upload_file" groesse={20} />Aus Excel oder CSV</Button>
		<Button disabled={!text.trim() || laeuft} onclick={hinzufuegen}><Icon name="add" groesse={20} />Hinzufügen</Button>
	</div>
</div>
