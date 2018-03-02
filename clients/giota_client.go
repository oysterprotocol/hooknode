package giotaClient

import (
	"os"
	"sync"
	"time"

	"github.com/iotaledger/giota"
	"fmt"
	"encoding/json"
	"bytes"
	"net/http"

	"github.com/getsentry/raven-go"
)

type PowJob struct {
	Transactions     []giota.Transaction
	TrunkTransaction giota.Trytes
	BranchTransacion giota.Trytes
	BroadcastNodes   []string
}

type broadcastRequest struct {
	Trytes []giota.Transaction `json:"trytes"`
}

// SendTrytes does attachToTangle and finally, it broadcasts the transactions.
func SendTrytes(trytes []giota.Trytes, trunk giota.Trytes, branch giota.Trytes, broadcastNodes []string, jobQueue chan PowJob) (
	ts []giota.Transaction, err error) {

	// Convert []Trytes to []Transaction
	txs := make([]giota.Transaction, len(trytes))
	for i, t := range trytes {
		tx, _ := giota.NewTransaction(t)
		txs[i] = *tx
	}

	powJobRequest := PowJob{
		BranchTransacion: branch,
		TrunkTransaction: trunk,
		Transactions:     txs,
		BroadcastNodes:   broadcastNodes,
	}

	jobQueue <- powJobRequest

	return txs, nil
}

func PowWorker(jobQueue <-chan PowJob, err error) {
	for powJobRequest := range jobQueue {
		// this is where we would call methods to deal with each job request
		fmt.Println("In PowWorker")

		provider := os.Getenv("PROVIDER")
		minDepth := int64(giota.DefaultNumberOfWalks)
		minWeightMag := int64(giota.DefaultMinWeightMagnitude)

		api := giota.NewAPI(provider, nil)

		err = doPow(powJobRequest.BranchTransacion, powJobRequest.TrunkTransaction, minDepth, powJobRequest.Transactions, minWeightMag)

		if err != nil {
			return
		}

		// Broadcast and store tx
		err = api.BroadcastTransactions(powJobRequest.Transactions)
		if err != nil {
			return
		}

		BroadcastTxs(&powJobRequest.Transactions, powJobRequest.BroadcastNodes)
	}
}

// Things below are copied from the giota lib since they are not public.
// https://github.com/iotaledger/giota/blob/master/transfer.go#L322

// (3^27-1)/2
const maxTimestampTrytes = "MMMMMMMMM"

// This mutex was added by us.
var mutex = &sync.Mutex{}

func doPow(branch giota.Trytes, trunk giota.Trytes, depth int64,
	trytes []giota.Transaction, mwm int64) error {

	var prev giota.Trytes
	var err error
	for i := len(trytes) - 1; i >= 0; i-- {
		switch {
		case i == len(trytes)-1:
			trytes[i].TrunkTransaction = trunk
			trytes[i].BranchTransaction = branch
		default:
			trytes[i].TrunkTransaction = prev
			trytes[i].BranchTransaction = trunk
		}

		timestamp := giota.Int2Trits(time.Now().UnixNano()/1000000, giota.TimestampTrinarySize).Trytes()
		trytes[i].AttachmentTimestamp = timestamp
		trytes[i].AttachmentTimestampLowerBound = ""
		trytes[i].AttachmentTimestampUpperBound = maxTimestampTrytes

		// We customized this to lock here.
		mutex.Lock()
		trytes[i].Nonce, err = giota.PowC(trytes[i].Trytes(), int(mwm))
		mutex.Unlock()

		if err != nil {
			return err
		}

		prev = trytes[i].Hash()
	}
	return nil
}

func BroadcastTxs(txs *[]giota.Transaction, nodes []string) {
	broadcastReq := broadcastRequest{
		Trytes: *txs,
	}
	jsonReq, err := json.Marshal(broadcastReq)
	if err != nil {
		raven.CaptureError(err, nil)
		return
	}
	reqBody := bytes.NewBuffer(jsonReq)

	for _, node := range nodes {
		nodeURL := "http://" + node + ":3000/broadcast/"

		// Async log
		// go segmentClient.Enqueue(analytics.Track{
		// 	Event:  "broadcast_to_other_hooknodes",
		// 	UserId: getLocalIP(),
		// 	Properties: analytics.NewProperties().
		// 		Set("addresses", mapTxsToAddrs(*txs)),
		// })

		// Async broadcasting
		go func() {
			_, err := http.Post(nodeURL, "application/json", reqBody)
			if err != nil {
				raven.CaptureError(err, nil)
			}
		}()

	}
}
