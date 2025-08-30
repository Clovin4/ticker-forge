package ui

import (
	"embed"
	"io/fs"
	"net/http"
)

var assets embed.FS

func TemplatesFS() fs.FS {
	sub, _ := fs.Sub(assets, "server/templates")
	return sub
}

func StaticFS() http.FileSystem {
	sub, _ := fs.Sub(assets, "server/static")
	return http.FS(sub)
}
