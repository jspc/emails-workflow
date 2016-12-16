package main

import (
	"encoding/base64"
	"fmt"
	"net/mail"
	"net/smtp"
	"os"

	owm "github.com/briandowns/openweathermap"
	"github.com/gincorp/gin/taskmanager"
)

func getLondonWeather(jn taskmanager.JobNotification) (output map[string]interface{}, err error) {
	w, err := owm.NewForecast("C", "en")
	if err != nil {
		return
	}

	location := jn.Context["location"]
	w.DailyByName(location, 1)
	day := w.List[0]

	output = make(map[string]interface{})
	output["minimum"] = day.Temp.Min
	output["maximum"] = day.Temp.Max

	return
}

func sendEmail(jn taskmanager.JobNotification) (output map[string]interface{}, err error) {
    username := context["username"]
    password := context["password"]
    mailhost := context["mailhost"]

    // send email
}
