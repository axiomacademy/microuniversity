package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dgraph-io/dgo/v200/protos/api"
)

// Accepts a challenge
// Requires: challengeId, energy, learnerId
// * Checks that the there is enough energy for the Challenge
// * Subtracts the energy cost and sets the challenge status to INPROGRESS
func (s *server) handleAcceptChallenge() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		luid := r.Header.Get("X-Uid-Claim")
		energy, _ := strconv.Atoi(r.Header.Get("X-Energy-Claim"))

		query := r.URL.Query()
		challengeId := query.Get("challengeId")

		if challengeId == "" {
			http.Error(w, "Invalid query parameters", http.StatusBadRequest)
			return
		}

		// Calculate energy depletion
		if e := energy - CHALLENGE_ENERGY_DEPLETION; e >= 0 {
			energy = e
		} else {
			fmt.Println("Not enough energy")
			http.Error(w, "Not enough energy", http.StatusBadRequest)
			return
		}

		const checkChallengeStatus = `
			query checkIfChallengeComplete($challengeId: string, $learnerId: string) {
				checkIfChallengeComplete(func: uid($learnerId)) {
					Learner.challenges @filter(uid($challengeId)) {
						uid
						LearnerChallenge.status
					}
				}
			}
		`

		txn := s.dg.NewTxn()
		defer txn.Discard(r.Context())

		resp, err := txn.QueryWithVars(r.Context(), checkChallengeStatus, map[string]string{
			"$challengeId": challengeId,
			"$learnerId":   luid,
		})
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var decode struct {
			CheckChallengeStatus []struct {
				Challenges []LearnerChallenge `json:"Learner.challenges"`
			}
		}

		if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(decode.CheckChallengeStatus[0].Challenges) != 1 {
			fmt.Println("Oops")
			http.Error(w, "oops", http.StatusInternalServerError)
			return
		}

		if decode.CheckChallengeStatus[0].Challenges[0].Status != "UNLOCKED" {
			fmt.Println("Incorrect status")
			http.Error(w, "Incorrect status", http.StatusBadRequest)
			return
		}

		// Set challenge to inprogress
		update := LearnerChallenge{
			Uid:    challengeId,
			Status: "INPROGRESS",
		}

		updateChallenge, err := json.Marshal(update)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println(string(updateChallenge))

		mu := &api.Mutation{
			SetJson: updateChallenge,
		}

		updateLearner := Learner{
			Uid:    luid,
			Energy: energy,
		}

		pl, err := json.Marshal(updateLearner)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		mu1 := &api.Mutation{
			SetJson: pl,
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

		return
	}
}

