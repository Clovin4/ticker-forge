package chart

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type yfChartResp struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Symbol    string `json:"symbol"`
				Timezone  string `json:"timezone"`
				Gmtoffset int64  `json:"gmtoffset"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Close  []float64 `json:"close"`
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Volume []int64   `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error any `json:"error"`
	} `json:"chart"`
}

func FetchIntraday(symbol, rng, interval string) ([]time.Time, []float64, error) {
	if rng == "" {
		rng = "1d"
	}
	if interval == "" {
		interval = "1m"
	}
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?range=%s&interval=%s", symbol, rng, interval)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (TickerForge)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("yahoo request: %w", err)
	}
	defer resp.Body.Close()

	var data yfChartResp
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, nil, fmt.Errorf("decode: %w", err)
	}
	if len(data.Chart.Result) == 0 || len(data.Chart.Result[0].Indicators.Quote) == 0 {
		return nil, nil, fmt.Errorf("no data for %s", symbol)
	}
	r := data.Chart.Result[0]
	q := r.Indicators.Quote[0]

	var times []time.Time
	var closes []float64
	for i, ts := range r.Timestamp {
		if i >= len(q.Close) {
			continue
		}
		c := q.Close[i]
		if c == 0 || (c != c) {
			continue
		}
		times = append(times, time.Unix(ts, 0))
		closes = append(closes, c)
	}
	return times, closes, nil
}
