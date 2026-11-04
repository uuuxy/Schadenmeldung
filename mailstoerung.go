package main

import "time"

// Stille Fehler sichtbar machen: Ändert sich etwa das Passwort des Absender-Postfachs, gehen keine Mails
// mehr hinaus, und niemand merkt es. Der Server merkt sich, seit wann der Versand scheitert; die
// Übersicht zeigt es allen im Backend (MailStoerung.svelte). Das steht nur im Speicher: Nach einem
// Neustart zeigt der nächste Versuch wieder, wie es steht.

// mailVersuch hält fest, ob ein Versand geklappt hat. Ändert sich der Stand, laden offene Übersichten neu.
func (a *app) mailVersuch(err error) {
	if err != nil {
		if a.mailStoerung.CompareAndSwap(0, a.jetzt().Unix()) {
			a.verteiler.melde()
		}
		return
	}
	if a.mailStoerung.Swap(0) != 0 {
		a.verteiler.melde()
	}
}

// mailStoerungSeit: seit wann der Versand scheitert; nil, solange alles läuft.
func (a *app) mailStoerungSeit() *time.Time {
	s := a.mailStoerung.Load()
	if s == 0 {
		return nil
	}
	t := time.Unix(s, 0)
	return &t
}
