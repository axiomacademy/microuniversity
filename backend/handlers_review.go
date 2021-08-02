package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/dgraph-io/dgo/v200/protos/api"
)

// Pass Review Card
// Requires: learnerId, reviewCardId
// 1. Remove review card from daily list
// 2. Cofigure the repeat count to be decremented until 0
func (s *server) handlePassReviewCard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l, ok := r.Context().Value("learner").(Learner)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

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

		txn := s.dg.NewTxn()
		defer txn.Discard(r.Context())

		resp, err := txn.QueryWithVars(r.Context(), getReviewCard, map[string]string{
			"$reviewCardId": reviewCardId,
			"$learnerId":    l.Uid,
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
}

// Fail Review Card
// Requires: learnerId, reviewCardId
// 1. Remove review card from daily list
// 2. Set the repeat count to 3
func (s *server) handleFailReviewCard() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l, ok := r.Context().Value("learner").(Learner)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

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

		txn := s.dg.NewTxn()
		defer txn.Discard(r.Context())

		resp, err := txn.QueryWithVars(r.Context(), getReviewCard, map[string]string{
			"$reviewCardId": reviewCardId,
			"$learnerId":    l.Uid,
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
}

// Mark daily review as complete
// Requires: learnerId
// * Sets last completed to the current time
func (s *server) handleCompleteReview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l, ok := r.Context().Value("learner").(Learner)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		updateLearner := Learner{
			Uid:           l.Uid,
			LastCompleted: time.Now().UTC(),
		}

		txn := s.dg.NewTxn()
		defer txn.Discard(r.Context())

		pb, err := json.Marshal(updateLearner)
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
}

// Get the daily review
// Requires: learnerId
// 1. Checks if the daily review is already complete
// 2. Gets all the cards that need to be repeated
// 3. Select 20 of them if repeat > 20
// 4. If repeat < 20, select a random number of other cards to make the number 20
func (s *server) handleDailyReview() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l, ok := r.Context().Value("learner").(Learner)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		const checkLastCompleted = `
		query checkLastCompleted($learnerId: string) {
			checkLastCompleted(func: uid($learnerId)) {
				Learner.lastCompleted
				Learner.timezone
			}
		}
	`

		txn := s.dg.NewTxn()
		defer txn.Discard(r.Context())

		resp, err := txn.QueryWithVars(r.Context(), checkLastCompleted, map[string]string{
			"$learnerId": l.Uid,
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
			"$learnerId": l.Uid,
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
}
