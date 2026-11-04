<script>
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { wenigBewegung } from '../../lib/weg.svelte.js';
	import Icon from '../Icon.svelte';
	import Button from '../ui/Button.svelte';
	import { installieren, alsApp, apfelGeraet, firefoxAndroid, hinweisWeg, hinweisAusblenden } from '../../lib/installieren.svelte.js';

	/**
	 * Gleich beim ersten Öffnen das Angebot, die Schadenmeldung als App auf den Startbildschirm zu legen
	 * (Anmeldung und Meldeseite): Auf dem iPhone ist der Anfang der beste Moment, wer zuerst installiert,
	 * meldet sich gleich in der App an. Chrome bietet das selbst an (ein Knopf genügt); auf dem iPhone
	 * geht es nur über „Teilen“, in Firefox auf Android über dessen Menü. Das Symbol der Karte zeigt,
	 * wonach man suchen muss. Einmal weggetippt, kommt das Angebot auf dem Gerät nicht wieder.
	 */
	let { class: klasse = '' } = $props();
	let weg = $state(alsApp() || hinweisWeg());
	// Erst nach dem Aufbau einblenden: Beim ersten Zeichnen spielt Svelte keine Animation ab.
	let bereit = $state(false);
	onMount(() => (bereit = true));
	const art = $derived(installieren.angebot ? 'knopf' : apfelGeraet() ? 'iphone' : firefoxAndroid() ? 'firefox' : null);
	const symbol = { knopf: 'mobile_arrow_down', iphone: 'ios_share', firefox: 'more_vert' };

	function ausblenden() {
		hinweisAusblenden();
		weg = true;
	}

	async function installiere() {
		installieren.angebot.prompt();
		await installieren.angebot.userChoice;
		installieren.angebot = null;
		ausblenden();
	}
</script>

{#if bereit && !weg && art}
	<section transition:fly={{ y: -16, duration: wenigBewegung.matches ? 0 : 250 }} class={['rounded-xl bg-primary-container p-4 text-left text-on-primary-container', klasse]}>
		<div class="flex items-start gap-3">
			<Icon name={symbol[art]} class="shrink-0" />
			<p class="text-body-medium">
				{#if art === 'iphone'}
					Als App auf den Home-Bildschirm: in der Leiste des Browsers auf das Teilen-Symbol tippen (oder zuerst auf
					die drei Punkte), dann auf „Zum Home-Bildschirm“ und die Schadenmeldung über das neue Symbol öffnen.
				{:else if art === 'firefox'}
					Als App installieren: im Menü von Firefox (die drei Punkte) auf „Installieren“ oder „Zum Startbildschirm
					hinzufügen“ tippen. Dann ist die Schadenmeldung mit einem Tipp vom Startbildschirm aus da.
				{:else}
					Als App installieren: Dann ist die Schadenmeldung mit einem Tipp vom Startbildschirm aus da.
				{/if}
			</p>
		</div>
		<div class="mt-3 flex justify-end gap-2">
			<Button variant="ghost" onclick={ausblenden}>{art === 'knopf' ? 'Nicht jetzt' : 'Verstanden'}</Button>
			{#if art === 'knopf'}<Button variant="tonal" onclick={installiere}>Installieren</Button>{/if}
		</div>
	</section>
{/if}
