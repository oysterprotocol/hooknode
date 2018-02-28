package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"

	"github.com/getsentry/raven-go"
	"github.com/iotaledger/giota"
	"github.com/joho/godotenv"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

type indexRequest struct {
	Trytes            []giota.Trytes `json:"trytes"`
	TrunkTransaction  giota.Trytes   `json:"trunkTransaction"`
	BranchTransaction giota.Trytes   `json:"branchTransaction"`
	Command           string         `json:"command"`
	BroadcastNodes    []string       `json:"broadcastingNodes"`
}

type broadcastRequest struct {
	Trytes []giota.Transaction `json:"trytes"`
}

func init() {
	// Load ENV variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Setup sentry
	raven.SetDSN(os.Getenv("SENTRY_DSN"))
}

func main() {
	raven.CapturePanic(func() {

		// Attach handlers
		http.HandleFunc("/", raven.RecoveryHandler(indexHandler))
		http.HandleFunc("/broadcast", raven.RecoveryHandler(broadcastHandler))
		http.HandleFunc("/stats", raven.RecoveryHandler(statsHandler))
		http.HandleFunc("/pow", powHandler)
		http.HandleFunc("/sentry", raven.RecoveryHandler(sentryHandler))

		// Fetch port from ENV
		port := os.Getenv("PORT")
		fmt.Printf("\nListening on http://localhost:%v\n", port)

		// Start listening
		http.ListenAndServe(":"+port, nil)

	}, nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Print("\nProcessing trytes\n")
	if r.Method == "POST" {

		// Unmarshal JSON
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			raven.CaptureError(err, nil)
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		req := indexRequest{}
		json.Unmarshal(b, &req)

		go attachAndBroadcastToTangle(&req)

		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func attachAndBroadcastToTangle(indexReq *indexRequest) {
	// provider := os.Getenv("PROVIDER")
	// minDepth, _ := strconv.ParseInt(os.Getenv("MIN_DEPTH"), 10, 64)
	// api := giota.NewAPI(provider, nil)
	minWeightMag, _ := strconv.ParseInt(os.Getenv("MIN_WEIGHT_MAGNITUDE"), 10, 0)

	// Convert []Trytes to []Transaction
	// txs := make([]giota.Transaction, len(indexReq.Trytes))
	// for i, t := range indexReq.Trytes {
	// 	tx, _ := giota.NewTransaction(t)
	// 	txs[i] = *tx
	// }

	// _, pow := giota.GetBestPoW()

	// attachReq := giota.AttachToTangleRequest{
	// 	Command:            "attachToTangle",
	// 	TrunkTransaction:   indexReq.TrunkTransaction,
	// 	BranchTransaction:  indexReq.BranchTransaction,
	// 	MinWeightMagnitude: minWeightMag,
	// 	Trytes:             txs,
	// }

	// attachRes, err := api.AttachToTangle(&attachReq)
	// if err != nil {
	// 	raven.CaptureError(err, nil)
	// 	return
	// }

	// Convert []Trytes to []Transaction
	txs := make([]giota.Transaction, len(indexReq.Trytes))
	for i, t := range indexReq.Trytes {
		powT, err := giota.PowSSE(t, int(minWeightMag))
		if err != nil {
			raven.CaptureError(err, nil)
			return
		}
		tx, _ := giota.NewTransaction(powT)
		txs[i] = *tx
	}

	// Broadcast trytes.

	// Broadcast on self
	go broadcastAndStore(&txs)

	// Broadcast to other hooknodes
	broadcastReq := broadcastRequest{
		Trytes: txs,
	}
	jsonReq, err := json.Marshal(broadcastReq)
	if err != nil {
		raven.CaptureError(err, nil)
		return
	}
	reqBody := bytes.NewBuffer(jsonReq)

	for _, node := range indexReq.BroadcastNodes {
		nodeURL := "http://" + node + ":3000/broadcast"
		// Async broadcasting
		go func() {
			_, err := http.Post(nodeURL, "application/json", reqBody)
			if err != nil {
				raven.CaptureError(err, nil)
			}
		}()

	}

}

func broadcastHandler(w http.ResponseWriter, r *http.Request) {
	// Unmarshal JSON
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		raven.CaptureError(err, nil)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	req := broadcastRequest{}
	json.Unmarshal(b, &req)

	go broadcastAndStore(&req.Trytes)

	w.WriteHeader(http.StatusNoContent)
}

func broadcastAndStore(txs *[]giota.Transaction) {
	provider := os.Getenv("PROVIDER")
	api := giota.NewAPI(provider, nil)

	// Broadcast
	err := api.BroadcastTransactions(*txs)
	if err != nil {
		raven.CaptureError(err, nil)
		return
	}

	// Store
	err = api.StoreTransactions(*txs)
	if err != nil {
		raven.CaptureError(err, nil)
		return
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

func statsHandler(w http.ResponseWriter, r *http.Request) {
	c, _ := cpu.Percent(0, false)
	l, _ := load.Avg()
	m, _ := mem.VirtualMemory()

	body := map[string]interface{}{
		"cpu": map[string]interface{}{
			"avgPercent": c[0],
		},
		"load":   l,
		"memory": m,
	}

	res, _ := json.Marshal(body)

	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func sentryHandler(w http.ResponseWriter, r *http.Request) {
	// TESTING Error
	err := errors.New("TESTING SENTRY")
	go raven.CaptureError(err, nil)

	w.WriteHeader(http.StatusNoContent)
}
