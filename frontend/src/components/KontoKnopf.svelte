<script>
	import Icon from './Icon.svelte';
	import Switch from './ui/Switch.svelte';
	import { api } from '../lib/api.js';
	import { zustand, vergissAlles, darfSehen, darfBearbeiten } from '../lib/zustand.svelte.js';
	import { rollenNamen } from '../lib/rollen.js';
	import { gehe } from '../lib/weg.svelte.js';
	import { darstellung } from '../lib/darstellung.svelte.js';
	import { zeige } from '../lib/snackbar.svelte.js';
	import { frage } from '../lib/rueckfrage.svelte.js';
	import { auffrischen } from '../lib/nachreichen.js';
	import { mitteilungenStand, mitteilungenEin, mitteilungenAus } from '../lib/mitteilungen.js';

	/**
	 * Kontomenü (M3 Menu). „Dunkles Design“ steht als Schalter darin, wie bei Google (Gemini: „Dark
	 * theme“ im Menü) und in der Bibliothek. Wer das Backend sieht, stellt hier auch Mails und
	 * Mitteilungen ein; Mitteilungen gelten je Gerät, darum „Auf diesem Gerät“.
	 */
	const rollen = { ...rollenNamen, kollegium: 'Kollegium' };
	let menue;
	let mail = $state(null);
	let mitteilung = $state('');
	let schaltet = $state(false);
	const mitteilungDetail = $derived(
		{ gesperrt: 'in den Einstellungen des Geräts gesperrt', homescreen: 'erst die App auf den Home-Bildschirm legen' }[mitteilung] ??
			'bei jeder neuen Meldung'
	);
	const initial = $derived((zustand.benutzer?.email ?? '?').charAt(0).toUpperCase());

	async function geoeffnet(e) {
		if (e.newState !== 'open' || !darfSehen()) return;
		try {
			mail = await api('/api/benachrichtigung');
		} catch {
			mail = null;
		}
		mitteilung = await mitteilungenStand().catch(() => '');
	}

	// Beide Schalter zeigen die Wahl sofort; klappt es nicht, springen sie zurück. Ohne diesen Schritt
	// bliebe der Schalter falsch stehen, denn er legt sich selbst um.
	async function mitteilungSetzen(an) {
		mitteilung = an ? 'an' : 'aus';
		schaltet = true;
		try {
			if (!an) await mitteilungenAus();
			else if (!(await mitteilungenEin())) zeige('Mitteilungen sind auf diesem Gerät gesperrt. Freigeben lässt sich das in den Einstellungen des Geräts.');
		} catch (e) {
			zeige(e.status !== undefined ? e.message : 'Mitteilungen ließen sich auf diesem Gerät nicht einschalten.');
		}
		mitteilung = await mitteilungenStand().catch(() => '');
		schaltet = false;
	}

	async function mailSetzen(aenderung) {
		const vorher = mail;
		mail = { ...mail, ...aenderung };
		try {
			mail = await api('/api/benachrichtigung', { methode: 'POST', json: aenderung });
		} catch (e) {
			mail = vorher;
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
		// Beim Abmelden bleibt nichts von dieser Person auf dem Gerät, auch nicht der Ausgang. Darum erst
		// losschicken, was wartet; bleibt etwas liegen (kein Netz), fragt die App, statt still zu löschen.
		if (zustand.ausgang.length || zustand.wartend.length) await auffrischen();
		const [m, a] = [zustand.ausgang.length, zustand.wartend.length];
		if (m + a) {
			const was = [m && `${m} ${m === 1 ? 'Meldung' : 'Meldungen'}`, a && `${a} ${a === 1 ? 'Änderung' : 'Änderungen'}`].filter(Boolean).join(' und ');
			const eins = m + a === 1;
			// Erst das Menü fertig ausblenden lassen (200 ms), sonst liegt es noch über dem Dialog. War die
			// Sitzung am Server schon beendet, ist das Menü beim Senden samt Kopfleiste verschwunden.
			await Promise.all((menue?.getAnimations() ?? []).map((x) => x.finished.catch(() => {})));
			const ja = await frage({
				titel: 'Trotzdem abmelden?',
				text: `${was} ${eins ? 'ist' : 'sind'} noch nicht beim Server angekommen. Abmelden löscht sie von diesem Gerät; bleibst du angemeldet, ${eins ? 'geht' : 'gehen'} sie von selbst raus, sobald wieder Netz da ist.`,
				ja: 'Abmelden'
			});
			if (!ja) return;
		}
		// Ein abgemeldetes Gerät bekommt keine Mitteilungen mehr.
		await mitteilungenAus().catch(() => {});
		await api(ueberall ? '/api/abmelden?ueberall=1' : '/api/abmelden', { methode: 'POST' }).catch(() => {});
		await vergissAlles();
		zustand.benutzer = null;
		gehe('/', { ersetzen: true });
	}
</script>

<!-- Ein Menüpunkt (M3 Menu item): 48 px hoch, Symbol vorn. -->
{#snippet punkt(icon, text, tun)}
	<button class="flex h-12 w-full items-center gap-3 px-4 text-body-large text-on-surface" onclick={tun}>
		<Icon name={icon} class="text-on-surface-variant" /> {text}
	</button>
{/snippet}

{#snippet zeile(id, text, detail, checked, onchange, gesperrt = false)}
	<div class="flex items-center gap-4 px-4 py-2">
		<label for={id} class="min-w-0 flex-1">
			<span class="block text-body-large text-on-surface">{text}</span>
			{#if detail}<span class="block text-body-small text-on-surface-variant">{detail}</span>{/if}
		</label>
		<Switch {id} {checked} {onchange} disabled={gesperrt} />
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

	{#if darfSehen() && mitteilung}
		<hr class="border-outline-variant" />
		<p class="px-4 pt-3 pb-1 text-label-medium text-on-surface-variant">Mitteilung aufs Handy</p>
		<div class="pb-1">
			{@render zeile('mitteilung-neu', 'Auf diesem Gerät', mitteilungDetail, mitteilung === 'an', mitteilungSetzen, schaltet || mitteilung === 'gesperrt' || mitteilung === 'homescreen')}
		</div>
	{/if}

	<hr class="border-outline-variant" />
	<div class="py-1">
		{#if darfBearbeiten()}
			{@render punkt('settings', 'Einstellungen', () => zu('/einstellungen'))}
		{:else if darfSehen()}
			<!-- Die Verwaltung druckt Aufkleber, ohne Räume zu ändern; die übrigen finden sie unter Einstellungen. -->
			{@render punkt('qr_code_2', 'QR-Aufkleber drucken', () => zu('/raeume/qr'))}
		{/if}
		{@render punkt('privacy_tip', 'Datenschutz', () => zu('/datenschutz'))}
		{@render punkt('logout', 'Abmelden', () => abmelden())}
		{@render punkt('devices', 'Auf allen Geräten abmelden', () => abmelden(true))}
	</div>
</div>
