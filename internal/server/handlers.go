package server

import (
	"fmt"
	"html/template"
	"net/http"

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
		html := fmt.Sprintf(
			`<iframe class="chart-frame" src="/chart?symbol=%s&range=%s&interval=%s" loading="lazy"></iframe>`,
			template.URLQueryEscaper(symbol),
			template.URLQueryEscaper(rng),
			template.URLQueryEscaper(interval),
		)
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, html)
	}
}

func Chart() gin.HandlerFunc {
	return func(c *gin.Context) {
		symbol := orDefault(c.Query("symbol"), "", "AAPL")
		rng := orDefault(c.Query("range"), "", "1d")
		interval := orDefault(c.Query("interval"), "", "1m")

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
