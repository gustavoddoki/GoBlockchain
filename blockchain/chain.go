package blockchain

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/dgraph-io/badger"
)

const (
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST"
	genesisData = "First Transaction from Genesis"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func DBexists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func (chain *BlockChain) AddBlock(transactions []*Transaction) {
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
	new_block := CreateBlock(transactions, last_hash)
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

func CreateBlockchain(address string) *BlockChain {
	var last_hash []byte

	if DBexists() {
		fmt.Println("Blockchain already exists.")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions("")
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CreateCoinbaseTx(address, genesisData)
		genesis := CreateGenesisBlock(cbtx)
		fmt.Println("Genesis Block created")

		err = txn.Set(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = txn.Set([]byte("lh"), genesis.Hash)

		last_hash = genesis.Hash

		return err
	})

	if err != nil {
		log.Panic(err)
	}

	blockchain := BlockChain{last_hash, db}
	return &blockchain
}

func ContinueBlockChain(address string) *BlockChain {

	if !DBexists() {
		fmt.Println("Blockchain not exists.")
		runtime.Goexit()
	}

	var last_hash []byte

	opts := badger.DefaultOptions("")
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			log.Panic(err)
		}
		last_hash, err = item.ValueCopy(nil)
		return err
	})
	if err != nil {
		log.Panic(nil)
	}
	chain := BlockChain{last_hash, db}
	return &chain
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

func (chain *BlockChain) FindUnspentTransactions(address string) []Transaction {
	var unspent_txs []Transaction
	spent_txs0 := make(map[string][]int)
	iter := chain.Iterator()

	for {
		block := iter.Next()
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for out_id, out := range tx.Outputs {
				if spent_txs0[txID] != nil {
					for _, spent_out := range spent_txs0[txID] {
						if spent_out == out_id {
							continue Outputs
						}
					}
				}
				if out.CanBeUnlocked(address) {
					unspent_txs = append(unspent_txs, *tx)
				}
				if !tx.FlagCoinbaseTx() {
					for _, in := range tx.Inputs {
						if in.CanUnlockOutput(address) {
							inTxID := hex.EncodeToString(in.ID)
							spent_txs0[inTxID] = append(spent_txs0[inTxID], in.Out)
						}
					}
				}
			}
		}
		if len(block.PreviousHash) == 0 {
			break
		}
	}
	return unspent_txs
}

func (chain *BlockChain) FindUXT0(address string) []TxOutput {
	var UTX0s []TxOutput
	unspent_txs := chain.FindUnspentTransactions(address)

	for _, tx := range unspent_txs {
		for _, out := range tx.Outputs {
			if out.CanBeUnlocked(address) {
				UTX0s = append(UTX0s, out)
			}
		}
	}
	return UTX0s
}

func (chain *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspent_outs := make(map[string][]int)
	unspent_txs := chain.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspent_txs {
		txID := hex.EncodeToString(tx.ID)
		for out_id, out := range tx.Outputs {
			if out.CanBeUnlocked(address) && accumulated < amount {
				accumulated += out.Value
				unspent_outs[txID] = append(unspent_outs[txID], out_id)
			}

			if accumulated >= amount {
				break Work
			}
		}
	}
	return accumulated, unspent_outs
}
