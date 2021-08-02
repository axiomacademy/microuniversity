package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dgraph-io/dgo/v200/protos/api"
)

// Enroll in a tutorial
// Requires: energy, learnerId, tutorialId
// 1. Check if enough energy, and tutorial is unlocked
// 2. Find an unfilled cohort and add them in
// 3. If unfilled cohort becomes filled, set the status
// 4. If there is no unfilled cohort, create a new cohort and add them in
func (s *server) handleEnrollTutorial() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l, ok := r.Context().Value("learner").(Learner)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		query := r.URL.Query()
		tutorialId := query.Get("tutorialId")

		if tutorialId == "" {
			http.Error(w, "Invalid query parameters", http.StatusBadRequest)
			return
		}

		// Calculate energy depletion
		if e := l.Energy - TUTORIAL_ENERGY_DEPLETION; e >= 0 {
			l.Energy = e
		} else {
			fmt.Println("Not enough energy")
			http.Error(w, "Not enough energy", http.StatusBadRequest)
			return
		}

		// Check that you've unlocked the tutorial
		const checkIfTutorialUnlocked = `
		query checkIfTutorialUnlocked($tutorialId: string, $learnerId: string) {
			checkIfTutorialUnlocked(func: uid($learnerId)) {
				Learner.unlockedTutorials @filter(uid($tutorialId)) {
					uid
				}
			}
		}
	`

		txn := s.dg.NewTxn()
		defer txn.Discard(r.Context())

		resp, err := txn.QueryWithVars(r.Context(), checkIfTutorialUnlocked, map[string]string{
			"$tutorialId": tutorialId,
			"$learnerId":  l.Uid,
		})
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var decode struct {
			CheckIfTutorialUnlocked []Learner
		}

		if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(decode.CheckIfTutorialUnlocked) != 1 {
			fmt.Println("Oops")
			http.Error(w, "oops", http.StatusInternalServerError)
			return
		}

		// Tutorial is unlocked so find a cohort
		const findACohort = `
		query findACohort($tutorialId: string) {
			findACohort(func: type("TutorialCohort")) @filter(uid_in(TutorialCohort.tutorial, $tutorialId) AND eq(TutorialCohort.status, "FILLING")) {
				uid
			}
		}
	`

		resp, err = txn.QueryWithVars(r.Context(), checkIfTutorialUnlocked, map[string]string{
			"$tutorialId": tutorialId,
		})
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var decode1 struct {
			FindACohort []TutorialCohort
		}

		if err := json.Unmarshal(resp.GetJson(), &decode1); err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var tc TutorialCohort

		if len(decode1.FindACohort) == 0 {
			// Create a cohort instead\
			tc = TutorialCohort{
				Uid: "_:new",
				Tutorial: Tutorial{
					Uid: tutorialId,
				},
				Status: "FILLING",
				Members: []Learner{
					{
						Uid: l.Uid,
					},
				},
			}
		} else {
			tc = TutorialCohort{
				Tutorial: Tutorial{
					Uid: decode1.FindACohort[0].Uid,
				},
				Status: "FILLING",
				Members: []Learner{
					{
						Uid: l.Uid,
					},
				},
			}

			if len(decode1.FindACohort[0].Members) == 2 {
				tc.Status = "FILLED"
			}
		}

		updateLearner := Learner{
			Uid:    l.Uid,
			Energy: l.Energy,
			ActiveCohorts: []TutorialCohort{
				tc,
			},
		}

		createTutorial, err := json.Marshal(tc)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		createLearnerLink, err := json.Marshal(updateLearner)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		mu := &api.Mutation{
			SetJson: createTutorial,
		}

		mu1 := &api.Mutation{
			SetJson: createLearnerLink,
		}

		req := &api.Request{Mutations: []*api.Mutation{mu, mu1}}

		_, err = txn.Do(r.Context(), req)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = txn.Commit(r.Context())
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte("true"))
		return
	}
}