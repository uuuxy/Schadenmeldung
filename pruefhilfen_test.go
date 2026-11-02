package main

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"mime/multipart"
	"regexp"
	"testing"
)

func (n *testNutzer) melde(felder map[string]string, foto []byte) (int, map[string]any) {
	n.u.t.Helper()
	var puffer bytes.Buffer
	mw := multipart.NewWriter(&puffer)
	for k, v := range felder {
		_ = mw.WriteField(k, v)
	}
	if foto != nil {
		fw, _ := mw.CreateFormFile("foto", "foto.jpg")
		_, _ = fw.Write(foto)
	}
	_ = mw.Close()
	return n.anfrage("POST", "/api/meldungen", &puffer, mw.FormDataContentType())
}

func eintraege(antwort map[string]any, schluessel string) []map[string]any {
	var aus []map[string]any
	roh, _ := antwort[schluessel].([]any)
	for _, e := range roh {
		aus = append(aus, e.(map[string]any))
	}
	return aus
}

// fotoMitExif baut ein JPEG mit einem EXIF-Block, wie ihn Handykameras samt GPS-Ort schreiben.
func fotoMitExif(t *testing.T, breite, hoehe int) []byte {
	t.Helper()
	bild := image.NewRGBA(image.Rect(0, 0, breite, hoehe))
	for x := range breite {
		bild.Set(x, hoehe/2, color.RGBA{R: 200, A: 255})
	}
	var jpg bytes.Buffer
	if err := jpeg.Encode(&jpg, bild, nil); err != nil {
		t.Fatal(err)
	}
	nutzlast := append([]byte("Exif\x00\x00"), []byte("GPS 12.3456N 65.4321E")...)
	segment := append([]byte{0xFF, 0xE1, byte((len(nutzlast) + 2) >> 8), byte(len(nutzlast) + 2)}, nutzlast...)
	roh := jpg.Bytes()
	return append(append(append([]byte{}, roh[:2]...), segment...), roh[2:]...)
}

func betreffe(mails []testMail) []string {
	var liste []string
	for _, m := range mails {
		liste = append(liste, m.Betreff)
	}
	return liste
}

var linkMuster = regexp.MustCompile(`/aktion/([A-Za-z0-9_-]+)\?tun=erledigt`)
