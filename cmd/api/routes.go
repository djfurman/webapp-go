package main

import (
	"net/http"

	chi "github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	return mux
}
