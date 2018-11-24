package tiingo

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
	token  string
	client http.Client
}

// NewAdapter NewAdapter
func NewAdapter(token string) *Adapter {
	return &Adapter{
		token:  token,
		client: http.Client{},
	}
}

func convert(tquotes []quote, asset model.Asset) []model.Quote {
	quotes := make([]model.Quote, len(tquotes))
	for i, q := range tquotes {
		quotes[i].Asset = asset
		quotes[i].Price = q.Close
		quotes[i].Date = q.Date
	}
	return quotes
}

// GetDailyAsc GetDailyAsc
func (adapter *Adapter) GetDailyAsc(from, to time.Time, assets ...model.Asset) ([][]model.Quote, error) {

	quotes := make([][]model.Quote, len(assets))
	for i, asset := range assets {
		url := fmt.Sprintf("https://api.tiingo.com/tiingo/daily/%s/prices", asset.Symbol)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		q := req.URL.Query()
		q.Add("startDate", from.Format(DateFormat))
		q.Add("endDate", to.Format(DateFormat))
		req.Header.Add("Authorization", "Token "+adapter.token)
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
		if res.StatusCode != 200 {
			var tErr tiingoError
			err := json.Unmarshal(body, &tErr)
			if err != nil {
				return nil, err
			}
			return nil, errors.New(tErr.Detail)
		}

		var response []quote
		err = json.Unmarshal(body, &response)
		if err != nil {
			return nil, err
		}
		quotes[i] = convert(response, asset)
	}
	return quotes, nil
}
