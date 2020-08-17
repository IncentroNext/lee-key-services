package main

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var cacheStorage = make(map[int]Invoice)

var invoicesBucket = getInvoicesBucket()

func handle(w http.ResponseWriter, r *http.Request) {
	defer LogPanic()
	log.Printf("request from %s", GetIdentification(r))

	if r.URL.Path == "/healthz" {
		handleHealth(w, r)
	} else if r.URL.Path == "/invoices" {
		handleCreateInvoice(w, r)
	} else {
		http.NotFound(w, r)
	}
}

func handleCreateInvoice(w http.ResponseWriter, r *http.Request) {
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

	i := &Invoice{
		Customer: o.Customer,
		Total:    NewMoney(o.Quantity*1100, o.Quantity*01),
	}
	i.InvoiceNumber, err = GetNextNumber("invoice")
	if err != nil {
		log.Printf("could not get number: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = saveInvoice(i)
	if err != nil {
		log.Printf("could not save order: %s", err)
		InternalServerError(w, "could not save order")
		return
	}

	OkJson(w, i)
}

func saveInvoice(i *Invoice) error {
	if localEnvironment != "" {
		cacheStorage[i.InvoiceNumber] = *i
	} else {
		ctx := context.Background()
		client, err := storage.NewClient(ctx)
		if err != nil {
			log.Printf("could not get storage client: %s", err)
			return err
		}
		name := fmt.Sprintf("invoice-%v", i.InvoiceNumber)
		w := client.Bucket(invoicesBucket).Object(name).NewWriter(context.Background())
		bs, _ := json.Marshal(i)
		_, err = w.Write(bs)
		if err != nil {
			return fmt.Errorf("error writing invoice to bucket: %s", err)
		}
		err = w.Close()
		if err != nil {
			return fmt.Errorf("error writing to invoice bucket: %s", err)
		}
	}
	return nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addAttacks()
	http.HandleFunc("/", handle)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
