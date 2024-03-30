package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"strconv"
	"time"
)

type Block struct {
	Hash         []byte
	Data         []byte
	PreviousHash []byte
	CreationTime int64
	Nonce        int
}

func (block *Block) CalculateHash() {
	creation_time := []byte(strconv.FormatInt(block.CreationTime, 10))
	info := bytes.Join([][]byte{block.PreviousHash, block.Data, creation_time}, []byte{})
	hash := sha256.Sum256(info)
	block.Hash = hash[:]
}

func CreateBlock(data string, previous_hash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), previous_hash, time.Now().Unix(), 0}
	pow := CreateProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func CreateGenesisBlock() *Block {
	return CreateBlock("Genesis Block", []byte{})
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
