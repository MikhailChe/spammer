package base

import (
	_ "embed"
	"html/template"
)

//go:embed public/base.html
var baseHtml string

var t = template.Must(template.New("base").Parse(baseHtml))

func Page(withBody string) (*template.Template, error) {
	return t.New("body").Parse(withBody)
}
