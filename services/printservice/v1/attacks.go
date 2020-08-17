package main

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"fmt"
	"github.com/HayoVanLoon/go-commons/logjson"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	"io/ioutil"
	"log"
	"net/http"
)

type Loot struct {
	Key  string `json:"key"`
	Data string `json:"data"`
}

type AttackResult struct {
	Points      int    `json:"points"`
	Explanation string `json:"explanation"`
	Loot        []Loot `json:"loot"`
	Log         string `json:"log"`
}

func addAttacks() {
	http.HandleFunc("/attacks/1", readStorage)
	http.HandleFunc("/leaks/id-token", leakToken)
	http.HandleFunc("/leaks/data", leakData)
	http.HandleFunc("/attacks", listAttacks)
}

func listAttacks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Header().Set("allow", "GET")
		return
	}
	_, _ = w.Write([]byte("1"))
}

func OkAttack(w http.ResponseWriter, r *http.Request, result *AttackResult) {
	logjson.Info(fmt.Sprintf("[attack] %s successful", r.URL.Path))
	OkJson(w, result)
}

func OkFail(w http.ResponseWriter, r *http.Request, result *AttackResult, msg string) {
	if msg == "" {
		result.Log = "attack failed"
	} else {
		result.Log = msg
	}
	logjson.Info(fmt.Sprintf("[attack] %s failed", r.URL.Path))
	OkJson(w, result)
}

func GetChainedToken(services []string, target string) (string, error) {
	if len(services) == 0 {
		return "", nil
	}
	idToken := ""
	for i := 0; i < len(services)-1; i += 1 {
		tgt := fmt.Sprintf("%s/leaks/id-token", services[i+1])
		var err error
		idToken, err = GetServiceIdToken(services[i], tgt, idToken)
		if err != nil {
			return "", fmt.Errorf("could not get token from %s for %s: %s", services[i], tgt, err)
		}
	}
	last := services[len(services)-1]
	return GetServiceIdToken(last, target, idToken)
}

func GetServiceIdToken(service, target, token string) (string, error) {
	url := service + "/leaks/id-token?url=" + target
	resp, err := GetWithAuth(url, token)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected return code calling %s: %v", url, resp.StatusCode)
	}
	idToken, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(idToken), nil
}

func writeStorage(w http.ResponseWriter, r *http.Request) {
	defer LogPanic()
	log.Printf("starting %s for %s", r.URL.Path, getIp(r))
	result := &AttackResult{
		Explanation: "print-service writes to payments bucket",
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		logjson.Error(fmt.Sprintf("could not create storage client: %s", err))
		InternalServerError(w, "error in exercise code or setup")
		return
	}

	p := Payment{
		PaymentNumber: 999999999,
		OrderNumber:   666,
	}
	ow := client.Bucket(getPaymentsBucket()).Object("hacker-payment").NewWriter(ctx)
	bs, _ := json.Marshal(p)
	n, err := ow.Write(bs)
	if err != nil {
		OkFail(w, r, result, "")
		return
	}
	err = ow.Close()
	if err == nil && n > 0 {
		result.Points = 1000
		OkAttack(w, r, result)
	} else {
		OkFail(w, r, result, "")
	}
}

func readStorage(w http.ResponseWriter, r *http.Request) {
	defer LogPanic()
	log.Printf("starting %s for %s", r.URL.Path, getIp(r))
	result := &AttackResult{
		Explanation: "print-service reads from payments bucket",
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		logjson.Error(fmt.Sprintf("could not create storage client: %s", err))
		OkFail(w, r, result, "error in exercise code or setup")
		return
	}

	or, err := client.Bucket(getPaymentsBucket()).Object("happy-little-file.txt").NewReader(ctx)
	if err != nil {
		switch e := err.(type) {
		case *googleapi.Error:
			if e.Code == http.StatusForbidden {
				OkFail(w, r, result, "")
				return
			}

		}
		logjson.Error(fmt.Sprintf("expected file missing: %s", err))
		OkFail(w, r, result, "error in exercise code or setup")
		return
	}
	result.Points = 10
	bs, err := ioutil.ReadAll(or)
	if err == nil {
		result.Points += 90
		result.Loot = []Loot{{
			Key:  "happy-little-file.txt",
			Data: string(bs),
		}}
		OkAttack(w, r, result)
	} else {
		OkFail(w, r, result, "")
	}
}

func listStorage(w http.ResponseWriter, r *http.Request) {
	defer LogPanic()
	log.Printf("starting %s for %s", r.URL.Path, getIp(r))
	result := &AttackResult{
		Explanation: "payment-service explores payments bucket",
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		logjson.Error(fmt.Sprintf("could not create storage client: %s", err))
		OkFail(w, r, result, "error in exercise code or setup")
		return
	}

	iter := client.Bucket(getPaymentsBucket()).Objects(ctx, &storage.Query{})
	n, err := iter.Next()
	if err != nil && err != iterator.Done {
		OkFail(w, r, result, "")
		return
	}
	if n != nil {
		result.Loot = []Loot{{
			Key:  "storage-file",
			Data: n.Name,
		}}
	}
	result.Points = 10

	OkAttack(w, r, result)
}

func leakToken(w http.ResponseWriter, r *http.Request) {
	defer LogPanic()
	url := r.URL.Query().Get("url")
	idToken, err := GetIdToken(url)
	if err != nil {
		logjson.Error(fmt.Sprintf("could not get id token: %s", err))
		InternalServerError(w, "error in exercise code or setup")
		return
	}
	_, _ = w.Write([]byte(idToken))
}

func leakData(w http.ResponseWriter, r *http.Request) {
	defer LogPanic()
	client, err := GetFirestore()
	if err != nil {
		logjson.Error(fmt.Sprintf("could not create firestore client: %s", err))
		InternalServerError(w, "error in exercise code or setup")
		return
	}

	var data interface{}
	q := client.Collection(baseCollection + "/secrets").Limit(1)
	di := q.Documents(context.Background())
	for {
		d, err := di.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			break
		}
		data = d.Data()
		break
	}

	OkJson(w, data)
}
