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
	//"errors"
)

type PowJob struct {
	Transactions      []giota.Transaction
	TrunkTransaction  giota.Trytes
	BranchTransaction giota.Trytes
	BroadcastNodes    []string
}

type broadcastRequest struct {
	Trytes []giota.Transaction `json:"trytes"`
}

// PublicNodes is a list of known public nodes from http://iotasupport.com/lightwallet.shtml.
var PublicNodes = []string{
		// "http://service.iotasupport.com:14265",
		"http://eugene.iota.community:14265",
		"http://eugene.iotasupport.com:14999",
		"http://eugeneoldisoft.iotasupport.com:14265",
		"http://mainnet.necropaz.com:14500",
		"http://iotatoken.nl:14265",
		// "http://iota.digits.blue:14265",
		// "http://wallets.iotamexico.com:80",
		"http://5.9.137.199:14265",
		"http://5.9.118.112:14265",
		"http://5.9.149.169:14265",
		"http://88.198.230.98:14265",
		"http://176.9.3.149:14265",
		"http://iota.bitfinex.com:80",
	}


var seed giota.Trytes

var provider = os.Getenv("PROVIDER")
var minDepth = int64(giota.DefaultNumberOfWalks)
var minWeightMag = int64(giota.DefaultMinWeightMagnitude)

var api = giota.NewAPI(provider, nil)
var bestPow giota.PowFunc
var powName string

// Things below are copied from the giota lib since they are not public.
// https://github.com/iotaledger/giota/blob/master/transfer.go#L322

// (3^27-1)/2
const maxTimestampTrytes = "MMMMMMMMM"

// This mutex was added by us.
var mutex = &sync.Mutex{}

func init() {
	seed = "OYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRLOYSTERPRL"

	powName, bestPow = giota.GetBestPoW()
}

// SendTrytes does prepareTransfers, then attachToTangle, and finally, it broadcasts the transactions.
func SendTrytes(transfers []giota.Transfer, trunk giota.Trytes, branch giota.Trytes, broadcastNodes []string, jobQueue chan PowJob) (
	err error) {

	defer oysterUtils.TimeTrack(time.Now(), "prepareTransfers_performed", analytics.NewProperties().
		Set("addresses", oysterUtils.MapTransfersToAddrs(transfers)))

	var bdl giota.Bundle

	bdl, err = giota.PrepareTransfers(api, seed, transfers, nil, "", 1)

	if err != nil {
		raven.CaptureError(err, nil)
		return
	}

	transactions := []giota.Transaction(bdl)

	powJobRequest := PowJob{
		BranchTransaction: branch,
		TrunkTransaction:  trunk,
		Transactions:      transactions,
		BroadcastNodes:    broadcastNodes,
	}

	jobQueue <- powJobRequest

	return nil
}

func PowWorker(jobQueue <-chan PowJob, err error) {
	for powJobRequest := range jobQueue {
		// this is where we would call methods to deal with each job request
		fmt.Println("PowWorker: Starting")

		err = doPowAndBroadcast(
			powJobRequest.BranchTransaction,
			powJobRequest.TrunkTransaction,
			minDepth,
			powJobRequest.Transactions,
			minWeightMag,
			bestPow,
			powJobRequest.BroadcastNodes)

		if err != nil {
			raven.CaptureError(err, nil)
			return
		}

		fmt.Println("PowWorker: Leaving")
	}
}

func doPowAndBroadcast(branch giota.Trytes, trunk giota.Trytes, depth int64,
	trytes []giota.Transaction, mwm int64, bestPow giota.PowFunc, broadcastNodes []string) error {

	defer oysterUtils.TimeTrack(time.Now(), "doPow_using_" + powName, analytics.NewProperties().
		Set("addresses", oysterUtils.MapTransactionsToAddrs(trytes)))

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
		trytes[i].Nonce, err = bestPow(trytes[i].Trytes(), int(mwm))
		mutex.Unlock()

		if err != nil {
			raven.CaptureError(err, nil)
			return err
		}

		prev = trytes[i].Hash()
	}

	go func(branch giota.Trytes, trunk giota.Trytes, depth int64,
		trytes []giota.Transaction, mwm int64, bestPow giota.PowFunc, broadcastNodes []string) {

		err = api.BroadcastTransactions(trytes)

		if err != nil {

			// Async log
			go oysterUtils.SegmentClient.Enqueue(analytics.Track{
				Event:  "broadcast_fail_redoing_pow",
				UserId: oysterUtils.GetLocalIP(),
				Properties: analytics.NewProperties().
					Set("addresses", oysterUtils.MapTransactionsToAddrs(trytes)),
			})

			raven.CaptureError(err, nil)
		} else {
			go BroadcastTxs(&trytes, broadcastNodes)

			go oysterUtils.SegmentClient.Enqueue(analytics.Track{
				Event:  "broadcast_success",
				UserId: oysterUtils.GetLocalIP(),
				Properties: analytics.NewProperties().
					Set("addresses", oysterUtils.MapTransactionsToAddrs(trytes)),
			})
		}
	}(branch, trunk, depth, trytes, mwm, bestPow, broadcastNodes)

	return nil
}

func BroadcastTxs(txs *[]giota.Transaction, nodes []string) {
	broadcastReq := broadcastRequest{
		Trytes: *txs,
	}

	/*
	The time tracking for the log below may not be accurate because
	the broadcasts occur in a go-routine
	 */

	//defer oysterUtils.TimeTrack(time.Now(), "Rebroadcast requests made", analytics.NewProperties().
	//	Set("addresses", oysterUtils.MapTransactionsToAddrs(*txs)).
	//	Set("hooknodes_rebroadcasting_to", nodes).
	//	Set("public_nodes_rebroadcasting_to", nodes))

	defer oysterUtils.TimeTrack(time.Now(), "rebroadcast_requests_made", analytics.NewProperties().
		Set("addresses", oysterUtils.MapTransactionsToAddrs(*txs)).
		Set("hooknodes_rebroadcasting_to", nodes))

	for _, node := range nodes {

		jsonReq, err := json.Marshal(broadcastReq)
		if err != nil {
			raven.CaptureError(err, nil)
			return
		}

		reqBody := bytes.NewBuffer(jsonReq)

		nodeURL := "http://" + node + ":3000/broadcast/"

		go func() {
			_, err := http.Post(nodeURL, "application/json", reqBody)
			if err != nil {
				raven.CaptureError(err, nil)
			}
		}()
	}

	// Leaving this in because we might want to re-enable it someday, but for now it
	// does not seem to help and in fact seems to make things worse

	//fmt.Println("BroadcastToPublic nodes")
	//for _, nodeURL := range PublicNodes {
	//
	//	go func(nodeURL string) {
	//		publicApi := giota.NewAPI(nodeURL, nil)
	//
	//		var err error
	//
	//		fmt.Println("public node: "+ nodeURL)
	//
	//		err = publicApi.BroadcastTransactions(*txs)
	//		if err != nil {
	//			//var newErr = errors.New("Bad public node or bad trytes.  Node: " + node + " " + err.Error())
	//			//raven.CaptureError(newErr, nil)
	//			fmt.Println("-----------")
	//			fmt.Println(err)
	//			fmt.Println(nodeURL)
	//			fmt.Println("-----------")
	//		}
	//	}(nodeURL)
	//}

	return
}
