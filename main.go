package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

const password = "Basic dGVzdC11c2VyOnRlc3QtdXNlcg=="

type addresses struct {
	BCH string `json:"BCH"`
	BTC string `json:"BTC"`
	LTC string `json:"LTC"`
	ZEC string `json:"ZEC"`
}

type output struct {
	Address string `json:"address"`
	Value   int64  `json:"value"`
	Index   int    `json:"index"`
	OrderID string `json:"orderid"`
}

type amount struct {
	Amount string `json:"amount"`
}

var m addresses
var n bool
var price amount

func getNewAddress(w http.ResponseWriter, r *http.Request) {

	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	json.Unmarshal(d, &price)
	fmt.Println(price.Amount)

	n = false
	var url = "http://localhost:4002/wallet/address"

	req, _ := http.NewRequest("GET", url, nil)

	//Should use token
	req.Header.Set("Authorization", password)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(data, &m)
	fmt.Fprintf(w, fmt.Sprintf(`{"BCH":"%s"}`, m.BCH))

}

func transactionReceived(w http.ResponseWriter, r *http.Request) {

	d, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	a, _ := regexp.Compile(string(m.BCH))
	b, _ := regexp.Compile(price.Amount)
	fmt.Println(string(d))
	n = a.MatchString(string(d))
	n = b.MatchString(string(d))

}

func saleComplete(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, fmt.Sprintf(`{"bool":"%s"}`, strconv.FormatBool(n)))
}

func main() {

	//Init Router
	r := mux.NewRouter()

	r.HandleFunc("/api/getnewaddress/", getNewAddress).Methods("POST")
	r.HandleFunc("/api/transactionreceived/", transactionReceived).Methods("POST")
	r.HandleFunc("/api/salecomplete/", saleComplete).Methods("GET")

	handler := cors.Default().Handler(r)

	log.Fatal(http.ListenAndServe(":8000", handler))
}
