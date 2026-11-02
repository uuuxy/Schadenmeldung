// Namen und Bedeutung der Rollen, wie die Oberfläche sie zeigt. Die Rechte selbst prüft der Server
// (anmeldung.go: darfBearbeiten, darfSehen), die Oberfläche blendet nur aus.
export const rollenNamen = { hausmeister: 'Hausmeister', leitung: 'Leitung', verwaltung: 'Verwaltung', admin: 'Admin' };

export const rollenText = {
	hausmeister: 'darf alles und bekommt die Mails',
	leitung: 'darf alles, auch Räume und Personen',
	verwaltung: 'sieht alle Meldungen und druckt QR-Aufkleber, ändert nichts',
	admin: 'darf alles, auch Räume und Personen'
};

export const rollenWahl = Object.entries(rollenNamen).map(([wert, text]) => ({ wert, text }));
