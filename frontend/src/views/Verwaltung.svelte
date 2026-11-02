<script>
	import { api } from '../lib/api.js';
	import { zustand } from '../lib/zustand.svelte.js';
	import { gehe } from '../lib/weg.svelte.js';
	import { zeige } from '../lib/snackbar.svelte.js';
	import Icon from '../components/Icon.svelte';
	import Blatt from '../components/Blatt.svelte';
	import KopfLeiste from '../components/KopfLeiste.svelte';
	import Button from '../components/ui/Button.svelte';
	import Feld from '../components/ui/Feld.svelte';
	import Reiter from '../components/ui/Reiter.svelte';
	import FilterChips from '../components/ui/FilterChips.svelte';

	/** Verwaltung: Räume mit QR-Aufklebern und die Personen, die mehr sehen als das Meldeformular. */
	let reiter = $state('raeume');

	// ── Räume ────────────────────────────────────────────────────────────────
	let text = $state('');
	let laeuft = $state(false);
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

	// ── Personen ─────────────────────────────────────────────────────────────
	const rollenNamen = { hausmeister: 'Hausmeister', leitung: 'Leitung', admin: 'Verwaltung' };
	const rollenText = {
		hausmeister: 'bearbeitet die Meldungen und bekommt Mails',
		leitung: 'sieht alle Meldungen, ändert nichts',
		admin: 'alles, auch Räume und Personen'
	};
	const rollenWahl = Object.entries(rollenNamen).map(([wert, text]) => ({ wert, text }));
	let personen = $state([]);
	let neueAdresse = $state('');
	let neueRolle = $state('hausmeister');
	let bearbeitet = $state(null);
	let blattOffen = $state(false);

	async function ladePersonen() {
		try {
			personen = await api('/api/rollen');
		} catch (e) {
			zeige(e.message);
		}
	}

	$effect(() => {
		if (reiter === 'personen') ladePersonen();
	});

	async function rolleSetzen(email, rolle) {
		try {
			personen = await api('/api/rollen', { methode: 'POST', json: { email, rolle } });
			return true;
		} catch (e) {
			zeige(e.message);
			return false;
		}
	}

	async function hinzufuegen() {
		if (!neueAdresse.trim()) return;
		if (await rolleSetzen(neueAdresse, neueRolle)) {
			zeige(`${rollenNamen[neueRolle]}: ${neueAdresse.trim()}`);
			neueAdresse = '';
		}
	}

	function bearbeiten(p) {
		bearbeitet = { ...p };
		blattOffen = true;
	}

	async function entfernen(p) {
		try {
			personen = await api(`/api/rollen/${encodeURIComponent(p.email)}`, { methode: 'DELETE' });
			blattOffen = false;
			zeige(`${p.email} ist wieder Kollegium`);
		} catch (e) {
			zeige(e.message);
		}
	}
</script>

<KopfLeiste titel="Räume & Personen" zurueck={() => gehe('/')} />

<div class="mx-auto max-w-3xl px-4">
	<Reiter
		etikett="Verwaltung"
		reiter={[
			{ id: 'raeume', label: 'Räume' },
			{ id: 'personen', label: 'Personen' }
		]}
		aktiv={reiter}
		onwahl={(id) => (reiter = id)}
	/>
</div>

{#if reiter === 'raeume'}
	<main class="mx-auto flex max-w-3xl flex-col gap-6 px-4 pt-6 pb-12">
		<section class="flex flex-col gap-4 rounded-xl bg-surface-container-low p-4">
			<div>
				<h2 class="text-title-medium text-on-surface">Räume hinzufügen</h2>
				<p class="mt-1 text-body-medium text-on-surface-variant">Eine Zeile je Raum, der Bereich nach einem Semikolon. Spalten aus einer Tabelle lassen sich direkt einfügen.</p>
			</div>
			<Feld etikett="Räume" placeholder={'204; Hauptgebäude · 2. OG\nTurnhalle; Sport'} bind:wert={text} mehrzeilig zeilen={5} />
			<Button class="self-end" disabled={!text.trim() || laeuft} onclick={raeumeHinzufuegen}><Icon name="add" groesse={20} />Hinzufügen</Button>
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
{:else}
	<main class="mx-auto flex max-w-3xl flex-col gap-6 px-4 pt-6 pb-12">
		<p class="text-body-medium text-on-surface-variant">
			Alle Postfächer der Schule können melden. Wer mehr sehen soll – Hausmeister, Schulleitung, Verwaltung –, steht hier.
		</p>

		<section class="flex flex-col gap-4 rounded-xl bg-surface-container-low p-4">
			<h2 class="text-title-medium text-on-surface">Person hinzufügen</h2>
			<Feld etikett="Name oder E-Mail" bind:wert={neueAdresse} autocomplete="off" autocapitalize="none" spellcheck="false" hinweis={zustand.einstellungen.domain ? `Der Name vor dem @ genügt – @${zustand.einstellungen.domain} ergänzt die App.` : ''} />
			<FilterChips etikett="Rolle" abwaehlbar={false} optionen={rollenWahl} wert={neueRolle} onwahl={(r) => (neueRolle = r)} />
			<p class="text-body-small text-on-surface-variant">{rollenNamen[neueRolle]}: {rollenText[neueRolle]}.</p>
			<Button class="self-end" disabled={!neueAdresse.trim()} onclick={hinzufuegen}><Icon name="add" groesse={20} />Hinzufügen</Button>
		</section>

		<ul class="overflow-hidden rounded-xl bg-surface-container-low">
			{#each personen as p (p.email)}
				<li>
					<button class="flex min-h-16 w-full items-center gap-4 px-4 py-2 text-left disabled:cursor-default" disabled={p.fest || p.email === zustand.benutzer?.email} onclick={() => bearbeiten(p)}>
						<span class="grid size-10 shrink-0 place-items-center rounded-full bg-primary-container text-title-medium text-on-primary-container">{p.email.charAt(0).toUpperCase()}</span>
						<span class="min-w-0 flex-1">
							<span class="block truncate text-body-large text-on-surface">{p.email}</span>
							<span class="block text-body-medium text-on-surface-variant">{rollenNamen[p.rolle]}{p.fest ? ' · fest eingetragen' : p.email === zustand.benutzer?.email ? ' · du' : ''}</span>
						</span>
						{#if !p.fest && p.email !== zustand.benutzer?.email}<Icon name="edit_note" class="text-on-surface-variant" />{/if}
					</button>
				</li>
			{:else}
				<li class="px-4 py-6 text-center text-body-medium text-on-surface-variant">Noch niemand eingetragen.</li>
			{/each}
		</ul>
	</main>
{/if}

<Blatt bind:offen={blattOffen} titel={bearbeitet?.email ?? ''}>
	{#if bearbeitet}
		<div class="flex flex-col gap-4 px-6 pb-6">
			<FilterChips etikett="Rolle" abwaehlbar={false} optionen={rollenWahl} wert={bearbeitet.rolle} onwahl={async (r) => { if (await rolleSetzen(bearbeitet.email, r)) bearbeitet.rolle = r; }} />
			<p class="text-body-medium text-on-surface-variant">{rollenNamen[bearbeitet.rolle]}: {rollenText[bearbeitet.rolle]}.</p>
			<div class="flex justify-between gap-2">
				<Button variant="danger" onclick={() => entfernen(bearbeitet)}><Icon name="delete" groesse={18} />Rolle entfernen</Button>
				<Button variant="ghost" onclick={() => (blattOffen = false)}>Fertig</Button>
			</div>
		</div>
	{/if}
</Blatt>
