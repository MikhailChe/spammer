package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mikhailche/spammer/pages/send"
	"github.com/mikhailche/spammer/pages/static"
)

func main() {
	mux := chi.NewMux()

	mux.Mount("/static/", static.Handler)
	mux.Mount("/", http.HandlerFunc(send.Page))

	http.ListenAndServe(":8080", mux)
}
