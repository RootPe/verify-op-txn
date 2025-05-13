package main

import (
	"encoding/json"
	"fmt"

	"github.com/RootPe/verify-op-txn/verify"
)

func PrettyPrint(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
}

func main() {
	const rpcURL = "http://localhost:9545"
	header, proof, txnIndex, err := verify.VerifyTransaction(rpcURL, "0x5e0262fa74c7c9dc436fedc5f414f7d7ed71bdd54c7cc305a20c09b450783220")
	if err != nil {
		fmt.Println(err)
		return
	}
	PrettyPrint(header)
	sproof := []string{}
	for _, b := range proof {
		sproof = append(sproof, fmt.Sprintf("%x", b))
	}
	PrettyPrint(sproof)
	fmt.Println("txnIndex", txnIndex)

}
