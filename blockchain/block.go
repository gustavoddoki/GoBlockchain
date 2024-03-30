package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Hash         []byte
	Transactions []*Transaction
	PreviousHash []byte
	CreationTime int64
	Nonce        int
}

func (block *Block) HashTransactions() []byte {
	var tx_hashes [][]byte
	var tx_hash [32]byte

	for _, tx := range block.Transactions {
		tx_hashes = append(tx_hashes, tx.ID)
	}
	tx_hash = sha256.Sum256(bytes.Join(tx_hashes, []byte{}))
	return tx_hash[:]
}

func CreateBlock(transactions []*Transaction, previous_hash []byte) *Block {
	block := &Block{[]byte{}, transactions, previous_hash, time.Now().Unix(), 0}
	pow := CreateProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func CreateGenesisBlock(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{})
}

func (block *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(block)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}
	return &block
}
