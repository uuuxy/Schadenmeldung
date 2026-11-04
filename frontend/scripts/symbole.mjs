// Schreibt die Pfade der benutzten Material Symbols (Rounded) nach src/lib/symbole.js. So landen nur
// diese Symbole im Bündel, und die Oberfläche lädt keine Symbolschrift. Neues Symbol: Namen in die
// Liste, dann `npm run symbole`.
import { existsSync, readFileSync, writeFileSync } from 'node:fs';

const namen = [
	// Arten des Schadens
	'lightbulb', 'heat', 'water_drop', 'door_front', 'chair', 'tv', 'cleaning_services', 'more_horiz',
	// Bedienung
	'meeting_room', 'location_on', 'search', 'photo_camera', 'close', 'check', 'check_circle', 'send', 'warning',
	'error', 'edit_note', 'edit', 'schedule', 'local_shipping', 'engineering', 'undo', 'delete', 'print', 'qr_code_2',
	'upload_file', 'download', 'picture_as_pdf', 'table_view', 'logout', 'devices', 'handyman', 'ios_share', 'share', 'mobile_arrow_down', 'more_vert', 'privacy_tip', 'arrow_back', 'task_alt',
	'visibility', 'visibility_off', 'build', 'add_comment', 'history', 'cloud_off', 'schedule_send', 'arrow_downward',
	'arrow_drop_down', 'chevron_right', 'add', 'settings', 'category', 'admin_panel_settings', 'open_in_new'
];

// Gefüllt zeigt M3 das gewählte Ziel und die gewählte Art; nur dafür kommt die zweite Form mit.
const gefuellt = new Set([
	'lightbulb', 'heat', 'water_drop', 'door_front', 'chair', 'tv', 'cleaning_services', 'more_horiz',
	'meeting_room', 'location_on', 'warning', 'check_circle', 'build'
]);

const ordner = new URL('../node_modules/@material-symbols/svg-400/rounded/', import.meta.url);
const pfade = {};
for (const name of namen) {
	for (const variante of gefuellt.has(name) ? [name, `${name}-fill`] : [name]) {
		const datei = new URL(`${variante}.svg`, ordner);
		if (!existsSync(datei)) {
			if (variante === name) throw new Error(`Symbol fehlt: ${name}`);
			continue;
		}
		const treffer = [...readFileSync(datei, 'utf8').matchAll(/ d="([^"]+)"/g)].map((t) => t[1]);
		pfade[variante] = treffer.join(' ');
	}
}

const kopf = '// Erzeugt von scripts/symbole.mjs aus @material-symbols/svg-400 (Apache-2.0). Nicht von Hand ändern.\n';
writeFileSync(new URL('../src/lib/symbole.js', import.meta.url), `${kopf}export const symbole = ${JSON.stringify(pfade, null, '\t')};\n`);
console.log(`${Object.keys(pfade).length} Pfade geschrieben.`);
