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
	Mail   MailPayload `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
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
	case "mail":
		app.sendMail(w, payload.Mail)
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
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		_ = app.errorJSON(w, errors.New("Error calling log service"))
		return
	}

	app.sendResponseFromService(w, response, "Logged")
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

	app.sendResponseFromService(w, response, "Authenticated")
}

func (app *Config) sendMail(w http.ResponseWriter, m MailPayload) {
	jsonData, err := json.MarshalIndent(m, "", "\t")
	if err != nil {
		_ = app.errorJSON(w, err)
		return
	}
	logUrl := "http://mailer-service/send"
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
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		_ = app.errorJSON(w, errors.New("Error calling mail service"))
		return
	}

	app.sendResponseFromService(w, response, "Message is sent")
}

func (app *Config) sendResponseFromService(w http.ResponseWriter, r *http.Response, message string) {
	var jsonFromService jsonRespone

	err := json.NewDecoder(r.Body).Decode(&jsonFromService)
	if err != nil {
		_ = app.errorJSON(w, err)
		return
	}

	if jsonFromService.Error {
		_ = app.errorJSON(w, err, http.StatusInternalServerError)
		return
	}

	var payload jsonRespone
	payload.Error = false
	payload.Message = message
	payload.Data = jsonFromService.Data

	_ = app.writeJSON(w, http.StatusAccepted, payload)
}
