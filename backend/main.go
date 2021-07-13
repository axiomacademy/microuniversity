package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

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
	auth.HandleFunc("/lecture/complete", completeLecture).Methods("POST", "OPTIONS") // Deplete energy
	auth.HandleFunc("/lecture/recommended", recommendedLectures).Methods("GET", "OPTIONS")

	// Review handlers
	auth.HandleFunc("/review", getDailyReview).Methods("GET", "OPTIONS")
	auth.HandleFunc("/review/complete", completeReview).Methods("POST", "OPTIONS")
	auth.HandleFunc("/review/card/pass", passReviewCard).Methods("POST", "OPTIONS")
	auth.HandleFunc("/review/card/fail", failReviewCard).Methods("POST", "OPTIONS")

	// Challenge handlers
	auth.HandleFunc("/challenge/accept", acceptChallenge).Methods("POST", "OPTIONS")     // Depletes energy
	auth.HandleFunc("/challenge/complete", completeChallenge).Methods("POST", "OPTIONS") // Mines knowledge

	// Tutorial handlers
	auth.HandleFunc("/tutorial/enroll", enrollTutorial).Methods("POST", "OPTIONS") // Depletes energy

	// Game handlers
	auth.HandleFunc("/planet/goto", gotoPlanet).Methods("POST", "OPTIONS")         // Deplete Energy
	auth.HandleFunc("/starsystem/goto", gotoStarsystem).Methods("POST", "OPTIONS") // Deplete Energy

	// Enabling middlewares
	r.Use(corsMiddleware)
	auth.Use(authMiddleware)

	log.Print("All setup running, and available on port 8003")
	log.Fatal(http.ListenAndServe(":8003", r))
}

