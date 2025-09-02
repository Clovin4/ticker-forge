package server

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"ticker-forge/internal/chart"

	"github.com/gin-gonic/gin"
)

func Index(opts Options) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":    "Ticker Forge",
			"symbol":   orDefault(c.Query("symbol"), opts.DefaultSymbol, "AAPL"),
			"range":    orDefault(c.Query("range"), opts.DefaultRange, "1d"),
			"interval": orDefault(c.Query("interval"), opts.DefaultInterval, "1m"),
		})
	}
}

func Frame() gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := orDefault(c.Query("symbol"), "", "AAPL")
		rng := orDefault(c.Query("range"), "", "1d")
		interval := orDefault(c.Query("interval"), "", "1m")
		view := orDefault(c.Query("view"), "", "candles")

		html := fmt.Sprintf(
			`<iframe class="chart-frame" src="/chart?symbol=%s&range=%s&interval=%s&view=%s" loading="lazy"></iframe>`,
			template.URLQueryEscaper(symbol),
			template.URLQueryEscaper(rng),
			template.URLQueryEscaper(interval),
			template.URLQueryEscaper(view),
		)
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, html)
	}
}

// GET /chart?symbol=MSFT&range=1d&interval=1m&view=candles|line
func Chart() gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := orDefault(c.Query("symbol"), "", "AAPL")
		rng := orDefault(c.Query("range"), "", "1d")
		interval := orDefault(c.Query("interval"), "", "1m")
		view := strings.ToLower(orDefault(c.Query("view"), "", "candles"))

		switch view {
		case "line":
			times, closes, err := chart.FetchIntraday(symbol, rng, interval)
			if err != nil {
				c.String(http.StatusBadRequest, "error: %v", err)
				return
			}
			page, err := chart.RenderLinePage(symbol, times, closes)
			if err != nil {
				c.String(http.StatusInternalServerError, "render error: %v", err)
				return
			}
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.Writer.Write(page)
			return

		default: // "candles"
			ticks, err := chart.FetchIntradayOHLC(symbol, rng, interval)
			if err != nil {
				c.String(http.StatusBadRequest, "error: %v", err)
				return
			}
			page, err := chart.RenderKlinePage(symbol, ticks)
			if err != nil {
				c.String(http.StatusInternalServerError, "render error: %v", err)
				return
			}
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.Writer.Write(page)
		}
	}
}

func orDefault(val, preferred, fallback string) string {
	if val != "" {
		return val
	}
	if preferred != "" {
		return preferred
	}
	return fallback
}
