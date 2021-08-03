package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"github.com/dgraph-io/dgo/v200"
	"github.com/dgraph-io/dgo/v200/protos/api"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

/*
 * The code oraganisation in this codebase is heavily inspired by Mat Ryer
 * Read the article @ https://pace.dev/blog/2018/05/09/how-I-write-http-services-after-eight-years.html
 *
 * All HTTP server code is handled inside/by the server struct in server.go
 */

// Constants
const PLANET_ENERGY_DEPLETION int = 100
const PLANET_REWARD int = 100
const TOTAL_PLANET_KNOWLEDGE int = 100
const STARSYSTEM_ENERGY_DEPLETION int = 100
const CHALLENGE_ENERGY_DEPLETION int = 100
const TUTORIAL_ENERGY_DEPLETION int = 100
const LECTURE_ENERGY_DEPLETION int = 100
const CHALLENGE_KNOWLEDGE int = 100

// Main abstraction to catch setup errors
func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

// Setup dependencies for the server and run it
func run() error {
	// Setting up the logger
	logger, _ := zap.NewProduction()
	sugar := logger.Sugar()

	sugar.Infow("Server initialising...")

	// Getting all the environmental variables
	DB_URL := os.Getenv("DB_URL")
	checkEnvVariable(DB_URL)

	// Loading up firebase
	var err error
	opt := option.WithCredentialsFile("./fb-creds.json")
	fb, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return err
	}

	// Initialise the dgraph database
	conn, err := grpc.Dial(DB_URL, grpc.WithInsecure())
	if err != nil {
		return err
	}

	dg := dgo.NewDgraphClient(
		api.NewDgraphClient(conn),
	)

	// Setup server
	s := server{
		dg:     dg,
		fb:     fb,
		router: mux.NewRouter(),
		logger: logger,
	}
	s.routes()

	sugar.Infow("All setup running, and availale on port 8003")
	http.ListenAndServe(":8003", s)

	return nil
}

/*
 * Utilities
 */
func checkEnvVariable(env string) {
	if env == "" {
		log.Panic("Some environmental variables are not populated")
		return
	}
}
