<script>
	import { zustand } from '../../lib/zustand.svelte.js';
	import Icon from '../Icon.svelte';

	/**
	 * Stille Fehler sichtbar machen (mailstoerung.go): Gehen keine Mails hinaus, steht das oben in der
	 * Übersicht, statt dass Meldungen tagelang unbemerkt liegen. Der Hinweis verschwindet von selbst,
	 * sobald wieder eine Mail hinausgeht. Eine feste Uhrzeit liest sich hier besser als „vor 5 Min.“.
	 */
	const uhr = new Intl.DateTimeFormat('de-DE', { hour: '2-digit', minute: '2-digit' });
	const tag = new Intl.DateTimeFormat('de-DE', { weekday: 'short', day: 'numeric', month: 'short' });
	const seit = $derived.by(() => {
		const d = new Date(zustand.mailStoerungSeit);
		return d.toDateString() === new Date().toDateString() ? `${uhr.format(d)} Uhr` : `${tag.format(d)}, ${uhr.format(d)} Uhr`;
	});
	const admin = $derived(zustand.benutzer?.rolle === 'admin');
</script>

{#if zustand.mailStoerungSeit}
	<section role="alert" class="mt-4 flex items-start gap-3 rounded-xl bg-error-container p-4 text-on-error-container">
		<Icon name="warning" class="shrink-0" />
		<p class="text-body-medium">
			Seit {seit} gehen keine Mails hinaus. Neue Meldungen stehen trotzdem hier.
			{admin ? 'Bitte die Zugangsdaten des Absender-Postfachs auf dem Server prüfen.' : 'Bitte dem Admin Bescheid geben.'}
		</p>
	</section>
{/if}
