package main

import (
	htmlvorlage "html/template"
	textvorlage "text/template"
)

// Schlicht und hell wie die App (M3-Farben der Bibliothek), mit Stil am Element: Viele Mailprogramme
// verwerfen Stylesheets im Kopf.

var mailHTML = htmlvorlage.Must(htmlvorlage.New("mail").Parse(`<!doctype html>
<html lang="de"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><meta name="color-scheme" content="light"><title>{{.Betreff}}</title></head>
<body style="margin:0;padding:0;background:#faf9fd;color:#1a1c1e;font-family:Roboto,'Segoe UI',Helvetica,Arial,sans-serif;">
<div style="max-width:560px;margin:0 auto;padding:24px 16px;">
<p style="margin:0 0 4px;font-size:14px;line-height:20px;color:#42474e;">Schadenmeldung{{if .Schule}} · {{.Schule}}{{end}}</p>
<h1 style="margin:0 0 20px;font-size:24px;line-height:32px;font-weight:400;">{{.Ueberschrift}}</h1>
{{with .Meldung}}<div style="background:#f4f3f7;border-radius:12px;padding:16px;margin:0 0 12px;">
{{if .Dringend}}<p style="margin:0 0 8px;"><span style="display:inline-block;background:#ffdad6;color:#93000a;border-radius:8px;padding:2px 8px;font-size:12px;line-height:16px;font-weight:500;">Dringend</span></p>{{end}}
<p style="margin:0;font-size:16px;line-height:24px;font-weight:500;">{{.Ort}}{{if .Bereich}}<span style="font-weight:400;color:#42474e;font-size:14px;"> · {{.Bereich}}</span>{{end}}</p>
<p style="margin:2px 0 0;font-size:16px;line-height:24px;">{{.Was}}</p>
{{if .Notiz}}<p style="margin:2px 0 0;font-size:14px;line-height:20px;color:#42474e;">{{.Notiz}}</p>{{end}}
<p style="margin:8px 0 0;font-size:12px;line-height:16px;color:#42474e;">{{.Wann}}{{if .Von}} · {{.Von}}{{end}}{{if .MitFoto}} · mit Foto{{end}}</p>
{{if $.MitKnoepfen}}<p style="margin:6px 0 0;">{{/* Kein Leerraum zwischen den Knöpfen: Der Abstand kommt allein vom Rand. */ -}}
<a href="{{.LinkGesehen}}" style="display:inline-block;margin:8px 8px 0 0;padding:9px 15px;border:1px solid #c2c7cf;border-radius:999px;background:#ffffff;color:#42474e;font-size:14px;line-height:20px;font-weight:500;text-decoration:none;">Gesehen</a>{{- /* */ -}}
<a href="{{.LinkFachfirma}}" style="display:inline-block;margin:8px 8px 0 0;padding:9px 15px;border:1px solid #c2c7cf;border-radius:999px;background:#ffffff;color:#42474e;font-size:14px;line-height:20px;font-weight:500;text-decoration:none;">Fachfirma</a>{{- /* */ -}}
<a href="{{.LinkErledigt}}" style="display:inline-block;margin:8px 0 0;padding:10px 16px;border-radius:999px;background:#0061a4;color:#ffffff;font-size:14px;line-height:20px;font-weight:500;text-decoration:none;">Erledigt</a>
</p>{{end}}
</div>
{{end}}<p style="margin:20px 0 0;font-size:14px;line-height:20px;"><a href="{{.LinkApp}}" style="color:#0061a4;font-weight:500;text-decoration:none;">Alle Aufgaben in der App</a><span style="color:#42474e;"> · {{.Offen}} offen</span></p>
<p style="margin:24px 0 0;font-size:12px;line-height:16px;color:#72777f;">{{if .MitKnoepfen}}„Gesehen“ zeigt dem Kollegium, dass Bescheid gesagt ist – ohne Zusage, wann. {{end}}Diese Mail kommt, weil die Adresse in der Schadenmeldung eingetragen ist. Abbestellen: in der App im Kontomenü.</p>
</div></body></html>`))

var mailText = textvorlage.Must(textvorlage.New("mail").Parse(`{{.Ueberschrift}}
{{with .Meldung}}
{{if .Dringend}}[DRINGEND] {{end}}{{.Ort}}{{if .Bereich}} ({{.Bereich}}){{end}}
{{.Was}}{{if .Notiz}}
{{.Notiz}}{{end}}
{{.Wann}}{{if .Von}} · {{.Von}}{{end}}
{{if $.MitKnoepfen}}Gesehen:   {{.LinkGesehen}}
Fachfirma: {{.LinkFachfirma}}
Erledigt:  {{.LinkErledigt}}
{{end}}{{end}}
Alle Aufgaben ({{.Offen}} offen): {{.LinkApp}}

Abbestellen: in der App im Kontomenü.
`))
