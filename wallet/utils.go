package wallet

import (
	"log"

	"github.com/mr-tron/base58"
)

// Base58 is a derivative from base65 but without
// the followig characters 0 O 1 l +

// Base58Encode Creates a base58 encoded string
func Base58Encode(input []byte) []byte {
	encode := base58.Encode(input)
	return []byte(encode)
}

// Base58 Decodes a byte slice into a base 58 byte slice
func Base58Decode(input []byte) []byte {
	decode, err := base58.Decode(string(input[:]))
	if err != nil {
		log.Panic(err)
	}

	return decode
}
