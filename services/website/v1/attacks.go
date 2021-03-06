package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/HayoVanLoon/go-commons/logjson"
	"google.golang.org/api/iterator"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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
	http.HandleFunc("/attacks/1", writeFirestore)
	http.HandleFunc("/attacks/2", readFirestore)
	http.HandleFunc("/attacks/3", connectToDeeperService)
	http.HandleFunc("/attacks/4", impersonatePaymentService)
	http.HandleFunc("/attacks/5", impersonatePaymentService2)
	http.HandleFunc("/attacks/6", shortChainToPrintService)
	http.HandleFunc("/attacks/7", longChainToPrintService)
	http.HandleFunc("/leaks/id-token", leakToken)
	http.HandleFunc("/attacks", listAttacks)
}

func listAttacks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Header().Set("allow", http.MethodGet)
		return
	}
	_, _ = w.Write([]byte("1,2,3,4,5,6,7"))
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

func leakToken(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	idToken, err := GetIdToken(url)
	if err != nil {
		logjson.Error(fmt.Sprintf("[attack] could not get id token: %s", err))
		InternalServerError(w, "error in exercise code or setup")
		return
	}
	_, _ = w.Write([]byte(idToken))
}

func connectToDeeperService(w http.ResponseWriter, r *http.Request) {
	defer LogPanic()
	log.Printf("starting %s for %s", r.URL.Path, getIp(r))
	result := &AttackResult{
		Explanation: "website-service reaches deeper services",
	}

	url := numberServiceScheme + "/ranges/payment"
	resp, err := GetWithAuth(url, "")
	if err != nil {
		logjson.Info(fmt.Sprintf("[attack] could not call %s: %s", url, err))
		OkFail(w, r, result, "")
	} else if resp.StatusCode != http.StatusOK {
		OkFail(w, r, result, fmt.Sprintf("[attack] could not perform call to %s", numberServiceScheme))
	} else {
		result.Points = 10
		OkAttack(w, r, result)
	}
}

func writeFirestore(w http.ResponseWriter, r *http.Request) {
	defer LogPanic()
	log.Printf("starting %s for %s", r.URL.Path, getIp(r))
	result := &AttackResult{
		Explanation: "website-service updates firestore",
	}

	client, err := GetFirestore()
	if err != nil {
		logjson.Error(fmt.Sprintf("[attack] could not create firestore client: %s", err))
		OkFail(w, r, result, "error in exercise code or setup")
		return
	}

	o := Order{
		Customer:    "hacker-" + getIp(r),
		Name:        "loot",
		Quantity:    1234567890,
		OrderNumber: 666,
	}
	d := client.Collection(baseCollection + "/orders").Doc(strconv.Itoa(o.OrderNumber))
	_, err = d.Create(context.Background(), o)
	if err == nil {
		result.Points = 1000
		OkAttack(w, r, result)
	} else {
		OkFail(w, r, result, "")
	}
}

func readFirestore(w http.ResponseWriter, r *http.Request) {
	defer LogPanic()
	log.Printf("starting %s for %s", r.URL.Path, getIp(r))
	result := &AttackResult{
		Explanation: "website-service reads from firestore",
	}

	client, err := GetFirestore()
	if err != nil {
		logjson.Error(fmt.Sprintf("[attack] could not create firestore client: %s", err))
		OkFail(w, r, result, "error in exercise code or setup")
		return
	}

	q := client.Collection("orders").Limit(1)
	di := q.Documents(context.Background())
	d, err := di.Next()
	if err != nil && err != iterator.Done {
		logjson.Info(fmt.Sprintf("[attack] error querying firestore: %s", err))
		OkFail(w, r, result, "")
		return
	}

	if d != nil {
		loot, err := json.Marshal(d.Data())
		if err == nil {
			result.Loot = []Loot{{"firestore record", string(loot)}}
		}
	}
	result.Points = 100
	OkAttack(w, r, result)
}

