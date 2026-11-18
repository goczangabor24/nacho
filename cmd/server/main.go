package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gaborgoczan/nacho/internal/static"
	"github.com/gaborgoczan/nacho/internal/templates"
)

var tmpl = template.Must(template.ParseFS(templates.FS, "layout.html", "login.html"))

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static.FS)))

	mux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		render(w, "layout", map[string]any{"Title": "Sign in"})
	})

	addr := ":" + port
	log.Printf("listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func render(w http.ResponseWriter, name string, data any) {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		log.Printf("render %q: %v", name, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}
