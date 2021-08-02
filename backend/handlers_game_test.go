package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/matryer/is"
	"go.uber.org/zap"
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
	logger, _ := zap.NewDevelopment()

	s := server{
		dg:     db.dg,
		router: mux.NewRouter(),
		logger: logger,
	}

	req := httptest.NewRequest("GET", "/gotoPlanet?planetId=8", nil)
	l := Learner{
		Uid:    "23",
		Energy: 1000,
	}

	ctx := context.WithValue(req.Context(), "learner", l)
	w := httptest.NewRecorder()

	s.handleGotoPlanet().ServeHTTP(w, req.WithContext(ctx))

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(body)
	is.Equal(resp.StatusCode, http.StatusOK) // Status code not okay
}
