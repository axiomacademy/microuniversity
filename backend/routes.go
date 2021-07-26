package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Setup all the server routes
func (s *server) routes() {

	// Lecture handlers
	s.router.HandleFunc("/lecture/complete", s.handleCompleteLecture()).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/lecture/recommended", s.handleRecommendedLectures()).Methods("GET", "OPTIONS")

	// Review handlers
	s.router.HandleFunc("/review", s.handleDailyReview()).Methods("GET", "OPTIONS")
	s.router.HandleFunc("/review/complete", s.handleCompleteReview()).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/review/card/pass", s.handlePassReviewCard()).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/review/card/fail", s.handleFailReviewCard()).Methods("POST", "OPTIONS")

	// Challenge handlers
	s.router.HandleFunc("/challenge/accept", s.handleAcceptChallenge()).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/challenge/complete", s.handleCompleteChallenge()).Methods("POST", "OPTIONS")

	// Tutorial handlers
	s.router.HandleFunc("/tutorial/enroll", s.handleEnrollTutorial()).Methods("POST", "OPTIONS")

	// Game handlers
	s.router.HandleFunc("/planet/goto", s.handleGotoPlanet()).Methods("POST", "OPTIONS")
	s.router.HandleFunc("/starsystem/goto", s.handleGotoStarsystem()).Methods("POST", "OPTIONS")

	s.router.Use(s.corsMiddleware)
	s.router.Use(s.authMiddleware)
}

/******************* MIDDLEWARES ****************************/
func (s *server) corsMiddleware(next http.Handler) http.Handler {
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

func (s *server) authMiddleware(next http.Handler) http.Handler {
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
