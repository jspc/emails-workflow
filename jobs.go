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

type financeitem struct {
	Name, Price string
}

type newsitem struct {
	Timestamp time.Time
	Title     string
	URL       string
}

type ByTime []newsitem

func (t ByTime) Len() int           { return len(t) }
func (t ByTime) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t ByTime) Less(i, j int) bool { return !t[i].Timestamp.Before(t[j].Timestamp) } // Actual do the opposite of Less to order by descending

func apiCall(url string) (d map[string]interface{}, err error) {
	var resp *http.Response

	if resp, err = http.Get(url); err != nil {
		return
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	defer resp.Body.Close()

	d = make(map[string]interface{})
	json.Unmarshal(buf.Bytes(), &d)

	return
}

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

	apiKey := jn.Context["apiKey"]

	newsItems := []newsitem{}

	for _, section := range sections {
		// get stuff
		var data map[string]interface{}
		if data, err = apiCall(guardianURL(apiKey, date, section)); err != nil {
			return
		}

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
	if len(newsItems) > 10 {
		output["articles"] = newsItems[:10]
	} else {
		output["articles"] = newsItems
	}

	return
}

func getCurrencyPrices(jn taskmanager.JobNotification) (output map[string]interface{}, err error) {
	respData, err := apiCall("http://finance.yahoo.com/webservice/v1/symbols/allcurrencies/quote?format=json")
	if err != nil {
		return
	}

	output = make(map[string]interface{})
	prices := []financeitem{}

	// Somewhere, and I suspect in java world, there's a fucking awful library turning xml into json
	// by trying to map xml classes/types to a json object. It means stupid shit like this format.
	//
	// Seriously: fuck anybody who does this.

	data := respData["list"].(map[string]interface{}) // ITS THE ONLY FUCKING KEY IN THE FUCKING OBJECT

	for _, resource := range data["resources"].([]interface{}) {
		actualResource := resource.(map[string]interface{})["resource"].(map[string]interface{}) // WANKERS
		fields := actualResource["fields"].(map[string]interface{})

		switch fields["name"].(string) {
		case "USD/GBP":
			prices = append(prices, financeitem{"$/£", fields["price"].(string)})
		case "USD/EUR":
			prices = append(prices, financeitem{"$/€", fields["price"].(string)})
		case "GOLD 1 OZ":
			prices = append(prices, financeitem{"Gold per ounce", fields["price"].(string)})
		}
	}
	output["prices"] = prices

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
