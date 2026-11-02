<script>
	import { fly } from 'svelte/transition';
	import { snackbar, schliesse } from '../lib/snackbar.svelte.js';
	import Icon from './Icon.svelte';
</script>

{#if snackbar.text}
	{#key snackbar.nummer}
		<div
			role="status"
			class="snackbar fixed inset-x-3 z-50 mx-auto flex max-w-md items-center gap-1 rounded-xs bg-inverse-surface py-1 pr-1 pl-4 text-body-medium text-inverse-on-surface shadow-3"
			transition:fly={{ y: 24, duration: 250 }}
		>
			<span class="flex-1 py-2.5">{snackbar.text}</span>
			{#if snackbar.aktion}
				<button
					class="h-10 shrink-0 rounded-full px-3 text-label-large text-inverse-primary"
					onclick={() => {
						snackbar.aktion.tun();
						schliesse();
					}}
				>
					{snackbar.aktion.name}
				</button>
				<!-- Mit Aktion bleibt die Snackbar stehen; das Kreuz schließt sie (M3: close affordance). -->
				<button class="icon-btn text-inverse-on-surface" aria-label="Schließen" onclick={schliesse}><Icon name="close" groesse={20} /></button>
			{/if}
		</div>
	{/key}
{/if}

<style>
	/* Unten in Daumenreichweite; eine Ansicht mit eigener Leiste am Boden (Melden, Detail) schiebt
	   sie mit --snackbar-abstand darüber, damit nichts verdeckt wird. */
	.snackbar {
		bottom: calc(var(--snackbar-abstand, 0px) + env(safe-area-inset-bottom) + 1rem);
	}
</style>
