package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// Global environmental variables
var DB_URL string
var JWT_SECRET string

// Twilio env vars
var TWILIO_ACCOUNT_SID string
var TWILIO_AUTH_TOKEN string
var TWILIO_VERIFY_SID string

// Global handlers for simplicity
var db *sql.DB
var twilio *TwilioApi

func main() {
	fmt.Println("Server initialising...")

	// Getting all the environmental variables
	DB_URL = os.Getenv("DB_URL")
	checkEnvVariable(DB_URL)
	JWT_SECRET = os.Getenv("JWT_SECRET")
	checkEnvVariable(JWT_SECRET)

	// Twilio env vars
	TWILIO_ACCOUNT_SID = os.Getenv("TWILIO_ACCOUNT_SID")
	checkEnvVariable(TWILIO_ACCOUNT_SID)
	TWILIO_AUTH_TOKEN = os.Getenv("TWILIO_AUTH_TOKEN")
	checkEnvVariable(TWILIO_AUTH_TOKEN)
	TWILIO_VERIFY_SID = os.Getenv("TWILIO_VERIFY_SID")
	checkEnvVariable(TWILIO_VERIFY_SID)

	var err error
	// Initialise the database
	db, err = sql.Open("postgres", DB_URL)
	PanicOnError(err)
	defer db.Close()

	// Checking for valid database connection
	err = db.Ping()
	PanicOnError(err)

	// Initialise the twilio client
	twilio = NewTwilioApi(TWILIO_ACCOUNT_SID, TWILIO_AUTH_TOKEN, TWILIO_VERIFY_SID)

	// Initialise the router
	r := mux.NewRouter()

	// Unauthenticated endpoints (user management)
	r.HandleFunc("/api/v0.2/enrollLearner", enrollModule).Methods("POST", "OPTIONS")

	r.HandleFunc("/api/v0.2/register", registerLearner).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/v0.2/verifyOtp", verifyOtp).Methods("POST", "OPTIONS")

	auth := r.PathPrefix("/api/v0.2").Subrouter()

	// Related to lessons
	auth.HandleFunc("/lessons/today", getLessonToday).Methods("GET", "OPTIONS")
	auth.HandleFunc("/lessons/past", getLessonsPast).Methods("GET", "OPTIONS")
	auth.HandleFunc("/lessons/complete", completeLesson).Methods("POST", "OPTIONS")
	auth.HandleFunc("/lessons/flashcards", getLessonFlashcards).Methods("GET", "OPTIONS")

	// Related to daily review
	auth.HandleFunc("/review", getDailyReview).Methods("GET", "OPTIONS")
	auth.HandleFunc("/review/complete", completeReview).Methods("POST", "OPTIONS")
	auth.HandleFunc("/flashcard/pass", passFlashcard).Methods("POST", "OPTIONS")
	auth.HandleFunc("/flashcard/fail", failFlashcard).Methods("POST", "OPTIONS")

	// Get self data
	auth.HandleFunc("/self", getSelf).Methods("GET", "OPTIONS")

	// Get tutorial schedule
	auth.HandleFunc("/tutorials", getUpcomingTutorials).Methods("GET", "OPTIONS")

	// Enabling middlewares
	r.Use(corsMiddleware)
	auth.Use(authMiddleware)

	log.Print("All setup running, and available on port 8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}

/******************* USER MANAGEMENT HANDLERS ****************************/

func registerLearner(w http.ResponseWriter, r *http.Request) {
	// Get lesson id from query params
	query := r.URL.Query()
	email := query.Get("email")

	if email == "" {
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	// Check email regex for verification
	if !isEmailValid(email) {
		http.Error(w, "Invalid email", http.StatusBadRequest)
		return
	}

	err := twilio.StartEmailVerification(email)
	if err != nil {
		http.Error(w, "Error starting email verification", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusAccepted)
}

type loginResponse struct {
	Jwt string `json:"jwt"`
	Id  string `json:"id"`
}

func verifyOtp(w http.ResponseWriter, r *http.Request) {
	var response loginResponse

	// Get lesson id from query params
	query := r.URL.Query()
	code := query.Get("code")
	email := query.Get("email")

	if code == "" || email == "" {
		http.Error(w, "Invalid query parameters", http.StatusBadRequest)
		return
	}

	// Check email regex for verification
	if !isEmailValid(email) {
		http.Error(w, "Invalid email", http.StatusBadRequest)
		return
	}

	err := twilio.VerifyCode(code, email)
	if err == TWILIO_API_ERROR {
		http.Error(w, "Invalid OTP code", http.StatusUnauthorized)
	} else if err != nil {
		http.Error(w, "Error verifying code", http.StatusInternalServerError)
	}

	// Reaching here means OTP code is valid

	var lid string

	// Get the user associated to the email if it exists
	sqlquery := `SELECT learner_id FROM learner WHERE email = $1`
	err = db.QueryRow(sqlquery, email).Scan(&lid)

	if err == sql.ErrNoRows {
		// Means that the user is new and has to be created
		sqlquery = `INSERT INTO learner(email) VALUES ($1) RETURNING learner_id`
		err = db.QueryRow(sqlquery, email).Scan(&lid)
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := createJWT(lid, JWT_SECRET)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.Jwt = token
	response.Id = lid

	res, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(res)
}

type userResponse struct {
	Id            string    `json:"id"`
	Email         string    `json:"email"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	LastCompleted time.Time `json:"last_completed"`
	Streak        int       `json:"streak"`
}

func getSelf(w http.ResponseWriter, r *http.Request) {
	learnerId := r.Header.Get("X-User-Claim")

	var res userResponse

	fmt.Println(learnerId)

	sql := `SELECT learner_id, email, first_name, last_name, last_completed, streak FROM learner WHERE learner_id = $1`
	if err := db.QueryRow(sql, learnerId).Scan(&res.Id, &res.Email, &res.FirstName, &res.LastName, &res.LastCompleted, &res.Streak); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check if the last completed is more than a 2 days away
	diff := time.Now().Sub(res.LastCompleted)
	if diff.Hours() > 48 {
		res.Streak = 0
		// Reset streak
		sql = `UPDATE learner SET streak = 0 WHERE learner_id = $1`
		stmt, err := db.Prepare(sql)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = stmt.Exec(learnerId)
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

/*************** LESSON HANDLERS ****************************/
type lessonResponse struct {
	Id            string `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	VideoLink     string `json:"video_link"`
	ScheduledDate string `json:"scheduled_date"`
	Module        string `json:"module"`
	Completed     bool   `json:"completed"`
}

func enrollModule(w http.ResponseWriter, r *http.Request) {

	// Get lesson id from query params
	query := r.URL.Query()
	module := query.Get("module")
	learnerId := query.Get("learnerId")

	// Get all of the lessons
	sql := `SELECT lesson_id, title, description, video_link, scheduled_date, module from lesson 
					WHERE module = $1`

	result, err := db.Query(sql, module)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer result.Close()

	for result.Next() {
		var lesson lessonResponse
		if err := result.Scan(&lesson.Id, &lesson.Title, &lesson.Description, &lesson.VideoLink, &lesson.ScheduledDate, &lesson.Module); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create the learner_lesson
		sql = `INSERT INTO learner_lesson(learner, lesson, completed) VALUES ($1, $2, $3)`
		stmt, err := db.Prepare(sql)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = stmt.Exec(learnerId, lesson.Id, false)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func getLessonToday(w http.ResponseWriter, r *http.Request) {

	learnerId := r.Header.Get("X-User-Claim")

	// Get lesson id from query params
	// query := r.URL.Query()
	// lessonId := query.Get("id")
	timezone := "Asia/Singapore"

	var res lessonResponse

	// Get the current date application is date specific, regardless of timezone they should be shown some lesson at some local date
	// So take reference to a no timezone date value and compare to their timezone date
	local := time.Now().UTC()
	location, err := time.LoadLocation(timezone)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	local = local.In(location)

	// All learners have access to the module, because there is only one
	query := `SELECT lesson_id, title, description, video_link, scheduled_date, module from lesson WHERE scheduled_date = $1`

	// There should only be one lesson
	err = db.QueryRow(query, local.Format("2006-01-02")).Scan(&res.Id, &res.Title, &res.Description, &res.VideoLink, &res.ScheduledDate, &res.Module)
	res.Completed = false

	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNoContent)
			return
		} else {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	var completed bool

	// Check if lesson is completed
	query = `SELECT completed FROM learner_lesson WHERE lesson = $1 AND learner = $2`
	err = db.QueryRow(query, res.Id, learnerId).Scan(&completed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if completed == true {
		w.WriteHeader(http.StatusNoContent)
		return
	} else {
		dres, err := json.Marshal(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(dres)
	}
}

func getLessonsPast(w http.ResponseWriter, r *http.Request) {
	learnerId := r.Header.Get("X-User-Claim")

	var res []lessonResponse

	// Selecting all lessons that happened earlier than today
	sql := `SELECT lesson_id, title, description, video_link, scheduled_date, module, completed from lesson
					LEFT JOIN learner_lesson
					ON lesson.lesson_id = learner_lesson.lesson AND learner_lesson.learner = $1
					WHERE scheduled_date <= CURRENT_DATE`

	result, err := db.Query(sql, learnerId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer result.Close()

	for result.Next() {
		var lesson lessonResponse
		if err := result.Scan(&lesson.Id, &lesson.Title, &lesson.Description, &lesson.VideoLink, &lesson.ScheduledDate, &lesson.Module, &lesson.Completed); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res = append(res, lesson)
	}

	// Marshal to JSON and return
	dres, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(dres)

}

func completeLesson(w http.ResponseWriter, r *http.Request) {
	learnerId := r.Header.Get("X-User-Claim")

	// Get lesson id from query params
	query := r.URL.Query()
	lessonId := query.Get("id")

	if lessonId == "" {
		http.Error(w, "No valid lesson id provided", http.StatusBadRequest)
		return
	}

	// Get all the flashcards associated to this lesson
	var flashcardIds []string
	sql := `SELECT flashcard_id FROM flashcard WHERE lesson = $1`
	result, err := db.Query(sql, lessonId)
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
		_, err = stmt.Exec(learnerId, flashcardId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}

	// Update the learner_lesson data
	sql = `UPDATE learner_lesson SET completed = true WHERE learner_lesson.learner = $1 AND learner_lesson.lesson = $2`
	stmt, err = db.Prepare(sql)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = stmt.Exec(learnerId, lessonId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type flashcardResponse struct {
	Id         string `json:"id"`
	TopSide    string `json:"top_side"`
	BottomSide string `json:"bottom_side"`
	LessonId   string `json:"lesson_id"`
}

// You need to ensure that the lesson flashcards exist for the user
func getLessonFlashcards(w http.ResponseWriter, r *http.Request) {

	var res []flashcardResponse
	// Get lesson id from query params
	query := r.URL.Query()
	lessonId := query.Get("id")

	if lessonId == "" {
		http.Error(w, "No valid lesson id provided", http.StatusBadRequest)
		return
	}

	sql := `SELECT flashcard_id, top_side, bottom_side, lesson FROM flashcard WHERE lesson = $1`
	result, err := db.Query(sql, lessonId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer result.Close()

	for result.Next() {
		var card flashcardResponse
		if err := result.Scan(&card.Id, &card.TopSide, &card.BottomSide, &card.LessonId); err != nil {
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

	sql := `SELECT tutorial_id, title, description, scheduled_datetime, module FROM tutorial WHERE scheduled_datetime > NOW() LIMIT 5`
	result, err := db.Query(sql)
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
	learnerId := r.Header.Get("X-User-Claim")

	var res []flashcardResponse

	// Retrieve and check last completed
	var user userResponse

	sql := `SELECT learner_id, email, first_name, last_name, last_completed, streak FROM learner WHERE learner_id = $1`
	if err := db.QueryRow(sql, learnerId).Scan(&user.Id, &user.Email, &user.FirstName, &user.LastName, &user.LastCompleted, &user.Streak); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// start and end of day
	timezone := "Asia/Singapore"

	now := time.Now().UTC()
	location, err := time.LoadLocation(timezone)

	// Last completed should already be in UTC because of TimezoneTZ
	if err == nil {
		now = now.In(location)
		user.LastCompleted = user.LastCompleted.In(location)
	}

	d1 := time.Date(user.LastCompleted.Year(), user.LastCompleted.Month(), user.LastCompleted.Day(), 0, 0, 0, 0, user.LastCompleted.Location())
	d2 := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	fmt.Println(d1)
	fmt.Println(d2)

	if d1.Unix() == d2.Unix() {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// First retrieve all the repeat due ones
	sql = `SELECT flashcard_id, top_side, bottom_side, lesson FROM flashcard RIGHT JOIN learner_flashcard ON flashcard.flashcard_id = learner_flashcard.flashcard WHERE learner = $1 AND repeat > 0`

	result, err := db.Query(sql, learnerId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer result.Close()

	for result.Next() {
		var flashcard flashcardResponse
		if err := result.Scan(&flashcard.Id, &flashcard.TopSide, &flashcard.BottomSide, &flashcard.LessonId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		res = append(res, flashcard)
	}

	var allFlashcards []flashcardResponse
	sql = `SELECT flashcard_id, top_side, bottom_side, lesson FROM flashcard RIGHT JOIN learner_flashcard ON flashcard.flashcard_id = learner_flashcard.flashcard WHERE learner = $1 AND repeat = 0`
	result, err = db.Query(sql, learnerId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer result.Close()

	for result.Next() {
		var flashcard flashcardResponse
		if err := result.Scan(&flashcard.Id, &flashcard.TopSide, &flashcard.BottomSide, &flashcard.LessonId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		allFlashcards = append(allFlashcards, flashcard)
	}

	// Check and calculate how much randomness to retrieve
	remaining := 20 - len(res)
	total := len(allFlashcards)

	if remaining >= total {
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

	// Marshal to JSON and return
	dres, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(dres)

}

func passFlashcard(w http.ResponseWriter, r *http.Request) {
	learnerId := r.Header.Get("X-User-Claim")

	// Get lesson id from query params
	query := r.URL.Query()
	flashcardId := query.Get("id")

	var repeat int

	sql := `SELECT repeat FROM learner_flashcard WHERE learner = $1 AND flashcard = $2`
	if err := db.QueryRow(sql, learnerId, flashcardId).Scan(&repeat); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Decrement repeat count
	if repeat > 0 {
		repeat -= 1
	}

	sql = `UPDATE learner_flashcard SET repeat = $1 WHERE learner = $2 AND flashcard = $3`
	stmt, err := db.Prepare(sql)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = stmt.Exec(repeat, learnerId, flashcardId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func failFlashcard(w http.ResponseWriter, r *http.Request) {
	learnerId := r.Header.Get("X-User-Claim")

	// Get lesson id from query params
	query := r.URL.Query()
	flashcardId := query.Get("id")

	// set repeat to 3 no matter what
	repeat := 3

	sql := `UPDATE learner_flashcard SET repeat = $1 WHERE learner = $2 AND flashcard = $3`
	stmt, err := db.Prepare(sql)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = stmt.Exec(repeat, learnerId, flashcardId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func completeReview(w http.ResponseWriter, r *http.Request) {
	learnerId := r.Header.Get("X-User-Claim")

	// Retrieve the learner
	var user userResponse

	sql := `SELECT learner_id, email, first_name, last_name, streak FROM learner WHERE learner_id = $1`
	if err := db.QueryRow(sql, learnerId).Scan(&user.Id, &user.Email, &user.FirstName, &user.LastName, &user.Streak); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sql = `UPDATE learner SET streak = $1, last_completed = $2 WHERE learner_id = $3`
	stmt, err := db.Prepare(sql)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = stmt.Exec(user.Streak+1, time.Now().UTC(), learnerId)
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

		splitToken := strings.Split(reqToken, "Bearer")

		if len(splitToken) != 2 {
			http.Error(w, "Malformed format for auth token", http.StatusForbidden)
			return
		}

		reqToken = strings.TrimSpace(splitToken[1])

		parsedToken, err := jwt.Parse(reqToken, func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("Invalid Signing Type")
			}

			return []byte(JWT_SECRET), nil
		})

		// Invalid JWT secret error
		if err != nil {
			http.Error(w, "Authentication failed", http.StatusForbidden)
			return
		}

		// Parsing the claims in the JWT token
		if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
			// If the claims doesn't include the Id or the UserType, throw an error
			if claims["id"] == nil {
				http.Error(w, "Authentication claims failed", http.StatusForbidden)
				return
			}

			uid := claims["id"].(string)

			r.Header.Set("X-User-Claim", uid)

			next.ServeHTTP(w, r)
		} else {
			http.Error(w, "Auth token invalid", http.StatusForbidden)
			return
		}
	})
}

/********* UTILITIES **************/
func createJWT(uid string, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": uid,
	})
	tokenString, err := token.SignedString([]byte(secret))

	return tokenString, err
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

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
