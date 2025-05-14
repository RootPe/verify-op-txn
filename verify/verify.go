package verify

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/trie/trienode"
)

func PrettyPrint(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
}

func GetTransactionProof(rpcURL string, txHash string) ([][]byte, *types.Header, uint64, error) {
	ctx := context.Background()
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to connect to RPC: %w", err)
	}

	// Get the transaction and receipt
	_, isPending, err := client.TransactionByHash(ctx, common.HexToHash(txHash))
	if err != nil {
		return nil, nil, 0, fmt.Errorf("tx not found: %w", err)
	}
	if isPending {
		return nil, nil, 0, fmt.Errorf("tx is pending")
	}

	receipt, err := client.TransactionReceipt(ctx, common.HexToHash(txHash))
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get receipt: %w", err)
	}

	block, err := client.BlockByHash(ctx, receipt.BlockHash)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to get block: %w", err)
	}

	header := block.Header()
	txs := block.Transactions()
	PrettyPrint(txs)
	txIndex := receipt.TransactionIndex
	fmt.Println("tex", txIndex)

	// Build a transaction trie using NewStackTrie
	txTrie := trie.NewEmpty(nil)

	for i := 0; i < len(txs); i++ {
		tx := txs[i]

		// Build trie key: RLP(index)
		key, err := rlp.EncodeToBytes(uint(i))
		if err != nil {
			log.Fatalf("Failed to encode index %d: %v", i, err)
		}

		// RLP-encode the transaction itself
		val, err := tx.MarshalBinary()
		if err != nil {
			log.Fatalf("Failed to encode transaction at index %d: %v", i, err)
		}

		fmt.Println("Index:", i, "Key:", hex.EncodeToString(key), "TxHash:", tx.Hash().Hex(), hex.EncodeToString(val))
		txTrie.Update(key, val)
	}

	reconstructedTxRoot := txTrie.Hash()
	targetTxKey, err := rlp.EncodeToBytes(uint(0))
	if err != nil {
		log.Fatalf("Failed to RLP encode target transaction index %d: %v", txIndex, err)
	}
	fmt.Println("reconstructedTxRoot", reconstructedTxRoot)
	proof := new(trienode.ProofList)
	fmt.Println("targetTxKey", targetTxKey, hex.EncodeToString(targetTxKey))
	txTrie.Prove(targetTxKey, proof)
	for i, node := range *proof {
		fmt.Printf("Node %d: 0x%x\n", i, node)
		fmt.Println(crypto.Keccak256Hash(node).String())
	}

	// db := memorydb.New()
	// for _, node := range *proof {
	// 	db.Put(crypto.Keccak256Hash(node).Bytes(), node)
	// }

	// aaa, err := trie.VerifyProof(reconstructedTxRoot, targetTxKey, nil)
	// if err != nil {
	// 	log.Fatalf("err %v", err)
	// }
	// fmt.Println("aaa", hex.EncodeToString(aaa))

	fmt.Println("reconstructedTxRoot", reconstructedTxRoot)
	return nil, header, uint64(txIndex), nil
}
