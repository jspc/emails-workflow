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

	log.Print(d.SaveWorkflow(postToEchoServer(), true))
	log.Print(d.SaveWorkflow(postToEchoServerAndLog(), true))
}

func postToEchoServer() workflow.Workflow {
	// Make a request against https://hub.docker.com/r/jspc/go-echo-http-server/

	vars := make(map[string]string)

	// hyphens in variable keys messes up text/template, which parses
	// context values below
	vars["echo_url"] = "http://localhost:8000/some-endpoint"
	vars["content_type"] = "application/json"

	return workflow.Workflow{
		Name:      "Post to Echo Server",
		Variables: vars,
		Steps: []workflow.Step{
			workflow.Step{
				Name: "Make Request",
				Type: "post-to-web",
				Context: map[string]string{
					"url":          "{{.Defaults.echo_url}}",
					"content-type": "{{.Defaults.content_type}}",
				},
				Register: "echo_data"},
		},
	}
}

func postToEchoServerAndLog() workflow.Workflow {
	// Make a request against https://hub.docker.com/r/jspc/go-echo-http-server/
	// Log the output

	vars := make(map[string]string)

	vars["echo_url"] = "http://localhost:8000/some-endpoint"
	vars["content_type"] = "application/json"

	return workflow.Workflow{
		Name:      "Post and Log Echo Server",
		Variables: vars,
		Steps: []workflow.Step{
			workflow.Step{
				Name: "Make Request",
				Type: "post-to-web",
				Context: map[string]string{
					"url":          "{{.Defaults.echo_url}}",
					"content-type": "{{.Defaults.content_type}}",
				},
				Register: "echo_data"},
			workflow.Step{
				Name: "Log Output",
				Type: "log",
				Context: map[string]string{
					// Note: there are workarounds for hyphenated keys, as below
					"message": "user agent set to: {{index .echo_data.Headers \"User-Agent\"}}",
				},
			},
		},
	}
}
