package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"

	"github.com/iotaledger/giota"
	"github.com/joho/godotenv"
)

type indexRequest struct {
	Trytes []giota.Trytes `json:"trytes"`
}

func main() {
	// Load ENV variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Attach handlers
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/pow", powHandler)

	// Fetch port from ENV
	port := os.Getenv("PORT")
	fmt.Print("Listing on http://localhost:" + port)

	// Start listening
	http.ListenAndServe(":"+port, nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		// Unmarshal JSON
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Invalid request method", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		req := indexRequest{}
		json.Unmarshal(b, &req)

		// Convert []Trytes to []Transaction
		txs := make([]giota.Transaction, len(req.Trytes))
		for i, t := range req.Trytes {
			tx, _ := giota.NewTransaction(t)
			txs[i] = *tx
		}

		// Get configuration.
		provider := os.Getenv("PROVIDER")
		minDepth, _ := strconv.ParseInt(os.Getenv("MIN_DEPTH"), 10, 64)
		minWeightMag, _ := strconv.ParseInt(os.Getenv("MIN_WEIGHT_MAGNITUDE"), 10, 64)

		// Async sendTrytes
		api := giota.NewAPI(provider, nil)
		_, pow := giota.GetBestPoW()

		go giota.SendTrytes(api, minDepth, txs, minWeightMag, pow)

		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func powHandler(w http.ResponseWriter, r *http.Request) {
	_, pow := giota.GetBestPoW()

	// TODO: Figure out how to print the func name.
	body, err :=
		json.Marshal(map[string]interface{}{"powAlgo": getFuncName(pow)})

	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func getFuncName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
