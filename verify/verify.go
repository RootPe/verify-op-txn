package verify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/RootPe/verify-op-txn/trie"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
)

type transaction struct {
	BlockNumberHex   string `json:"blockNumber"`
	From             string `json:"from"`
	Hash             string `json:"hash"`
	Input            string `json:"input"`
	To               string `json:"to"`
	TransactionIndex string `json:"transactionIndex"`
}

type rpcResponse struct {
	JSONRPC string       `json:"jsonrpc"`
	ID      int          `json:"id"`
	Result  *transaction `json:"result,omitempty"`
	Error   *rpcError    `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func transactionByHash(rpcURL, txnHash string) (*transaction, error) {
	reqBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "eth_getTransactionByHash",
		"params":  []interface{}{txnHash},
		"id":      1,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", rpcURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var rpcResp rpcResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error %d: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}
	if rpcResp.Result == nil {
		return nil, errors.New("transaction not found")
	}

	return rpcResp.Result, nil
}

func VerifyTransaction(rpcURL, txnHash string) (*types.Header, [][]byte, uint64, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, [][]byte{}, 0, err
	}
	txn, err := transactionByHash(rpcURL, txnHash)
	if err != nil {
		return nil, [][]byte{}, 0, err
	}

	if txn.BlockNumberHex == "" {
		return nil, [][]byte{}, 0, errors.New("transaction is pending")
	}

	n := new(big.Int)
	_, ok := n.SetString(txn.BlockNumberHex[2:], 16)
	if !ok {
		return nil, [][]byte{}, 0, errors.New("invalid block number")
	}
	b, err := client.BlockByNumber(context.Background(), n)
	if err != nil {
		return nil, [][]byte{}, 0, err
	}
	txns := b.Transactions()

	type data struct {
		hash common.Hash
		blob []byte
	}
	kv := make(map[string]data)
	options := trie.NewStackTrieOptions()

	options.WithWriter(func(path []byte, hash common.Hash, blob []byte) {
		kv[common.Bytes2Hex(path)] = data{hash, blob}
	})

	t := trie.NewStackTrie(options)
	rootHash := trie.DeriveSha(txns, t)
	if rootHash != b.Header().TxHash {
		return nil, [][]byte{}, 0, errors.New("trie build error")
	}

	bi := new(big.Int)
	bi.SetString(txn.TransactionIndex[2:], 16)
	k := bi.Uint64()
	var indexBuf []byte
	indexBuf = rlp.AppendUint64(indexBuf[:0], k)

	key := trie.KeybytesToHex(indexBuf)
	pdb := trie.NewProofDB()
	for i := 0; i < len(key); i++ {
		data, ok := kv[common.Bytes2Hex(key[:i])]
		if ok {
			pdb.Put(data.hash.Bytes(), data.blob)
		}
	}

	return b.Header(), pdb.Serialize(), k, nil
}
