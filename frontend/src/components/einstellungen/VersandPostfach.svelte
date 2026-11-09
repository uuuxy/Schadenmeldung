<script>
	import { onMount } from 'svelte';
	import { api } from '../../lib/api.js';
	import { zustand, ladeEinstellungen } from '../../lib/zustand.svelte.js';
	import { zeige } from '../../lib/snackbar.svelte.js';
	import { frage } from '../../lib/rueckfrage.svelte.js';
	import Icon from '../Icon.svelte';
	import Ladekreis from '../Ladekreis.svelte';
	import Button from '../ui/Button.svelte';
	import Feld from '../ui/Feld.svelte';

	/**
	 * Das Postfach, über das die App ihre Mails verschickt. Am besten trägt es die Person ein, der es
	 * gehört: Das Passwort tippt sie selbst, die App prüft es mit einer Testmail an sie und speichert
	 * es verschlüsselt. Angezeigt wird es nie wieder, nur dass es gespeichert ist.
	 */
	let stand = $state(null);
	let adresse = $state('');
	let passwort = $state('');
	let aendern = $state(false);
	let laeuft = $state(false);
	const seit = (wann) => new Date(wann).toLocaleDateString('de-DE', { day: 'numeric', month: 'long', year: 'numeric' });

	onMount(() =>
		api('/api/postfach')
			.then((s) => (stand = s))
			.catch((e) => zeige(e.message))
	);

	// Danach stimmt auch „Mail an mich“ im Kontomenü, das nur bei laufendem Versand erscheint.
	async function uebernimm(neu, text) {
		stand = neu;
		passwort = '';
		aendern = false;
		zeige(text);
		await ladeEinstellungen().catch(() => {});
	}

	async function speichern() {
		laeuft = true;
		try {
			await uebernimm(await api('/api/postfach', { methode: 'PUT', json: { adresse, passwort } }), 'Postfach gespeichert – die Testmail ist raus');
		} catch (e) {
			zeige(e.message);
		} finally {
			laeuft = false;
		}
	}

	async function entfernen() {
		const ja = await frage({ titel: 'Postfach entfernen?', text: 'Danach verschickt die App keine Mails mehr, bis wieder ein Postfach eingetragen ist.', ja: 'Entfernen' });
		if (!ja) return;
		try {
			await uebernimm(await api('/api/postfach', { methode: 'DELETE' }), 'Postfach entfernt');
		} catch (e) {
			zeige(e.message);
		}
	}
</script>

{#snippet hinweis(text)}
	<p class="rounded-xl bg-surface-container-low p-4 text-body-large text-on-surface">{text}</p>
{/snippet}

<main class="mx-auto flex max-w-3xl flex-col gap-6 px-4 pt-6 pb-12">
	<p class="text-body-medium text-on-surface-variant">
		Über dieses Postfach verschickt die App ihre Mails: neue Meldungen an den Hausmeister und die Nachricht an die Meldenden, sobald etwas erledigt ist.
	</p>
	{#if !stand}
		<div class="grid place-items-center py-12 text-primary"><Ladekreis groesse={32} /></div>
	{:else if !stand.mailserver}
		{@render hinweis('Mails sind auf dem Server ausgeschaltet: In der .env fehlt SMTP_HOST.')}
	{:else if stand.fest || (stand.adresse && !aendern)}
		<section class="flex flex-col gap-4 rounded-xl bg-surface-container-low p-4">
			<div class="flex items-center gap-4">
				<span class="grid size-10 shrink-0 place-items-center rounded-full bg-success-container text-on-success-container"><Icon name="check" /></span>
				<div class="min-w-0 flex-1">
					<p class="truncate text-body-large text-on-surface">{stand.adresse}</p>
					<p class="text-body-medium text-on-surface-variant">
						{stand.fest ? 'Fest auf dem Server eingestellt (.env)' : `Passwort verschlüsselt gespeichert · seit ${seit(stand.eingetragen)}`}
					</p>
				</div>
			</div>
			{#if !stand.fest}
				<div class="flex flex-wrap justify-end gap-2">
					<Button variant="ghost" onclick={entfernen}>Entfernen</Button>
					<Button variant="tonal" onclick={() => ((adresse = stand.adresse), (aendern = true))}>Ändern</Button>
				</div>
			{/if}
		</section>
	{:else if !stand.speicherbar}
		{@render hinweis('Zum sicheren Speichern fehlt auf dem Server SCHLUESSEL in der .env. ./update.sh legt ihn beim nächsten Aufspielen an.')}
	{:else}
		<section class="flex flex-col gap-4 rounded-xl bg-surface-container-low p-4">
			<Feld etikett="E-Mail des Postfachs" bind:wert={adresse} autocomplete="off" autocapitalize="none" spellcheck="false" hinweis={zustand.einstellungen.domain ? `Der Name vor dem @ genügt – @${zustand.einstellungen.domain} ergänzt die App.` : ''} />
			<Feld etikett="Passwort" typ="password" bind:wert={passwort} autocomplete="new-password" hinweis="Wird verschlüsselt gespeichert und hier nie wieder angezeigt." />
			<p class="text-body-small text-on-surface-variant">
				Zur Probe geht eine Mail an dich. Bitte genau tippen: Nach mehreren falschen Passwörtern sperrt der Mailserver der Schule die App für eine Weile.
			</p>
			<div class="flex flex-wrap justify-end gap-2">
				{#if aendern}<Button variant="ghost" onclick={() => ((aendern = false), (passwort = ''))}>Abbrechen</Button>{/if}
				<Button disabled={laeuft || !adresse.trim() || !passwort} onclick={speichern}>
					{#if laeuft}<Ladekreis />{:else}<Icon name="check" groesse={20} />{/if}Prüfen und speichern
				</Button>
			</div>
		</section>
	{/if}
</main>
