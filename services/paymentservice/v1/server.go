package main

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	"github.com/HayoVanLoon/go-commons/logjson"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var cacheStorage = make(map[int]Payment)

var paymentsBucket = getPaymentsBucket()

func handleCreatePayment(w http.ResponseWriter, r *http.Request) {
	defer LogPanic()
	logjson.Info(fmt.Sprintf("request from %s", GetIdentification(r)))

	if filterOutMethod([]string{http.MethodPost}, w, r) {
		return
	}

	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logjson.Warn(fmt.Sprintf("error reading request body: %s", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	p := &Payment{}
	err = json.Unmarshal(data, p)
	if err != nil {
		logjson.Warn(fmt.Sprintf("error parsing request: %s", err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !orderExists(p.OrderNumber) {
		w.WriteHeader(http.StatusNotFound)
		msg := fmt.Sprintf(`{"message": "unknown order number %v"}`, p.OrderNumber)
		_, _ = w.Write([]byte(msg))
		return
	}

	p.PaymentNumber, err = GetNextNumber("payment")
	if err != nil {
		logjson.Warn(fmt.Sprintf("could not get random number: %s", err))
		InternalServerError(w, "could not get random number")
		return
	}

	err = savePayment(p)
	if err != nil {
		logjson.Warn(fmt.Sprintf("could not save payment: %s", err))
		InternalServerError(w, "could not save payment")
		return
	}

	OkJson(w, p)
}

func orderExists(o int) bool {
	r, err := HeadWithAuth(fmt.Sprintf("%s/orders/%v", orderServiceScheme, o))
	if err != nil {
		return false
	}
	return r.StatusCode == http.StatusOK
}

func savePayment(p *Payment) error {
	if localEnvironment != "" {
		cacheStorage[p.OrderNumber] = *p
	} else {
		client, err := storage.NewClient(context.Background())
		if err != nil {
			return fmt.Errorf("could not get storage client: %s", err)
		}
		name := fmt.Sprintf("payment-%v", p.PaymentNumber)
		w := client.Bucket(paymentsBucket).Object(name).NewWriter(context.Background())
		bs, _ := json.Marshal(p)
		_, err = w.Write(bs)
		if err != nil {
			return fmt.Errorf("error writing payment to bucket: %s", err)
		}
		err = w.Close()
		if err != nil {
			return fmt.Errorf("error writing to payments bucket: %s", err)
		}
	}
	return nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if orderServiceScheme == "" {
		logjson.Critical("ORDER_SERVICE not set")
	}
	if numberServiceScheme == "" {
		logjson.Critical("NUMBER_SERVICE not set")
	}

	addAttacks()
	http.HandleFunc("/payments", handleCreatePayment)
	http.HandleFunc("/healthz", handleHealth)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
