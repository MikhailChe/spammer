package static

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*
var files embed.FS
var Handler http.Handler

func init() {
	sub, err := fs.Sub(files, "static")
	if err != nil {
		panic(err)
	}
	Handler = http.StripPrefix("/static/", http.FileServer(http.FS(sub)))
}
