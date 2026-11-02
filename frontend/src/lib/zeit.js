const uhrzeit = new Intl.DateTimeFormat('de-DE', { hour: '2-digit', minute: '2-digit' });
const tag = new Intl.DateTimeFormat('de-DE', { weekday: 'short', day: 'numeric', month: 'numeric' });

const gleicherTag = (a, b) => a.toDateString() === b.toDateString();

/** „gerade eben“, „vor 12 Min.“, „heute, 09:41“, „gestern, 14:20“ oder „Mo., 28.10., 08:15“. */
export function wann(iso, jetzt = new Date()) {
	const d = new Date(iso);
	const sekunden = (jetzt - d) / 1000;
	if (sekunden < 60) return 'gerade eben';
	if (sekunden < 3600) return `vor ${Math.floor(sekunden / 60)} Min.`;
	if (gleicherTag(d, jetzt)) return `heute, ${uhrzeit.format(d)}`;
	const gestern = new Date(jetzt);
	gestern.setDate(jetzt.getDate() - 1);
	if (gleicherTag(d, gestern)) return `gestern, ${uhrzeit.format(d)}`;
	return `${tag.format(d)}, ${uhrzeit.format(d)}`;
}

/** Aus „m.mustermann@schule.de“ wird „m.mustermann“ – für die Rückfrage reicht der Name. */
export const kurzname = (email) => (email ?? '').split('@')[0];
