<script>
	import { onMount } from 'svelte';
	import { fly } from 'svelte/transition';
	import { wenigBewegung } from '../../lib/weg.svelte.js';
	import Icon from '../Icon.svelte';
	import Button from '../ui/Button.svelte';
	import { installieren, alsApp, anleitung, hinweisWeg, hinweisAusblenden } from '../../lib/installieren.svelte.js';

	/**
	 * Gleich beim ersten Öffnen das Angebot, die Schadensmeldung als App auf den Startbildschirm zu legen
	 * (Anmeldung und Meldeseite): Auf dem iPhone ist der Anfang der beste Moment, wer zuerst installiert,
	 * meldet sich gleich in der App an. Bietet der Browser es selbst an (Chrome), genügt ein Knopf. Sonst
	 * führen nummerierte Schritte durch sein Menü, damit es auch gelingt, wer so etwas nie gemacht hat:
	 * ein Tipp je Schritt, die Knöpfe fett und so benannt, wie sie dort heißen, das gesuchte Symbol mitten
	 * im Satz. Einmal weggetippt, kommt das Angebot auf dem Gerät nicht wieder.
	 */
	let { class: klasse = '' } = $props();
	let weg = $state(alsApp() || hinweisWeg());
	// Erst nach dem Aufbau einblenden: Beim ersten Zeichnen spielt Svelte keine Animation ab.
	let bereit = $state(false);
	onMount(() => (bereit = true));
	const vonHand = anleitung();
	const karte = $derived(installieren.angebot ? { titel: 'Als App installieren', danach: 'Dann ist die Schadensmeldung mit einem Tipp da.' } : vonHand);

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

<!-- Eine Zeile der Anleitung: {icon} wird zum Symbol, „Knopfnamen“ werden fett. -->
{#snippet zeile(s)}
	{#each s.text.split('{icon}') as teil, i}{#if i}<Icon name={s.icon} groesse={18} class="inline align-[-4px]" />{#if s.name}<span class="sr-only">{s.name}</span>{/if}{/if}{#each teil.split(/(„[^“]*“)/) as stueck, j}{#if j % 2}<strong class="font-medium">{stueck}</strong>{:else}{stueck}{/if}{/each}{/each}
{/snippet}

{#if bereit && !weg && karte}
	<section transition:fly={{ y: -16, duration: wenigBewegung.matches ? 0 : 250 }} class={['rounded-xl bg-primary-container p-4 text-left text-on-primary-container', klasse]}>
		<div class="flex items-start gap-3">
			<Icon name="mobile_arrow_down" class="shrink-0" />
			<div class="min-w-0 flex-1">
				<p class="text-title-small">{karte.titel}</p>
				{#if karte.schritte}
					<ol class="mt-2 flex list-decimal flex-col gap-1 pl-5 text-body-medium marker:font-medium">
						{#each karte.schritte as s (s.text)}<li>{@render zeile(s)}</li>{/each}
					</ol>
					{#if karte.hinweis}<p class="mt-2 text-body-small">{@render zeile(karte.hinweis)}</p>{/if}
				{/if}
				<p class="mt-2 text-body-medium">{karte.danach}</p>
			</div>
		</div>
		<div class="mt-3 flex justify-end gap-2">
			<Button variant="ghost" onclick={ausblenden}>{installieren.angebot ? 'Nicht jetzt' : 'Verstanden'}</Button>
			{#if installieren.angebot}<Button variant="tonal" onclick={installiere}>Installieren</Button>{/if}
		</div>
	</section>
{/if}
