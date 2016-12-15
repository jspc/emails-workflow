package main

import (
    "net/http"

    "github.com/gincorp/gin/taskmanager"
    owm "github.com/briandowns/openweathermap"
)

func getLondonWeather(jn taskmanager.JobNotification) (output map[string]interface{}, err error) {
    w, err := owm.NewForecast("C", "en")
    if err != nil {
        return
    }

    w.DailyByName("London,UK",1)
    day := w.List[0]

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
