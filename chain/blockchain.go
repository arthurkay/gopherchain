package chain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/dgraph-io/badger/v3"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
	dbPath = os.Getenv("DB_PATH")
	dbFile = os.Getenv("DB_FILE")
}

var (
	dbPath      string
	dbFile      string
	genesisData = "The gopherchain is born"
)

// BlockChainIterator loops through the
// Badger DB instance records
type BlockChainIterator struct {
	CurrentHash []byte
	Datatbase   *badger.DB
}

// BlockChain struct implementation
// of the connection between different chains
// and how these chains link to the one before and after
type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

// DBExists checks whether or not a db file exists on the system
func DBExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

// ContinueBlockChain takes in a string address and returns a
// pointer to the blockchain instance
// the method adds a new address to the block chain
func ContinueBlockChain(address string) *BlockChain {
	if !DBExists() {
		fmt.Println("No already existing gopherchain found, creating one")
		runtime.Goexit()
	}

	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	HandleErr(err)

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		HandleErr(err)
		err = item.Value(func(val []byte) error {
			lastHash = append([]byte{}, val...)
			return nil
		})
		return err
	})

	HandleErr(err)
	chain := BlockChain{lastHash, db}
	return &chain
}

// FindUnspentTransactions takes in a string address and returns an array
// of transactions available to the provided address
func (chain *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTxs []Transaction

	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				if out.IsLockedWithKey(pubKeyHash) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}

			if !tx.IsCoinBase() {
				for _, in := range tx.Inputs {
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.ID)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
					}
				}
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}

	}
	return unspentTxs
}

// FindUTXO takes in a string address and returns an array of
// transaction outputs
func (chain *BlockChain) FindUTXO(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput
	unspentTransactions := chain.FindUnspentTransactions(pubKeyHash)
	for _, tx := range unspentTransactions {
		for _, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}

// FindSpendableOutputs takes in a string address and n integer amount
// then returns an accumulated amount and a slice of the spendable outputs
func (chain *BlockChain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTxs := chain.FindUnspentTransactions(pubKeyHash)
	acumulated := 0

Work:
	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Outputs {
			if out.IsLockedWithKey(pubKeyHash) && acumulated < amount {
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if acumulated > amount {
					break Work
				}
			}
		}
	}
	return acumulated, unspentOutputs
}

// InitBlockChain initialises the blockchain if
// there is no db file on the current system.
// Otherwise the system just shutsdown
func InitBlockChain(address string) *BlockChain {
	var lastHash []byte

	if DBExists() {
		fmt.Println("gopherchain has a db file on this machine")
		runtime.Goexit()
	}
	opts := badger.DefaultOptions(dbPath)

	db, err := badger.Open(opts)
	HandleErr(err)

	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinBaseTx(address, genesisData)
		genesis := Genesis(cbtx)
		fmt.Println("Genesis Data created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		HandleErr(err)
		err = txn.Set([]byte("lh"), genesis.Hash)
		lastHash = genesis.Hash
		return err
	})
	HandleErr(err)

	blockchain := BlockChain{LastHash: lastHash, Database: db}
	return &blockchain
}

// AddBlock takes in a pointer of an array of transactions and adds them
// to the badger DB data store
func (chain *BlockChain) AddBlock(transactions []*Transaction) {
	var lastHash []byte

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		HandleErr(err)
		err = item.Value(func(val []byte) error {
			fmt.Printf("Got value: %x", val)
			lastHash = append([]byte{}, val...)
			return nil
		})
		return err
	})

	HandleErr(err)

	newBlock := CreateBlock(transactions, lastHash)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		HandleErr(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash
		return err
	})

	HandleErr(err)
}

// Iterator returns a blockchain iterrator instance
func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{chain.LastHash, chain.Database}
	return iter
}

// Next gets the blockchain's iterator and returns a block
// pointer instance that is next in the sequence
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

func (bc *BlockChain) FindTransaction(id []byte) (Transaction, error) {
	iter := bc.Iterator()

	for {
		block := iter.Next()
		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, id) {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("Transaction doesn't exist")
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		HandleErr(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}
