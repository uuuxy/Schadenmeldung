<script>
	import Icon from '../Icon.svelte';
	import Button from '../ui/Button.svelte';
	import { installieren, alsApp, apfelGeraet, hinweisWeg, hinweisAusblenden } from '../../lib/installieren.svelte.js';

	/**
	 * Nach dem Melden der Tipp, die Schadenmeldung als App auf den Startbildschirm zu legen: Dann ist
	 * sie beim nächsten Schaden mit einem Tipp da. Chrome bietet das selbst an (ein Knopf genügt), auf
	 * dem iPhone geht es nur über „Teilen“. Einmal weggetippt, kommt der Tipp auf dem Gerät nicht wieder.
	 */
	let weg = $state(alsApp() || hinweisWeg());
	const art = $derived(installieren.angebot ? 'knopf' : apfelGeraet() ? 'iphone' : null);

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

{#if !weg && art}
	<section class="rounded-xl bg-surface-container-low p-4 text-left">
		<div class="flex items-start gap-3">
			<Icon name={art === 'iphone' ? 'ios_share' : 'mobile_arrow_down'} class="shrink-0 text-primary" />
			<p class="text-body-medium text-on-surface">
				{#if art === 'iphone'}
					Als App auf den Home-Bildschirm: auf „Teilen“ tippen (das Symbol links), dann auf „Zum Home-Bildschirm“.
					Beim nächsten Schaden genügt ein Tipp.
				{:else}
					Als App installieren: Beim nächsten Schaden ist die Schadenmeldung mit einem Tipp da.
				{/if}
			</p>
		</div>
		<div class="mt-3 flex justify-end gap-2">
			<Button variant="ghost" onclick={ausblenden}>{art === 'iphone' ? 'Verstanden' : 'Nicht jetzt'}</Button>
			{#if art === 'knopf'}<Button variant="tonal" onclick={installiere}>Installieren</Button>{/if}
		</div>
	</section>
{/if}
