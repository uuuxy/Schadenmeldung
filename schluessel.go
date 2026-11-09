package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
)

// schluesselLaenge: AES-256 braucht 32 Byte, in der .env als 64 Hex-Zeichen.
const schluesselLaenge = 32

// schluesselAus liest SCHLUESSEL aus der .env. Leer heißt: Die App kann kein Postfach speichern,
// alles andere läuft. ./update.sh legt ihn beim ersten Mal zufällig an.
func schluesselAus(text string) ([]byte, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}
	s, err := hex.DecodeString(text)
	if err != nil || len(s) != schluesselLaenge {
		return nil, errors.New("SCHLUESSEL muss aus 64 Hex-Zeichen bestehen; ./update.sh legt ihn an")
	}
	return s, nil
}

// verschluesseleGeheim schützt ein Geheimnis mit AES-256-GCM; die zufällige Nonce steht vorn im
// Ergebnis. So steht das Passwort des Versand-Postfachs weder in der Datenbank noch in einer
// Sicherung lesbar. (verschluessele in webpush.go ist die Verschlüsselung der Mitteilungen.)
func verschluesseleGeheim(schluessel, klar []byte) ([]byte, error) {
	gcm, err := gcmAus(schluessel)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, klar, nil), nil
}

// entschluesseleGeheim kehrt verschluesseleGeheim um; ein anderer Schlüssel oder veränderte Daten
// fallen auf.
func entschluesseleGeheim(schluessel, geheim []byte) ([]byte, error) {
	gcm, err := gcmAus(schluessel)
	if err != nil {
		return nil, err
	}
	if len(geheim) < gcm.NonceSize() {
		return nil, errors.New("verschlüsselter Wert ist zu kurz")
	}
	nonce, rest := geheim[:gcm.NonceSize()], geheim[gcm.NonceSize():]
	return gcm.Open(nil, nonce, rest, nil)
}

func gcmAus(schluessel []byte) (cipher.AEAD, error) {
	if len(schluessel) != schluesselLaenge {
		return nil, errors.New("SCHLUESSEL fehlt in der .env")
	}
	block, err := aes.NewCipher(schluessel)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
