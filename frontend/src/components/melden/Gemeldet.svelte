<script>
	import { fade } from 'svelte/transition';
	import { zustand, darfSehen } from '../../lib/zustand.svelte.js';
	import { gehe } from '../../lib/weg.svelte.js';
	import Icon from '../Icon.svelte';
	import Button from '../ui/Button.svelte';

	/**
	 * Bestätigung statt Formular: Die Meldung ist sicher angekommen oder sicher im Ausgang. Der Text
	 * verspricht nur, was stimmt: eine Mail nur, wenn der Server Mails verschickt.
	 */
	let { fertig, onweiter } = $props();
	const gemeldet = $derived(fertig.art === 'gemeldet');
	const text = $derived.by(() => {
		if (!gemeldet) return 'Gerade ist kein Netz. Die Meldung geht automatisch raus, sobald wieder Verbindung besteht.';
		if (zustand.einstellungen.mailAktiv) return 'Sie ist beim Hausmeister angekommen, auch per Mail.';
		return 'Sie ist beim Hausmeister angekommen.';
	});
</script>

<main class="mx-auto flex max-w-xl flex-col items-center px-6 pt-16 pb-12 text-center" in:fade={{ duration: 200 }}>
	<span class={['mb-6 grid size-24 place-items-center rounded-2xl', gemeldet ? 'bg-success-container text-on-success-container' : 'bg-surface-container-highest text-on-surface-variant']}>
		<Icon name={gemeldet ? 'task_alt' : 'schedule_send'} groesse={48} />
	</span>
	<h2 class="text-headline-small text-on-surface">{gemeldet ? 'Gemeldet – danke!' : 'Gespeichert'}</h2>
	<p class="mt-2 text-body-large text-on-surface-variant">{fertig.was} · {fertig.ort}</p>
	<p class="mt-4 max-w-sm text-balance text-body-medium text-on-surface-variant">{text}</p>
	<div class="mt-10 flex flex-col gap-3 self-stretch">
		<Button size="lg" onclick={onweiter}><Icon name="add" />Noch etwas melden</Button>
		{#if darfSehen()}
			<Button variant="ghost" onclick={() => gehe('/')}>Zur Übersicht</Button>
		{/if}
	</div>
</main>
