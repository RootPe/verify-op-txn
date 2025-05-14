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
	Value    string   `json:"value"`
	TxnIndex uint64   `json:"txn_index"`
	TxnCount uint64   `json:"txn_count"`
	BlockNum uint64   `json:"block_num"`
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

	root, proof, txnIndex, txnCount, value, blockNum, err := verify.GetTransactionProof(*rpcURL, *txHash)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Verification error:", err)
		os.Exit(1)
	}

	out := Output{
		Root:     root,
		Proof:    proof,
		TxnIndex: txnIndex,
		TxnCount: txnCount,
		Value:    value,
		BlockNum: blockNum,
	}
	json.NewEncoder(os.Stdout).Encode(out)
}
