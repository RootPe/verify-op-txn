package trie

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

type ProofDB struct {
	kv map[string][]byte
}

func NewProofDB() *ProofDB {
	return &ProofDB{
		kv: make(map[string][]byte),
	}
}

func (w *ProofDB) Put(key []byte, value []byte) error {
	keyS := fmt.Sprintf("%x", key)
	w.kv[keyS] = value
	return nil
}

func (w *ProofDB) Delete(key []byte) error {
	keyS := fmt.Sprintf("%x", key)
	delete(w.kv, keyS)
	return nil
}

func (w *ProofDB) Has(key []byte) (bool, error) {
	keyS := fmt.Sprintf("%x", key)
	_, ok := w.kv[keyS]
	return ok, nil
}

func (w *ProofDB) Get(key []byte) ([]byte, error) {
	keyS := fmt.Sprintf("%x", key)
	val, ok := w.kv[keyS]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return val, nil
}

func (w *ProofDB) Serialize() [][]byte {
	nodes := make([][]byte, 0, len(w.kv)*2)
	for key, value := range w.kv {
		nodes = append(nodes, common.Hex2Bytes(key))
		nodes = append(nodes, value)
	}
	return nodes
}
