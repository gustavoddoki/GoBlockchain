package blockchain

import (
	"fmt"
	"log"

	"github.com/dgraph-io/badger"
)

const (
	dbPath = "./tmp/blocks"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func (chain *BlockChain) AddBlock(data string) {
	var last_hash []byte
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			log.Panic(err)
		}
		last_hash, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		log.Panic(err)
	}
	new_block := CreateBlock(data, last_hash)
	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(new_block.Hash, new_block.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = txn.Set([]byte("lh"), new_block.Hash)
		chain.LastHash = new_block.Hash
		return err
	})
	if err != nil {
		log.Panic(err)
	}
}

func CreateBlockchain() *BlockChain {
	var last_hash []byte

	opts := badger.DefaultOptions("")
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {

			fmt.Println("Blockchain not exists")

			genesis := CreateGenesisBlock()
			fmt.Println("Genesis Block created")

			err = txn.Set(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}
			err = txn.Set([]byte("lh"), genesis.Hash)

			last_hash = genesis.Hash

			return err
		} else {
			item, err := txn.Get([]byte("lh"))
			if err != nil {
				log.Panic(err)
			}
			last_hash, err = item.ValueCopy(nil)
			return err
		}
	})

	if err != nil {
		log.Panic(err)
	}

	blockchain := BlockChain{last_hash, db}
	return &blockchain
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{chain.LastHash, chain.Database}
	return iter
}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block
	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		if err != nil {
			log.Panic(err)
		}
		encoded_block, err := item.ValueCopy(nil)
		block = Deserialize(encoded_block)
		return err
	})
	if err != nil {
		log.Panic(err)
	}
	iter.CurrentHash = block.PreviousHash
	return block
}
