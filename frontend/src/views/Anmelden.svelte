<script>
	import { api } from '../lib/api.js';
	import { zustand, ladeAlles } from '../lib/zustand.svelte.js';
	import { gehe } from '../lib/weg.svelte.js';
	import Icon from '../components/Icon.svelte';
	import Ladekreis from '../components/Ladekreis.svelte';
	import Button from '../components/ui/Button.svelte';
	import Feld from '../components/ui/Feld.svelte';
	import AlsApp from '../components/melden/AlsApp.svelte';

	/**
	 * Anmeldung mit dem Schul-Postfach, nur einmal je Gerät: Danach bleibt das Handy angemeldet, und
	 * ein QR-Code führt direkt ins Formular. Kam jemand über einen QR-Code hierher, bleibt der Weg
	 * (/r/…) stehen, und nach der Anmeldung ist der Raum schon gewählt.
	 */
	let email = $state('');
	let passwort = $state('');
	let zeigen = $state(false);
	let laeuft = $state(false);
	let fehler = $state('');
	const domain = $derived(zustand.einstellungen.domain);

	async function absenden(e) {
		e.preventDefault();
		if (laeuft) return;
		fehler = '';
		laeuft = true;
		try {
			const benutzer = await api('/api/anmelden', { methode: 'POST', json: { email, passwort } });
			passwort = '';
			await ladeAlles().catch(() => {});
			zustand.benutzer = benutzer;
		} catch (err) {
			fehler = err.message;
		} finally {
			laeuft = false;
		}
	}
</script>

<main class="mx-auto flex min-h-dvh max-w-sm flex-col justify-center px-6 pt-[calc(env(safe-area-inset-top)+2.5rem)] pb-10">
	<div class="mb-10 flex flex-col items-center text-center">
		<span class="mb-6 grid size-20 place-items-center rounded-2xl bg-primary-container text-on-primary-container">
			<Icon name="build" groesse={40} gefuellt />
		</span>
		<h1 class="text-headline-medium text-on-surface">Schadenmeldung</h1>
		{#if zustand.einstellungen.schulname}<p class="mt-1 text-body-large text-on-surface-variant">{zustand.einstellungen.schulname}</p>{/if}
	</div>
	<AlsApp class="mb-6" />

	<form class="flex flex-col gap-5" onsubmit={absenden}>
		<Feld
			etikett={domain ? 'Name oder E-Mail' : 'E-Mail der Schule'}
			bind:wert={email}
			typ={domain ? 'text' : 'email'}
			autocomplete="username"
			autocapitalize="none"
			spellcheck="false"
			enterkeyhint="next"
			required
			hinweis={domain ? `Der Name vor dem @ genügt – @${domain} ergänzt die App.` : ''}
		/>
		<Feld etikett="Passwort des Schul-Postfachs" bind:wert={passwort} typ={zeigen ? 'text' : 'password'} autocomplete="current-password" enterkeyhint="go" required>
			{#snippet nachlaufend()}
				<Button variant="symbol" type="button" aria-label={zeigen ? 'Passwort verbergen' : 'Passwort zeigen'} onclick={() => (zeigen = !zeigen)}>
					<Icon name={zeigen ? 'visibility_off' : 'visibility'} />
				</Button>
			{/snippet}
		</Feld>

		{#if fehler}
			<p class="flex gap-2 rounded-md bg-error-container px-4 py-3 text-body-medium text-on-error-container" role="alert">
				<Icon name="warning" groesse={20} gefuellt />{fehler}
			</p>
		{/if}

		<Button type="submit" size="lg" class="mt-2" disabled={laeuft}>
			{#if laeuft}<Ladekreis />{:else}Anmelden{/if}
		</Button>
	</form>

	<p class="mt-8 text-center text-body-small text-on-surface-variant">
		Nur einmal je Gerät. Das Passwort prüft der Mailserver der Schule; gespeichert wird es nirgends.
	</p>
	<Button variant="ghost" class="mt-2 self-center" onclick={() => gehe('/datenschutz')}><Icon name="privacy_tip" groesse={18} />Datenschutz</Button>
</main>
