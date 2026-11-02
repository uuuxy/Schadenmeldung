<script>
	import Icon from '../Icon.svelte';
	import { stati } from '../../lib/status.js';

	/**
	 * @component StatusChip — ein Zustand, den man nur liest, als Chip nach M3 (wie in der Bibliothek):
	 * 24 px hoch, Ecke 8 px, getönte Fläche statt Rahmen, Farben aus den container/on-container-Paaren.
	 * `knapp` nimmt den kurzen Namen („Bestellt“ statt „Material bestellt“), wo eine Zeile eng ist.
	 */
	let { status, knapp = false } = $props();
	const s = $derived(stati[status] ?? stati.offen);
	const toene = {
		neutral: 'bg-surface-container text-on-surface-variant',
		primaer: 'bg-primary-container text-on-primary-container',
		warten: 'bg-warning-container text-on-warning-container',
		erfolg: 'bg-success-container text-on-success-container'
	};
</script>

<span class={['inline-flex h-6 shrink-0 items-center gap-1 rounded-sm px-2 text-label-medium whitespace-nowrap', toene[s.ton]]}>
	<Icon name={s.icon} groesse={16} gefuellt={status === 'erledigt'} />{knapp ? s.knapp : s.name}
</span>
