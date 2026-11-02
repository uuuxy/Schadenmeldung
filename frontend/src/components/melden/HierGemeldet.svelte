<script>
	import { slide } from 'svelte/transition';
	import { zustand } from '../../lib/zustand.svelte.js';
	import { kategorieNach } from '../../lib/kategorien.js';
	import { wann } from '../../lib/zeit.js';
	import Icon from '../Icon.svelte';
	import StatusChip from '../ui/StatusChip.svelte';

	/** Was im gewählten Raum schon gemeldet ist und wie weit es ist: So meldet niemand doppelt. */
	let { raumId } = $props();
	const gemeldet = $derived(zustand.offen.filter((m) => m.raumId === raumId));
	const wartend = $derived(zustand.ausgang.filter((a) => a.raumId === raumId));
</script>

{#if gemeldet.length || wartend.length}
	<div class="mt-3 rounded-xl bg-surface-container-low py-1" transition:slide={{ duration: 250 }}>
		<p class="px-4 pt-2 pb-1 text-label-large text-on-surface-variant">Hier schon gemeldet</p>
		<ul>
			{#each wartend as a (a.id)}
				<li class="flex items-center gap-3 px-4 py-2.5">
					<Icon name={kategorieNach(a.kategorie).icon} groesse={20} class={kategorieNach(a.kategorie).farbe.symbol} />
					<span class="min-w-0 flex-1 truncate text-body-large text-on-surface">{a.was}</span>
					<span class="inline-flex h-6 items-center gap-1 rounded-sm bg-surface-container px-2 text-label-medium text-on-surface-variant"><Icon name="schedule_send" groesse={16} />Wartet auf Netz</span>
				</li>
			{/each}
			{#each gemeldet as m (m.id)}
				<li class="flex items-center gap-3 px-4 py-2.5">
					<Icon name={kategorieNach(m.kategorie).icon} groesse={20} class={kategorieNach(m.kategorie).farbe.symbol} />
					<span class="min-w-0 flex-1">
						<span class="block truncate text-body-large text-on-surface">{m.was}</span>
						<span class="block text-body-small text-on-surface-variant">{wann(m.gemeldetAm)}</span>
					</span>
					<StatusChip status={m.status} />
				</li>
			{/each}
		</ul>
	</div>
{/if}
