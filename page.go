package main

import (
	_ "embed"
	"net/http"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed index.html
var index string

type Page struct {
	tmpl     *template.Template
	file     string
	addr     string
	renderer Renderer
}

func NewPage(file string, addr string, renderer Renderer) (*Page, error) {
	tmpl, err := template.New("index").Parse(index)
	if err != nil {
		return nil, err
	}

	return &Page{tmpl, file, addr, renderer}, nil
}

type PageData struct {
	FileName string
	Content  string
	Addr     string
	Styles   []string
}

func (p *Page) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	file := p.file
	content, err := os.ReadFile(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	rendered, err := p.renderer.Render(string(content))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	styles, err := ListStyles()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	err = p.tmpl.Execute(w, PageData{FileName: filepath.Base(p.file), Content: rendered, Addr: p.addr, Styles: styles})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
}
