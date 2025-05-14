package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/RootPe/verify-op-txn/verify"
)

type Output struct {
	Root     string   `json:"root"`
	Proof    []string `json:"proof"`
	TxnIndex uint64   `json:"txnIndex"`
	TxnCount uint64   `json:"txnCount"`
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

	proof, _, _, err := verify.GetTransactionProof(*rpcURL, *txHash)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Verification error:", err)
		os.Exit(1)
	}

	sproof := []string{}
	for _, b := range proof {
		sproof = append(sproof, fmt.Sprintf("%x", b))
	}

	out := Output{
		// Root:     header.TxHash.String(),
		// Proof:    sproof,
		// TxnIndex: txnIndex,
		// TxnCount: txnCount,
	}
	json.NewEncoder(os.Stdout).Encode(out)
}
