package main

import "time"

type GqlCompleteLectureReward struct {
	Cards      []GqlLearnerReviewCard `json:"cards,omitempty"`
	Challenges []GqlLearnerChallenge  `json:"challenges,omitempty"`
}

type GqlLearnerReviewCard struct {
	Id         string        `json:"id,omitempty"`
	ReviewCard GqlReviewCard `json:"reviewCard"`
	Selected   time.Time     `json:"selected"`
	Repeat     int           `json:"repeat"`
}

type GqlLearnerChallenge struct {
	Id        string       `json:"id,omitempty"`
	Status    string       `json:"status,omitempty"`
	Challenge GqlChallenge `json:"challenge,omitempty"`
}

type GqlChallenge struct {
	Id               string        `json:"id,omitempty"`
	Title            string        `json:"title,omitempty"`
	Description      string        `json:"description,omitempty"`
	UnlocksTutorials []GqlTutorial `json:"unlocksTutorials,omitempty"`
}

type GqlTutorial struct {
	Id          string `json:"id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
}

type GqlReviewCard struct {
	Id         string `json:"id,omitempty"`
	TopText    string `json:"topText,omitempty"`
	BottomText string `json:"bottomText,omitempty"`
}

type GqlLearner struct {
	Id                string                 `json:"id,omitempty"`
	Cards             []GqlLearnerReviewCard `json:"cards,omitempty"`
	CompletedLectures []GqlLecture           `json:"completedLectures,omitempty"`
}

type GqlLecture struct {
	Id                string       `json:"id,omitempty"`
	Title             string       `json:"title,omitempty"`
	CompletedLearners []GqlLearner `json:"completedLearners,omitempty"`
}
