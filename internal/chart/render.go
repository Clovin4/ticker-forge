package chart

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func RenderLinePage(symbol string, times []time.Time, closes []float64) ([]byte, error) {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			PageTitle: fmt.Sprintf("%s Intraday", symbol),
			Width:     "100%",
			Height:    "520px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    fmt.Sprintf("%s – Intraday", symbol),
			Subtitle: "Data: Yahoo Finance (unofficial)",
			Left:     "center",
		}),
		charts.WithTooltipOpts(opts.Tooltip{Show: opts.Bool(true), Trigger: "axis"}),
		charts.WithXAxisOpts(opts.XAxis{Type: "time"}),
		charts.WithYAxisOpts(opts.YAxis{
			Type: "value",
			AxisLabel: &opts.AxisLabel{Formatter: "{value}"},
			Scale: opts.Bool(true),
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type: "inside", Start: 0, End: 100, Throttle: 50,
		}),
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
	)

	items := make([]opts.LineData, 0, len(times))
	for i, t := range times {
		items = append(items, opts.LineData{Value: []any{t, closes[i]}})
	}
	line.SetXAxis(times).
		AddSeries("Close", items).
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
