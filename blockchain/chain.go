package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
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
		fmt.Println("Blockchain does not exist.")
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

func (chain *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
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
				if out.IsLockedWithKey(pubKeyHash) {
					unspent_txs = append(unspent_txs, *tx)
				}
			}
			if !tx.FlagCoinbaseTx() {
				for _, in := range tx.Inputs {
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.ID)
						spent_txs0[inTxID] = append(spent_txs0[inTxID], in.Out)
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

func (chain *BlockChain) FindUXT0(pubKeyHash []byte) []TxOutput {
	var UTX0s []TxOutput
	unspent_txs := chain.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspent_txs {
		for _, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) {
				UTX0s = append(UTX0s, out)
			}
		}
	}
	return UTX0s
}

func (chain *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspent_outs := make(map[string][]int)
	unspent_txs := chain.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspent_txs {
		txID := hex.EncodeToString(tx.ID)
		for out_id, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
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

func (blockchain *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iter := blockchain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return *tx, nil
			}
		}

		if len(block.PreviousHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction does not exist")
}

func (blockchain *BlockChain) SignTransaction(transaction *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, in := range transaction.Inputs {
		prevTX, err := blockchain.FindTransaction(in.ID)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	transaction.Sign(privKey, prevTXs)
}

func (blockchain *BlockChain) VerifyTransaction(transaction *Transaction) bool {
	prevTXs := make(map[string]Transaction)

	for _, in := range transaction.Inputs {
		prevTX, err := blockchain.FindTransaction(in.ID)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	return transaction.Verify(prevTXs)
}
