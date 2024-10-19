package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonRespone{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) Handle(w http.ResponseWriter, r *http.Request) {
	var payload RequestPayload

	err := app.readJSON(w, r, &payload)
	if err != nil {
		_ = app.errorJSON(w, err)
		return
	}

	switch payload.Action {
	case "auth":
		app.authenticate(w, payload.Auth)
	case "log":
		app.log(w, payload.Log)
	default:
		_ = app.errorJSON(w, errors.New("Invalid action"))
	}
}

func (app *Config) log(w http.ResponseWriter, l LogPayload) {
	jsonData, err := json.MarshalIndent(l, "", "\t")
	if err != nil {
		_ = app.errorJSON(w, err)
		return
	}
	logUrl := "http://logger-service/log"
	request, err := http.NewRequest("POST", logUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		_ = app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		_ = app.errorJSON(w, err)
		return
	}

	if response.StatusCode != http.StatusAccepted {
		_ = app.errorJSON(w, errors.New("Error calling log service"))
	}

	var payload jsonRespone
	payload.Error = false
	payload.Message = "logged"

	_ = app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	jsonData, err := json.MarshalIndent(a, "", "\t")
	if err != nil {
		_ = app.errorJSON(w, err)
		return
	}

	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		_ = app.errorJSON(w, err)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		_ = app.errorJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		_ = app.errorJSON(w, errors.New("Invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		_ = app.errorJSON(w, errors.New("Error calling auth service"))
		return
	}

	var jsonFromService jsonRespone

	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		_ = app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		_ = app.errorJSON(w, err, http.StatusUnauthorized)
		return
	}

	var payload jsonRespone
	payload.Error = false
	payload.Message = "Authenticated"
	payload.Data = jsonFromService.Data

	_ = app.writeJSON(w, http.StatusAccepted, payload)
}
