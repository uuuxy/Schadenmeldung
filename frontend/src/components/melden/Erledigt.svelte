<script>
	import { untrack } from 'svelte';
	import { zustand } from '../../lib/zustand.svelte.js';
	import Icon from '../Icon.svelte';
	import Button from '../ui/Button.svelte';

	/**
	 * Rückmeldung an die Meldenden (rueckmeldung.go): Was seit dem letzten Öffnen erledigt wurde, steht
	 * einmal oben im Formular. Wer sieht, dass Melden etwas bewirkt, meldet wieder. Gemerkt wird nur auf
	 * diesem Gerät, bis wann schon Bescheid gegeben wurde; beim ersten Mal zählen die letzten sieben Tage.
	 */
	const SCHLUESSEL = 'schaden.erledigtGesehen';
	let zeigen = $state([]);

	function gesehenBis() {
		try {
			return localStorage.getItem(SCHLUESSEL);
		} catch {
			return null;
		}
	}

	function merke(zeit) {
		try {
			localStorage.setItem(SCHLUESSEL, zeit);
		} catch {
			/* Speicher gesperrt: Dann kommt die Rückmeldung eben noch einmal. */
		}
	}

	$effect(() => {
		const liste = zustand.meineErledigt;
		if (!liste.length) return;
		const bis = new Date(gesehenBis() ?? Date.now() - 7 * 86_400_000);
		const schon = untrack(() => zeigen);
		const neu = liste.filter((m) => new Date(m.erledigtAm) > bis && !schon.some((z) => z.id === m.id));
		if (!neu.length) return;
		zeigen = [...schon, ...neu];
		merke(liste[0].erledigtAm);
	});
</script>

{#if zeigen.length}
	<section class="rounded-xl bg-surface-container-high p-4 text-on-surface">
		<div class="flex items-start gap-3">
			<Icon name="task_alt" class="shrink-0 text-primary" />
			<div class="min-w-0">
				<p class="text-title-small">
					{zeigen.length === 1 ? 'Erledigt, danke fürs Melden!' : `${zeigen.length} Meldungen erledigt, danke fürs Melden!`}
				</p>
				<ul class="mt-1 text-body-medium text-on-surface-variant">
					{#each zeigen as m (m.id)}<li>{m.was} · {m.ort}</li>{/each}
				</ul>
			</div>
		</div>
		<div class="mt-3 flex justify-end">
			<Button variant="ghost" onclick={() => (zeigen = [])}>OK</Button>
		</div>
	</section>
{/if}
