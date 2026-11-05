<script>
	import { aufklappen, zuklappen } from '../../lib/bewegung.js';
	import { zustand } from '../../lib/zustand.svelte.js';
	import { kategorieNach } from '../../lib/kategorien.js';
	import { wann } from '../../lib/zeit.js';
	import Icon from '../Icon.svelte';
	import StatusChip from '../ui/StatusChip.svelte';

	/**
	 * Was im gewählten Raum schon gemeldet ist und wie weit es ist: So meldet niemand doppelt. Das muss ins
	 * Auge springen, auch beim QR-Code, wo die nächste Frage gleich darunter steht: oben ein farbiger Kopf
	 * wie der Hinweis zum QR-Code (warning) mit dem, was zu tun ist, darunter die Meldungen auf ruhiger
	 * Fläche, damit ihre Status-Chips lesbar bleiben. Die Karte federt herein (global: auch dann, wenn
	 * sie mit dem Raum zusammen erscheint), denn Bewegung zieht den Blick.
	 */
	let { raumId } = $props();
	const gemeldet = $derived(zustand.offen.filter((m) => m.raumId === raumId));
	const wartend = $derived(zustand.ausgang.filter((a) => a.raumId === raumId));
</script>

{#if gemeldet.length || wartend.length}
	<section class="mt-3 overflow-hidden rounded-xl bg-surface-container-low" aria-labelledby="hier-gemeldet" in:aufklappen|global={{ schnell: true }} out:zuklappen>
		<div class="flex gap-3 bg-warning-container px-4 py-3 text-on-warning-container">
			<Icon name="info" groesse={24} class="mt-0.5" />
			<div>
				<h3 id="hier-gemeldet" class="text-title-medium">Hier schon gemeldet</h3>
				<p class="text-body-medium">Ist es dasselbe? Dann musst du es nicht noch einmal melden.</p>
			</div>
		</div>
		<ul class="py-1">
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
					<StatusChip status={m.status} knapp />
				</li>
			{/each}
		</ul>
	</section>
{/if}
