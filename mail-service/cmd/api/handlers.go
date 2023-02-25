package main

import "net/http"

func (app *Config) sendMail(w http.ResponseWriter, r *http.Request) {
	type mailMessage struct {
		To      string `json:"to"`
		From    string `json:"from"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}
	var requestPayload mailMessage

	err := app.readJson(w, r, &requestPayload)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	msg := Message{
		To:      requestPayload.To,
		From:    requestPayload.From,
		Subject: requestPayload.Subject,
		Data:    requestPayload.Message,
	}
	err = app.Mailer.SendSMTPMessage(msg)
	if err != nil {
		app.errorJson(w, err)
		return
	}

	payload := jsonReponse{
		Error:   false,
		Message: "Mail sent successfully to " + requestPayload.To,
	}

	err = app.writeJson(w, http.StatusAccepted, payload)

}
