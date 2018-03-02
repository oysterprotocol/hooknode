package oysterUtils

import (
	"os"
	"github.com/iotaledger/giota"
	"github.com/joho/godotenv"
	"gopkg.in/segmentio/analytics-go.v3"
	"log"
	"net"
)

var SegmentClient analytics.Client

func init() {
	// Load ENV variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Setup Segment
	SegmentClient = analytics.New(os.Getenv("SEGMENT_WRITE_KEY"))
}


func MapTransactionsToAddrs(txs []giota.Transaction) (addrs []giota.Address) {

	addrs = make([]giota.Address, 0, len(txs))

	for _, tx := range txs {
		addrs = append(addrs, tx.Address)
	}

	return
}

func MapTransfersToAddrs(txs []giota.Transfer) (addrs []giota.Address) {

	addrs = make([]giota.Address, 0, len(txs))

	for _, tx := range txs {
		addrs = append(addrs, tx.Address)
	}

	return
}

func GetLocalIP() string {
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
