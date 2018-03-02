package giotaClient

import (
	"os"
	"sync"
	"time"

	"github.com/iotaledger/giota"
)

// SendTrytes does attachToTangle and finally, it broadcasts the transactions.
func SendTrytes(trytes []giota.Trytes, trunk giota.Trytes, branch giota.Trytes) (
	ts []giota.Transaction, err error) {
	// Get configuration.
	provider := os.Getenv("PROVIDER")
	minDepth := int64(giota.DefaultNumberOfWalks)
	minWeightMag := int64(giota.DefaultMinWeightMagnitude)
	// minDepth, _ := strconv.ParseInt(os.Getenv("MIN_DEPTH"), 10, 64)
	// minWeightMag, _ := strconv.ParseInt(os.Getenv("MIN_WEIGHT_MAGNITUDE"), 10, 64)

	api := giota.NewAPI(provider, nil)
	// _, powFn := giota.GetBestPoW()

	// Convert []Trytes to []Transaction
	txs := make([]giota.Transaction, len(trytes))
	for i, t := range trytes {
		tx, _ := giota.NewTransaction(t)
		txs[i] = *tx
	}

	getTxsRes, err := api.GetTransactionsToApprove(minDepth, giota.DefaultNumberOfWalks, "")
	if err != nil {
		return
	}

	// getTxsRes := giota.GetTransactionsToApproveResponse{
	// 	TrunkTransaction:  trunk,
	// 	BranchTransaction: branch,
	// }

	// err = doPow(getTxsRes, minDepth, txs, minWeightMag, powFn)
	// if err != nil {
	// 	return
	// }

	at := giota.AttachToTangleRequest{
		TrunkTransaction:   getTxsRes.TrunkTransaction,
		BranchTransaction:  getTxsRes.BranchTransaction,
		MinWeightMagnitude: minWeightMag,
		Trytes:             txs,
	}
	attached, err := api.AttachToTangle(&at)
	if err != nil {
		return
	}
	txs = attached.Trytes

	// Broadcast and store tx
	err = api.BroadcastTransactions(txs)
	if err != nil {
		return
	}

	return txs, nil
}

// Things below are copied from the giota lib since they are not public.
// https://github.com/iotaledger/giota/blob/master/transfer.go#L322

// (3^27-1)/2
const maxTimestampTrytes = "MMMMMMMMM"

// This mutex was added by us.
var mutex = &sync.Mutex{}

func doPow(tra *giota.GetTransactionsToApproveResponse, depth int64,
	trytes []giota.Transaction, mwm int64, pow giota.PowFunc) error {

	var prev giota.Trytes
	var err error
	for i := len(trytes) - 1; i >= 0; i-- {
		switch {
		case i == len(trytes)-1:
			trytes[i].TrunkTransaction = tra.TrunkTransaction
			trytes[i].BranchTransaction = tra.BranchTransaction
		default:
			trytes[i].TrunkTransaction = prev
			trytes[i].BranchTransaction = tra.TrunkTransaction
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
	return nil
}
