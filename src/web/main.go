package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

type (
	page struct {
		title string
		body  []byte
	}

	loginController struct {
	}

	User struct {
		UserName string
	}

	adminController struct {
	}
)

// Login Controller
func (this *loginController) IndexAction(w http.ResponseWriter, r *http.Request) {
	if t, err := template.ParseFiles("template/html/login/index.html"); err == nil {
		t.Execute(w, nil)
	}
}

func (this *adminController) IndexAction(w http.ResponseWriter,
	r *http.Request, user string) {
	t, err := template.ParseFiles("template/html/admin/index.html")
	if err != nil {
		t.Execute(w, &User{user})
	}
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

///////
func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Redirect(w, r, "/login/index", http.StatusFound)
	}

	if t, err := template.ParseFiles("template/html/404.html"); err != nil {
		t.Execute(w, nil)
	}
}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	pathInfo := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(pathInfo, "/")
	action := ""
	if len(parts) > 1 {
		action = strings.Title(parts[1]) + "Action"
	}

	login := &loginController{}
	controller := reflect.ValueOf(login)
	method := controller.MethodByName(action)
	if !method.IsValid() {
		method = controller.MethodByName(strings.Title("index") + "Action")
	}
	requestValue := reflect.ValueOf(r)
	responseValue := reflect.ValueOf(w)
	method.Call([]reflect.Value{responseValue, requestValue})
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("admin_name")
	if err != nil && cookie.Value == "" {
		http.Redirect(w, r, "/login/index", http.StatusNotFound)

		pathInfo := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(pathInfo, "/")

		action := ""
		if len(parts) > 1 {
			action = strings.Title(parts[1]) + "Action"
		}

		admin := &adminController{}
		controller := reflect.ValueOf(admin)
		method := controller.MethodByName(action)
		if !method.IsValid() {
			method = controller.MethodByName(strings.Title("index") + "Action")
		}
		requestValue := reflect.ValueOf(r)
		responseValue := reflect.ValueOf(w)
		userValue := reflect.ValueOf(cookie.Value)
		method.Call([]reflect.Value{responseValue, requestValue, userValue})

	}
}
func main() {
	http.Handle("/css/", http.FileServer(http.Dir("template")))
	http.Handle("/js/", http.FileServer(http.Dir("template")))

	http.HandleFunc("/admin/", adminHandler)
	http.HandleFunc("/login/", loginHandler)
	http.HandleFunc("/ajax/", ajaxHandler)
	http.HandleFunc("/", notFoundHandler)
	http.ListenAndServe(":8888", nil)
}
