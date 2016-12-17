package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
	"sort"
	"time"

	owm "github.com/briandowns/openweathermap"
	"github.com/gincorp/gin/taskmanager"
)

type newsitem struct {
	Timestamp time.Time
	Title     string
	URL       string
}

type ByTime []newsitem

func (t ByTime) Len() int           { return len(t) }
func (t ByTime) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t ByTime) Less(i, j int) bool { return !t[i].Timestamp.Before(t[j].Timestamp) } // Actual do the opposite of Less to order by descending

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

func getNews(jn taskmanager.JobNotification) (output map[string]interface{}, err error) {
	// I don't have any non-hacky access to what one of these objects definitively looks like
	// so I'm using interfaces and type assertions. One day I'll get around to fixing this.
	//
	// Until then, then, it'll remain incredibly ugly and a tad inefficient.

	date := time.Now().Format("2006-01-02")
	sections := []string{
		"football",
		"law",
		"media",
		"money",
		"politics",
		"society",
		"world",
	}

	var resp *http.Response

	newsItems := []newsitem{}

	apiKey := jn.Context["apiKey"]
	for _, section := range sections {
		// get stuff
		if resp, err = http.Get(guardianURL(apiKey, date, section)); err != nil {
			return
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		defer resp.Body.Close()

		data := make(map[string]interface{})
		json.Unmarshal(buf.Bytes(), &data)

		response := data["response"].(map[string]interface{})

		switch response["results"].(type) {
		case []interface{}:
			for _, item := range response["results"].([]interface{}) {
				itemMap := item.(map[string]interface{})

				timestamp, _ := time.Parse("2006-01-02T15:04:05Z", itemMap["webPublicationDate"].(string))
				title := itemMap["webTitle"].(string)
				url := itemMap["fields"].(map[string]interface{})["shortUrl"].(string)

				newsItems = append(newsItems, newsitem{
					Timestamp: timestamp,
					Title:     title,
					URL:       url,
				})
			}
		}
	}

	sort.Sort(ByTime(newsItems))

	output = make(map[string]interface{})
	output["articles"] = newsItems[:10]

	return
}

func guardianURL(apiKey, date, section string) string {
	return fmt.Sprintf("https://content.guardianapis.com/search?api-key=%s&show-fields=short-url&from-date=%s&page-size=50&production-office=uk&section=%s",
		apiKey,
		date,
		section,
	)
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
