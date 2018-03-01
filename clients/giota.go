package giotaClient

import (
	"os"
	"strconv"
	"time"

	"github.com/iotaledger/giota"
)

// SendTrytes does attachToTangle and finally, it broadcasts the transactions.
func SendTrytes(trytes []giota.Trytes, trunk giota.Trytes, branch giota.Trytes) (err error) {
	// Get configuration.
	provider := os.Getenv("PROVIDER")
	minDepth, _ := strconv.ParseInt(os.Getenv("MIN_DEPTH"), 10, 64)
	minWeightMag, _ := strconv.ParseInt(os.Getenv("MIN_WEIGHT_MAGNITUDE"), 10, 64)

	api := giota.NewAPI(provider, nil)
	_, powFn := giota.GetBestPoW()

	// Convert []Trytes to []Transaction
	txs := make([]giota.Transaction, len(trytes))
	for i, t := range trytes {
		tx, _ := giota.NewTransaction(t)
		txs[i] = *tx
	}

	getTxsRes := giota.GetTransactionsToApproveResponse{}
	err = doPow(&getTxsRes, minDepth, txs, minWeightMag, powFn)
	if err != nil {
		return
	}

	// Broadcast and store tx
	return api.BroadcastTransactions(txs)
}

/*
	These are copied from the giota lib since they are not public.
*/
// This is copied from giota lib since it is not public.

// (3^27-1)/2
const maxTimestampTrytes = "MMMMMMMMM"

// https://github.com/iotaledger/giota/blob/master/transfer.go#L322
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

		trytes[i].Nonce, err = pow(trytes[i].Trytes(), int(mwm))
		if err != nil {
			return err
		}

		prev = trytes[i].Hash()
	}
	return nil
}
