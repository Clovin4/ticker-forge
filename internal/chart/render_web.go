package chart

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// RenderLinePage renders a simple line chart of closes over time.
func RenderLinePage(symbol string, times []time.Time, closes []float64) ([]byte, error) {
	if len(times) != len(closes) || len(closes) == 0 {
		return nil, fmt.Errorf("RenderLinePage: mismatched/empty data")
	}

	x := make([]string, 0, len(times))
	y := make([]opts.LineData, 0, len(closes))
	for i, t := range times {
		x = append(x, t.Format("2006-01-02 15:04"))
		y = append(y, opts.LineData{Value: closes[i]})
	}

	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			PageTitle: fmt.Sprintf("%s · Line", symbol),
			Width:     "100%",
			Height:    "560px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    fmt.Sprintf("%s – Close", symbol),
			Subtitle: "Data: Yahoo Finance (unofficial)",
			Left:     "center",
		}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}),
		charts.WithDataZoomOpts(opts.DataZoom{Type: "inside", Start: 0, End: 100}),
		charts.WithXAxisOpts(opts.XAxis{Type: "category"}),
		charts.WithYAxisOpts(opts.YAxis{Type: "value", Scale: opts.Bool(true)}),
	)
	line.SetXAxis(x).AddSeries("Close", y).
		SetSeriesOptions(
			charts.WithLineChartOpts(opts.LineChart{Smooth: opts.Bool(true)}),
			charts.WithAreaStyleOpts(opts.AreaStyle{Opacity: opts.Float(0.15)}),
		)

	var buf bytes.Buffer
	if err := line.Render(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenderKlinePage renders OHLC candles (K-line).
// Uses chart.Tick from your chart package (T, O, H, L, C, V).
func RenderKlinePage(symbol string, ticks []Tick) ([]byte, error) {
	if len(ticks) == 0 {
		return nil, fmt.Errorf("RenderKlinePage: empty data")
	}

	x := make([]string, 0, len(ticks))
	y := make([]opts.KlineData, 0, len(ticks))
	for _, k := range ticks {
		x = append(x, k.T.Format("2006-01-02 15:04"))
		// Kline expects [open, close, low, high] in that order
		y = append(y, opts.KlineData{Value: []any{k.O, k.C, k.L, k.H}})
	}

	k := charts.NewKLine()
	k.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			PageTitle: fmt.Sprintf("%s · Candles", symbol),
			Width:     "100%",
			Height:    "560px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    fmt.Sprintf("%s – Candlesticks", symbol),
			Subtitle: "Data: Yahoo Finance (unofficial)",
			Left:     "center",
		}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}),
		charts.WithDataZoomOpts(opts.DataZoom{Type: "inside", Start: 0, End: 100}),
		charts.WithXAxisOpts(opts.XAxis{Type: "category"}),
	)
	k.SetXAxis(x).AddSeries("kline", y)

	var buf bytes.Buffer
	if err := k.Render(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
