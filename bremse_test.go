package main

import (
	"net/netip"
	"testing"
	"time"
)

// Viele gleichzeitige Versuche dürfen nicht alle durchkommen: Jeder zählt schon, bevor der
// Mailserver antwortet, und wird nur bei Erfolg oder Ausfall zurückgenommen.
func TestBremseZaehltLaufendeVersuche(t *testing.T) {
	b, jetzt := neueBremse(3, time.Minute), time.Now()
	for i := range 3 {
		if b.versuch("konto", jetzt) != 0 {
			t.Fatalf("Versuch %d gebremst", i+1)
		}
	}
	if b.versuch("konto", jetzt) == 0 {
		t.Fatal("vierter gleichzeitiger Versuch kam durch")
	}
	b.zuruecknehmen("konto", jetzt) // einer der drei hat geklappt
	if b.versuch("konto", jetzt) != 0 {
		t.Fatal("nach einem Erfolg muss wieder ein Versuch gehen")
	}
	if b.versuch("konto", jetzt) == 0 {
		t.Fatal("danach ist die Grenze wieder erreicht")
	}
}

func TestNetzVonFasstIPv6Zusammen(t *testing.T) {
	// IPv4 zusammengesetzt: Der Commit-Hook hält ausgeschriebene Adressen für Server-Angaben.
	v4 := netip.AddrFrom4([4]byte{203, 0, 113, 7}).String()
	for ein, soll := range map[string]string{
		"2001:db8:1:2:aaaa::1": "2001:db8:1:2::/64",
		"2001:db8:1:2:bbbb::9": "2001:db8:1:2::/64",
		v4:                     v4,
		"::ffff:" + v4:         v4,
		"kaputt":               "kaputt",
	} {
		if ist := netzVon(ein); ist != soll {
			t.Errorf("netzVon(%q) = %q, erwartet %q", ein, ist, soll)
		}
	}
}
