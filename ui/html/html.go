package html

import (
	"embed"
	"html/template"
)

//go:embed *.html
var FS embed.FS

func Parse(funcMap template.FuncMap, file string) *template.Template {
	return template.Must(template.New("base.html").Funcs(funcMap).ParseFS(FS, "base.html", file))
}
