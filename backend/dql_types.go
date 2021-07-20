package main

import "time"

type Challenge struct {
	Uid              string     `json:"uid,omitempty"`
	Title            string     `json:"Challenge.title,omitempty"`
	Description      string     `json:"Challenge.description,omitempty"`
	RequiredLectures []Lecture  `json:"Challenge.requiredLectures,omitempty"`
	UnlocksTutorials []Tutorial `json:"Challenge.unlocksTutorials,omitempty"`
}

func (c *Challenge) toGql() GqlChallenge {
	var tutorials []GqlTutorial

	for _, tutorial := range c.UnlocksTutorials {
		tutorials = append(tutorials, tutorial.toGql())
	}

	return GqlChallenge{
		Id:               c.Uid,
		Title:            c.Title,
		Description:      c.Description,
		UnlocksTutorials: tutorials,
	}
}

type LearnerChallenge struct {
	Uid       string    `json:"uid,omitempty"`
	DType     []string  `json:"dgraph.type,omitempty"`
	Status    string    `json:"LearnerChallenge.status,omitempty"`
	Challenge Challenge `json:"LearnerChallenge.challenge,omitempty"`
	Learner   Learner   `json:"LearnerChallenge.learner,omitempty"`
}

func (lc *LearnerChallenge) toGql() GqlLearnerChallenge {
	return GqlLearnerChallenge{
		Id:        lc.Uid,
		Status:    lc.Status,
		Challenge: lc.Challenge.toGql(),
	}
}

type Tutorial struct {
	Uid                string      `json:"uid,omitempty"`
	Title              string      `json:"Tutorial.title, omitempty"`
	Description        string      `json:"Tutorial.description"`
	RequiredChallenges []Challenge `json:"Tutorial.requiredChallenges"`
}

func (t *Tutorial) toGql() GqlTutorial {
	return GqlTutorial{
		Id:          t.Uid,
		Title:       t.Title,
		Description: t.Description,
	}
}

type TutorialCohort struct {
	Uid      string    `json:"uid,omitempty"`
	Tutorial Tutorial  `json:"TutorialCohort.tutorial,omitempty"`
	Status   string    `json:"TutorialCohort.status,omitempty"`
	Members  []Learner `json:"TutorialCohort.members,omitempty"`
}

type Learner struct {
	Uid string `json:"uid,omitempty"`

	Energy        int           `json:"Learner.energy,omitempty"`
	Coins         int           `json:"Learner.coins,omitempty"`
	CurrentPlanet LearnerPlanet `json:"Learner.currentPlanet,omitempty"`

	Cards             []LearnerReviewCard `json:"Learner.cards,omitempty"`
	Challenges        []LearnerChallenge  `json:"Learner.challenges,omitempty"`
	LastCompleted     time.Time           `json:"Learner.lastCompleted,omitempty"`
	CompletedLectures []Lecture           `json:"Learner.completedLectures,omitempty"`
	UnlockedTutorials []Tutorial          `json:"Learner.unlockedTutorials,omitempty"`
	ActiveCohorts     []TutorialCohort    `json:"Learner.activeCohorts,omitempty"`
}

func (l *Learner) toGql() GqlLearner {
	return GqlLearner{
		Id: l.Uid,
	}
}

type Planet struct {
	Uid            string     `json:"uid,omitempty"`
	StarSystem     StarSystem `json:"Planet.starSystem,omitempty"`
	TotalKnowledge int        `json:"Planet.totalKnowledge,omitempty"`
	Reward         int        `json:"Planet.reward,omitempty"`
}

type LearnerPlanet struct {
	Uid            string  `json:"uid,omitempty"`
	Planet         Planet  `json:"LearnerPlanet.planet,omitempty"`
	Learner        Learner `json:"LearnerPlanet.learner,omitempty"`
	MinedKnowledge int     `json:"LearnerPlanet.minedKnowledge,omitempty"`
	Completed      bool    `json:"LearnerPlanet.completed,omitempty"`
}

type StarSystem struct {
	Uid           string       `json:"uid,omitempty"`
	Name          string       `json:"StarSystem.name,omitempty"`
	Planets       []Planet     `json:"StarSystem.planets,omitempty"`
	NearbySystems []StarSystem `json:"StarSystem.nearbySystems,omitempty"`
}

type Lecture struct {
	Uid               string      `json:"uid,omitempty"`
	Title             string      `json:"Lecture.title,omitempty"`
	CompletedLearners []Learner   `json:"Lecture.completedLearners,omitempty"`
	UnlocksChallenges []Challenge `json:"Lecture.unlocksChallenges,omitempty"`
}

func (l *Lecture) toGql() GqlLecture {
	var learners []GqlLearner

	for _, learner := range l.CompletedLearners {
		learners = append(learners, learner.toGql())
	}

	return GqlLecture{
		Id:                l.Uid,
		Title:             l.Title,
		CompletedLearners: learners,
	}
}

type ReviewCard struct {
	Uid        string `json:"uid,omitempty"`
	TopText    string `json:"ReviewCard.topText,omitempty"`
	BottomText string `json:"ReviewCard.bottomText,omitempty"`
}

func (card *ReviewCard) toGql() GqlReviewCard {
	return GqlReviewCard{
		Id:         card.Uid,
		TopText:    card.TopText,
		BottomText: card.BottomText,
	}
}

type LearnerReviewCard struct {
	Uid        string     `json:"uid,omitempty"`
	DType      []string   `json:"dgraph.type,omitempty"`
	Learner    Learner    `json:"LearnerReviewCard.learner,omitempty"`
	ReviewCard ReviewCard `json:"LearnerReviewCard.reviewCard,omitempty"`
	Selected   time.Time  `json:"LearnerReviewCard.selected,omitempty"`
	Repeat     int        `json:"LearnerReviewCard.repeat,omitempty"`
}

func (lCard *LearnerReviewCard) toGql() GqlLearnerReviewCard {
	return GqlLearnerReviewCard{
		Id:         lCard.Uid,
		ReviewCard: lCard.ReviewCard.toGql(),
		Selected:   lCard.Selected,
		Repeat:     lCard.Repeat,
	}
}
