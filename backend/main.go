package main

import (
	"encoding/json"
	"net/http"

	"github.com/dgraph-io/dgo/v200"
	"github.com/dgraph-io/dgo/v200/protos/api"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go"

	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

// Global environmental variables
var DB_URL string

var fb *firebase.App
var c *dgo.Dgraph

// Constants
const PLANET_ENERGY_DEPLETION int = 100
const PLANET_REWARD int = 100
const TOTAL_PLANET_KNOWLEDGE int = 100
const STARSYSTEM_ENERGY_DEPLETION int = 100
const CHALLENGE_ENERGY_DEPLETION int = 100
const TUTORIAL_ENERGY_DEPLETION int = 100
const LECTURE_ENERGY_DEPLETION int = 100
const CHALLENGE_KNOWLEDGE int = 100

func main() {
	fmt.Println("Server initialising...")

	// Getting all the environmental variables
	DB_URL = os.Getenv("DB_URL")
	checkEnvVariable(DB_URL)

	// Loading up firebase
	var err error
	opt := option.WithCredentialsFile("./fb-creds.json")
	fb, err = firebase.NewApp(context.Background(), nil, opt)
	PanicOnError(err)

	// Initialise the dgraph database
	d, err := grpc.Dial(DB_URL, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	c = dgo.NewDgraphClient(
		api.NewDgraphClient(d),
	)

	// Initialise the router
	r := mux.NewRouter()
	auth := r.PathPrefix("/api/v0.3").Subrouter()

	// Lecture handlers
	auth.HandleFunc("/lecture/complete", completeLecture).Methods("POST", "OPTIONS")
	auth.HandleFunc("/lecture/recommended", recommendedLectures).Methods("GET", "OPTIONS")

	// Review handlers
	auth.HandleFunc("/review", getDailyReview).Methods("GET", "OPTIONS")
	auth.HandleFunc("/review/complete", completeReview).Methods("POST", "OPTIONS")
	auth.HandleFunc("/review/card/pass", passReviewCard).Methods("POST", "OPTIONS")
	auth.HandleFunc("/review/card/fail", failReviewCard).Methods("POST", "OPTIONS")

	// Challenge handlers
	auth.HandleFunc("/challenge/accept", acceptChallenge).Methods("POST", "OPTIONS")
	auth.HandleFunc("/challenge/complete", completeChallenge).Methods("POST", "OPTIONS")

	// Tutorial handlers
	auth.HandleFunc("/tutorial/enroll", enrollTutorial).Methods("POST", "OPTIONS")

	// Game handlers
	auth.HandleFunc("/planet/goto", gotoPlanet).Methods("POST", "OPTIONS")
	auth.HandleFunc("/starsystem/goto", gotoStarsystem).Methods("POST", "OPTIONS")

	// Enabling middlewares
	r.Use(corsMiddleware)
	auth.Use(authMiddleware)

	log.Print("All setup running, and available on port 8003")
	log.Fatal(http.ListenAndServe(":8003", r))
}

/******************* MIDDLEWARES ****************************/
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(200)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqToken := r.Header.Get("Authorization")

		if reqToken == "" {
			fmt.Println("Pew")
			http.Error(w, "No auth token", http.StatusForbidden)
			return
		}

		ctx := context.Background()
		client, err := fb.Auth(ctx)
		if err != nil {
			http.Error(w, "Error validating auth token", http.StatusInternalServerError)
			return
		}

		token, err := client.VerifyIDToken(ctx, reqToken)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, "Auth token invalid", http.StatusForbidden)
			return
		}

		// Valid auth token received check if user exists
		email := token.Claims["email"].(string)

		// Retrieve the uid
		const userUid = `
			query userUid($email: string) {
					userUid(func: eq(Learner.email, $email)) {
						uid
						Learner.coins
						Learner.energy
					}
			}
		`

		txn := c.NewTxn()
		defer txn.Discard(r.Context())

		resp, err := txn.QueryWithVars(r.Context(), userUid, map[string]string{
			"$email": email,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var decode struct {
			UserUid []struct {
				Uid    string
				Coins  int `json:"Learner.coins,omitempty"`
				Energy int `json:"Learner.energy,omitempty"`
			}
		}

		if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println(decode.UserUid[0].Uid)

		r.Header.Set("X-User-Claim", email)
		r.Header.Set("X-Uid-Claim", decode.UserUid[0].Uid)
		r.Header.Set("X-Coins-Claim", string(decode.UserUid[0].Coins))
		r.Header.Set("X-Energy-Claim", string(decode.UserUid[0].Energy))
		next.ServeHTTP(w, r)
	})
}

/********* UTILITIES **************/
func PanicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func checkEnvVariable(env string) {
	if env == "" {
		log.Panic("Some environmental variables are not populated")
		return
	}
}
