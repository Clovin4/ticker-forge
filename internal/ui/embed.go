package ui

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
)

//go:embed server/static/* server/templates/*
var assets embed.FS

func TemplatesFS() fs.FS {
    sub, err := fs.Sub(assets, "server/templates")
    if err != nil {
        panic(err) // Handle error properly in production
    }

    fs.WalkDir(sub, ".", func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            fmt.Println("Error:", err)
            return err
        }
        fmt.Println("Found file:", path)
        return nil
    })

    return sub
}

func StaticFS() http.FileSystem {
	sub, _ := fs.Sub(assets, "server/static")
	return http.FS(sub)
}
