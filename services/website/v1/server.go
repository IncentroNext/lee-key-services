package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func handleProxy(w http.ResponseWriter, r *http.Request) {
	defer LogPanic()
	log.Printf("request from %s", GetIdentification(r))

	if r.Method != http.MethodPost {
		w.Header().Set("allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var scheme string
	if r.URL.Path == "/orders" {
		scheme = orderServiceScheme
	} else {
		scheme = paymentServiceScheme
	}
	url := scheme + r.URL.Path

	resp, err := DoWithAuth(r.Method, url, r.Body, r.Header.Get("content-type"), "")
	if err != nil {
		log.Printf("error proxying to order scheme at %s: %s", orderServiceScheme, err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message": "request failed"}`))
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("unexpected status code: %v", resp.StatusCode)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message": "request failed"}`))
		return
	}

	w.Header().Set("content-type", resp.Header.Get("content-type"))
	body, _ := ioutil.ReadAll(resp.Body)
	_, _ = w.Write(body)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if orderServiceScheme == "" {
		log.Fatal("ORDER_SERVICE not set")
	}
	if paymentServiceScheme == "" {
		log.Fatal("PAYMENT_SERVICE not set")
	}

	addAttacks()
	http.HandleFunc("/orders", handleProxy)
	http.HandleFunc("/payments", handleProxy)
	http.HandleFunc("/healthz", handleHealth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "files"+r.URL.Path)
	})
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
