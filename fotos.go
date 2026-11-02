package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	_ "image/png"
	"io"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const (
	fotoHoechstGroesse = 15 << 20
	fotoHoechstPixel   = 60_000_000
	fotoKante          = 1600
)

var errKeinFoto = errors.New("keine lesbare Bilddatei")

// bereiteFotoAuf liest ein Bild, verkleinert es auf höchstens fotoKante Pixel und schreibt es neu als
// JPEG. Beim Neuschreiben fallen alle Metadaten weg: GPS-Ort, Gerät und Aufnahmezeit bleiben nicht
// erhalten. Die Oberfläche verkleinert schon vor dem Hochladen; das hier ist die Absicherung.
func bereiteFotoAuf(quelle io.Reader) ([]byte, error) {
	roh, err := io.ReadAll(io.LimitReader(quelle, fotoHoechstGroesse+1))
	if err != nil || len(roh) > fotoHoechstGroesse {
		return nil, errKeinFoto
	}
	// Die Maße zuerst: Ein winziges PNG kann sich auf Gigabytes entpacken.
	abmessung, _, err := image.DecodeConfig(bytes.NewReader(roh))
	if err != nil || abmessung.Width*abmessung.Height > fotoHoechstPixel {
		return nil, errKeinFoto
	}
	bild, _, err := image.Decode(bytes.NewReader(roh))
	if err != nil {
		return nil, errKeinFoto
	}
	var aus bytes.Buffer
	if err := jpeg.Encode(&aus, aufWeiss(bild, fotoKante), &jpeg.Options{Quality: 82}); err != nil {
		return nil, err
	}
	return aus.Bytes(), nil
}

// aufWeiss zeichnet das Bild verkleinert auf weißen Grund. JPEG kennt keine Transparenz; ohne den
// Grund würden durchsichtige Flächen eines PNG schwarz.
func aufWeiss(bild image.Image, kante int) image.Image {
	q := bild.Bounds()
	b, h := q.Dx(), q.Dy()
	if b > kante || h > kante {
		if b >= h {
			b, h = kante, max(1, h*kante/b)
		} else {
			b, h = max(1, b*kante/h), kante
		}
	}
	ziel := image.NewRGBA(image.Rect(0, 0, b, h))
	draw.Draw(ziel, ziel.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.CatmullRom.Scale(ziel, ziel.Bounds(), bild, q, draw.Over, nil)
	return ziel
}

func zufallsHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
