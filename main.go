package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/RootPe/verify-op-txn/verify"
)

type Output struct {
	Header   interface{} `json:"header"`
	Proof    []string    `json:"proof"`
	TxnIndex uint64      `json:"txnIndex"`
}

func main() {
	rpcURL := flag.String("rpc", "", "Ethereum RPC URL")
	txHash := flag.String("tx", "", "Transaction hash to verify")

	flag.Parse()

	if *txHash == "" {
		fmt.Fprintln(os.Stderr, "Error: -tx flag is required")
		flag.Usage()
		os.Exit(1)
	}

	if *rpcURL == "" {
		fmt.Fprintln(os.Stderr, "Error: -rpc flag is required")
		flag.Usage()
		os.Exit(1)
	}

	header, proof, txnIndex, err := verify.VerifyTransaction(*rpcURL, *txHash)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Verification error:", err)
		os.Exit(1)
	}

	sproof := []string{}
	for _, b := range proof {
		sproof = append(sproof, fmt.Sprintf("%x", b))
	}

	out := Output{
		Header:   header,
		Proof:    sproof,
		TxnIndex: txnIndex,
	}
	json.NewEncoder(os.Stdout).Encode(out)
}
