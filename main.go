package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
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
	"gopkg.in/segmentio/analytics-go.v3"
)

type indexRequest struct {
	Trytes            []giota.Trytes `json:"trytes"`
	//Trytes            []string `json:"trytes"`
	TrunkTransaction  giota.Trytes   `json:"trunkTransaction"`
	BranchTransaction giota.Trytes   `json:"branchTransaction"`
	Command           string         `json:"command"`
	BroadcastNodes    []string       `json:"broadcastingNodes"`
}

type broadcastRequest struct {
	Trytes []giota.Transaction `json:"trytes"`
}

var segmentClient analytics.Client

func init() {
	// Load ENV variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Setup sentry
	raven.SetDSN(os.Getenv("SENTRY_DSN"))

	// Setup Segment
	segmentClient = analytics.New(os.Getenv("SEGMENT_WRITE_KEY"))
}

func main() {
	raven.CapturePanic(func() {

		// Attach handlers
		//http.HandleFunc("/attach/", raven.RecoveryHandler(attachHandler))
		http.HandleFunc("/attach/", raven.RecoveryHandler(attachHandler2))
		http.HandleFunc("/broadcast/", raven.RecoveryHandler(broadcastHandler))
		http.HandleFunc("/stats/", raven.RecoveryHandler(statsHandler))
		http.HandleFunc("/pow/", powHandler)
		http.HandleFunc("/sentry/", raven.RecoveryHandler(sentryHandler))
		http.HandleFunc("/version/", raven.RecoveryHandler(versionHandler))

		// Fetch port from ENV
		port := os.Getenv("PORT")
		fmt.Printf("\nListening on http://localhost:%v\n", port)

		// Start listening
		http.ListenAndServe(":"+port, nil)

	}, nil)
}

func attachHandler2(w http.ResponseWriter, r *http.Request) {
	fmt.Print("\nattachHandler2\n")
	fmt.Print("\nProcessing trytes\n")
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

		//
		fmt.Println(len(req.Trytes))
		//

		for i, t := range req.Trytes {
			tx, _ := giota.NewTransaction(t)
			txs[i] = *tx
		}

		fmt.Println(txs)

		// Get configuration.
		provider := os.Getenv("PROVIDER")
		minDepth, _ := strconv.ParseInt(os.Getenv("MIN_DEPTH"), 10, 64)
		minWeightMag, _ := strconv.ParseInt(os.Getenv("MIN_WEIGHT_MAGNITUDE"), 10, 64)

		// Async sendTrytes
		api := giota.NewAPI(provider, nil)
		_, pow := giota.GetBestPoW()

		fmt.Print("Sending Transactions...\n")
		go func() {
			e := CustomSendTrytes(api, minDepth, txs, minWeightMag, pow)
			raven.CaptureError(e, nil)
		}()

		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

// SendTrytes does attachToTangle and finally, it broadcasts the transactions.
func CustomSendTrytes(api *API, depth int64, trytes []Transaction, mwm int64, pow PowFunc) error {
	tra, err := api.GetTransactionsToApprove(depth, DefaultNumberOfWalks, "")
	if err != nil {
		return err
	}

	switch {
	case pow == nil:
		at := AttachToTangleRequest{
			TrunkTransaction:   tra.TrunkTransaction,
			BranchTransaction:  tra.BranchTransaction,
			MinWeightMagnitude: mwm,
			Trytes:             trytes,
		}

		// attach to tangle - do pow
		attached, err := api.AttachToTangle(&at)
		if err != nil {
			return err
		}

		trytes = attached.Trytes
	default:
		err := doPow(tra, depth, trytes, mwm, pow)
		if err != nil {
			return err
		}
	}

	// Broadcast and store tx
	return api.BroadcastTransactions(trytes)
}

func attachHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Print("\nattachHandler\n")
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

		w.Header().Set("Content-Type", "application/json")
		w.Write(successJSON())
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func attachAndBroadcastToTangle(indexReq *indexRequest) {
	minWeightMag, _ := strconv.ParseInt(os.Getenv("MIN_WEIGHT_MAGNITUDE"), 10, 0)

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

		// Async log
		go func() {
			segmentClient.Enqueue(analytics.Track{
				Event:  "broadcast_to_other_hooknodes",
				UserId: getLocalIP(),
				Properties: analytics.NewProperties().
					Set("addresses", mapTxsToAddrs(txs)),
			})
		}()

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
	fmt.Print("\nbrodcastHandler\n")

	// HACK: Using sentry for logs.
	err := errors.New("/broadcast")
	go raven.CaptureError(err, nil)

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

	w.Header().Set("Content-Type", "application/json")
	w.Write(successJSON())
}

func broadcastAndStore(txs *[]giota.Transaction) {
	provider := os.Getenv("PROVIDER")
	api := giota.NewAPI(provider, nil)

	// Async log
	go func() {
		segmentClient.Enqueue(analytics.Track{
			Event:  "broadcast_transactions",
			UserId: getLocalIP(),
			Properties: analytics.NewProperties().
				Set("addresses", mapTxsToAddrs(*txs)),
		})
	}()

	// Broadcast
	err := api.BroadcastTransactions(*txs)
	if err != nil {
		raven.CaptureError(err, nil)
		return
	}

	// Async log
	go func() {
		segmentClient.Enqueue(analytics.Track{
			Event:  "store_transactions",
			UserId: getLocalIP(),
			Properties: analytics.NewProperties().
				Set("addresses", mapTxsToAddrs(*txs)),
		})
	}()

	// Store
	err = api.StoreTransactions(*txs)
	if err != nil {
		raven.CaptureError(err, nil)
		return
	}
}

func powHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Print("\npowHandler\n")

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
	fmt.Print("\nstatsHandler\n")

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
	fmt.Print("\nsentryHandler\n")

	// TESTING Error
	err := errors.New("TESTING SENTRY")
	go raven.CaptureError(err, nil)

	w.Header().Set("Content-Type", "application/json")
	w.Write(successJSON())
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Print("\nversionHandler\n")

	gitCommit := os.Getenv("GIT_COMMIT")
	if gitCommit == "" {
		gitCommit = "Error: GIT_COMMIT not set"
	}

	res, _ := json.Marshal(map[string]string{"lastGitCommit": gitCommit})
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func successJSON() (res []byte) {
	res, _ = json.Marshal(map[string]bool{"success": true})
	return
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func mapTxsToAddrs(txs []giota.Transaction) (addrs []giota.Address) {
	for i, tx := range txs {
		addrs[i] = tx.Address
	}

	return
}
