package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
)

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		_ = app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	if requestPayload.Email == "" || requestPayload.Password == "" {
		_ = app.errorJSON(w, errors.New("email and password are required"), http.StatusBadRequest)
		return
	}

	u, err := app.Models.User.GetByEmail(requestPayload.Email)
	if err != nil {
		_ = app.errorJSON(w, errors.New("invalid email or password"), http.StatusBadRequest)
		return
	}

	ok, err := app.Models.User.PasswordMatches(u.Password)
	if err != nil || !ok {
		log.Println("Authentication failed for user:", requestPayload.Email)
		_ = app.errorJSON(w, errors.New("invalid email or password"), http.StatusBadRequest)
		return
	}

	response := jsonRespone{
		Error:   false,
		Message: fmt.Sprintf("User %v authenticated successfully", requestPayload.Email),
		Data:    u,
	}
	_ = app.writeJSON(w, http.StatusOK, response)
}
