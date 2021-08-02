package main

import (
	"context"
	"encoding/json"
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

/*
 * Middlewares
 */
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
			http.Error(w, "No auth token", http.StatusForbidden)
			return
		}

		client, err := s.fb.Auth(r.Context())
		if err != nil {
			http.Error(w, "Error validating auth token", http.StatusInternalServerError)
			return
		}

		token, err := client.VerifyIDToken(r.Context(), reqToken)
		if err != nil {
			s.logger.Error(err.Error())
			http.Error(w, "Auth token invalid", http.StatusForbidden)
			return
		}

		// Valid auth token received check if user exists
		email := token.Claims["email"].(string)

		// Retrieve the uid
		const learner = `
			query learner($email: string) {
					learner(func: eq(Learner.email, $email)) {
						uid
						Learner.coins
						Learner.energy
					}
			}
		`

		txn := s.dg.NewTxn()
		defer txn.Discard(r.Context())

		resp, err := txn.QueryWithVars(r.Context(), learner, map[string]string{
			"$email": email,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var decode struct {
			Learner []Learner
		}

		if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(decode.Learner) == 0 {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		l := decode.Learner[0]
		s.logger.Info(l.Uid)

		r.Header.Set("X-User-Claim", email)

		ctx := context.WithValue(r.Context(), "learner", l)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
