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
	"github.com/oysterprotocol/hooknode/utils"
	"gopkg.in/segmentio/analytics-go.v3"

	"github.com/getsentry/raven-go"
)

type PowJob struct {
	Transfers        []giota.Transfer
	TrunkTransaction giota.Trytes
	BranchTransacion giota.Trytes
	BroadcastNodes   []string
}

type broadcastRequest struct {
	Trytes []giota.Transaction `json:"trytes"`
}

// SendTrytes does attachToTangle and finally, it broadcasts the transactions.
func SendTrytes(transfers []giota.Transfer, trunk giota.Trytes, branch giota.Trytes, broadcastNodes []string, jobQueue chan PowJob) (
	err error) {

	powJobRequest := PowJob{
		BranchTransacion: branch,
		TrunkTransaction: trunk,
		Transfers:        transfers,
		BroadcastNodes:   broadcastNodes,
	}

	jobQueue <- powJobRequest

	return nil
}

func PowWorker(jobQueue <-chan PowJob, err error) {
	for powJobRequest := range jobQueue {
		// this is where we would call methods to deal with each job request
		fmt.Println("In PowWorker")

		provider := os.Getenv("PROVIDER")
		minDepth := int64(giota.DefaultNumberOfWalks)
		minWeightMag := int64(giota.DefaultMinWeightMagnitude)

		api := giota.NewAPI(provider, nil)
		_, pow := giota.GetBestPoW()

		var seed giota.Trytes
		seed = "OYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRL"

		var bdl giota.Bundle

		bdl, err = giota.PrepareTransfers(api, seed, powJobRequest.Transfers, nil, "", 2)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-1)
		}

		transactions := []giota.Transaction(bdl)

		err = doPow(powJobRequest.BranchTransacion, powJobRequest.TrunkTransaction, minDepth, transactions, minWeightMag, pow)

		if err != nil {
			return
		}

		// Broadcast and store tx

		err = api.BroadcastTransactions(transactions)
		if err != nil {
			return
		}

		go BroadcastTxs(&transactions, powJobRequest.BroadcastNodes)
	}
}

// Things below are copied from the giota lib since they are not public.
// https://github.com/iotaledger/giota/blob/master/transfer.go#L322

// (3^27-1)/2
const maxTimestampTrytes = "MMMMMMMMM"

// This mutex was added by us.
var mutex = &sync.Mutex{}

func doPow(branch giota.Trytes, trunk giota.Trytes, depth int64,
	trytes []giota.Transaction, mwm int64, pow giota.PowFunc) error {

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
		trytes[i].Nonce, err = pow(trytes[i].Trytes(), int(mwm))
		mutex.Unlock()

		if err != nil {
			return err
		}

		prev = trytes[i].Hash()
	}

	// Async log
	go oysterUtils.SegmentClient.Enqueue(analytics.Track{
		Event:  "performed_pow",
		UserId: oysterUtils.GetLocalIP(),
		Properties: analytics.NewProperties().
			Set("addresses", oysterUtils.MapTransactionsToAddrs(trytes)),
	})

	return nil
}

func BroadcastTxs(txs *[]giota.Transaction, nodes []string) {
	broadcastReq := broadcastRequest{
		Trytes: *txs,
	}

	for _, node := range nodes {

		jsonReq, err := json.Marshal(broadcastReq)
		if err != nil {
			raven.CaptureError(err, nil)
			return
		}

		reqBody := bytes.NewBuffer(jsonReq)

		nodeURL := "http://" + node + ":3000/broadcast/"

		// Async log
		 go oysterUtils.SegmentClient.Enqueue(analytics.Track{
		 	Event:  "broadcast_to_other_hooknodes",
		 	UserId: oysterUtils.GetLocalIP(),
		 	Properties: analytics.NewProperties().
		 		Set("addresses", oysterUtils.MapTransactionsToAddrs(*txs)),
		 })

		// Async broadcasting
		go func() {
			_, err := http.Post(nodeURL, "application/json", reqBody)
			if err != nil {
				raven.CaptureError(err, nil)
			}
		}()

	}
}
