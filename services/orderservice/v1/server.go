package main

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

var cacheStorage = make(map[int]Order)

func handle(w http.ResponseWriter, r *http.Request) {
	defer LogPanic()
	log.Printf("request from %s", GetIdentification(r))

	if r.URL.Path == "/healthz" {
		handleHealth(w, r)
	} else if r.URL.Path == "/orders" {
		handleCreateOrder(w, r)
	} else {
		handleGetOrder(w, r)
	}
}

func handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	if filterOutMethod([]string{http.MethodPost}, w, r) {
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("error reading request body: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	o := &Order{}
	err = json.Unmarshal(data, o)
	if err != nil {
		log.Printf("error parsing request: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	o.OrderNumber, err = GetNextNumber("order")
	if err != nil {
		log.Printf("could not get random number: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = saveOrder(o)
	if err != nil {
		log.Printf("could not save order: %s", err)
		InternalServerError(w, "could not save order")
		return
	}

	OkJson(w, o)
}

func saveOrder(o *Order) error {
	if localEnvironment != "" {
		cacheStorage[o.OrderNumber] = *o
	} else {
		client, err := GetFirestore()
		if err != nil {
			log.Printf("could not get firestore client: %s", err)
		} else {
			ctx := context.Background()
			d := client.Collection(baseCollection + "/orders").Doc(strconv.Itoa(o.OrderNumber))
			_, err := d.Set(ctx, o)
			if err != nil {
				return fmt.Errorf("error creating new document: %s", err)
			}
		}
	}
	return nil
}

func handleGetOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodHead && r.Method != http.MethodGet {
		w.Header().Set("allow", "HEAD,GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	on, err := strconv.Atoi(r.URL.Path[len("/orders/"):])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(fmt.Sprintf("invalid order number %v", on)))
		return
	}

	var o *Order
	if localEnvironment != "" {
		if cached, ok := cacheStorage[on]; ok {
			o = &cached
		} else {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(fmt.Sprintf("unknown order number %v", on)))
			return
		}
	} else {
		client, err := GetFirestore()
		if err != nil {
			log.Printf("could not get firestore client: %s", err)
			InternalServerError(w, "could not store order")
			return
		} else {
			ctx := context.Background()
			dr, err := client.Doc(fmt.Sprintf("%s/orders/%v", baseCollection, on)).Get(ctx)
			if err != nil {
				if status.Code(err) == codes.NotFound {
					w.WriteHeader(http.StatusNotFound)
					_, _ = w.Write([]byte(fmt.Sprintf("unknown order number %v", on)))
					return
				}
				log.Printf("error reading order from firestore: %s", err)
				InternalServerError(w, "error reading from Firestore")
				return
			}
			o = &Order{}
			err = dr.DataTo(o)
			if err != nil {
				log.Printf("error parsing Firestore data: %s", err)
				InternalServerError(w, "error parsing Firestore data")
				return
			}
		}
	}

	OkJson(w, o)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if numberServiceScheme == "" {
		log.Fatal("NUMBER_SERVICE not set")
	}

	addAttacks()
	http.HandleFunc("/", handle)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
