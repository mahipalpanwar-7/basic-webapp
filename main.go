package main

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/go-redis/redis" // Redis for fast data storage
	"github.com/gorilla/mux"    // used for URI parameters insted of query parameters
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var client *redis.Client                                  // global variable for all the handlers
var store = sessions.NewCookieStore([]byte("w6b-s3cr3t")) //session store
var templates *template.Template                          // need templates to be accessible from different routes  for simplicity global object is created

func main() {

	// instantiating client object
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	templates = template.Must(template.ParseGlob("templates/*.html")) // instantiate the template object (for parse the code from the folder)
	r := mux.NewRouter()
	r.HandleFunc("/", AuthRequired(indexGetHandler)).Methods("GET")
	r.HandleFunc("/", AuthRequired(indexPostHandler)).Methods("POST")
	r.HandleFunc("/login", loginGetHandler).Methods("GET")
	r.HandleFunc("/login", loginPostHandler).Methods("POST")
	r.HandleFunc("/register", registerGetHandler).Methods("GET")
	r.HandleFunc("/register", registerPostHandler).Methods("POST")

	r.HandleFunc("/test", testGetHandler).Methods("GET")

	// static files access
	fs := http.FileServer(http.Dir("./static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static", fs))
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

// middleware function
func AuthRequired(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session")
		_, ok := session.Values["username"]
		if !ok {
			http.Redirect(w, r, "/login", 302)
			return
		}
		handler.ServeHTTP(w, r)
	}
}

func indexGetHandler(w http.ResponseWriter, r *http.Request) {
	comments, err := client.LRange("comments", 0, 10).Result()
	if err != nil { // proper error handling
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}
	fmt.Println(comments)
	templates.ExecuteTemplate(w, "index.html", comments)

}

func indexPostHandler(w http.ResponseWriter, r *http.Request) {
	// parse the form from the request body
	r.ParseForm()
	comment := r.PostForm.Get("comment")
	err := client.LPush("comments", comment).Err()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}
	http.Redirect(w, r, "/", 302) // redirecting to the same page
}

func loginGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "login.html", nil)
}

func loginPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.PostForm.Get("username") // Getting the username from the form
	password := r.PostForm.Get("password")

	hash, err := client.Get("user:" + username).Bytes() // user key from Redis to get the stored hash for that user
	if err == redis.Nil {                               // error when user is not registered
		templates.ExecuteTemplate(w, "login.html", "unknown user")
		return

	} else if err != nil { //error if redis down/ database issue
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}
	err = bcrypt.CompareHashAndPassword(hash, []byte(password)) // check for password validation

	if err != nil {
		templates.ExecuteTemplate(w, "login.html", "invalid login")
		return
	}

	session, _ := store.Get(r, "session") // Getting the current session
	session.Values["username"] = username // Storing the username so it can be used across multiple requests.
	session.Save(r, w)                    // Saves the session by sending the session cookie to the browser.

	http.Redirect(w, r, "/", 302)

}

// for session testing purpose only

func testGetHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")

	untyped, ok := session.Values["username"]
	if !ok {
		return
	}
	username, ok := untyped.(string)
	if !ok {
		return
	}
	w.Write([]byte(username))
}

func registerGetHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "register.html", nil)
}

func registerPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.PostForm.Get("username")
	password := r.PostForm.Get("password")

	// it provides a sensible balance between login speed and password security.
	cost := bcrypt.DefaultCost // default cost 10 (good speed with good security) [brute-force attack]
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}

	err = client.Set("user:"+username, hash, 0).Err() // if redis failed while adding the user
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
		return
	}
	http.Redirect(w, r, "/login", 302)
}
