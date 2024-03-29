package main

import (
	"database/sql"
	"encoding/json"
	"math/rand"
	"net/http"
	"regexp"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go"

	"google.golang.org/api/option"
)

// Global environmental variables
var DB_URL string
var JWT_SECRET string

// Global handlers for simplicity
var db *sql.DB

var fb *firebase.App

func main() {
	fmt.Println("Server initialising...")

	// Getting all the environmental variables
	DB_URL = os.Getenv("DB_URL")
	checkEnvVariable(DB_URL)
	JWT_SECRET = os.Getenv("JWT_SECRET")
	checkEnvVariable(JWT_SECRET)

	// Loading up firebase
	var err error
	opt := option.WithCredentialsFile("./fb-creds.json")
	fb, err = firebase.NewApp(context.Background(), nil, opt)
	PanicOnError(err)

	// Initialise the database
	db, err = sql.Open("postgres", DB_URL)
	PanicOnError(err)
	defer db.Close()

	// Checking for valid database connection
	err = db.Ping()
	PanicOnError(err)

	// Initialise the router
	r := mux.NewRouter()

	// Unauthenticated endpoints (user management)
	r.HandleFunc("/api/v0.2/cohort/start", startCohort).Methods("POST", "OPTIONS")

	// Retrieve all the existing modules
	r.HandleFunc("/api/v0.2/modules", getModules).Methods("GET", "OPTIONS")

	auth := r.PathPrefix("/api/v0.2").Subrouter()

	// Related to cohorts
	auth.HandleFunc("/cohorts", getSelfCohorts).Methods("GET", "OPTIONS")
	auth.HandleFunc("/cohorts/available", getCohortsForModule).Methods("GET", "OPTIONS")
	auth.HandleFunc("/cohort/join", joinCohort).Methods("POST", "OPTIONS")
	auth.HandleFunc("/cohort/self", getModuleCohort).Methods("GET", "OPTIONS")
	auth.HandleFunc("/cohort/leave", leaveModuleCohort).Methods("DELETE", "OPTIONS")

	// Related to lectures
	auth.HandleFunc("/lectures/today", getLectureToday).Methods("GET", "OPTIONS")
	auth.HandleFunc("/lectures/past", getLecturesPast).Methods("GET", "OPTIONS")
	auth.HandleFunc("/lectures/complete", completeLecture).Methods("POST", "OPTIONS")
	auth.HandleFunc("/lectures/flashcards", getLectureFlashcards).Methods("GET", "OPTIONS")

	// Related to daily review
	auth.HandleFunc("/review", getDailyReview).Methods("GET", "OPTIONS")
	auth.HandleFunc("/review/complete", completeReview).Methods("POST", "OPTIONS")
	auth.HandleFunc("/flashcard/pass", passFlashcard).Methods("POST", "OPTIONS")
	auth.HandleFunc("/flashcard/fail", failFlashcard).Methods("POST", "OPTIONS")

	// Get self data
	auth.HandleFunc("/self", getSelf).Methods("GET", "OPTIONS")
	auth.HandleFunc("/self", updateSelf).Methods("PUT", "OPTIONS")

	// Get tutorial schedule
	auth.HandleFunc("/tutorials", getUpcomingTutorials).Methods("GET", "OPTIONS")

	// Enabling middlewares
	r.Use(corsMiddleware)
	auth.Use(authMiddleware)

	log.Print("All setup running, and available on port 8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}

type userResponse struct {
	Email         string    `json:"email"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	LastCompleted time.Time `json:"last_completed"`
	Streak        int       `json:"streak"`
	Timezone      string    `json:"timezone"`
}

func getSelf(w http.ResponseWriter, r *http.Request) {
	lemail := r.Header.Get("X-User-Claim")

	var res userResponse

	fmt.Println(lemail)

	sqlquery := `SELECT email, first_name, last_name, last_completed, streak, timezone FROM learner WHERE email = $1`
	if err := db.QueryRow(sqlquery, lemail).Scan(&res.Email, &res.FirstName, &res.LastName, &res.LastCompleted, &res.Streak, &res.Timezone); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if the last completed is more than a 2 days away
	diff := time.Now().Sub(res.LastCompleted)
	if diff.Hours() > 48 {
		res.Streak = 0
		// Reset streak
		sqlquery = `UPDATE learner SET streak = 0 WHERE email = $1`
		stmt, err := db.Prepare(sqlquery)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = stmt.Exec(lemail)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	dres, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(dres)
}

type updateSelfRequest struct {
	Firstname string `json:"first_name"`
	Lastname  string `json:"last_name"`
	Timezone  string `json:"timezone"`
}

func updateSelf(w http.ResponseWriter, r *http.Request) {
	lemail := r.Header.Get("X-User-Claim")

	var req updateSelfRequest

	// Parsing request body
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sqlquery := `UPDATE learner SET first_name = $1, last_name = $2, timezone = $3 WHERE email = $4`
	stmt, err := db.Prepare(sqlquery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = stmt.Exec(req.Firstname, req.Lastname, req.Timezone, lemail)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

/************** MODULE HANDLERS ******************************/
type moduleResponse struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Image       string `json:"image"`
	Description string `json:"description"`
	Duration    int    `json:"duration"`
}

func getModules(w http.ResponseWriter, r *http.Request) {
	var res []moduleResponse

	sqlquery := `SELECT module_id, title, image, description, duration FROM module`

	result, err := db.Query(sqlquery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer result.Close()

	for result.Next() {
		var module moduleResponse
		if err := result.Scan(&module.Id, &module.Title, &module.Image, &module.Description, &module.Duration); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res = append(res, module)
	}

	// Marshal to JSON and return
	dres, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(dres)
}

/*************** COHORT HANDLERS ***************************/
type cohortResponse struct {
	ModuleId    string    `json:"id"`
	Title       string    `json:"title"`
	Image       string    `json:"image"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	Duration    int       `json:"duration"`
	Status      int       `json:"status"`
}

func getSelfCohorts(w http.ResponseWriter, r *http.Request) {
	var res []cohortResponse

	lemail := r.Header.Get("X-User-Claim")

	sqlquery := `SELECT module, start_date, status FROM cohort
	INNER JOIN learner_cohort ON learner_cohort.cohort=cohort.cohort_id AND learner_cohort.learner=$1`

	result, err := db.Query(sqlquery, lemail)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer result.Close()

	for result.Next() {
		var cohort cohortResponse
		if err := result.Scan(&cohort.ModuleId, &cohort.StartDate, &cohort.Status); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Now retrieve the remaining module data
		modulequery := `SELECT title, image, description, duration FROM module WHERE module.module_id=$1`
		if err := db.QueryRow(modulequery, cohort.ModuleId).Scan(&cohort.Title, &cohort.Image, &cohort.Description, &cohort.Duration); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res = append(res, cohort)
	}

	if len(res) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Marshal to JSON and return
	dres, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(dres)
}

type cohortData struct {
	Id                 string
	Module             string
	Status             int
	StartDate          time.Time
	WeeklyTutorialDay  int
	WeeklyTutorialTime int
}

// Relative date is the number of days from the cohort start date
type lectureDate struct {
	Id           string
	RelativeDate int
	AbsoluteDate time.Time
}

// Week is which week the tutorial is on relative to the start date
// Week 0 is the first week
type tutorialDate struct {
	Id               string
	Week             int
	AbsoluteDateTime time.Time
}

type moduleCohortRes struct {
	Id           string `json:"id"`
	TutorialDay  int    `json:"tutorial_day"`
	TutorialTime int    `json:"tutorial_time"`
	LearnerCount int    `json:"learner_count"`
}

func getCohortsForModule(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	moduleId := query.Get("module")
	lemail := r.Header.Get("X-User-Claim")

	if moduleId == "" {
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	var dummy string
	// Check if learner is already enrolled in a cohort for the module
	sqlquery := `SELECT cohort FROM learner_cohort INNER JOIN cohort ON learner_cohort.cohort = cohort.cohort_id
								WHERE learner_cohort.learner=$1 AND cohort.module=$2`
	if err := db.QueryRow(sqlquery, lemail, moduleId).Scan(&dummy); err != sql.ErrNoRows {
		if err == nil {
			w.WriteHeader(http.StatusNoContent)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	sqlquery = `SELECT cohort_id, weekly_tutorial_day, weekly_tutorial_time, COUNT(learner_cohort.learner) learner_count
							FROM cohort LEFT JOIN learner_cohort ON learner_cohort.cohort = cohort.cohort_id
							WHERE cohort.module = $1 AND cohort.status = 0
							GROUP BY cohort_id, cohort.weekly_tutorial_day, cohort.weekly_tutorial_time`

	var res []moduleCohortRes

	result, err := db.Query(sqlquery, moduleId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer result.Close()

	for result.Next() {
		var cohort moduleCohortRes
		if err := result.Scan(&cohort.Id, &cohort.TutorialDay, &cohort.TutorialTime, &cohort.LearnerCount); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res = append(res, cohort)
	}

	// Marshal to JSON and return
	dres, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(dres)
}

func joinCohort(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	cohortId := query.Get("cohort")
	lemail := r.Header.Get("X-User-Claim")

	if cohortId == "" {
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	var cohort cohortData
	// CHeck that the cohort is open
	sqlquery := `SELECT cohort_id, weekly_tutorial_time, weekly_tutorial_day, module, status FROM cohort WHERE cohort_id = $1`
	if err := db.QueryRow(sqlquery, cohortId).Scan(&cohort.Id, &cohort.WeeklyTutorialTime, &cohort.WeeklyTutorialDay, &cohort.Module, &cohort.Status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if cohort.Status != 0 {
		http.Error(w, "Invalid cohort id", http.StatusBadRequest)
		return
	}

	var dummy string
	// Check if learner is already enrolled in a cohort for the module
	sqlquery = `SELECT cohort FROM learner_cohort INNER JOIN cohort ON learner_cohort.cohort = cohort.cohort_id
								WHERE learner_cohort.learner=$1 AND cohort.module=$2`
	if err := db.QueryRow(sqlquery, lemail, cohort.Module).Scan(&dummy); err != sql.ErrNoRows {
		if err == nil {
			http.Error(w, "Already enrolled in a cohort for the module", http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	sqlquery = `INSERT INTO learner_cohort(learner, cohort) VALUES ($1, $2)`
	if _, err := db.Exec(sqlquery, lemail, cohortId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var cohortLearnerCount int
	// Check total count
	sqlquery = `SELECT COUNT(learner_cohort.learner) learner_count
							FROM learner_cohort INNER JOIN cohort ON learner_cohort.cohort = cohort.cohort_id
							WHERE cohort.cohort_id = $1
							GROUP BY learner_cohort.cohort`
	if err := db.QueryRow(sqlquery, cohortId).Scan(&cohortLearnerCount); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Close enrollment into the cohort oonce full
	if cohortLearnerCount >= 15 {
		// Close enrollment
		sqlquery = `UPDATE cohort SET status=1 WHERE cohort_id=$1`
		if _, err := db.Exec(sqlquery, cohort.Id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// Checks which cohort you've enrolled in for the module and leaves it
func leaveModuleCohort(w http.ResponseWriter, r *http.Request) {
	// First retrieve the cohort
	lemail := r.Header.Get("X-User-Claim")
	query := r.URL.Query()
	moduleId := query.Get("module")

	if moduleId == "" {
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	// Check if they're even enrolled in any cohort
	sqlquery := `SELECT cohort_id, status FROM learner_cohort INNER JOIN cohort ON learner_cohort.cohort = cohort.cohort_id 
								WHERE learner_cohort.learner = $1 AND cohort.module = $2`

	var cohortId string
	var cohortStatus int

	if err := db.QueryRow(sqlquery, lemail, moduleId).Scan(&cohortId, &cohortStatus); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if cohortStatus != 0 {
		// It's too late to de-enroll
		http.Error(w, "The cohort has already been finalised", http.StatusBadRequest)
		return
	}

	// Else, all is good we cann de-enroll you
	sqlquery = `DELETE FROM learner_cohort WHERE learner = $1 AND cohort = $2`
	if _, err := db.Exec(sqlquery, lemail, cohortId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

type getModuleCohortRes struct {
	Id                 string    `json:"id"`
	Module             string    `json:"module"`
	Status             int       `json:"status"`
	StartDate          time.Time `json:"start_date"`
	WeeklyTutorialDay  int       `json:"tutorial_day"`
	WeeklyTutorialTime int       `json:"tutorial_time"`
	LearnerCount       int       `json:"learner_count"`
}

func getModuleCohort(w http.ResponseWriter, r *http.Request) {
	// First retrieve the cohort
	lemail := r.Header.Get("X-User-Claim")

	// Check if they're even enrolled in any cohort should only be one cohort
	sqlquery := `SELECT cohort_id, module, status, start_date, weekly_tutorial_day, weekly_tutorial_time FROM learner_cohort INNER JOIN cohort ON learner_cohort.cohort = cohort.cohort_id WHERE learner_cohort.learner = $1`

	var res getModuleCohortRes

	if err := db.QueryRow(sqlquery, lemail).Scan(&res.Id, &res.Module, &res.Status, &res.StartDate, &res.WeeklyTutorialDay, &res.WeeklyTutorialTime); err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNoContent)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Get the count
	sqlquery = `SELECT COUNT(learner_cohort.learner) learner_count
							FROM cohort LEFT JOIN learner_cohort ON learner_cohort.cohort = cohort.cohort_id
							WHERE cohort_id = $1
							GROUP BY cohort_id, cohort.weekly_tutorial_day, cohort.weekly_tutorial_time`

	if err := db.QueryRow(sqlquery, res.Id).Scan(&res.LearnerCount); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dres, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(dres)
}

func startCohort(w http.ResponseWriter, r *http.Request) {
	// First retrieve the cohort
	query := r.URL.Query()
	cohortId := query.Get("cohort")

	if cohortId == "" {
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	var cohort cohortData
	cohort.Id = cohortId

	cohortQuery := `SELECT module, status, start_date, weekly_tutorial_day, weekly_tutorial_time FROM cohort WHERE cohort_id=$1`

	if err := db.QueryRow(cohortQuery, cohort.Id).Scan(&cohort.Module, &cohort.Status, &cohort.StartDate, &cohort.WeeklyTutorialDay, &cohort.WeeklyTutorialTime); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculating absolute lecture dates
	var lectureDates []lectureDate
	lectureQuery := `SELECT lecture_id, date_offset FROM lecture WHERE module=$1`
	result, err := db.Query(lectureQuery, cohort.Module)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer result.Close()

	for result.Next() {
		var lDate lectureDate
		if err := result.Scan(&lDate.Id, &lDate.RelativeDate); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Calculate the absolute date
		lDate.AbsoluteDate = cohort.StartDate.AddDate(0, 0, lDate.RelativeDate)

		lectureDates = append(lectureDates, lDate)
	}

	// Calculating absolute tutorial dates
	var tutorialDates []tutorialDate

	// Weekly Tutorial Day is relative to Monday, being 0 and 6 on Sunday
	firstTutorialDate := cohort.StartDate.AddDate(0, 0, cohort.WeeklyTutorialDay)
	firstTutorialDateTime := firstTutorialDate.Add(time.Minute * time.Duration(cohort.WeeklyTutorialTime))

	tutorialQuery := `SELECT tutorial_id, week FROM tutorial WHERE module=$1`
	result, err = db.Query(tutorialQuery, cohort.Module)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer result.Close()

	for result.Next() {
		var tDate tutorialDate
		if err := result.Scan(&tDate.Id, &tDate.Week); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Calculate the absolute date
		tDate.AbsoluteDateTime = firstTutorialDateTime.AddDate(0, 0, tDate.Week*7)
		tutorialDates = append(tutorialDates, tDate)
	}

	// Get the learners in the cohort
	learnerQuery := `SELECT learner FROM learner_cohort WHERE cohort=$1`
	enrollLectureQuery := `INSERT INTO learner_lecture(learner, lecture, scheduled_date) VALUES ($1, $2, $3)`
	enrollTutorialQuery := `INSERT INTO learner_tutorial(learner, tutorial, scheduled_datetime) VALUES ($1, $2, $3)`

	// Preparing the statements
	enrollLectureStmt, err := db.Prepare(enrollLectureQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer enrollLectureStmt.Close()

	enrollTutorialStmt, err := db.Prepare(enrollTutorialQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer enrollTutorialStmt.Close()

	result, err = db.Query(learnerQuery, cohort.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer result.Close()

	for result.Next() {
		var learnerId string
		if err := result.Scan(&learnerId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Enrolling them into the lecture
		for _, lecture := range lectureDates {
			if _, err := enrollLectureStmt.Exec(learnerId, lecture.Id, lecture.AbsoluteDate); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Enrolling them into the tutorials
		for _, tutorial := range tutorialDates {
			if _, err := enrollTutorialStmt.Exec(learnerId, tutorial.Id, tutorial.AbsoluteDateTime); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
}

/*************** LECTURE HANDLERS ****************************/
type lectureResponse struct {
	Id            string `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	VideoLink     string `json:"video_link"`
	ScheduledDate string `json:"scheduled_date"`
	Module        string `json:"module"`
	Completed     bool   `json:"completed"`
}

func enrollModule(w http.ResponseWriter, r *http.Request) {

	// Get lecture id from query params
	query := r.URL.Query()
	module := query.Get("module")
	lemail := query.Get("email")

	// Get all of the lectures
	sql := `SELECT lecture_id, title, description, video_link, scheduled_date, module from lecture 
					WHERE module = $1`

	result, err := db.Query(sql, module)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer result.Close()

	for result.Next() {
		var lecture lectureResponse
		if err := result.Scan(&lecture.Id, &lecture.Title, &lecture.Description, &lecture.VideoLink, &lecture.ScheduledDate, &lecture.Module); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create the learner_lecture
		sql = `INSERT INTO learner_lecture(learner, lecture, completed) VALUES ($1, $2, $3)`
		stmt, err := db.Prepare(sql)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = stmt.Exec(lemail, lecture.Id, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func getLectureToday(w http.ResponseWriter, r *http.Request) {

	lemail := r.Header.Get("X-User-Claim")
	timezone := r.Header.Get("X-Timezone-Claim")

	var res lectureResponse

	// Get the current date application is date specific, regardless of timezone they should be shown some lecture at some local date
	// So take reference to a no timezone date value and compare to their timezone date
	local := time.Now().UTC()
	location, err := time.LoadLocation(timezone)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	local = local.In(location)

	query := `SELECT lecture_id, title, description, video_link, scheduled_date, completed, module from lecture 
	INNER JOIN learner_lecture ON learner_lecture.lecture=lecture.lecture_id AND learner_lecture.learner=$1 
	WHERE scheduled_date = $2 AND completed=$3`

	// There should only be one lecture
	err = db.QueryRow(query, lemail, local.Format("2006-01-02"), false).Scan(&res.Id, &res.Title, &res.Description, &res.VideoLink, &res.ScheduledDate, &res.Completed, &res.Module)

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNoContent)
			return
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	dres, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(dres)
}

func getLecturesPast(w http.ResponseWriter, r *http.Request) {
	lemail := r.Header.Get("X-User-Claim")
	timezone := r.Header.Get("X-Timezone-Claim")

	query := r.URL.Query()
	moduleId := query.Get("module")

	if moduleId == "" {
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	var res []lectureResponse

	// Get the current date application is date specific, regardless of timezone they should be shown some lecture at some local date
	// So take reference to a no timezone date value and compare to their timezone date
	local := time.Now().UTC()
	location, err := time.LoadLocation(timezone)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	local = local.In(location)

	// Selecting all lectures that happened earlier than today
	sqlquery := `SELECT lecture_id, title, description, video_link, scheduled_date, module, completed from lecture
					LEFT JOIN learner_lecture
					ON lecture.lecture_id = learner_lecture.lecture AND learner_lecture.learner = $1
					WHERE scheduled_date <= $2 AND module = $3`

	result, err := db.Query(sqlquery, lemail, local.Format("2006-01-02"), moduleId)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer result.Close()

	for result.Next() {
		var lecture lectureResponse
		if err := result.Scan(&lecture.Id, &lecture.Title, &lecture.Description, &lecture.VideoLink, &lecture.ScheduledDate, &lecture.Module, &lecture.Completed); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res = append(res, lecture)
	}

	if len(res) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Marshal to JSON and return
	dres, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(dres)
}

func completeLecture(w http.ResponseWriter, r *http.Request) {
	lemail := r.Header.Get("X-User-Claim")

	// Get lecture id from query params
	query := r.URL.Query()
	lectureId := query.Get("id")

	if lectureId == "" {
		http.Error(w, "No valid lecture id provided", http.StatusBadRequest)
		return
	}

	// Get all the flashcards associated to this lecture
	var flashcardIds []string
	sql := `SELECT flashcard_id FROM flashcard WHERE lecture = $1`
	result, err := db.Query(sql, lectureId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer result.Close()

	for result.Next() {
		var flashcardId string
		if err := result.Scan(&flashcardId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		flashcardIds = append(flashcardIds, flashcardId)
	}

	// Create the new learner_flashcard entries
	sql = `INSERT INTO learner_flashcard(learner, flashcard) VALUES ($1, $2)`
	stmt, err := db.Prepare(sql)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, flashcardId := range flashcardIds {
		_, err = stmt.Exec(lemail, flashcardId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}

	// Update the learner_lecture data
	sql = `UPDATE learner_lecture SET completed = true WHERE learner_lecture.learner = $1 AND learner_lecture.lecture = $2`
	stmt, err = db.Prepare(sql)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = stmt.Exec(lemail, lectureId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type flashcardResponse struct {
	Id         string `json:"id"`
	TopSide    string `json:"top_side"`
	BottomSide string `json:"bottom_side"`
	LectureId  string `json:"lecture_id"`
}

// You need to ensure that the lecture flashcards exist for the user
func getLectureFlashcards(w http.ResponseWriter, r *http.Request) {

	var res []flashcardResponse
	// Get lecture id from query params
	query := r.URL.Query()
	lectureId := query.Get("id")

	if lectureId == "" {
		http.Error(w, "No valid lecture id provided", http.StatusBadRequest)
		return
	}

	sql := `SELECT flashcard_id, top_side, bottom_side, lecture FROM flashcard WHERE lecture = $1`
	result, err := db.Query(sql, lectureId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer result.Close()

	for result.Next() {
		var card flashcardResponse
		if err := result.Scan(&card.Id, &card.TopSide, &card.BottomSide, &card.LectureId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res = append(res, card)
	}

	// Marshal to JSON and return
	dres, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(dres)
}

/******************* TUTORIAL HANDLERS **********************/
type tutorialResponse struct {
	Id            string    `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	ScheduledTime time.Time `json:"scheduled_time"`
	Module        string    `json:"module"`
}

func getUpcomingTutorials(w http.ResponseWriter, r *http.Request) {
	var res []tutorialResponse

	lemail := r.Header.Get("X-User-Claim")

	query := r.URL.Query()
	moduleId := query.Get("module")
	if moduleId == "" {
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	sqlquery := `SELECT tutorial_id, title, description, scheduled_datetime, module FROM tutorial
		INNER JOIN learner_tutorial ON learner_tutorial.tutorial=tutorial.tutorial_id AND learner_tutorial.learner=$1
		WHERE scheduled_datetime > NOW() AND module = $2 LIMIT 5`
	result, err := db.Query(sqlquery, lemail, moduleId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer result.Close()

	for result.Next() {
		var tutorial tutorialResponse
		if err := result.Scan(&tutorial.Id, &tutorial.Title, &tutorial.Description, &tutorial.ScheduledTime, &tutorial.Module); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res = append(res, tutorial)
	}

	if len(res) == 0 {
		w.WriteHeader(http.StatusNoContent)
	}

	// Marshal to JSON and return
	dres, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(dres)
}

/******************* DAILY REVIEW HANDLERS ******************/
func getDailyReview(w http.ResponseWriter, r *http.Request) {
	lemail := r.Header.Get("X-User-Claim")

	var res []flashcardResponse

	// Retrieve and check last completed
	var lastCompleted time.Time
	var timezoneStr string

	sqlquery := `SELECT last_completed, timezone FROM learner WHERE email = $1`
	if err := db.QueryRow(sqlquery, lemail).Scan(&lastCompleted, &timezoneStr); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := time.Now().UTC()
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
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Check for existing scheduled flashcards
	sqlquery = `SELECT flashcard_id, top_side, bottom_side, lecture FROM flashcard RIGHT JOIN learner_flashcard ON flashcard.flashcard_id = learner_flashcard.flashcard WHERE learner = $1 AND selected = $2`

	result, err := db.Query(sqlquery, lemail, d2)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer result.Close()

	for result.Next() {
		var flashcard flashcardResponse
		if err := result.Scan(&flashcard.Id, &flashcard.TopSide, &flashcard.BottomSide, &flashcard.LectureId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res = append(res, flashcard)
	}

	if len(res) != 0 {
		// Just return the response because we've already scheduled stuff for today
		// Marshal to JSON and return
		dres, err := json.Marshal(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(dres)
		return
	}

	// Need to request for the schedule
	// First retrieve all the repeat due ones
	sqlquery = `SELECT flashcard_id, top_side, bottom_side, lecture FROM flashcard RIGHT JOIN learner_flashcard ON flashcard.flashcard_id = learner_flashcard.flashcard WHERE learner = $1 AND repeat > 0`

	result, err = db.Query(sqlquery, lemail)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer result.Close()

	for result.Next() {
		var flashcard flashcardResponse
		if err := result.Scan(&flashcard.Id, &flashcard.TopSide, &flashcard.BottomSide, &flashcard.LectureId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res = append(res, flashcard)
	}

	var allFlashcards []flashcardResponse
	sqlquery = `SELECT flashcard_id, top_side, bottom_side, lecture FROM flashcard RIGHT JOIN learner_flashcard ON flashcard.flashcard_id = learner_flashcard.flashcard WHERE learner = $1 AND repeat = 0`
	result, err = db.Query(sqlquery, lemail)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer result.Close()

	for result.Next() {
		var flashcard flashcardResponse
		if err := result.Scan(&flashcard.Id, &flashcard.TopSide, &flashcard.BottomSide, &flashcard.LectureId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		allFlashcards = append(allFlashcards, flashcard)
	}

	// Check and calculate how much randomness to retrieve
	remaining := 20 - len(res)
	total := len(allFlashcards)

	if remaining <= 0 {
		// do nothing
	} else if remaining >= total {
		// just return all the cards
		for _, card := range allFlashcards {
			res = append(res, card)
		}
	} else {

		// Get a random permutation
		v := rand.Perm(total)[0:remaining]

		for _, value := range v {
			res = append(res, allFlashcards[value])
		}
	}

	if len(res) == 0 {
		// Means there are no flashcards at all, user didn't do any microlectures
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Update the database to show that these are marked to be done today
	sqlquery = `UPDATE learner_flashcard SET selected = $1 WHERE learner = $2 AND flashcard = $3`
	stmt, err := db.Prepare(sqlquery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, flashcard := range res {
		_, err = stmt.Exec(d2, lemail, flashcard.Id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Marshal to JSON and return
	dres, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(dres)

}

func passFlashcard(w http.ResponseWriter, r *http.Request) {
	lemail := r.Header.Get("X-User-Claim")

	// Get lecture id from query params
	query := r.URL.Query()
	flashcardId := query.Get("id")

	var repeat int

	sql := `SELECT repeat FROM learner_flashcard WHERE learner = $1 AND flashcard = $2`
	if err := db.QueryRow(sql, lemail, flashcardId).Scan(&repeat); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Decrement repeat count
	if repeat > 0 {
		repeat -= 1
	}

	sql = `UPDATE learner_flashcard SET repeat = $1, selected = NULL WHERE learner = $2 AND flashcard = $3`
	stmt, err := db.Prepare(sql)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = stmt.Exec(repeat, lemail, flashcardId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func failFlashcard(w http.ResponseWriter, r *http.Request) {
	lemail := r.Header.Get("X-User-Claim")

	// Get lecture id from query params
	query := r.URL.Query()
	flashcardId := query.Get("id")

	// set repeat to 3 no matter what
	repeat := 3

	sql := `UPDATE learner_flashcard SET repeat = $1, selected = NULL WHERE learner = $2 AND flashcard = $3`
	stmt, err := db.Prepare(sql)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = stmt.Exec(repeat, lemail, flashcardId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func completeReview(w http.ResponseWriter, r *http.Request) {
	lemail := r.Header.Get("X-User-Claim")

	// Retrieve the learner
	var streak int

	sql := `SELECT streak FROM learner WHERE email = $1`
	if err := db.QueryRow(sql, lemail).Scan(&streak); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sql = `UPDATE learner SET streak = $1, last_completed = $2 WHERE email = $3`
	stmt, err := db.Prepare(sql)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = stmt.Exec(streak+1, time.Now().UTC(), lemail)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
			http.Error(w, "Auth token invalid", http.StatusForbidden)
			return
		}

		// Valid auth token received check if user exists
		email := token.Claims["email"].(string)
		var timezone string

		// Get the user associated to the email if it exists
		sqlquery := `SELECT email, timezone FROM learner WHERE email = $1`
		err = db.QueryRow(sqlquery, email).Scan(&email, &timezone)

		if err == sql.ErrNoRows {
			// Means that the user is new and has to be created
			sqlquery = `INSERT INTO learner(email) VALUES ($1)`

			stmt, err := db.Prepare(sqlquery)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			_, err = stmt.Exec(email)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Successfully created user go create the Header
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		r.Header.Set("X-User-Claim", email)
		r.Header.Set("X-Timezone-Claim", timezone)
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

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// isEmailValid checks if the email provided passes the required structure and length.
func isEmailValid(e string) bool {
	if len(e) < 3 && len(e) > 254 {
		return false
	}
	return emailRegex.MatchString(e)
}
