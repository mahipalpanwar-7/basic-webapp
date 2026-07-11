package main

import (
	"html/template"
	"net/http"

	"github.com/gorilla/mux" // used for URI parameters insted of query parameters
)

var templates *template.Template // need templates to be accessible from routes  for simplicity global object is created

func main() {
	templates = template.Must(template.ParseGlob("templates/*.html")) // instantiate the template object (for parse the code from the folder)
	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler).Methods("GET")
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.html", nil)
}
