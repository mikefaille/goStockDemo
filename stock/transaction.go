package stock

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"

	"github.com/syndtr/goleveldb/leveldb"
)

type Transaction struct {
	s      Stock
	action string
}

func (t Transaction) Vente() {
	t.action = "vente"
}

func (t Transaction) Achat() {
	t.action = "achat"
}
func (t Transaction) GetStock() Stock {
	return t.s
}

func (t Transaction) SetStock(s Stock) {
	t.s = s
}

type Transactions struct {
	//	transactions []Transaction
}

func (Transactions) Put(db leveldb.DB, t Transaction) {

	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	err := enc.Encode(t)
	check(err)

	db.Put(b.Bytes(), nil, nil)

}
func (Transactions) Get(db leveldb.DB) {

	var t Transaction

	iter := db.NewIterator(nil, nil)

	for iter.Next() {
		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		key := iter.Key()
		transactionBuff := bytes.NewBuffer(key)
		dec := gob.NewDecoder(transactionBuff)
		err := dec.Decode(&t)
		if err != nil {
			log.Fatal("decode error 1:", err)
		}
		//value := iter.Value()

	}
	iter.Release()
	err := iter.Error()
	check(err)
}

type Stock struct {
	Date  []byte
	Value float64
}

func check(e error) {
	if e != nil {
		fmt.Println(e)
	}
}
