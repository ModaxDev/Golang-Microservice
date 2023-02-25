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

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
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
	payload := jsonReponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJson(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmissions(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload

	err := app.readJson(w, r, &requestPayload)
	if err != nil {
		app.errorJson(w, err, http.StatusBadRequest)
		return
	}

	switch requestPayload.Action {
	case "auth":
		app.authenticate(w, requestPayload.Auth)
	case "log":
		app.logItem(w, requestPayload.Log)
	case "mail":
		app.sendMail(w, requestPayload.Mail)
	default:
		app.errorJson(w, errors.New("invalid action"), http.StatusBadRequest)
	}
}

func (app *Config) logItem(w http.ResponseWriter, entry LogPayload) {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceUrl := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJson(w, err, http.StatusInternalServerError)
		return
	}

	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJson(w, err, http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJson(w, err)
		return
	}

	var payload jsonReponse
	payload.Error = false
	payload.Message = "Logged"

	app.writeJson(w, http.StatusAccepted, payload)
}

func (app *Config) authenticate(w http.ResponseWriter, a AuthPayload) {
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJson(w, err, http.StatusInternalServerError)
		return
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJson(w, err, http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUnauthorized {
		app.errorJson(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errorJson(w, errors.New("error calling authentication service"+err.Error()))
		return
	}

	var jsonFromService jsonReponse
	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.errorJson(w, err, http.StatusInternalServerError)
		return
	}

	if jsonFromService.Error {
		app.errorJson(w, err, http.StatusUnauthorized)
		return
	}

	var payload jsonReponse
	payload.Error = false
	payload.Message = "Authenticated"
	payload.Data = jsonFromService.Data

	app.writeJson(w, http.StatusAccepted, payload)
}

func (app *Config) sendMail(w http.ResponseWriter, msg MailPayload) {
	jsonData, _ := json.MarshalIndent(msg, "", "\t")

	request, err := http.NewRequest("POST", "http://mailer-service/send", bytes.NewBuffer(jsonData))
	if err != nil {
		app.errorJson(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errorJson(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errorJson(w, err)
		return
	}

	var payload jsonReponse
	payload.Error = false
	payload.Message = "Mail sent successfully to " + msg.To

	app.writeJson(w, http.StatusAccepted, payload)
}
