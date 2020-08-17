package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

var counters = make(map[string]int)
var mux = sync.RWMutex{}

func handleRanges(w http.ResponseWriter, r *http.Request) {
	defer LogPanic()
	log.Printf("request from %s", GetIdentification(r))

	if filterOutMethod([]string{http.MethodGet}, w, r) {
		return
	}

	key := r.URL.Path[len("/ranges/"):]
	if key == "" {
		http.NotFound(w, r)
		return
	}

	mux.Lock()
	defer mux.Unlock()
	c := counters[key] + 1
	counters[key] = c

	_, _ = w.Write([]byte(strconv.Itoa(c)))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addAttacks()
	http.HandleFunc("/ranges/", handleRanges)
	http.HandleFunc("/healthz", handleHealth)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
