<script>
	import Icon from './Icon.svelte';
	import Switch from './ui/Switch.svelte';
	import { api } from '../lib/api.js';
	import { zustand, vergissAlles, darfSehen, darfBearbeiten } from '../lib/zustand.svelte.js';
	import { rollenNamen } from '../lib/rollen.js';
	import { gehe } from '../lib/weg.svelte.js';
	import { darstellung } from '../lib/darstellung.svelte.js';
	import { zeige } from '../lib/snackbar.svelte.js';

	/**
	 * Kontomenü (M3 Menu). „Dunkles Design“ steht als Schalter darin, wie bei Google (Gemini: „Dark
	 * theme“ im Menü) und in der Bibliothek. Wer das Backend sieht, stellt hier auch seine Mails ein.
	 */
	const rollen = { ...rollenNamen, kollegium: 'Kollegium' };
	let menue;
	let mail = $state(null);
	const initial = $derived((zustand.benutzer?.email ?? '?').charAt(0).toUpperCase());

	async function geoeffnet(e) {
		if (e.newState !== 'open' || !darfSehen()) return;
		try {
			mail = await api('/api/benachrichtigung');
		} catch {
			mail = null;
		}
	}

	async function mailSetzen(aenderung) {
		try {
			mail = await api('/api/benachrichtigung', { methode: 'POST', json: aenderung });
		} catch (e) {
			zeige(e.message);
		}
	}

	function zu(pfad) {
		menue.hidePopover();
		gehe(pfad);
	}

	/** Mit ueberall enden auch die Sitzungen auf anderen Geräten, etwa nach einem verlorenen Handy. */
	async function abmelden(ueberall = false) {
		menue.hidePopover();
		await api(ueberall ? '/api/abmelden?ueberall=1' : '/api/abmelden', { methode: 'POST' }).catch(() => {});
		await vergissAlles();
		zustand.benutzer = null;
		gehe('/', { ersetzen: true });
	}
</script>

{#snippet zeile(id, text, detail, checked, onchange)}
	<div class="flex items-center gap-4 px-4 py-2">
		<label for={id} class="min-w-0 flex-1">
			<span class="block text-body-large text-on-surface">{text}</span>
			{#if detail}<span class="block text-body-small text-on-surface-variant">{detail}</span>{/if}
		</label>
		<Switch {id} {checked} {onchange} />
	</div>
{/snippet}

<button class="icon-btn size-12" popovertarget="konto-menue" aria-label="Konto und Einstellungen">
	<span class="grid size-8 place-items-center rounded-full bg-primary text-title-small text-on-primary">{initial}</span>
</button>

<div id="konto-menue" popover bind:this={menue} ontoggle={geoeffnet} class="menue fixed top-[calc(env(safe-area-inset-top)+3.75rem)] right-2 w-[min(20rem,calc(100vw-1rem))]">
	<div class="px-4 pt-2 pb-3">
		<p class="truncate text-body-large text-on-surface">{zustand.benutzer?.email}</p>
		<p class="text-body-small text-on-surface-variant">{rollen[zustand.benutzer?.rolle] ?? ''}</p>
	</div>
	<hr class="border-outline-variant" />

	<div class="py-1">
		{@render zeile('dunkles-design', 'Dunkles Design', '', darstellung.dunkel, (an) => darstellung.schalte(an))}
	</div>

	{#if darfSehen() && mail?.mailAktiv}
		<hr class="border-outline-variant" />
		<p class="px-4 pt-3 pb-1 text-label-medium text-on-surface-variant">Mail an mich</p>
		<div class="pb-1">
			{@render zeile('mail-neu', 'Bei jeder neuen Meldung', 'sofort, eine Mail je Meldung', mail.mail, (an) => mailSetzen({ mail: an }))}
		</div>
	{/if}

	<hr class="border-outline-variant" />
	<div class="py-1">
		{#if darfBearbeiten()}
			<button class="flex h-12 w-full items-center gap-3 px-4 text-body-large text-on-surface" onclick={() => zu('/einstellungen')}>
				<Icon name="settings" class="text-on-surface-variant" /> Einstellungen
			</button>
		{:else if darfSehen()}
			<!-- Die Verwaltung druckt Aufkleber, ohne Räume zu ändern; die übrigen finden sie unter Einstellungen. -->
			<button class="flex h-12 w-full items-center gap-3 px-4 text-body-large text-on-surface" onclick={() => zu('/raeume/qr')}>
				<Icon name="qr_code_2" class="text-on-surface-variant" /> QR-Aufkleber drucken
			</button>
		{/if}
		<button class="flex h-12 w-full items-center gap-3 px-4 text-body-large text-on-surface" onclick={() => zu('/datenschutz')}>
			<Icon name="privacy_tip" class="text-on-surface-variant" /> Datenschutz
		</button>
		<button class="flex h-12 w-full items-center gap-3 px-4 text-body-large text-on-surface" onclick={() => abmelden()}>
			<Icon name="logout" class="text-on-surface-variant" /> Abmelden
		</button>
		<button class="flex h-12 w-full items-center gap-3 px-4 text-body-large text-on-surface" onclick={() => abmelden(true)}>
			<Icon name="devices" class="text-on-surface-variant" /> Auf allen Geräten abmelden
		</button>
	</div>
</div>
