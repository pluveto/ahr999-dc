package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env")
	token := os.Getenv("INFLUXDB_TOKEN")
	if token == "" {
		log.Fatal("INFLUXDB_TOKEN environment variable not set")
	}

	url := os.Getenv("INFLUXDB_URL")
	if url == "" {
		url = "http://localhost:8086"
	}

	client := influxdb2.NewClient(url, token)

	org := os.Getenv("INFLUXDB_ORG")
	if org == "" {
		org = "MyOrg"
	}

	bucket := os.Getenv("INFLUXDB_BUCKET")
	if bucket == "" {
		bucket = "MyBucket"
	}

	interval := 8 // hours
	if os.Getenv("INTERVAL") != "" {
		interval_, err := strconv.Atoi(os.Getenv("INTERVAL"))
		if err != nil {
			log.Fatal("Invalid INTERVAL value:", os.Getenv("INTERVAL"))
		} else {
			interval = interval_
		}
	}

	log.Println("Configured with INFLUXDB_URL:", url, "INTERVAL:", interval, "INFLUXDB_ORG:", org, "INFLUXDB_BUCKET:", bucket)

	writeAPI := client.WriteAPIBlocking(org, bucket)

	var lastTimestamp int64 = 0

	for {
		data, err := fetchData()
		if err != nil {
			log.Println("Failed to fetch data:", err)
			continue
		}

		if data.Code != 200 || !data.Success {
			log.Println("Failed to fetch data:", data.Message)
			continue
		}

		if len(data.Data) == 0 {
			log.Println("No data received")
			continue
		}

		if data.Data[0].Date == lastTimestamp {
			log.Println("Data timestamp is the same as last fetched, skipping...")
			continue
		}

		ahr999 := data.Data[0].Ahr999
		avg := data.Data[0].Avg
		value, err := strconv.ParseFloat(data.Data[0].Value, 64)
		if err != nil {
			log.Println("Failed to parse value:", err)
			continue
		}

		// Convert date from milliseconds to time.Time
		timestamp := time.Unix(0, data.Data[0].Date*int64(time.Millisecond))

		point := influxdb2.NewPointWithMeasurement("ahr999").
			AddField("ahr999", ahr999).
			AddField("avg", avg).
			AddField("value", value).
			SetTime(timestamp)

		if err := writeAPI.WritePoint(context.Background(), point); err != nil {
			log.Fatal("Failed to write data point:", err)
		}

		lastTimestamp = data.Data[0].Date // Update last timestamp

		log.Printf("Data written for timestamp: %v", timestamp)

		if interval == 0 {
			break
		}

		time.Sleep(time.Duration(interval) * time.Hour)
	}
}

type APIResponse struct {
	Code    int    `json:"code"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    []struct {
		Ahr999      float64 `json:"ahr999"`
		AhrChange   float64 `json:"ahrChange"`
		Avg         float64 `json:"avg"`
		AvgChange   float64 `json:"avgChange"`
		Value       string  `json:"value"`
		ValueChange float64 `json:"valueChange"`
		Date        int64   `json:"date"`
	} `json:"data"`
}

func fetchData() (*APIResponse, error) {
	resp, err := http.Get("https://coinank.com/indicatorapi/getAhr999Table")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &data, nil
}
