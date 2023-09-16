package main

import (
	"embed"
	"net/http"
)

//go:embed js/*
var js embed.FS

func ListJs() ([]string, error) {
	entries, err := styles.ReadDir("js")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names, nil
}

func JsHandler() http.Handler {
	return http.FileServer(http.FS(js))
}
