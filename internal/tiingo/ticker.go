package tiingo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var baseURL = "https://api.tiingo.com"

type PriceInfo struct {
	Date      time.Time `json:"date"`
	Open      float32   `json:"open"`
	Close     float32   `json:"close"`
	High      float32   `json:"high"`
	Low       float32   `json:"low"`
	Volume    int64     `json:"volume"`
	AdjOpen   float32   `json:"adjOpen"`
	AdjClose  float32   `json:"adjClose"`
	AdjHigh   float32   `json:"adjHigh"`
	AdjLow    float32   `json:"adjLow"`
	AdjVolume int64     `json:"adjVolume"`
	Dividend  float32   `json:"divCash"`
	Split     float32   `json:"splitFactor"`
}

func GetTickerInfo(token string, ticker string) ([]PriceInfo, error) {
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	resp, err := fetch(headers, fmt.Sprintf("%v/tiingo/daily/%v/prices?token=%v", baseURL, ticker, token))
	if err != nil {
		return []PriceInfo{}, err
	}

	var priceInfo []PriceInfo
	err = json.Unmarshal(resp, &priceInfo)
	if err != nil {
		return []PriceInfo{}, err
	}

	return priceInfo, nil
}

func fetch(headers map[string]string, url string) ([]byte, error) {
	req, err := http.NewRequest(
		"GET",
		url,
		bytes.NewBuffer([]byte("")),
	)
	if err != nil {
		log.Fatalf("building request: %v", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("making request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed with statusCode: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("reading body: %v", err)
	}

	return body, nil
}
