package chart

import (
	"strings"

	"github.com/guptarohit/asciigraph"
)

func RenderLineASCII(closes []float64, width, height int, header, caption, footer string) string {
	if width <= 0 { width = 100 }
	if height <= 0 { height = 30 }
	chartW := max(40, width-4)
	chartH := max(10, height-8)

	graph := asciigraph.Plot(
		closes,
		asciigraph.Width(chartW),
		asciigraph.Height(chartH),
		asciigraph.Caption(caption),
		asciigraph.Offset(1),
		asciigraph.SeriesColors(asciigraph.Green),
	)
	var b strings.Builder
	b.WriteString(header + "\n")
	b.WriteString(graph)
	b.WriteString("\n" + footer + "\n")
	return b.String()
}


func RenderCandlesASCII(ticks []Tick, width, height int, header, caption, footer string) string {
	if width <= 0 { width = 100 }
	if height <= 0 { height = 30 }
	chartW := max(50, width-4)
	chartH := max(12, height-8)

	// one column per tick (use most recent if narrow)
	if len(ticks) > chartW {
		ticks = ticks[len(ticks)-chartW:]
	}

	lo, hi := ticks[0].L, ticks[0].H
	for _, k := range ticks {
		if k.L < lo { lo = k.L }
		if k.H > hi { hi = k.H }
	}
	if hi == lo { hi = lo + 1 }

	// canvas rows: top→bottom; cols: left→right
	canvas := make([][]rune, chartH)
	for i := range canvas {
		row := make([]rune, len(ticks))
		for j := range row { row[j] = ' ' }
		canvas[i] = row
	}
	yScale := func(p float64) int {
		rel := (p - lo) / (hi - lo)
		y := int(float64(chartH-1) - rel*float64(chartH-1))
		if y < 0 { y = 0 }
		if y >= chartH { y = chartH-1 }
		return y
	}

	upCol := make([]bool, len(ticks))
	for x, k := range ticks {
		yH, yL := yScale(k.H), yScale(k.L)
		yO, yC := yScale(k.O), yScale(k.C)
		for y := yH; y <= yL; y++ { canvas[y][x] = '│' } // wick
		top, bot := yO, yC
		if bot < top { top, bot = bot, top }
		for y := top; y <= bot; y++ { canvas[y][x] = '█' } // body
		upCol[x] = k.C >= k.O
	}

	var b strings.Builder
	b.WriteString(header + "\n")
	b.WriteString(caption + "\n")
	for y := 0; y < chartH; y++ {
		for x := 0; x < len(ticks); x++ {
			r := canvas[y][x]
			if r == '█' || r == '│' {
				if upCol[x] { b.WriteString("\x1b[32m") } else { b.WriteString("\x1b[31m") }
				b.WriteRune(r)
				b.WriteString("\x1b[0m")
			} else {
				b.WriteByte(' ')
			}
		}
		b.WriteByte('\n')
	}
	b.WriteString("\n" + footer + "\n")
	return b.String()
}