/************************* GAME HANDLERS ************************************/
func gotoPlanet(w http.ResponseWriter, r *http.Request) {
	luid := r.Header.Get("X-Uid-Claim")
	energy := r.Header.Get("X-Energy-Claim")

	query := r.URL.Query()
	planetId := query.Get("planetId")

	if planetId == "" {
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	// Calculate energy depletion
	if e := energy - PLANET_ENERGY_DEPLETION; e >= 0 {
		energy = e
	} else {
		fmt.Println("Not enough energy")
		http.Error(w, "Not enough energy", http.StatusBadRequest)
		return
	}

	// Make sure the planet is visitable
	const checkPlanetNearby = `
		query checkPlanetNearby($planetId: string, $learnerId: string) {
			checkPlanetNearby(func: uid($learnerId)) @cascade {
				Learner.currentPlanet @cascade {
					Planet.starSystem @cascade {
						StarSystem.planets @filter(uid($planetId)) {
							uid
						}
					}
				}
			}
		}
	`

	txn := c.NewTxn()
	defer txn.Discard(r.Context())

	resp, err := txn.QueryWithVars(r.Context(), checkPlanetNearby, map[string]string{
		"$planetId":  planetId,
		"$learnerId": luid,
	})
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var d struct {
		CheckPlanetNearby []Learner
	}

	if err := json.Unmarshal(resp.GetJson(), &d); err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If there's nothing means it's not nearby
	if len(d.CheckPlanetNearby) == 0 {
		fmt.Println("Planet not nearby")
		http.Error(w, "Planet not nearby", http.StatusBadRequest)
		return
	}

	l := Learner{
		Uid:    luid,
		Energy: energy,
		CurrentPlanet: Planet{
			Uid: planetId,
		},
	}

	pl, err := json.Marshal(l)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mu := &api.Mutation{
		SetJson: pl,
	}

	_, err = txn.Mutate(r.Context(), mu)
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

func gotoStarsystem(w http.ResponseWriter, r *http.Request) {
	luid := r.Header.Get("X-Uid-Claim")
	energy := r.Header.Get("X-Energy-Claim")

	query := r.URL.Query()
	planetId := query.Get("planetId")

	if planetId == "" {
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	// Calculate energy depletion
	if e := energy - STARSYSTEM_ENERGY_DEPLETION; e >= 0 {
		energy = e
	} else {
		fmt.Println("Not enough energy")
		http.Error(w, "Not enough energy", http.StatusBadRequest)
		return
	}

	// Make sure the planet is visitable
	const checkSystemNearby = `
query checkPlanetNearby($systemId: string, $planetId: string, $learnerId: string) {
			checkPlanetNearby(func: uid($learnerId)) @cascade {
				Learner.currentPlanet @cascade {
					Planet.starSystem @cascade {
						StarSystem.nearbySystems @cascade {
							uid
							StarSystem.name
							StarSystem.planets @filter(uid($planetId)) {
								uid
							}
						}
					}
				}
			}
		}
	`

	txn := c.NewTxn()
	defer txn.Discard(r.Context())

	resp, err := txn.QueryWithVars(r.Context(), checkSystemNearby, map[string]string{
		"$systemId":  systemId,
		"$planetId":  planetId,
		"$learnerId": luid,
	})
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var d struct {
		CheckSystemNearby []Learner
	}

	if err := json.Unmarshal(resp.GetJson(), &d); err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If there's nothing means it's not nearby
	if len(d.CheckSystemNearby) == 0 {
		fmt.Println("System not nearby")
		http.Error(w, "System not nearby", http.StatusBadRequest)
		return
	}

	// TODO: generate planets and systems programatically

	l := Learner{
		Uid:    luid,
		Energy: energy,
		CurrentPlanet: Planet{
			Uid: planetId,
		},
	}

	pl, err := json.Marshal(l)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mu := &api.Mutation{
		SetJson: pl,
	}

	_, err = txn.Mutate(r.Context(), mu)
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

/************************* TUTORIAL HANDLERS ************************************/
func enrollTutorial(w http.ResponseWriter, r *http.Request) {
	luid := r.Header.Get("X-Uid-Claim")
	energy := r.Header.Get("X-Energy-Claim")

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

/************************* CHALLENGE HANDLERS ***********************************/
func acceptChallenge(w http.ResponseWriter, r *http.Request) {
	luid := r.Header.Get("X-Uid-Claim")
	energy := r.Header.Get("X-Energy-Claim")

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

	txn := c.NewTxn()
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
	_, err := txn.Do(r.Context(), req)
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

func completeChallenge(w http.ResponseWriter, r *http.Request) {
	luid := r.Header.Get("X-Uid-Claim")
	lcoins := r.Header.Get("X-Coins-Claim")

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
					Planet.totalKnowledge
					Planet.minedKnowledge
					Planet.reward
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

	txn := c.NewTxn()
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
		CheckIfChallengeComplete []struct {
			Challenges []LearnerChallenge `json:"Learner.challenges"`
		}
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
	if mineLevel < currentPlanet.TotalKnowledge {
		currentPlanet.MinedKnowledge = mineLevel
	} else {
		currentPlanet.MinedKnowledge = currentPlanet.totalKnowledge
		currentPlanet.Completed = true
	}

	// Add unlocked tutorials
	l := Learner{
		Uid:               luid,
		UnlockedTutorials: unlockedTutorials,
		CurrentPlanet:     currentPlanet,
	}

	if currentPlanet.Completed {
		l.Coins = lcoins + currentPlanet.reward
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

/************************* REVIEW HANDLERS ***********************************/

func passReviewCard(w http.ResponseWriter, r *http.Request) {
	luid := r.Header.Get("X-Uid-Claim")
	query := r.URL.Query()
	reviewCardId := query.Get("reviewCardId")

	const getReviewCard = `
		query getReviewCard($reviewCardId: string, $learnerId: string) {
			getReviewCard(func: uid($reviewCardId)) @cascade {
				LearnerReviewCard.repeat
				LearnerReviewCard.learner @filter(uid($learnerId)) {}
			}
		}
	`

	txn := c.NewTxn()
	defer txn.Discard(r.Context())

	resp, err := txn.QueryWithVars(r.Context(), getReviewCard, map[string]string{
		"$reviewCardId": reviewCardId,
		"$learnerId":    luid,
	})
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var decode struct {
		GetReviewCard []struct {
			Repeat int `json:"LearnerReviewCard.repeat"`
		}
	}

	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(decode.GetReviewCard) != 1 {
		fmt.Println("Oops")
		http.Error(w, "oops", http.StatusInternalServerError)
		return
	}

	// Make the mutation
	var repeat int
	if decode.GetReviewCard[0].Repeat-1 >= 0 {
		repeat = decode.GetReviewCard[0].Repeat - 1
	} else {
		repeat = 0
	}

	emptyTime, err := time.Parse(time.RFC3339, "")
	update := LearnerReviewCard{
		Uid:      reviewCardId,
		Repeat:   repeat,
		Selected: emptyTime,
	}

	pb, err := json.Marshal(update)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mu := &api.Mutation{
		SetJson: pb,
	}

	_, err = txn.Mutate(r.Context(), mu)
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

func failReviewCard(w http.ResponseWriter, r *http.Request) {
	luid := r.Header.Get("X-Uid-Claim")
	query := r.URL.Query()
	reviewCardId := query.Get("reviewCardId")

	const getReviewCard = `
		query getReviewCard($reviewCardId: string, $learnerId: string) {
			getReviewCard(func: uid($reviewCardId)) @cascade {
				LearnerReviewCard.repeat
				LearnerReviewCard.learner @filter(uid($learnerId)) {}
			}
		}
	`

	txn := c.NewTxn()
	defer txn.Discard(r.Context())

	resp, err := txn.QueryWithVars(r.Context(), getReviewCard, map[string]string{
		"$reviewCardId": reviewCardId,
		"$learnerId":    luid,
	})
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var decode struct {
		GetReviewCard []struct {
			Repeat int `json:"LearnerReviewCard.repeat"`
		}
	}

	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(decode.GetReviewCard) != 1 {
		fmt.Println("Oops")
		http.Error(w, "oops", http.StatusInternalServerError)
		return
	}

	// Make the mutation
	emptyTime, err := time.Parse(time.RFC3339, "")
	update := LearnerReviewCard{
		Uid:      reviewCardId,
		Repeat:   3,
		Selected: emptyTime,
	}

	pb, err := json.Marshal(update)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mu := &api.Mutation{
		SetJson: pb,
	}

	_, err = txn.Mutate(r.Context(), mu)
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

func completeReview(w http.ResponseWriter, r *http.Request) {
	luid := r.Header.Get("X-Uid-Claim")

	l := Learner{
		Uid:           luid,
		LastCompleted: time.Now().UTC(),
	}

	txn := c.NewTxn()
	defer txn.Discard(r.Context())

	pb, err := json.Marshal(l)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mu := &api.Mutation{
		SetJson: pb,
	}

	_, err = txn.Mutate(r.Context(), mu)
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

func getDailyReview(w http.ResponseWriter, r *http.Request) {
	luid := r.Header.Get("X-Uid-Claim")

	const checkLastCompleted = `
		query checkLastCompleted($learnerId: string) {
			checkLastCompleted(func: uid($learnerId)) {
				Learner.lastCompleted
				Learner.timezone
			}
		}
	`

	txn := c.NewTxn()
	defer txn.Discard(r.Context())

	resp, err := txn.QueryWithVars(r.Context(), checkLastCompleted, map[string]string{
		"$learnerId": luid,
	})
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var decode struct {
		CheckLastCompleted []struct {
			Timezone      string    `json:"Learner.timezone"`
			LastCompleted time.Time `json:"Learner.lastCompleted"`
		}
	}

	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(decode.CheckLastCompleted) == 0 {
		fmt.Println("User doesn't exist?")
		http.Error(w, "User doesn't exist?", http.StatusInternalServerError)
		return
	}

	timezoneStr := decode.CheckLastCompleted[0].Timezone
	lastCompleted := decode.CheckLastCompleted[0].LastCompleted

	now := time.Now().UTC()
	lastCompleted = lastCompleted.UTC()

	location, err := time.LoadLocation(timezoneStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Last completed should already be in UTC because of TimezoneTZ
	now = now.In(location)
	lastCompleted = lastCompleted.In(location)

	d1 := time.Date(lastCompleted.Year(), lastCompleted.Month(), lastCompleted.Day(), 0, 0, 0, 0, lastCompleted.Location())
	d2 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	fmt.Println(d1)
	fmt.Println(d2)

	// Today's review is already completed
	if d1.Unix() == d2.Unix() {
		res := []GqlLearnerReviewCard{}

		// Marshal to JSON and return
		dres, err := json.Marshal(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(dres)
		return
	}

	const getExistingReview = `
		query getExistingReviewCards($learnerId: string, $today: string) {
			getExistingReviewCards(func: uid($learnerId)) {
				Learner.cards @filter(eq(LearnerReviewCard.selected, $today)){
					uid
					LearnerReviewCard.repeat
					LearnerReviewCard.selected
					LearnerReviewCard.reviewCard {
						ReviewCard.topText
						ReviewCard.bottomText
					}
				}
			}
		}

		query getRepeatReviewCards($learnerId: string) {
			getRepeatReviewCards(func: uid($learnerId)) {
				Learner.cards @filter(gt(LearnerReviewCard.repeat, 0)){
					uid
					LearnerReviewCard.repeat
					LearnerReviewCardselected
					LearnerReviewCard.reviewCard {
						ReviewCard.topText
						ReviewCard.bottomText
					}
				}
			}
		}
		
		query getRemainingReviewCards($learnerId: string) {
			getRemainingReviewCards(func: uid($learnerId)) {
				Learner.cards @filter(eq(LearnerReviewCard.repeat, 0)){
					uid
					LearnerReviewCard.repeat
					LearnerReviewCard.selected
					LearnerReviewCard.reviewCard {
						ReviewCard.topText
						ReviewCard.bottomText
					}
				}
			}
		}
	`

	resp, err = txn.QueryWithVars(r.Context(), getExistingReview, map[string]string{
		"$learnerId": luid,
		"$today":     d2.Format(time.RFC3339),
	})
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var decode1 struct {
		GetExistingReviewCards []struct {
			Cards []LearnerReviewCard `json:"Learner.cards"`
		}
		GetRepeatReviewCards []struct {
			Cards []LearnerReviewCard `json:"Learner.cards"`
		}
		GetRemainingReviewCards []struct {
			Cards []LearnerReviewCard `json:"Learner.cards"`
		}
	}

	if err := json.Unmarshal(resp.GetJson(), &decode1); err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println(decode1.GetExistingReviewCards)

	if len(decode1.GetExistingReviewCards) > 0 {
		res := []GqlLearnerReviewCard{}

		for _, card := range decode1.GetExistingReviewCards[0].Cards {
			res = append(res, card.toGql())
		}

		// Marshal to JSON and return
		dres, err := json.Marshal(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(dres)
		return
	}

	var selectedCards []LearnerReviewCard
	var repeatCards []LearnerReviewCard
	var remainingCards []LearnerReviewCard

	if len(decode1.GetRepeatReviewCards) > 0 {
		repeatCards = decode1.GetRepeatReviewCards[0].Cards
	} else {
		repeatCards = []LearnerReviewCard{}
	}
	if len(decode1.GetRemainingReviewCards) > 0 {
		remainingCards = decode1.GetRemainingReviewCards[0].Cards
	} else {
		remainingCards = []LearnerReviewCard{}
	}

	remainingCount := 20 - len(repeatCards)
	availableCount := len(remainingCards)

	fmt.Println(remainingCards)
	fmt.Println(repeatCards)

	if remainingCount <= 0 {
		// Pick random 20
		v := rand.Perm(len(repeatCards))[0:20]
		for _, value := range v {
			selectedCards = append(selectedCards, repeatCards[value])
		}
	} else if remainingCount >= availableCount {
		// just return all the cards
		for _, card := range repeatCards {
			selectedCards = append(selectedCards, card)
		}
		for _, card := range remainingCards {
			selectedCards = append(selectedCards, card)
		}
	} else {

		// Get a random permutation
		v := rand.Perm(availableCount)[0:remainingCount]

		for _, card := range repeatCards {
			selectedCards = append(selectedCards, card)
		}

		for _, value := range v {
			selectedCards = append(selectedCards, remainingCards[value])
		}
	}

	fmt.Println(selectedCards)

	// Mark all the selected cards
	for _, card := range selectedCards {
		card.Selected = d2
		pb, err := json.Marshal(card)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		mu := &api.Mutation{
			SetJson: pb,
		}

		_, err = txn.Mutate(r.Context(), mu)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = txn.Commit(r.Context())
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the selectedCards
	res := []GqlLearnerReviewCard{}

	for _, card := range selectedCards {
		res = append(res, card.toGql())
	}

	fmt.Println(res)

	// Marshal to JSON and return
	dres, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(dres)
	return
}

/************************* LECTURE HANDLERS ***********************************/

func recommendedLectures(w http.ResponseWriter, r *http.Request) {
	luid := r.Header.Get("X-Uid-Claim")

	const getRecommendedLectures = `
		query getRecommendedLectures($learnerId: string) {
			var(func: uid("0x2")) {
				Learner.completedLectures {
				 A as Lecture.preReqs @filter(NOT uid_in(Lecture.completedLearners, "0x2")) {
						uid
						title: Lecture.title
					}
				 B as Lecture.postReqs @filter(NOT uid_in(Lecture.completedLearners, "0x2")) {
						uid
						title: Lecture.title
					}
				}
			}

			getRecommendedLectures(func: uid(A,B)) {
				uid
				Lecture.title
			}
		}
	`

	txn := c.NewTxn()
	defer txn.Discard(r.Context())

	resp, err := txn.QueryWithVars(r.Context(), getRecommendedLectures, map[string]string{
		"$learnerId": luid,
	})
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println(string(resp.GetJson()))

	var decode struct {
		GetRecommendedLectures []Lecture `json:"getRecommendedLectures"`
	}

	if err := json.Unmarshal(resp.GetJson(), &decode); err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(decode.GetRecommendedLectures) == 0 {
		http.Error(w, "oops", http.StatusInternalServerError)
		return
	}

	fmt.Println(decode)

	// Return the selectedCards
	res := []GqlLecture{}

	for _, lecture := range decode.GetRecommendedLectures {
		res = append(res, lecture.toGql())
	}

	fmt.Println(res)

	// Marshal to JSON and return
	dres, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(dres)
	return
}

/***************************** LECTURE HANDLERS **************************************/

func completeLecture(w http.ResponseWriter, r *http.Request) {
	luid := r.Header.Get("X-Uid-Claim")
	energy := r.Header.Get("X-Energy-Claim")

	query := r.URL.Query()
	lectureId := query.Get("lectureId")

	if lectureId == "" {
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	// Calculate energy depletion
	if e := energy - LECTURE_ENERGY_DEPLETION; e >= 0 {
		energy = e
	} else {
		fmt.Println("Not enough energy")
		http.Error(w, "Not enough energy", http.StatusBadRequest)
		return
	}

	// Check if lecture is already complete
	const getLearner = `
		query getLearner($lectureId: string, $learnerId: string) {
			getLearner(func: uid($learnerId)) @cascade {
				uid
				Learner.completedLectures @filter(uid($lectureId)) {}
			}
		}
	`

	txn := c.NewTxn()
	defer txn.Discard(r.Context())

	resp, err := txn.QueryWithVars(r.Context(), getLearner, map[string]string{
		"$lectureId": lectureId,
		"$learnerId": luid,
	})
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Decode
	var d struct {
		GetLearner []Learner
	}

	if err := json.Unmarshal(resp.GetJson(), &d); err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// If there's nothing means it's incomplete
	if len(d.GetLearner) != 0 {
		fmt.Println("Already complete")
		http.Error(w, "Already complete", http.StatusBadRequest)
		return
	}

	lecture := Lecture{
		Uid: lectureId,
		CompletedLearners: []Learner{{
			Uid: luid,
		}},
	}

	plecture, err := json.Marshal(lecture)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mu2 := &api.Mutation{
		SetJson: plecture,
	}

	_, err = txn.Mutate(r.Context(), mu2)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Only returns challenges that are ready to be added
	const lCards = `
		query lectureCard($lectureId: string, $learnerId: string) {
			lectureCards(func: uid($lectureId)) {
				cards: Lecture.cards {
					uid
					ReviewCard.topText
					ReviewCard.bottomText
				}
				challenges: Lecture.unlocksChallenges {
					uid
					Challenge.title
					Challenge.description
					Challenge.requiredLectures @filter(NOT uid_in(Lecture.completedLearners, 0x2)) {
						uid
					}
				}
			}
		}
	`

	resp, err = txn.QueryWithVars(r.Context(), lCards, map[string]string{
		"$lectureId": lectureId,
		"$learnerId": luid,
	})
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Decode
	var d1 struct {
		LectureCards []struct {
			Cards      []ReviewCard
			Challenges []Challenge
		}
	}

	if err := json.Unmarshal(resp.GetJson(), &d1); err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(d1.LectureCards) > 0 {
		var filteredChallenges []Challenge

		for _, challenge := range d1.LectureCards[0].Challenges {
			if len(challenge.RequiredLectures) == 0 {
				filteredChallenges = append(filteredChallenges, challenge)
			}
		}

		// Iterate through any unlocked challenges and generate the mutations to create
		// and link challenges to the learner
		var mList []*api.Mutation
		for _, challenge := range filteredChallenges {
			c := LearnerChallenge{
				DType:     []string{"LearnerChallenge"},
				Learner:   Learner{Uid: luid},
				Challenge: Challenge{Uid: challenge.Uid},
				Status:    "UNLOCKED",
			}

			pc, err := json.Marshal(c)
			if err != nil {
				fmt.Println(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			mu := &api.Mutation{
				SetJson: pc,
			}
			mList = append(mList, mu)
		}

		// Iterate through the cards and generate the mutations to create and link
		// cards to the lecture
		for _, card := range d1.LectureCards[0].Cards {
			p := LearnerReviewCard{
				DType: []string{"LearnerReviewCard"},
				Learner: Learner{
					Uid: luid,
				},
				ReviewCard: ReviewCard{
					Uid: card.Uid,
				},
				Repeat: 0,
			}

			pb, err := json.Marshal(p)
			if err != nil {
				fmt.Println(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			mu := &api.Mutation{
				SetJson: pb,
			}
			mList = append(mList, mu)
		}

		// Commit all the mutations
		req := &api.Request{Mutations: mList}
		_, err := txn.Do(r.Context(), req)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Link the lecture to the cards
	const lReviewCards = `
		query learnerReviewCards($lectureId: string, $learnerId: string) {
			learnerReviewCards(func: type("LearnerReviewCard")) @cascade {
				uid
				LearnerReviewCard.learner @filter(uid($learnerId)) {}
				LearnerReviewCard.reviewCard @cascade {
					uid
					ReviewCard.topText
					ReviewCard.bottomText
					ReviewCard.lecture @filter(uid($lectureId)) {}
				}
			}
			learnerChallenges(func: type("LearnerChallenge")) @cascade {
				uid
				LearnerChallenge.status
				LearnerChallenge.learner @filter(uid($learnerId)) {}
				LearnerChallenge.challenge @cascade {
					uid
					Challenge.title
					Challenge.description
					Challenge.unlocksTutorials {
						Tutorial.title
					}
					Challenge.requiredLectures @filter(uid($lectureId)) {}
				}
			}
		}	
	`

	resp, err = txn.QueryWithVars(r.Context(), lReviewCards, map[string]string{
		"$lectureId": lectureId,
		"$learnerId": luid,
	})

	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Decode
	var d2 struct {
		LearnerReviewCards []LearnerReviewCard
		LearnerChallenges  []LearnerChallenge
	}

	if err := json.Unmarshal(resp.GetJson(), &d2); err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Adding the reverse edges
	l := Learner{
		Uid:        luid,
		Energy:     energy,
		Cards:      d2.LearnerReviewCards,
		Challenges: d2.LearnerChallenges,
		CompletedLectures: []Lecture{{
			Uid: lectureId,
		}},
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

	var cards []GqlLearnerReviewCard
	var challenges []GqlLearnerChallenge

	for _, card := range d2.LearnerReviewCards {
		cards = append(cards, card.toGql())
	}

	for _, challenge := range d2.LearnerChallenges {
		challenges = append(challenges, challenge.toGql())
	}

	// Returning the correct values
	httpRes := GqlCompleteLectureReward{
		Cards:      cards,
		Challenges: challenges,
	}

	dres, err := json.Marshal(httpRes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println(string(dres))

	w.Write(dres)
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
		r.Header.Set("X-Coins-Claim", decode.UserUid[0].Coins)
		r.Header.Set("X-Energy-Claim", decode.UserUid[0].Energy)
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
