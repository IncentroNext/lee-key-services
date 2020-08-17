package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/HayoVanLoon/go-commons/logjson"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type AttackSummary struct {
	Score         int            `json:"score"`
	AttackResults []AttackResult `json:"attackResults"`
}

type AttackInfo struct {
	Component string `json:"component"`
	Number    int    `json:"number"`
	Url       string `json:"url"`
}

const (
	websiteService = "website-service"
	orderService   = "order-service"
	paymentService = "payment-service"
	numberService  = "number-service"
	printService   = "print-service"
)

var services = []AttackInfo{
	{Component: websiteService, Number: 0, Url: websiteServiceScheme},
	{Component: orderService, Number: 100, Url: orderServiceScheme},
	{Component: paymentService, Number: 200, Url: paymentServiceScheme},
	{Component: numberService, Number: 300, Url: numberServiceScheme},
	{Component: printService, Number: 400, Url: printServiceScheme},
}

var routes = map[string][]string{
	websiteService: {},
	orderService:   {websiteServiceScheme},
	paymentService: {websiteServiceScheme},
	numberService:  {websiteServiceScheme, orderServiceScheme},
	printService:   {websiteServiceScheme, orderServiceScheme},
}

type handler struct {
	lastCheck time.Time
	attacks   map[int]AttackInfo
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer LogPanic()
	logjson.Info(fmt.Sprintf("incoming call from %s: %s", GetIdentification(r), r.URL.Path))

	if h.lastCheck.Add(60 * time.Second).Before(time.Now()) {
		h.attacks = discoverAttacks()
	}
	if len(r.URL.Path) >= 8 && r.URL.Path[:8] == "/attacks" {
		h.handleAttacks(w, r)
	} else if r.URL.Path == "/tests/normal" {
		h.handleTests(w, r)
	} else {
		http.ServeFile(w, r, "files"+r.URL.Path)
	}
}

func discoverAttacks() map[int]AttackInfo {
	atks := make(map[int]AttackInfo)
	for _, info := range services {
		target := info.Url + "/attacks"
		route := routes[info.Component]
		idToken, err := GetChainedToken(route, target)
		if err != nil {
			logjson.Error(fmt.Sprintf("could not fetch attack list from %s: %s", info.Url, err))
			continue
		}
		resp, err := GetWithAuth(target, idToken)
		if err != nil {
			logjson.Error(fmt.Sprintf("could not fetch attack list from %s: %s", info.Url, err))
			continue
		}
		if resp.StatusCode != http.StatusOK {
			logjson.Error(fmt.Sprintf("could not fetch attack list from %s: status %v", info.Url, resp.StatusCode))
			continue
		}
		bs, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logjson.Error(fmt.Sprintf("could decode attack list response from %s: %s", info.Component, err))
			continue
		}
		for _, a := range strings.Split(string(bs), ",") {
			an, err := strconv.Atoi(a)
			if err != nil {
				logjson.Error(fmt.Sprintf("not a number: %s", a))
				continue
			}
			atks[info.Number+an] = AttackInfo{
				Component: info.Component,
				Number:    an + info.Number,
				Url:       fmt.Sprintf("%s/attacks/%s", info.Url, a),
			}
		}
	}
	return atks
}

