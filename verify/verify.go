package verify

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethdb/memorydb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/ethereum/go-ethereum/trie/trienode"
)

func GetTransactionProof(rpcURL string, txHash string) (string, []string, uint64, uint64, string, uint64, error) {
	ctx := context.Background()
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return "", nil, 0, 0, "", 0, fmt.Errorf("failed to connect to RPC: %w", err)
	}

	t, isPending, err := client.TransactionByHash(ctx, common.HexToHash(txHash))
	if err != nil {
		return "", nil, 0, 0, "", 0, fmt.Errorf("tx not found: %w", err)
	}

	txBytes, err := t.MarshalBinary()
	if err != nil {
		return "", nil, 0, 0, "", 0, fmt.Errorf("failed to encode transaction: %w", err)
	}

	if isPending {
		return "", nil, 0, 0, "", 0, fmt.Errorf("tx is pending")
	}

	receipt, err := client.TransactionReceipt(ctx, common.HexToHash(txHash))
	if err != nil {
		return "", nil, 0, 0, "", 0, fmt.Errorf("failed to get receipt: %w", err)
	}

	block, err := client.BlockByHash(ctx, receipt.BlockHash)
	if err != nil {
		return "", nil, 0, 0, "", 0, fmt.Errorf("failed to get block: %w", err)
	}

	txs := block.Transactions()
	txTrie := trie.NewEmpty(nil)

	for i, tx := range txs {
		key, err := rlp.EncodeToBytes(uint(i))
		if err != nil {
			return "", nil, 0, 0, "", 0, fmt.Errorf("failed to encode index %d: %w", i, err)
		}

		val, err := tx.MarshalBinary()
		if err != nil {
			return "", nil, 0, 0, "", 0, fmt.Errorf("failed to encode transaction at index %d: %w", i, err)
		}

		txTrie.Update(key, val)
	}

	reconstructedTxRoot := txTrie.Hash()
	txIndex := receipt.TransactionIndex

	targetTxKey, err := rlp.EncodeToBytes(uint(txIndex))
	if err != nil {
		return "", nil, 0, 0, "", 0, fmt.Errorf("failed to encode transaction index: %w", err)
	}

	proof := new(trienode.ProofList)
	if err := txTrie.Prove(targetTxKey, proof); err != nil {
		return "", nil, 0, 0, "", 0, fmt.Errorf("failed to generate proof: %w", err)
	}

	sproof := make([]string, len(*proof))
	db := memorydb.New()
	for i, node := range *proof {
		sproof[i] = hex.EncodeToString(node)
		db.Put(crypto.Keccak256Hash(node).Bytes(), node)
	}

	_, err = trie.VerifyProof(reconstructedTxRoot, targetTxKey, db)
	if err != nil {
		return "", nil, 0, 0, "", 0, fmt.Errorf("failed to verify proof: %w", err)
	}

	return reconstructedTxRoot.String(), sproof, uint64(txIndex), uint64(len(txs)), hex.EncodeToString(txBytes), block.NumberU64(), nil
}
