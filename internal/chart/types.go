package chart

import "time"

// Tick = OHLCV bar
type Tick struct {
	T time.Time
	O float64
	H float64
	L float64
	C float64
	V int64
}

// MakeTicksFromCloses synthesizes OHLC from consecutive closes.
// (Prev close → open; hi/lo = max/min(prev,curr))
func MakeTicksFromCloses(ts []time.Time, closes []float64) []Tick {
	if len(closes) < 2 {
		return nil
	}
	n := len(closes) - 1
	out := make([]Tick, 0, n)
	prev := closes[0]
	for i := 1; i < len(closes); i++ {
		cur := closes[i]
		o, c := prev, cur
		h, l := o, o
		if c > h { h = c }
		if c < l { l = c }
		var t time.Time
		if i < len(ts) { t = ts[i] }
		out = append(out, Tick{T: t, O: o, H: h, L: l, C: c})
		prev = cur
	}
	return out
}
