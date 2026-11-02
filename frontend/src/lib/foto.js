/**
 * Verkleinert ein Handyfoto vor dem Hochladen auf höchstens 1600 px und rechnet es in JPEG um. Das
 * spart im schwachen WLAN Sekunden, und beim Umrechnen fallen GPS-Ort und Gerätedaten weg. Der
 * Server rechnet zur Sicherheit noch einmal um.
 */
export async function verkleinereFoto(datei, kante = 1600) {
	const url = URL.createObjectURL(datei);
	try {
		const bild = new Image();
		bild.decoding = 'async';
		bild.src = url;
		await bild.decode();
		// Moderne Browser zeichnen das Foto schon richtig herum gedreht (EXIF-Ausrichtung).
		const faktor = Math.min(1, kante / Math.max(bild.naturalWidth, bild.naturalHeight));
		const breite = Math.round(bild.naturalWidth * faktor);
		const hoehe = Math.round(bild.naturalHeight * faktor);
		const leinwand = document.createElement('canvas');
		leinwand.width = breite;
		leinwand.height = hoehe;
		const ctx = leinwand.getContext('2d');
		ctx.fillStyle = '#fff';
		ctx.fillRect(0, 0, breite, hoehe);
		ctx.drawImage(bild, 0, 0, breite, hoehe);
		return await new Promise((fertig, fehler) =>
			leinwand.toBlob((blob) => (blob ? fertig(blob) : fehler(new Error('Umrechnen fehlgeschlagen'))), 'image/jpeg', 0.82)
		);
	} finally {
		URL.revokeObjectURL(url);
	}
}
