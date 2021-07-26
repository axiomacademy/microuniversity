package main

import (
	"net/http"

	firebase "firebase.google.com/go"
	"github.com/dgraph-io/dgo/v200"
	"github.com/gorilla/mux"
)

// Abstracts away any global state
type server struct {
	dg     *dgo.Dgraph
	fb     *firebase.App
	router *mux.Router
}

// Implements HTTP.Handler
func (s server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Pass control to the router
	s.router.ServeHTTP(w, r)
}
