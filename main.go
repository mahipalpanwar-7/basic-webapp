package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-redis/redis" // Redis for fast data storage
	"github.com/gorilla/mux"    // used for URI parameters insted of query parameters
)

var client *redis.Client

var templates *template.Template // need templates to be accessible from different routes  for simplicity global object is created

func main() {

	// instantiating client object
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	templates = template.Must(template.ParseGlob("templates/*.html")) // instantiate the template object (for parse the code from the folder)
	r := mux.NewRouter()
	r.HandleFunc("/", indexGetHandler).Methods("GET")
	r.HandleFunc("/", indexPostHandler).Methods("POST")
	// static files access 
	fs := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static",fs))
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

func indexGetHandler(w http.ResponseWriter, r *http.Request) {
	comments, err := client.LRange("comments", 0, 10).Result()
	if err != nil {
		return
	}
	fmt.Println(comments)
	templates.ExecuteTemplate(w, "index.html", comments)
}

func indexPostHandler(w http.ResponseWriter, r *http.Request) {
	// parse the form from the request body
	r.ParseForm()
	comment := r.PostForm.Get("comment")
	client.LPush("comments", comment)
	http.Redirect(w, r, "/", 302) // redirecting to the same page
}
