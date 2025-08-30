package server

import (
	"html/template"
	"log"

	"ticker-forge/internal/ui"

	"github.com/gin-gonic/gin"
)

type Options struct {
	Port     string
	DefaultSymbol  string
	DefaultRange   string
	DefaultInterval string
}

func NewRouter(opts Options) *gin.Engine {
	r := gin.Default()

	// Static files (from embed)
	r.StaticFS("/static", ui.StaticFS())

	// Templates (from embed)
	tfs := ui.TemplatesFS()
	tpl := template.Must(template.New("").ParseFS(tfs, "*.html"))
	r.SetHTMLTemplate(tpl)

	// Routes
	r.GET("/", Index(opts))
	r.GET("/frame", Frame())
	r.GET("/chart", Chart())

	return r
}

func ListenAndServe(opts Options) error {
	if opts.Port == "" {
		opts.Port = "8080"
	}
	r := NewRouter(opts)
	log.Printf("listening on http://localhost:%s", opts.Port)
	return r.Run(":" + opts.Port)
}
