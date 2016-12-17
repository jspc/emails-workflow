// Copyright 2016, gincorp.
//
// This project is under the MIT licence;
// found in LICENCE.md in the parent directory.
//
// This script is expected to be run via go run wf.go
// It configures a simple workflow designed to be run with the
// job manager implemenetation in the parent directory.
//
//
// This workflow will compile certain interesting bits of info:
// weather forecasts, news headlines, exchange rates. It'll then
// email this information to a comfigured email address.
//
// ENV Vars:
// SENDER_ADDRESS=''     - the sender address on an email
// RECEIPIENT_ADDRESS='' - the receipient of the email
// GUARDIAN_API_KEY=''   - the API key to be used against the Guardian api

package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/gincorp/gin/datastore"
	"github.com/gincorp/gin/workflow"
)

var (
	redisURI *string
)

func init() {
	redisURI = flag.String("redis", "redis://localhost:6379/0", "URI of redis node")
	flag.Parse()
}

func main() {
	d, err := datastore.NewDatastore(*redisURI)
	if err != nil {
		panic(err)
	}

	log.Print(d.SaveWorkflow(sendDailyEmail(), true))
}

func sendDailyEmail() workflow.Workflow {
	// Grab some information about some stuff, email it

	vars := make(map[string]string)
	vars["location"] = "London,uk"
	vars["mail_host"] = "smtp.gmail.com"
	vars["mail_port"] = "587"
	vars["mail_from"] = os.Getenv("SENDER_ADDRESS")
	vars["mail_to"] = os.Getenv("RECEIPIENT_ADDRESS")
	vars["guardian_key"] = os.Getenv("GUARDIAN_API_KEY")

	emailBody := []string{
		"Greetings,",
		"",
		"Weather today in {{.Defaults.location}}:",
		"Minimum: {{.weather.minimum}} celsius",
		"Maximum: {{.weather.maximum}} celsius",
		"",
		"In the news:",
		"{{ range .news.articles }}",
		"{{ .Timestamp }}: {{ .Title }} - {{ .URL }}",
		"{{ end }}",
		"",
		"Regards,",
		"gin",
	}

	return workflow.Workflow{
		Name:      "Send daily email",
		Variables: vars,
		Steps: []workflow.Step{
			workflow.Step{
				Name: "Get Weather",
				Type: "get-day-temperature",
				Context: map[string]string{
					"location": "{{.Defaults.location}}",
				},
				Register: "weather",
			},
			workflow.Step{
				Name: "Log Minimum",
				Type: "log",
				Context: map[string]string{
					"message": "Minimum Temperature: {{ .weather.minimum }}",
				},
			},
			workflow.Step{
				Name: "Log Maximum",
				Type: "log",
				Context: map[string]string{
					"message": "Maximum Temperature: {{ .weather.maximum }}",
				},
			},
			workflow.Step{
				Name: "Get Today's News",
				Type: "get-news",
				Context: map[string]string{
					"apiKey": "{{.Defaults.guardian_key}}",
				},
				Register: "news",
			},
			workflow.Step{
				Name: "Send Email",
				Type: "send-email",
				Context: map[string]string{
					"host": "{{ .Defaults.mail_host }}",
					"port": "{{ .Defaults.mail_port }}",

					"from":    "{{ .Defaults.mail_from }}",
					"to":      "{{ .Defaults.mail_to }}",
					"subject": "Daily Update Email",
					"body":    strings.Join(emailBody[:], "\r\n"),
				},
			},
		},
	}
}
