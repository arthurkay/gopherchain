package chain

import (
	"fmt"

	"github.com/dgraph-io/badger/v3"
)

const (
	dbPath = "./tmp/blocks"
)

type BlockChainIterator struct {
	CurrentHash []byte
	Datatbase   *badger.DB
}

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

func InitBlockChain() *BlockChain {
	var lastHash []byte
	opts := badger.DefaultOptions(dbPath)

	db, err := badger.Open(opts)
	HandleErr(err)

	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain found")
			genesis := Genesis()
			fmt.Println("The Gopher Chain is born!!")

			err := txn.Set(genesis.Hash, genesis.Serialize())
			HandleErr(err)
			err = txn.Set([]byte("lh"), genesis.Hash)
			lastHash = genesis.Hash
			return err
		} else {
			item, err := txn.Get([]byte("lh"))
			HandleErr(err)
			err = item.Value(func(val []byte) error {
				fmt.Printf("Got value: %s \n", val)
				lastHash = append([]byte{}, val...)
				return nil
			})
			return err
		}
	})

	HandleErr(err)

	blockchain := BlockChain{LastHash: lastHash, Database: db}
	return &blockchain
}

func (chain *BlockChain) AddBlock(data string) {
	var lastHash []byte

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		HandleErr(err)
		err = item.Value(func(val []byte) error {
			fmt.Printf("Got value: %s", val)
			lastHash = append([]byte{}, val...)
			return nil
		})
		return err
	})

	HandleErr(err)

	newBlock := CreateBlock(data, lastHash)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		HandleErr(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash
		return err
	})

	HandleErr(err)
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{chain.LastHash, chain.Database}
	return iter
}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block
	err := iter.Datatbase.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		HandleErr(err)
		err = item.Value(func(val []byte) error {
			block = Deserialize(val)
			return err
		})
		return err
	})
	HandleErr(err)
	iter.CurrentHash = block.PrevHash
	return block
}
