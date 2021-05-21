package chain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
)

type Block struct {
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
	Nonce        int
}

func HandleErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func (b *Block) HasTransactions() []byte {
	var txsHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txsHashes = append(txsHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txsHashes, []byte{}))

	return txHash[:]
}

func CreateBlock(txs []*Transaction, prevHash []byte) *Block {
	block := &Block{[]byte{}, txs, prevHash, 0}
	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

func Genesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{})
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)

	HandleErr(err)

	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)
	HandleErr(err)
	return &block
}
