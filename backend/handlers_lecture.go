package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dgraph-io/dgo/v200/protos/api"
)

// Generate the recommended lectures
// Requires: learnerId
// 1. Fetches all the completed lectures whose direct pre-reqs/post-reqs are incomplete
func (s *server) handleRecommendedLectures() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l, ok := r.Context().Value("learner").(Learner)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

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

		txn := s.dg.NewTxn()
		defer txn.Discard(r.Context())

		resp, err := txn.QueryWithVars(r.Context(), getRecommendedLectures, map[string]string{
			"$learnerId": l.Uid,
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
}

// Complete a lecture
// Required: lectureId, learnerId, energy
// 1. Check sufficient energy
// 2. Check that it is incomplete
// 3. Add review cards to the user
// 4. Check and add unlocked challenges
// 5. Mark lecture as complete and deplete energy
func (s *server) handleCompleteLecture() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l, ok := r.Context().Value("learner").(Learner)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		query := r.URL.Query()
		lectureId := query.Get("lectureId")

		if lectureId == "" {
			http.Error(w, "Invalid query parameters", http.StatusBadRequest)
			return
		}

		// Calculate energy depletion
		if e := l.Energy - LECTURE_ENERGY_DEPLETION; e >= 0 {
			l.Energy = e
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

		txn := s.dg.NewTxn()
		defer txn.Discard(r.Context())

		resp, err := txn.QueryWithVars(r.Context(), getLearner, map[string]string{
			"$lectureId": lectureId,
			"$learnerId": l.Uid,
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
				Uid: l.Uid,
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
			"$learnerId": l.Uid,
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
					Learner:   Learner{Uid: l.Uid},
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
						Uid: l.Uid,
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
		const lCardsAndChallenges = `
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

		resp, err = txn.QueryWithVars(r.Context(), lCardsAndChallenges, map[string]string{
			"$lectureId": lectureId,
			"$learnerId": l.Uid,
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
		updateLearner := Learner{
			Uid:        l.Uid,
			Energy:     l.Energy,
			Cards:      d2.LearnerReviewCards,
			Challenges: d2.LearnerChallenges,
			CompletedLectures: []Lecture{{
				Uid: lectureId,
			}},
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
}
