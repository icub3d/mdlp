package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Renderer interface {
	Render(original string) (string, error)
	Close() error
}

type CommandRenderer struct {
	command string
	args    []string
}

func NewCommandRenderer(command string, args []string) *CommandRenderer {
	return &CommandRenderer{command: command, args: args}
}

func (r *CommandRenderer) Render(original string) (string, error) {
	cmd := exec.Command(r.command, r.args...)

	cmd.Stdin = strings.NewReader(original)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("running command '%v %v' (%v): %v", r.command, r.args, stderr.String(), err)
	}
	return out.String(), nil
}

func (r *CommandRenderer) Close() error { return nil }

type GitHubApiRenderer struct {
	host  string
	token string
}

func NewGitHubApiRenderer(host string, token string) *GitHubApiRenderer {
	return &GitHubApiRenderer{host: host, token: token}
}

type GitHubBody struct {
	Text string `json:"text"`
	Mode string `json:"mode"`
}

func (r *GitHubApiRenderer) Render(original string) (string, error) {
	// Create the body.
	body := GitHubBody{Text: original, Mode: "markdown"}
	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshalling body: %v", err)
	}

	// Create the request.
	url := fmt.Sprintf("https://%v/markdown", r.host)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("creating request: %v", err)
	}

	// Set the headers.
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if r.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", r.token))
	}

	// Send the request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response.
	var out bytes.Buffer
	_, err = out.ReadFrom(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %v", err)
	}

	// check the status code.
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("got status code %v: %v", resp.StatusCode, out.String())
	}

	// Return the response.
	return out.String(), nil
}

func (r *GitHubApiRenderer) Close() error { return nil }

type MermaidRenderer struct {
	command  string
	tempdir  string
	renderer Renderer
}

func NewMermaidRenderer(command string, renderer Renderer) (*MermaidRenderer, error) {
	tempdir, err := os.MkdirTemp("", "mdlp")
	if err != nil {
		return nil, fmt.Errorf("creating tempdir: %v", err)
	}
	return &MermaidRenderer{command: command, tempdir: tempdir, renderer: renderer}, nil
}

func (r *MermaidRenderer) execMmdc(diagram string, count int) error {
	cmd := exec.Command(r.command, "-i", "-", "-e", "png", "-o", filepath.Join(r.tempdir, fmt.Sprintf("mermaid-%v.png", count)))
	cmd.Stdin = strings.NewReader(diagram)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("running command '%v' (%v): %v %v", r.command, err, stdout.String(), stderr.String())
	}
	return nil
}

func (r *MermaidRenderer) Render(original string) (string, error) {
	// We need to find all the code blocks that are mermaid
	// (```mermaid) and their end (```) and replace them with a
	// markdown link to the image.\
	count := 0
	for {
		count++
		cur := strings.Index(original, "```mermaid")
		if cur == -1 {
			break
		}

		end := strings.Index(original[cur+11:], "```")
		if end == -1 {
			return "", fmt.Errorf("found opening ```mermaid without closing ```")
		}

		// Make the diagram.
		err := r.execMmdc(original[cur+11:cur+11+end], count)
		if err != nil {
			return "", fmt.Errorf("making diagram: %v", err)
		}

		// update the original string.
		original = original[:cur] + fmt.Sprintf("![mermaid diagram](/mermaid/mermaid-%v.png)", count) + original[cur+11+end+3:]
	}

	return r.renderer.Render(original)
}

func (r *MermaidRenderer) Close() error {
	return os.RemoveAll(r.tempdir)
}

func (r *MermaidRenderer) Handler() http.Handler {
	return http.StripPrefix("/mermaid/", http.FileServer(http.Dir(r.tempdir)))
}
