// Die Arten von Schäden mit typischen Fällen zum Antippen. Die Schlüssel prüft auch der Server
// (server.go, kategorien); eine neue Art braucht beide Stellen.
export const kategorien = [
	{
		key: 'licht',
		name: 'Licht & Strom',
		icon: 'lightbulb',
		faelle: ['Lampe defekt', 'Licht flackert', 'Steckdose ohne Strom', 'Schalter defekt']
	},
	{
		key: 'heizung',
		name: 'Heizung & Lüftung',
		icon: 'heat',
		faelle: ['Zu kalt', 'Zu warm', 'Heizkörper tropft', 'Lüftung defekt']
	},
	{
		key: 'wasser',
		name: 'Wasser & WC',
		icon: 'water_drop',
		faelle: ['Wasserhahn tropft', 'Abfluss verstopft', 'WC verstopft', 'Kein Wasser']
	},
	{
		key: 'fenster',
		name: 'Fenster & Türen',
		icon: 'door_front',
		faelle: ['Schließt nicht', 'Klemmt', 'Schloss defekt', 'Scheibe kaputt', 'Rollo defekt']
	},
	{
		key: 'moebel',
		name: 'Möbel & Tafel',
		icon: 'chair',
		faelle: ['Stuhl kaputt', 'Tisch wackelt', 'Tafel / Whiteboard', 'Schrank / Regal']
	},
	{
		key: 'technik',
		name: 'Medientechnik',
		icon: 'tv',
		faelle: ['Beamer / Display aus', 'Kein Ton', 'Kabel / Adapter fehlt', 'Dokumentenkamera']
	},
	{
		key: 'reinigung',
		name: 'Reinigung',
		icon: 'cleaning_services',
		faelle: ['Stark verschmutzt', 'Mülleimer voll', 'Spender leer', 'Graffiti']
	},
	{ key: 'sonstiges', name: 'Sonstiges', icon: 'more_horiz', faelle: [] }
];

export const kategorieNach = (key) => kategorien.find((k) => k.key === key) ?? kategorien.at(-1);
