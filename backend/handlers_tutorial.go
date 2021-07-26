package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dgraph-io/dgo/v200/protos/api"
)

// Enroll in a tutorial
// Requires: energy, learnerId, tutorialId
// * Check if enough energy, and tutorial is unlocked
// * Find an unfilled cohort and add them in
// * If unfilled cohort becomes filled, set the status
// * If there is no unfilled cohort, create a new cohort and add them in
func (s *server) handleEnrollTutorial() http.HandleFunc {
	return func(w http.Response, r *http.Request) {
		luid := r.Header.Get("X-Uid-Claim")
		energy, _ := strconv.Atoi(r.Header.Get("X-Energy-Claim"))

		query := r.URL.Query()
		tutorialId := query.Get("tutorialId")

		if tutorialId == "" {
			http.Error(w, "Invalid query parameters", http.StatusBadRequest)
			return
		}

		// Calculate energy depletion
		if e := energy - TUTORIAL_ENERGY_DEPLETION; e >= 0 {
			energy = e
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

		txn := c.NewTxn()
		defer txn.Discard(r.Context())

		resp, err := txn.QueryWithVars(r.Context(), checkIfTutorialUnlocked, map[string]string{
			"$tutorialId": tutorialId,
			"$learnerId":  luid,
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
					Learner{
						Uid: luid,
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
					Learner{
						Uid: luid,
					},
				},
			}

			if len(decode1.FindACohort[0].Members) == 2 {
				tc.Status = "FILLED"
			}
		}

		l := Learner{
			Uid:    luid,
			Energy: energy,
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

		createLearnerLink, err := json.Marshal(l)
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

		/*
			err = txn.Commit(r.Context())
			if err != nil {
				fmt.Println(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		*/

		w.Write([]byte("true"))
		return
	}
}
