// Die Arten von Schäden mit typischen Fällen zum Antippen. Die Schlüssel prüft auch der Server
// (server.go, kategorien); eine neue Art braucht beide Stellen und ihre Farbrollen kat-<key> in
// styles/rollen.css. Die Klassen stehen ausgeschrieben, sonst findet Tailwind sie nicht.
export const kategorien = [
	{
		key: 'licht',
		name: 'Licht & Strom',
		icon: 'lightbulb',
		farbe: { symbol: 'text-kat-licht', flaeche: 'bg-kat-licht-container text-on-kat-licht-container' },
		faelle: ['Lampe defekt', 'Licht flackert', 'Steckdose ohne Strom', 'Schalter defekt']
	},
	{
		key: 'heizung',
		name: 'Heizung & Lüftung',
		icon: 'heat',
		farbe: { symbol: 'text-kat-heizung', flaeche: 'bg-kat-heizung-container text-on-kat-heizung-container' },
		faelle: ['Zu kalt', 'Zu warm', 'Heizkörper tropft', 'Lüftung defekt']
	},
	{
		key: 'wasser',
		name: 'Wasser & WC',
		icon: 'water_drop',
		farbe: { symbol: 'text-kat-wasser', flaeche: 'bg-kat-wasser-container text-on-kat-wasser-container' },
		faelle: ['Wasserhahn tropft', 'Abfluss verstopft', 'WC verstopft', 'Kein Wasser']
	},
	{
		key: 'fenster',
		name: 'Fenster & Türen',
		icon: 'door_front',
		farbe: { symbol: 'text-kat-fenster', flaeche: 'bg-kat-fenster-container text-on-kat-fenster-container' },
		faelle: ['Schließt nicht', 'Klemmt', 'Schloss defekt', 'Scheibe kaputt', 'Rollo defekt']
	},
	{
		key: 'moebel',
		name: 'Möbel',
		icon: 'chair',
		farbe: { symbol: 'text-kat-moebel', flaeche: 'bg-kat-moebel-container text-on-kat-moebel-container' },
		faelle: ['Stuhl kaputt', 'Tisch wackelt', 'Schrank / Regal']
	},
	{
		key: 'technik',
		name: 'Medientechnik',
		icon: 'tv',
		farbe: { symbol: 'text-kat-technik', flaeche: 'bg-kat-technik-container text-on-kat-technik-container' },
		faelle: ['Beamer / Display aus', 'Tafel / Whiteboard', 'Kein Ton', 'Kabel / Adapter fehlt', 'Dokumentenkamera']
	},
	{
		key: 'reinigung',
		name: 'Reinigung',
		icon: 'cleaning_services',
		farbe: { symbol: 'text-kat-reinigung', flaeche: 'bg-kat-reinigung-container text-on-kat-reinigung-container' },
		faelle: ['Stark verschmutzt', 'Mülleimer voll', 'Spender leer', 'Graffiti']
	},
	{
		key: 'sonstiges',
		name: 'Sonstiges',
		icon: 'more_horiz',
		// Ohne eigene Bedeutung keine eigene Farbe: Sonstiges bleibt bei der gewöhnlichen Auswahl.
		farbe: { symbol: 'text-secondary', flaeche: 'bg-secondary-container text-on-secondary-container' },
		faelle: []
	}
];

export const kategorieNach = (key) => kategorien.find((k) => k.key === key) ?? kategorien.at(-1);

/** Die Arten, die das Formular zeigt: alle außer den unter Einstellungen → Arten abgeschalteten. */
export const sichtbareArten = (aus = []) => kategorien.filter((k) => !aus.includes(k.key));
