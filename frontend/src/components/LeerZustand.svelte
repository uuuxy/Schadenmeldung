<script>
	import { fade } from 'svelte/transition';
	import Icon from './Icon.svelte';

	/**
	 * Ein leerer Zustand nach M3 (empty state): großes Symbolfeld, Titel, ein Satz, darunter auf Wunsch
	 * ein Knopf. ton „erfolg“ für „Alles erledigt“, sonst ruhig.
	 */
	let { icon = null, ton = 'ruhig', titel, text = '', children } = $props();
</script>

<div class="flex flex-col items-center px-6 py-16 text-center" in:fade={{ duration: 200 }}>
	{#if icon}
		<span class={['mb-5 grid size-20 place-items-center rounded-2xl', ton === 'erfolg' ? 'bg-success-container text-on-success-container' : 'bg-surface-container-highest text-on-surface-variant']}>
			<Icon name={icon} groesse={40} />
		</span>
	{/if}
	<p class="text-title-large text-on-surface">{titel}</p>
	{#if text}<p class="mt-2 max-w-xs text-body-medium text-on-surface-variant">{text}</p>{/if}
	{#if children}<div class="mt-8">{@render children()}</div>{/if}
</div>
