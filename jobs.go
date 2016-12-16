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
	// Hat tip: https://gist.github.com/andelf/5004821
	username := os.Getenv("MAIL_USERNAME")
	password := os.Getenv("MAIL_PASSWORD")
	mailhost := jn.Context["host"]
	mailport := jn.Context["port"]

	auth := smtp.PlainAuth("",
		username,
		password,
		mailhost,
	)

	from := mail.Address{"gin", jn.Context["from"]}
	to := mail.Address{"", jn.Context["to"]}

	header := make(map[string]string)

	header["From"] = from.String()
	header["To"] = to.String()
	header["Subject"] = jn.Context["subject"]
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(jn.Context["body"]))

	err = smtp.SendMail(
		mailhost+":"+mailport,
		auth,
		from.Address,
		[]string{to.Address},
		[]byte(message),
	)

	output = make(map[string]interface{})
	return
}
