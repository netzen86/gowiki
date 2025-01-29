// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore

package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
)

type Page struct {
	Title string
	Body  []byte
}

type Storage struct {
	Data map[string]Page
}

func newStorage() *Storage {
	return &Storage{Data: make(map[string]Page)}
}

func (s *Storage) save(p Page) {
	s.Data[p.Title] = p
}

func (s *Storage) loadPage(title string) (Page, error) {
	page, ok := s.Data[title]
	if !ok {
		return Page{}, fmt.Errorf("page not found")
	}
	return page, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string, s *Storage) {
	log.Println("VIEWED")

	p, err := s.loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string, s *Storage) {
	log.Println("EDITED")

	p, err := s.loadPage(title)
	if err != nil {
		p = Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string, s *Storage) {
	log.Println("SAVED")

	body := r.FormValue("body")
	s.save(Page{Title: title, Body: []byte(body)})
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

var templates = template.Must(template.ParseFiles("./templates/edit.html", "./templates/view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func (s *Storage) makeHandler(fn func(http.ResponseWriter, *http.Request, string, *Storage)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2], s)
	}
}

func main() {
	log.Println("RUNNED")
	storage := newStorage()
	http.HandleFunc("/view/", storage.makeHandler(viewHandler))
	http.HandleFunc("/edit/", storage.makeHandler(editHandler))
	http.HandleFunc("/save/", storage.makeHandler(saveHandler))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
