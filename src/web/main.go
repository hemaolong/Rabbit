package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type page struct {
	title string
	body  []byte
}

func (p *page) save() error {
	fname := p.title + ".txt"
	return ioutil.WriteFile(fname, p.body, 0600)
}

func loadPage(title string) (*page, error) {
	fname := title + ".txt"
	println(fname)
	body, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	return &page{title: title, body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[1:]
	if p, err := loadPage(title); err == nil {
		fmt.Fprintf(w, "<h1>%s</h1><div>%s</div>", p.title, string(p.body))
	}
}
func main() {
	http.HandleFunc("/view/", viewHandler)
	http.ListenAndServe(":8080", nil)
}
