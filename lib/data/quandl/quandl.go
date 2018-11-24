package quandl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/MitchK/autorobin/lib/model"
)

const (
	// UserAgent UserAgent
	UserAgent = "autorobin"

	// DateFormat DateFormat
	DateFormat = "2006-01-02"
)

// Adapter Adapter
type Adapter struct {
	apiKey string
	client http.Client
}

// NewAdapter NewAdapter
func NewAdapter(apiKey string) *Adapter {
	return &Adapter{
		apiKey: apiKey,
		client: http.Client{},
	}
}

func convert(datasetData datasetData, asset model.Asset) []model.Quote {
	quotes := make([]model.Quote, len(datasetData.Data))
	indices := make(map[string]int, len(datasetData.ColumnNames))
	for i, c := range datasetData.ColumnNames {
		indices[c] = i
	}
	for i, d := range datasetData.Data {
		quotes[i].Asset = asset
		quotes[i].Price = d[indices["Close"]].(float64)
	}
	return quotes
}

// GetDailyAsc GetDailyAsc
func (adapter *Adapter) GetDailyAsc(from, to time.Time, assets ...model.Asset) ([][]model.Quote, error) {

	quotes := make([][]model.Quote, len(assets))
	for i, asset := range assets {
		url := fmt.Sprintf("https://www.quandl.com/api/v3/datasets/WIKI/%s/data.json", asset.Symbol)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		q := req.URL.Query()
		q.Add("api_key", adapter.apiKey)
		q.Add("order", "asc")
		q.Add("start_date", from.Format(DateFormat))
		q.Add("end_date", to.Format(DateFormat))
		req.URL.RawQuery = q.Encode()
		req.Header.Set("User-Agent", UserAgent)
		res, err := adapter.client.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		response := response{}
		err = json.Unmarshal(body, &response)
		if err != nil {
			return nil, err
		}
		if res.StatusCode != 200 {
			return nil, errors.New(response.QuandlError.Message)
		}
		quotes[i] = convert(response.DatasetData, asset)
	}
	return quotes, nil
}
