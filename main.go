package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/theplant/blackfriday"
	"gopkg.in/unrolled/render.v1"
)

var Render = render.New(render.Options{
	Layout:        "layout",
	IsDevelopment: true,
})

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fs := http.FileServer(http.Dir("."))
	handler := func(rw http.ResponseWriter, r *http.Request) {
		// Render the index as the main readme
		if r.URL.Path == "/" {
			if err := renderMarkdown(rw, "README.md"); err != nil {
				return
			}
			// Render markdown files
		} else if strings.HasSuffix(r.URL.Path, ".md") {
			if err := renderMarkdown(rw, r.URL.Path[1:]); err != nil {
				return
			}
		} else if strings.HasSuffix(r.URL.Path, ".go") {

			// Do we want to run code or render it?
			if strings.HasPrefix(r.URL.Path, "/run") {
				if err := runCode(rw, r.URL.Path[5:]); err != nil {
					log.Println("Error:", err)
				}
				return
			}
			if err := renderCode(rw, r.URL.Path[1:]); err != nil {
				return
			}
		} else {
			fs.ServeHTTP(rw, r)
		}

	}

	fmt.Println("Listening on port", port)
	http.ListenAndServe(":"+port, http.HandlerFunc(handler))
}

func renderMarkdown(rw http.ResponseWriter, name string) error {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		http.Error(rw, "Unable to read file", 500)
		return err
	}

	output := blackfriday.MarkdownCommon(data)
	Render.HTML(rw, 200, "slide", template.HTML(output))
	return nil
}

func renderCode(rw http.ResponseWriter, name string) error {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		http.Error(rw, "Unable to read file", 500)
		return err
	}

	d := struct {
		Code string
		File string
	}{
		Code: string(data),
		File: name,
	}

	Render.HTML(rw, 200, "code", d)
	return nil
}

func runCode(rw http.ResponseWriter, name string) error {
	log.Println("Running file", name)

	cmd := exec.Command("go", "run", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(rw, string(out), http.StatusInternalServerError)
		return err
	}

	rw.Write(out)

	return nil
}
