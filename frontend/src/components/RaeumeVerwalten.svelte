<script>
	import { api } from '../lib/api.js';
	import { zustand } from '../lib/zustand.svelte.js';
	import { gehe } from '../lib/weg.svelte.js';
	import { zeige } from '../lib/snackbar.svelte.js';
	import Icon from './Icon.svelte';
	import Button from './ui/Button.svelte';
	import Feld from './ui/Feld.svelte';

	/** Räume anlegen (Zeilen, eingefügte Tabellenspalten oder eine Datei), QR-Aufkleber, Räume löschen. */
	let text = $state('');
	let laeuft = $state(false);
	let dateiWahl = $state();
	const gruppen = $derived.by(() => {
		const bereiche = new Map();
		for (const r of zustand.raeume) {
			const b = r.bereich || 'Ohne Bereich';
			if (!bereiche.has(b)) bereiche.set(b, []);
			bereiche.get(b).push(r);
		}
		return [...bereiche].map(([bereich, raeume]) => ({ bereich, raeume }));
	});

	async function raeumeHinzufuegen() {
		if (!text.trim() || laeuft) return;
		laeuft = true;
		try {
			const antwort = await api('/api/raeume', { methode: 'POST', json: { text } });
			zustand.raeume = antwort.raeume;
			zeige(antwort.neu === 1 ? '1 Raum angelegt' : `${antwort.neu} Räume angelegt`);
			text = '';
		} catch (e) {
			zeige(e.message);
		} finally {
			laeuft = false;
		}
	}

	// Eine Excel- oder CSV-Datei liest der Server; die Räume landen als Zeilen im Feld, damit man vor
	// dem Anlegen sieht, was kommt. Angelegt wird wie beim Eintippen mit „Hinzufügen“.
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

	async function raumLoeschen(r) {
		if (!confirm(`Raum „${r.name}“ löschen? Sein QR-Code funktioniert danach nicht mehr. Meldungen bleiben erhalten.`)) return;
		try {
			await api(`/api/raeume/${r.id}`, { methode: 'DELETE' });
			zustand.raeume = zustand.raeume.filter((x) => x.id !== r.id);
			zeige(`„${r.name}“ gelöscht`);
		} catch (e) {
			zeige(e.message);
		}
	}
</script>

<main class="mx-auto flex max-w-3xl flex-col gap-6 px-4 pt-6 pb-12">
	<section class="flex flex-col gap-4 rounded-xl bg-surface-container-low p-4">
		<div>
			<h2 class="text-title-medium text-on-surface">Räume hinzufügen</h2>
			<p class="mt-1 text-body-medium text-on-surface-variant">
				Eine Zeile je Raum, der Bereich nach einem Semikolon. Spalten aus einer Tabelle lassen sich direkt einfügen, oder eine Excel- oder CSV-Datei wählen: erste Spalte Raum, zweite Bereich.
			</p>
		</div>
		<Feld etikett="Räume" placeholder={'204; Hauptgebäude · 2. OG\nTurnhalle; Sport'} bind:wert={text} mehrzeilig zeilen={5} />
		<input bind:this={dateiWahl} type="file" accept=".xlsx,.csv,.txt,text/csv,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet" class="hidden" onchange={dateiGewaehlt} />
		<div class="flex flex-wrap justify-end gap-2">
			<Button variant="ghost" disabled={laeuft} onclick={() => dateiWahl.click()}><Icon name="upload_file" groesse={20} />Aus Excel oder CSV</Button>
			<Button disabled={!text.trim() || laeuft} onclick={raeumeHinzufuegen}><Icon name="add" groesse={20} />Hinzufügen</Button>
		</div>
	</section>

	{#if zustand.raeume.length}
		<Button variant="tonal" size="lg" onclick={() => gehe('/raeume/qr')}><Icon name="qr_code_2" />QR-Aufkleber drucken · {zustand.raeume.length}</Button>

		<section>
			{#each gruppen as g (g.bereich)}
				<h3 class="px-1 pt-4 pb-2 text-title-small text-primary">{g.bereich}</h3>
				<ul class="overflow-hidden rounded-xl bg-surface-container-low">
					{#each g.raeume as r (r.id)}
						<li class="flex min-h-14 items-center gap-4 pr-1 pl-4">
							<Icon name="meeting_room" class="text-on-surface-variant" />
							<span class="min-w-0 flex-1 truncate text-body-large text-on-surface">{r.name}</span>
							<Button variant="symbol" aria-label="Raum {r.name} löschen" onclick={() => raumLoeschen(r)}><Icon name="delete" /></Button>
						</li>
					{/each}
				</ul>
			{/each}
		</section>
	{/if}
</main>