func impersonatePaymentService(w http.ResponseWriter, r *http.Request) {
	defer LogPanic()
	log.Printf("starting %s for %s", r.URL.Path, getIp(r))
	result := &AttackResult{
		Explanation: "website-service abuses leaked id token",
	}

	target := numberServiceScheme + "/ranges/payment"
	idToken, err := GetServiceIdToken(paymentServiceScheme, target, "")
	if err != nil {
		logjson.Error(fmt.Sprintf("[attack] could not fetch id token from %s: %s", paymentServiceScheme, err))
		OkFail(w, r, result, "")
		return
	}

	resp, err := GetWithAuth(target, idToken)
	if err == nil && resp.StatusCode == http.StatusOK {
		result.Points = 10
		OkAttack(w, r, result)
	} else {
		OkFail(w, r, result, "")
	}
}

func impersonatePaymentService2(w http.ResponseWriter, r *http.Request) {
	defer LogPanic()
	log.Printf("starting %s for %s", r.URL.Path, getIp(r))
	result := &AttackResult{
		Explanation: "website-service abuses leaked id token and gets data",
	}

	target := orderServiceScheme + "/leaks/data"
	idToken, err := GetServiceIdToken(paymentServiceScheme, target, "")
	if err != nil {
		msg := fmt.Sprintf("[attack] could not fetch id token from %s: %s", paymentServiceScheme, err)
		logjson.Error(msg)
		OkFail(w, r, result, msg)
		return
	}

	resp, err := GetWithAuth(target, idToken)
	if err == nil && resp.StatusCode == http.StatusOK {
		bs, _ := ioutil.ReadAll(resp.Body)
		result.Loot = []Loot{{Key: "data", Data: string(bs)}}
		result.Points = 100
		OkAttack(w, r, result)
	} else {
		OkFail(w, r, result, "")
	}
}

func shortChainToPrintService(w http.ResponseWriter, r *http.Request) {
	defer LogPanic()
	log.Printf("starting %s for %s", r.URL.Path, getIp(r))
	result := &AttackResult{
		Explanation: "website-service chains leaked id tokens and writes invoice data",
	}
	target := printServiceScheme + "/invoices"
	idToken, err := GetServiceIdToken(orderServiceScheme, target, "")
	if err != nil {
		logjson.Error(fmt.Sprintf("could not fetch idToken from %s: %s", orderServiceScheme, err))
		OkFail(w, r, result, "")
		return
	}

	o := &Order{
		Customer:    "hackert",
		Name:        "flag",
		Quantity:    666,
		OrderNumber: 666,
	}
	bs, _ := json.Marshal(o)
	resp, err := PostJsonWithAuth(target, bytes.NewReader(bs), idToken)
	if err != nil {
		logjson.Error(fmt.Sprintf("could upload data to %s: %s", printServiceScheme, err))
		OkFail(w, r, result, "")
		return
	}
	if resp.StatusCode == 200 {
		result.Points = 1000
		bs, _ = ioutil.ReadAll(resp.Body)
		result.Loot = []Loot{{"invoice", string(bs)}}
		OkAttack(w, r, result)
	} else {
		OkFail(w, r, result, "")
	}
}

func longChainToPrintService(w http.ResponseWriter, r *http.Request) {
	defer LogPanic()
	log.Printf("starting %s for %s", r.URL.Path, getIp(r))
	result := &AttackResult{
		Explanation: "website-service chains leaked id tokens and writes distant invoice data",
	}
	target := printServiceScheme + "/invoices"
	idToken, err := GetChainedToken([]string{paymentServiceScheme, orderServiceScheme}, target)
	if err != nil {
		logjson.Error(fmt.Sprintf("[attack] could not fetch id token from chain: %s", err))
		OkFail(w, r, result, "")
		return
	}

	o := &Order{
		Customer:    "hackert",
		Name:        "flag",
		Quantity:    666,
		OrderNumber: 666,
	}
	bs, _ := json.Marshal(o)
	resp, err := PostJsonWithAuth(target, bytes.NewReader(bs), idToken)
	if err != nil {
		logjson.Error(fmt.Sprintf("[attack] could not upload data to %s: %s", printServiceScheme, err))
		OkFail(w, r, result, "")
		return
	}
	if resp.StatusCode == 200 {
		result.Points = 1000
		bs, _ = ioutil.ReadAll(resp.Body)
		result.Loot = []Loot{{"invoice", string(bs)}}
		OkAttack(w, r, result)
	} else {
		OkFail(w, r, result, "")
	}
}
