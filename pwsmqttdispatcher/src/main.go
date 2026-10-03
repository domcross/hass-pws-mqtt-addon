package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xhhuango/json"

	"github.com/PuerkitoBio/goquery"
	"github.com/cdzombak/libwx"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type WeatherData struct {
	ReceiverTime         string  `json:"receiverTime"`
	ReceiverTimestamp    int64   `json:"receiverTimestamp"`
	TemperatureIndoor    float64 `json:"temperatureIndoor"`
	HumidityIndoor       float64 `json:"humidityIndoor"`
	PressureAbsolute     float64 `json:"pressureAbsolute"`
	PressureRelative     float64 `json:"pressureRelative"`
	Temperature          float64 `json:"temperature"`
	Humidity             float64 `json:"humidity"`
	DewPoint             float64 `json:"dewPoint"`
	WindDir              float64 `json:"windDir"`
	WindDirCardinal      string  `json:"windDirCardinal"`
	WindSpeed            float64 `json:"windSpeed"`
	WindGust             float64 `json:"windGust"`
	WindChill            float64 `json:"windChill"`
	SolarRadiation       float64 `json:"solarRadiation"`
	Uv                   float64 `json:"uv"`
	Uvi                  float64 `json:"uvi"`
	PrecipHourlyRate     float64 `json:"precipHourlyRate"`
	PrecipDaily          float64 `json:"precipDaily"`
	PrecipWeekly         float64 `json:"precipWeekly"`
	PrecipMonthly        float64 `json:"precipMonthly"`
	PrecipYearly         float64 `json:"precipYearly"`
	HeatIndex            float64 `json:"heatIndex"`
	IndoorSensorId       string  `json:"indoorSensorId"`
	OutdoorSensorId      string  `json:"outdoorSensorId"`
	IndoorSensorBattery  string  `json:"indoorSensorBattery"`
	OutdoorSensorBattery string  `json:"outdoorSensorBattery"`
}

var pwsIp string
var fetchInterval int
var debugEnabled bool

func fetchDocumentFromPws() *goquery.Document {
	pwsUrl := fmt.Sprintf("http://%s/livedata.htm", pwsIp)

	// Read the HTML file
	resp, err := http.Get(pwsUrl)
	if err != nil {
		log.Printf("Failed to connect to PWS: %s\n", err)
		return nil
	}
	defer resp.Body.Close()

	htmlData, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to fetch data from PWS: %s\n", err)
		return nil
	}

	// Parse the HTML file with goquery
	doc, err := goquery.NewDocumentFromReader(
		strings.NewReader(string(htmlData)),
	)
	if err != nil {
		log.Printf("Failed to process HTML: %s\n", err)
		return nil
	}

	if debugEnabled {
		log.Printf("Fetched data from PWS\n")
	}

	return doc
}

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return math.NaN()
	}

	return f
}

func parseHtml(doc *goquery.Document) WeatherData {
	var weatherData WeatherData

	doc.Find("table").Each(func(i int, s *goquery.Selection) {
		rows := s.Find("tr")

		// Parse the table rows and extract the data
		rows.Each(func(i int, s *goquery.Selection) {
			inputs := s.Find("input")
			value := inputs.AttrOr("value", "")

			switch i {
			case 9:
				weatherData.IndoorSensorId = value
				weatherData.IndoorSensorBattery =
					inputs.Eq(1).AttrOr("value", "")

			case 10:
				weatherData.OutdoorSensorId = value
				weatherData.OutdoorSensorBattery =
					inputs.Eq(1).AttrOr("value", "")

			case 12:
				weatherData.TemperatureIndoor = parseFloat(value)

			case 13:
				weatherData.HumidityIndoor = parseFloat(value)

			case 14:
				weatherData.PressureAbsolute = parseFloat(value)

			case 15:
				weatherData.PressureRelative = parseFloat(value)

			case 16:
				weatherData.Temperature = parseFloat(value)

			case 17:
				weatherData.Humidity = parseFloat(value)

			case 18:
				weatherData.WindDir = parseFloat(value)

			case 19:
				weatherData.WindSpeed = parseFloat(value)

			case 20:
				weatherData.WindGust = parseFloat(value)

			case 21:
				weatherData.SolarRadiation = parseFloat(value)

			case 22:
				weatherData.Uv = parseFloat(value)

			case 23:
				weatherData.Uvi = parseFloat(value)

			case 24:
				weatherData.PrecipHourlyRate = parseFloat(value)

			case 25:
				weatherData.PrecipDaily = parseFloat(value)

			case 26:
				weatherData.PrecipWeekly = parseFloat(value)

			case 27:
				weatherData.PrecipMonthly = parseFloat(value)

			case 28:
				weatherData.PrecipYearly = parseFloat(value)

			default:
				return
			}
		})
	})

	return weatherData
}

func weatherDataAsJson(wd WeatherData) []byte {
	// Convert the variables to JSON
	jsonData, err := json.Marshal(wd)
	if err != nil {
		log.Printf("Unable to marshal JSON: %s\n", err)
		return nil
	}

	return jsonData
}

func windDirToCardinal(windDirDeg int) string {
	dir := []string{
		"N ⬇️",
		"NNE ⬇️",
		"NE ↙️",
		"ENE ⬅️",
		"E ⬅️",
		"ESE ⬅️",
		"SE ↖️",
		"SSE ⬆️",
		"S ⬆️",
		"SSW ⬆️",
		"SW ↗️",
		"WSW ➡️
