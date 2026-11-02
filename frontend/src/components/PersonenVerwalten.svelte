<script>
	import { api } from '../lib/api.js';
	import { zustand } from '../lib/zustand.svelte.js';
	import { zeige } from '../lib/snackbar.svelte.js';
	import { rollenNamen, rollenText, rollenWahl } from '../lib/rollen.js';
	import Icon from './Icon.svelte';
	import Blatt from './Blatt.svelte';
	import Button from './ui/Button.svelte';
	import Feld from './ui/Feld.svelte';
	import FilterChips from './ui/FilterChips.svelte';

	/** Wer mehr sieht als das Meldeformular: Personen eintragen, Rolle ändern, Rolle entfernen. */
	let personen = $state([]);
	let neueAdresse = $state('');
	let neueRolle = $state('hausmeister');
	let bearbeitet = $state(null);
	let blattOffen = $state(false);

	$effect(() => {
		api('/api/rollen')
			.then((liste) => (personen = liste))
			.catch((e) => zeige(e.message));
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

	const aenderbar = (p) => !p.fest && p.email !== zustand.benutzer?.email;
</script>

<main class="mx-auto flex max-w-3xl flex-col gap-6 px-4 pt-6 pb-12">
	<p class="text-body-medium text-on-surface-variant">
		Alle Postfächer der Schule können melden. Wer mehr sehen soll – Hausmeister, Schulleitung, Sekretariat, Admins –, steht hier.
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
				<button class="flex min-h-16 w-full items-center gap-4 px-4 py-2 text-left disabled:cursor-default" disabled={!aenderbar(p)} onclick={() => bearbeiten(p)}>
					<span class="grid size-10 shrink-0 place-items-center rounded-full bg-primary-container text-title-medium text-on-primary-container">{p.email.charAt(0).toUpperCase()}</span>
					<span class="min-w-0 flex-1">
						<span class="block truncate text-body-large text-on-surface">{p.email}</span>
						<span class="block text-body-medium text-on-surface-variant">{rollenNamen[p.rolle]}{p.fest ? ' · fest eingetragen' : p.email === zustand.benutzer?.email ? ' · du' : ''}</span>
					</span>
					{#if aenderbar(p)}<Icon name="edit_note" class="text-on-surface-variant" />{/if}
				</button>
			</li>
		{:else}
			<li class="px-4 py-6 text-center text-body-medium text-on-surface-variant">Noch niemand eingetragen.</li>
		{/each}
	</ul>
</main>

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
