package main

import (
	"embed"
	"net/http"
)

//go:embed styles/*
var styles embed.FS

func ListStyles() ([]string, error) {
	entries, err := styles.ReadDir("styles")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	return names, nil
}

func StylesHandler() http.Handler {
	return http.FileServer(http.FS(styles))
}

//go:embed octicons/*
var octicons embed.FS

func OcticonsHandler() http.Handler {
	return http.FileServer(http.FS(octicons))
}
