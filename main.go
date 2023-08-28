package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var addr string
var renderer string
var githubHost string
var githubToken string
var command string
var mermaid bool
var mmdcPath string

func init() {
	flag.StringVar(&addr, "addr", "localhost:0", "address on which to listen")
	flag.StringVar(&renderer, "renderer", "github", "renderer to use: github, command")
	flag.StringVar(&githubHost, "github-host", "api.github.com", "github host to use")
	flag.StringVar(&githubToken, "github-token", "", "github token to use, not required but will increase rate limits")
	flag.StringVar(&command, "command", "markdown_py", "command to use for rendering")
	flag.BoolVar(&mermaid, "mermaid", true, "enable mermaid rendering support")
	flag.StringVar(&mmdcPath, "mmdc-path", "mmdc", "path to mmdc binary")

	flag.Usage = func() {
		fmt.Println("Usage: mdlp [options] <file>")
		fmt.Println("Render a live preview of the given markdown file. Uses a command or GitHub API to render the markdown.")
		fmt.Println("Options:")
		flag.PrintDefaults()
	}
}

func main() {
	// Handle our command line flags.
	flag.Parse()
	file := flag.Arg(0)
	if file == "" {
		flag.Usage()
		return
	}

	// Create a listener.
	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("listening on %v: %v\n", addr, err)
		return
	}
	defer l.Close()
	// update addr to reflect the actual address, in case we were
	// given ":0".
	addr = l.Addr().String()

	// Setup our watcher.
	watcher, err := NewWatcher(file)
	if err != nil {
		fmt.Printf("creating watcher for %v: %v\n", file, err)
		return
	}
	defer watcher.Close()
	fileChanged := watcher.Watch()

	// Setup our renderer.
	var r Renderer
	switch renderer {
	case "github":
		r = NewGitHubApiRenderer(githubHost, githubToken)
	case "command":
		r = NewCommandRenderer(command, []string{})
	default:
		fmt.Printf("unknown renderer: %v\n", renderer)
		return
	}
	defer r.Close()

	// Setup our mermaid wrapper
	if mermaid {
		mermaid, err := NewMermaidRenderer(mmdcPath, r)
		if err != nil {
			fmt.Printf("creating mermaid renderer: %v\n", err)
			return
		}
		http.Handle("/mermaid/", mermaid.Handler())
		defer mermaid.Close()
		r = mermaid
	}

	// Setup our websocket handler.
	http.Handle("/ws", WebSocketHandler(file, r, fileChanged))

	// Setup styles and octicons server.
	http.Handle("/styles/", StylesHandler())
	http.Handle("/octicons/", OcticonsHandler())

	// Setup rendering the main page
	page, err := NewPage(file, addr, r)
	if err != nil {
		fmt.Printf("creating page: %v\n", err)
		return
	}
	http.Handle("/", page)

	// Handle signals.
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Serve in a goroutine so we can handle signals.
	srv := &http.Server{
		Addr:    l.Addr().String(),
		Handler: http.DefaultServeMux,
	}
	go func() {
		fmt.Printf("serving at http://%s/\n", addr)
		srv.Serve(l)
	}()

	<-done
	fmt.Println("shutting down")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		fmt.Printf("shutting down: %v\n", err)
	}
}