// Completes a challenge
// Requires: challengeId, coins, learnerId
// * Checks if a challenge is complete,
// * Check currentPlanet can still be mined
// * Set challenge to complete and check for unlocked Tutorials
// * Add unlocked tutorials to learner and increment current plant mining
// * Check if planet is fully mined, and if so increment the coins
func (s *server) handleCompleteChallenge() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		luid := r.Header.Get("X-Uid-Claim")
		lcoins, _ := strconv.Atoi(r.Header.Get("X-Coins-Claim"))

		query := r.URL.Query()
		challengeId := query.Get("challengeId")

		if challengeId == "" {
			http.Error(w, "Invalid query parameters", http.StatusBadRequest)
			return
		}

		// Check if challenge is already complete
		const checkIfChallengeComplete = `
	query checkIfChallengeComplete($challengeId: string, $learnerId: string) {
			checkIfChallengeComplete(func: uid($learnerId)) {
				Learner.currentPlanet {
					uid
					LearnerPlanet.minedKnowledge
					LearnerPlanet.completed
				}
				Learner.challenges @filter(uid($challengeId)) {
					uid
					LearnerChallenge.status
					LearnerChallenge.challenge {
						Challenge.unlocksTutorials {
							uid
							Tutorial.title
							Tutorial.requiredChallenges {
								uid
								Challenge.title
							}
						}
					}
				}
			}
		}
	`

		txn := s.dg.NewTxn()
		defer txn.Discard(r.Context())

		resp, err := txn.QueryWithVars(r.Context(), checkIfChallengeComplete, map[string]string{
			"$challengeId": challengeId,
			"$learnerId":   luid,
		})
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var decode struct {
			CheckIfChallengeComplete []Learner
		}

		if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(decode.CheckIfChallengeComplete[0].Challenges) != 1 {
			fmt.Println("Oops")
			http.Error(w, "oops", http.StatusInternalServerError)
			return
		}

		// Pull out current planet details
		currentPlanet := decode.CheckIfChallengeComplete[0].CurrentPlanet

		// Check if planet has already been completely mined
		if currentPlanet.Completed {
			fmt.Println("Already mined planet")
			http.Error(w, "Already mined planet", http.StatusBadRequest)
			return
		}

		if decode.CheckIfChallengeComplete[0].Challenges[0].Status == "COMPLETED" {
			fmt.Println("Already complete")
			http.Error(w, "Already complete", http.StatusBadRequest)
			return
		}

		// Set challenge to completed
		update := LearnerChallenge{
			Uid:    challengeId,
			Status: "COMPLETED",
		}

		updateChallenge, err := json.Marshal(update)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println(string(updateChallenge))

		mu := &api.Mutation{
			SetJson: updateChallenge,
		}

		_, err = txn.Mutate(r.Context(), mu)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Check all the tutorials to see if anything should be unlocks
		const getTutorialUnlocked = `
		query getTutorialUnlocked($tutorialId: string, $learnerId: string) {
			var(func: uid($tutorialId)) {
				A as Tutorial.requiredChallenges {
					uid
					Challenge.title
				}
			}
	
			getLearnerChallenges(func: uid($learnerId)) {
				uid
				Learner.challenges @filter(uid_in(LearnerChallenge.challenge, uid(A)) AND eq(LearnerChallenge.status, "COMPLETED")) {
					uid
					LearnerChallenge.challenge {
						uid
						Challenge.title
					}
				}
			}
		}
	`

		tutorials := decode.CheckIfChallengeComplete[0].Challenges[0].Challenge.UnlocksTutorials
		var unlockedTutorials []Tutorial

		for _, tutorial := range tutorials {
			fmt.Println(tutorial.Uid)
			resp, err := txn.QueryWithVars(r.Context(), getTutorialUnlocked, map[string]string{
				"$tutorialId": tutorial.Uid,
				"$learnerId":  luid,
			})
			if err != nil {
				fmt.Println(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			fmt.Println(resp)

			var decode1 struct {
				GetLearnerChallenges []Learner
			}

			if err := json.Unmarshal(resp.GetJson(), &decode1); err != nil {
				fmt.Println(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			fmt.Println(decode1)

			if len(decode1.GetLearnerChallenges) != 1 {
				fmt.Println("Oops")
				http.Error(w, "oops", http.StatusInternalServerError)
				return
			}

			completedCount := len(decode1.GetLearnerChallenges[0].Challenges)
			unlockCount := len(tutorial.RequiredChallenges)

			fmt.Println(completedCount)
			fmt.Println(unlockCount)

			// Then we're ready to unlock
			if completedCount == unlockCount {
				unlockedTutorials = append(unlockedTutorials, tutorial)
			}
		}

		// Check currrent planet mining levels
		mineLevel := currentPlanet.MinedKnowledge + CHALLENGE_KNOWLEDGE
		if mineLevel < TOTAL_PLANET_KNOWLEDGE {
			currentPlanet.MinedKnowledge = mineLevel
		} else {
			currentPlanet.MinedKnowledge = TOTAL_PLANET_KNOWLEDGE
			currentPlanet.Completed = true
		}

		// Add unlocked tutorials
		l := Learner{
			Uid:               luid,
			UnlockedTutorials: unlockedTutorials,
			CurrentPlanet:     currentPlanet,
		}

		if currentPlanet.Completed {
			l.Coins = lcoins + PLANET_REWARD
		}

		pl, err := json.Marshal(l)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		mu1 := &api.Mutation{
			SetJson: pl,
		}

		_, err = txn.Mutate(r.Context(), mu1)
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

		var gqlTutorials []GqlTutorial

		for _, tutorial := range unlockedTutorials {
			gqlTutorials = append(gqlTutorials, tutorial.toGql())
		}

		dres, err := json.Marshal(gqlTutorials)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(dres)
	}
}
