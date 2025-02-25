package main

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"os"
	"reward-service/data"
	"strconv"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	Password  string    `json:"-"`
	Active    int       `json:"active"`
	Score     int       `json:"score"`
	Referrer  string    `json:"referrer,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type contextKey string

const userIDKey contextKey = "userID"

// getIDFromRequest gets id from the URL
func (app *Config) getIDFromRequest(w http.ResponseWriter, r *http.Request) (int, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		app.errorJSON(w, errors.New("couldn't convert id string to int"), http.StatusBadRequest)
		return 0, err
	}
	return id, nil
}

// Registrate insert new user to the database
func (app *Config) Registrate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Password  string `json:"password"`
		Active    int    `json:"active,omitempty"`
		Score     int    `json:"score,omitempty"`
		Referrer  string `json:"referrer,omitempty"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	if len(requestPayload.Password) < 8 {
		app.errorJSON(w, errors.New("password must be at least 8 characters long"), http.StatusBadRequest)
		return
	}
	user := User{
		Email:     requestPayload.Email,
		FirstName: requestPayload.FirstName,
		LastName:  requestPayload.LastName,
		Password:  requestPayload.Password,
		Active:    requestPayload.Active,
		Score:     requestPayload.Score,
		Referrer:  requestPayload.Referrer,
	}
	id, err := app.Repo.Insert(data.User(user))
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Succesfully created new user, id: %d", id),
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

// GetLeaderboard retrieves all users from the database, sort them by points
func (app *Config) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	users, err := app.Repo.GetAll()
	if err != nil {
		app.errorJSON(w, errors.New("couldn't fetch All users"), http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Fetched all users"),
		Data:    users,
	}

	app.writeJSON(w, http.StatusAccepted, payload)

}

// Authenticate authenticates user by provided email and password, provides tokens to access
func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	user, err := app.Repo.EmailCheck(requestPayload.Email)
	if err != nil {
		app.errorJSON(w, errors.New("user with this email does not exist"), http.StatusBadRequest)
		return
	}

	valid, err := app.Repo.PasswordMatches(requestPayload.Password, *user)
	if err != nil || !valid {
		app.errorJSON(w, errors.New("invalid password"), http.StatusBadRequest)
		return
	}

	secretKey := os.Getenv("SECRET_KEY")
	userData, err := generateTokens(user.ID, secretKey)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    userData.AccessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(15 * time.Minute),
	})
	err = validateRefreshToken(userData.HashedRefreshToken, userData.RefreshToken)
	if err != nil {
		app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}
	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Welcome back, %s!", user.FirstName),
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

// someTask some blank task
func (app *Config) someTask(w http.ResponseWriter, r *http.Request) {
	app.completeTask(w, r, 100)
}

// completeTask completes various task and adding some point to the user
func (app *Config) completeTask(w http.ResponseWriter, r *http.Request, points int) {
	id, err := app.getIDFromRequest(w, r)
	if err != nil {
		return
	}
	err = app.Repo.AddPoints(id, points)
	if err != nil {
		app.errorJSON(w, errors.New("couldn't add points to the user"), http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("complete task worked for user with id %d, added points %d", id, points),
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

// completeTelegramSign completes telegram sign to add points to the user
func (app *Config) completeTelegramSign(w http.ResponseWriter, r *http.Request) {
	app.completeTask(w, r, 50)
}

// completeTelegramSign completes X sign to add points to the user
func (app *Config) completeXSign(w http.ResponseWriter, r *http.Request) {
	app.completeTask(w, r, 75)
}

// Kuarhodron special task to add 10k points
func (app *Config) Kuarhodron(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		SecretWaterPassword string `json:"water_password"`
	}
	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	if requestPayload.SecretWaterPassword == "KUARHODRON" {
		app.completeTask(w, r, 10000)
	}
}

// retrieveOne retrieves one user from the database by id
func (app *Config) retrieveOne(w http.ResponseWriter, r *http.Request) {

	id, err := app.getIDFromRequest(w, r)
	if err != nil {
		return
	}
	user, err := app.Repo.GetOne(id)
	if err != nil {
		app.errorJSON(w, errors.New("couldn't fetch user"), http.StatusBadRequest)
		return
	}

	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Retrieved one user from the database"),
		Data:    user,
	}

	app.writeJSON(w, http.StatusAccepted, payload)

}

// redeemReferrer redeems referrer for the owner of the referrer and for the user, who used it base on id and referrer
func (app *Config) redeemReferrer(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Referrer string `json:"referrer"`
	}
	id, err := app.getIDFromRequest(w, r)
	if err != nil {
		return
	}
	err = app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	err = app.Repo.RedeemReferrer(id, requestPayload.Referrer)
	if err != nil {
		app.errorJSON(w, errors.New("couldn't redeem referrer"), http.StatusBadRequest)
		return
	}
	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Referrer redeemed"),
	}

	app.writeJSON(w, http.StatusAccepted, payload)

}

// DeleteUser delets user from the DB
func (app *Config) DeleteUser(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		id int `json:"id"`
	}
	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	err = app.Repo.DeleteByID(requestPayload.id)
	if err != nil {
		app.errorJSON(w, errors.New("couldn't delete user"), http.StatusBadRequest)
		return
	}
	payload := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("User deleted successfully"),
	}

	app.writeJSON(w, http.StatusAccepted, payload)

}
