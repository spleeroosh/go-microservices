package main

import (
	"authentication/data"
	"bytes"
	"encoding/json"
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

	ok, err := u.PasswordMatches(requestPayload.Password)
	if err != nil || !ok {
		log.Println("Authentication failed for user:", requestPayload.Email)
		_ = app.errorJSON(w, errors.New("invalid email or password"), http.StatusBadRequest)
		return
	}

	app.log(u)

	response := jsonRespone{
		Error:   false,
		Message: fmt.Sprintf("User %v authenticated successfully", requestPayload.Email),
		Data:    u,
	}
	_ = app.writeJSON(w, http.StatusAccepted, response)
}

func (app *Config) log(u *data.User) {
	logUrl := "http://logger-service/log"
	var payload struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}
	payload.Name = "authentication"
	payload.Data = u.Email

	jsonPayload, err := json.MarshalIndent(payload, "", "\t")
	if err != nil {
		log.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	request, err := http.NewRequest("POST", logUrl, bytes.NewBuffer([]byte(jsonPayload)))
	if err != nil {
		log.Printf("Error creating request: %v\n", err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		log.Printf("Error sending request: %v\n", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		log.Printf("Unexpected status code: %d\n", response.StatusCode)
	}
}