func (h handler) handleTests(w http.ResponseWriter, r *http.Request) {
	o := Order{
		Customer: "test-customer",
		Name:     "test",
		Quantity: 42,
	}
	bs, _ := json.Marshal(o)
	or, err := http.Post(websiteServiceScheme+"/orders", "application/json", bytes.NewReader(bs))
	if err != nil {
		msg := fmt.Sprintf("failed to create order: %s", err)
		logjson.Info(msg)
		_, _ = w.Write([]byte(msg))
		return
	} else if or.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("failed to create order: %v", or.StatusCode)
		logjson.Info(msg)
		_, _ = w.Write([]byte(msg))
		return
	}
	bs, err = ioutil.ReadAll(or.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to read order creation result: %s", err)
		logjson.Info(msg)
		_, _ = w.Write([]byte(msg))
		return
	}
	o2 := &Order{}
	err = json.Unmarshal(bs, o2)
	if err != nil {
		msg := fmt.Sprintf("failed to parse order creation result: %s", err)
		logjson.Info(msg)
		_, _ = w.Write([]byte(msg))
		return
	}
	p := Payment{
		OrderNumber: o2.OrderNumber,
	}
	bs, _ = json.Marshal(p)
	pr, err := http.Post(websiteServiceScheme+"/payments", "application/json", bytes.NewReader(bs))
	if err != nil {
		msg := fmt.Sprintf("failed to create payment at payment service: %s", err)
		logjson.Info(msg)
		_, _ = w.Write([]byte(msg))
		return
	}
	if pr.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("failed to create payment, got %v", pr.StatusCode)
		logjson.Info(msg)
		_, _ = w.Write([]byte(msg))
		return
	}
	bs, err = ioutil.ReadAll(pr.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to read response from payment creation call: %s", err)
		logjson.Info(msg)
		_, _ = w.Write([]byte(msg))
		return
	}
	p2 := &Payment{}
	err = json.Unmarshal(bs, p2)
	if err != nil {
		msg := fmt.Sprintf("failed to parse response from payment creation call: %s", err)
		logjson.Info(msg)
		_, _ = w.Write([]byte(msg))
		return
	}
	_, _ = w.Write([]byte("Success"))
}

func (h handler) handleAttacks(w http.ResponseWriter, r *http.Request) {
	sum := AttackSummary{}
	if r.URL.Path == "/attacks" {
		if r.Method == http.MethodGet {
			h.listAttacks(w, r)
			return
		}
		if filterOutMethod([]string{"GET", "POST"}, w, r) {
			return
		}
		for atk := range h.attacks {
			res, err := h.launchAttack(atk)
			if err != nil {
				res = &AttackResult{Explanation: fmt.Sprintf("error on attack %v: %s", atk, err)}
			}
			sum.AttackResults = append(sum.AttackResults, *res)
		}
	} else {
		if filterOutMethod([]string{"POST"}, w, r) {
			return
		}
		raw := r.URL.Path[len("/attacks/"):]
		attack, err := strconv.Atoi(raw)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(fmt.Sprintf("invalid attack name ", raw)))
			return
		}
		a, err := h.launchAttack(attack)
		if err != nil {
			log.Printf("attack %v failed with error: %s", attack, err)
		}
		sum.AttackResults = []AttackResult{*a}
	}
	for _, r := range sum.AttackResults {
		sum.Score += r.Points
	}

	OkJson(w, sum)
}

func (h handler) listAttacks(w http.ResponseWriter, _ *http.Request) {
	OkJson(w, h.attacks)
}

func (h handler) launchAttack(attack int) (*AttackResult, error) {
	info, ok := h.attacks[attack]
	if !ok {
		return nil, fmt.Errorf("unknown attack %v", attack)
	}
	route := routes[info.Component]
	idToken, err := GetChainedToken(route, info.Url)
	if err != nil {
		return nil, fmt.Errorf("could not fetch proper token: %s", err)
	}
	resp, err := GetWithAuth(info.Url, idToken)
	if err != nil {
		return nil, fmt.Errorf("could not initiate attack: %s", err)
	}
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading attack response: %s", err)
	}
	a := &AttackResult{}
	err = json.Unmarshal(bs, a)
	if err != nil {
		return nil, fmt.Errorf("error parsing attack response from %s: %s", info.Component, err)
	}
	return a, nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if websiteServiceScheme == "" {
		log.Fatal("WEBSITE_SERVICE not set")
	}

	h := &handler{
		lastCheck: time.Now(),
		attacks:   discoverAttacks(),
	}

	logjson.Info(fmt.Sprintf("fetched %v attacks", len(h.attacks)))

	http.Handle("/", h)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
