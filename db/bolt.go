package db

import (
	"gopherchain/utils"
	"time"

	"github.com/boltdb/bolt"
)

type BoltDB struct {
	DBPath string
}

// Get gets the value of the key provided as a slice of bytes
func (d *BoltDB) Get(key string) []byte {
	var value []byte
	db, err := bolt.Open(d.DBPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	utils.HandleError(err)

	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("gopherchain"))
		value = b.Get([]byte(key))
		return nil
	})
	return value
}

// Put store the key value pair as bytes into the data store
func (d *BoltDB) Put(key string, value string) error {
	db, err := bolt.Open(d.DBPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	utils.HandleError(err)

	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("gopherchain"))
		if err != nil {
			return err
		}
		err = b.Put([]byte(key), []byte(value))
		if err != nil {
			return err
		}
		return nil
	})
	return nil
}
