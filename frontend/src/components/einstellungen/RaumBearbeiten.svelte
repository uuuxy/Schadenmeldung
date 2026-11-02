<script>
	import { api } from '../../lib/api.js';
	import { zustand } from '../../lib/zustand.svelte.js';
	import { zeige } from '../../lib/snackbar.svelte.js';
	import { frage } from '../../lib/rueckfrage.svelte.js';
	import Icon from '../Icon.svelte';
	import Blatt from '../Blatt.svelte';
	import Button from '../ui/Button.svelte';
	import Feld from '../ui/Feld.svelte';

	/**
	 * Einen Raum umbenennen oder einem anderen Bereich zuordnen. Der QR-Code bleibt dabei gleich,
	 * der Aufkleber an der Tür gilt also weiter. Löschen fragt einmal nach, denn danach führt der
	 * Aufkleber ins Leere.
	 */
	let { raum = $bindable() } = $props();
	let offen = $state(false);
	let name = $state('');
	let bereich = $state('');
	let laeuft = $state(false);

	$effect(() => {
		if (!raum) return;
		name = raum.name;
		bereich = raum.bereich;
		offen = true;
	});
	$effect(() => {
		if (!offen) raum = null;
	});

	async function speichern() {
		laeuft = true;
		try {
			zustand.raeume = await api(`/api/raeume/${raum.id}`, { methode: 'PUT', json: { name, bereich } });
			zeige('Raum gespeichert');
			offen = false;
		} catch (e) {
			zeige(e.message);
		} finally {
			laeuft = false;
		}
	}

	async function loeschen() {
		const r = raum;
		if (!(await frage({ titel: `„${r.name}“ löschen?`, text: 'Sein QR-Aufkleber führt danach ins Leere. Gemeldete Schäden bleiben erhalten.', ja: 'Löschen' }))) return;
		try {
			await api(`/api/raeume/${r.id}`, { methode: 'DELETE' });
			zustand.raeume = zustand.raeume.filter((x) => x.id !== r.id);
			zeige(`„${r.name}“ gelöscht`);
			offen = false;
		} catch (e) {
			zeige(e.message);
		}
	}
</script>

<Blatt bind:offen titel="Raum bearbeiten">
	<form class="flex flex-col gap-4 px-6 pb-6" onsubmit={(e) => (e.preventDefault(), speichern())}>
		<Feld etikett="Name" bind:wert={name} maxlength="60" autocomplete="off" enterkeyhint="done" required />
		<Feld etikett="Bereich" bind:wert={bereich} maxlength="80" autocomplete="off" enterkeyhint="done" hinweis="Der QR-Aufkleber bleibt gültig." />
		<div class="mt-2 flex items-center justify-between gap-2">
			<Button variant="warnend" type="button" class="-ml-3" onclick={loeschen}><Icon name="delete" groesse={18} />Löschen</Button>
			<Button type="submit" disabled={laeuft || !name.trim()}>Speichern</Button>
		</div>
	</form>
</Blatt>
