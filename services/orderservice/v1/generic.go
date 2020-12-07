package main

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/HayoVanLoon/go-commons/logjson"
	"github.com/HayoVanLoon/metadataemu"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var websiteServiceScheme = os.Getenv("WEBSITE_SERVICE")
var orderServiceScheme = os.Getenv("ORDER_SERVICE")
var paymentServiceScheme = os.Getenv("PAYMENT_SERVICE")
var numberServiceScheme = os.Getenv("NUMBER_SERVICE")
var printServiceScheme = os.Getenv("PRINT_SERVICE")

var localEnvironment = os.Getenv("LOCAL_ENVIRONMENT")
var localMetadataKey = os.Getenv("LOCAL_METADATA_KEY")
var localMetadataPort = os.Getenv("LOCAL_METADATA_PORT")
var metadataClient = metadataemu.NewClient(localMetadataPort, localMetadataKey, localEnvironment == "")

const baseCollection = "exercises/leekeyservices"

type Order struct {
	Customer    string `json:"customer"`
	Name        string `json:"name"`
	Quantity    int    `json:"quantity"`
	OrderNumber int    `json:"orderNumber"`
}

type Payment struct {
	OrderNumber   int `json:"orderNumber"`
	PaymentNumber int `json:"paymentNumber"`
}

type Money struct {
	Value    int
	Decimals int
	Currency string
}

func NewMoney(a, b int) Money {
	val := (a + b/100) * 100
	val += b - b/100*100
	return Money{
		Value:    val,
		Decimals: 2,
		Currency: "bottle caps",
	}
}

func (m Money) String() string {
	if m.Value < 100 {
		b := strconv.Itoa(100 + m.Value)[:2]
		return fmt.Sprintf("0.%s %s", b, m.Currency)
	} else {
		s := strconv.Itoa(m.Value)
		return fmt.Sprintf("%s.%s %s", s[2:], s[:2], m.Currency)
	}
}

type Invoice struct {
	Customer      string
	InvoiceNumber int
	Total         Money
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("OK"))
}

func getProjectId() string {
	project := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if project == "" {
		project, _ = metadataClient.ProjectID()
	}
	return project
}

func GetFirestore() (*firestore.Client, error) {
	project := getProjectId()
	if project == "" {
		return nil, fmt.Errorf("no project id available")
	}
	c, err := firestore.NewClient(context.Background(), project)
	return c, err
}

func LogPanic() {
	if r := recover(); r != nil {
		logjson.Error(fmt.Sprintf("%s", r))
	}
}

func InternalServerError(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("content-type", "application/json")
	_, _ = w.Write([]byte(fmt.Sprintf(`{"message": "%s"}`, msg)))
}

func OkJson(w http.ResponseWriter, v interface{}) {
	w.Header().Set("content-type", "application/json")
	bs, _ := json.Marshal(v)
	_, _ = w.Write(bs)
}

func GetIdToken(target string) (string, error) {
	path := fmt.Sprintf("%s?audience=%s", metadataemu.EndPointIdToken, target)
	return metadataClient.Get(path)
}

func HeadWithAuth(url string) (*http.Response, error) {
	return DoWithAuth(http.MethodHead, url, nil, "", "")
}

func GetWithAuth(url, token string) (*http.Response, error) {
	return DoWithAuth(http.MethodGet, url, nil, "", token)
}

func PostJsonWithAuth(url string, body io.Reader, token string) (*http.Response, error) {
	return DoWithAuth(http.MethodPost, url, body, "application/json", token)
}

func DoWithAuth(method, url string, body io.Reader, contentType, idToken string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err)
	}

	if idToken == "" {
		idToken, err = GetIdToken(url)
		if err != nil {
			return nil, err
		}
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", idToken))

	if contentType != "" {
		req.Header.Set("content-type", contentType)
	}
	return http.DefaultClient.Do(req)
}

func getIp(r *http.Request) string {
	forw := r.Header.Get("x-forwarded-for")
	if forw != "" {
		return forw
	}
	return r.RemoteAddr
}

func GetIdentification(r *http.Request) string {
	email, _ := GetOpenIdEmail(r)
	if email != "" {
		return email
	}
	sub, _ := GetOpenIdSub(r)
	if sub != "" {
		return email
	}
	return getIp(r)
}

func GetOpenIdSub(r *http.Request) (string, error) {
	return GetOpenIdString(r, "sub")
}

func GetOpenIdEmail(r *http.Request) (string, error) {
	return GetOpenIdString(r, "email")
}

func GetOpenIdString(r *http.Request, field string) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return "", fmt.Errorf("no authorization header")
	}
	if strings.ToLower(auth[0:7]) != "bearer " {
		return "", fmt.Errorf("no bearer token")
	}

	ss := strings.Split(auth[7:], ".")
	if len(ss) < 2 {
		return "", fmt.Errorf("less than two parts in encoded token")
	}
	s, err := base64.RawURLEncoding.DecodeString(ss[1])
	if err != nil {
		return "", fmt.Errorf("error decoding token: %s", err)
	}

	var i interface{}
	err = json.Unmarshal(s, &i)
	if err == nil {
		switch v := i.(type) {
		case map[string]interface{}:
			if v2, ok := v[field]; ok {
				switch value := v2.(type) {
				case string:
					return value, nil
				default:
					return "", fmt.Errorf("%s is not a string value: %v", field, s)
				}
			}
		default:
			return "", fmt.Errorf("invalid token content: %s", s)
		}
	}
	return "", nil
}

func getInvoicesBucket() string {
	return getProjectId() + "-leekeyservices-invoices"
}

func getPaymentsBucket() string {
	return getProjectId() + "-leekeyservices-payments"
}

func filterOutMethod(allow []string, w http.ResponseWriter, r *http.Request) bool {
	for _, m := range allow {
		if r.Method == m {
			return false
		}
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Header().Set("allow", strings.Join(allow, ","))
	return true
}

func GetNextNumber(key string) (int, error) {
	r, err := GetWithAuth(fmt.Sprintf("%s/ranges/%s", numberServiceScheme, key), "")
	if err != nil {
		return 0, fmt.Errorf("error calling random service at %s: %s", numberServiceScheme, err)
	}
	if r.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code %v", r.StatusCode)
	}
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return 0, fmt.Errorf("error reading response body: %s", err)
	}
	num, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("invalid response content %s", num)
	}
	return num, nil
}
