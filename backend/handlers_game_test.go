package main

import (
	"testing"

	"github.com/gorilla/mux"
	"github.com/matryer/is"
)

//Each test function has 4 discrete steps
//1. Setup dependencies and server
//2. Make the requests
//3. Check the reponses
//4. Reset databasestate

func TestHandleGotoPlanet(t *testing.T) {
	// Setting up dependencies
	is := is.New(t)
	db := newTestDb()
	defer db.Teardown()

	s := server{
		dg:     db.dg,
		router: mux.NewRouter(),
	}
	s.routes()
}
