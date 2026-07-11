package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "webapp")
}
func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", handler).Methods("GET")
	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
