package stock

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
)

type Transaction struct {
	S      Stock
	Action string
}

func (t Transaction) Vente() {
	t.Action = "vente"
}

func (t Transaction) Achat() {
	t.Action = "achat"
}
func (t Transaction) Stock() Stock {
	return t.S
}

func (t Transaction) SetStock(s Stock) {

	t.S = s
}

type Transactions struct {
	//	transactions []Transaction
}

func (t Transaction) MarshalBinary() ([]byte, error) {
	// A simple encoding: plain text.

	var b bytes.Buffer
	fmt.Fprintln(&b, t.S, t.Action)
	return b.Bytes(), nil
}

// UnmarshalBinary modifies the receiver so it must take a pointer receiver.
func (t *Transaction) UnmarshalBinary(data []byte) error {
	// A simple encoding: plain text.
	b := bytes.NewBuffer(data)
	_, err := fmt.Fscanln(b, &t.S, &t.Action)
	return err
}

func (Transactions) Put(db leveldb.DB, t Transaction) {
	var network bytes.Buffer
	//	buf := new(bytes.Buffer)

	enc := gob.NewEncoder(&network)
	err := enc.Encode(&t)
	if err != nil {
		log.Fatal("encode:", err)
	}

	// err := dec.Decode(&t)
	// if err != nil {
	// 	log.Fatal("decode error 1:", err)
	// }
	//value := iter.Value()
	//	transactions = append(transactions, t)
	err = db.Put(network.Bytes(), nil, nil)
	check(err)

}

func (Transactions) Put2(f *os.File, t Transaction) {

	var network bytes.Buffer
	//	buf := new(bytes.Buffer)

	enc := gob.NewEncoder(&network)
	err := enc.Encode(&t)
	if err != nil {
		log.Fatal("encode:", err)
	}

	n3, err := f.Write(network.Bytes())
	_, err = f.WriteString("\n")
	f.Sync()
	fmt.Printf("wrote %d bytes\n", n3)
}
func (Transactions) Get(db leveldb.DB) []Transaction {

	var transactions []Transaction

	iter := db.NewIterator(nil, nil)

	for iter.Next() {
		k := bytes.NewBuffer(iter.Key())
		dec := gob.NewDecoder(k)
		var t Transaction
		err := dec.Decode(&t)
		check(err)
		// if err != nil {
		// 	log.Fatal("decode:", err)
		// }
		// Remember that the contents of the returned slice should not be modified, an

		//value := iter.Value()
		//		fmt.Println(string(key))
		fmt.Println("best date :", t)

		// err := dec.Decode(&t)
		// if err != nil {
		// 	log.Fatal("decode error 1:", err)
		// }
		//value := iter.Value()
		transactions = append(transactions, t)
	}
	iter.Release()
	err := iter.Error()
	check(err)
	return transactions
}

func (Transactions) Get2(f *os.File) []Transaction {
	w := bufio.NewWriter(f)
	n4, err := w.WriteString("buffered\n")
	check(err)
	fmt.Printf("wrote %d bytes\n", n4)
	var transactions []Transaction

	r := bufio.NewReader(f)

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {

		out := scanner.Bytes()

		k := bytes.NewBuffer(out)
		dec := gob.NewDecoder(k)
		var t Transaction
		err := dec.Decode(&t)
		check(err)
		// if err != nil {
		// 	log.Fatal("decode:", err)
		// }
		// Remember that the contents of the returned slice should not be modified, an

		//value := iter.Value()
		//		fmt.Println(string(key))
		fmt.Println("best date :", t)
		transactions = append(transactions, t)
	}
	// err := dec.Decode(&t)
	// if err != nil {
	// 	log.Fatal("decode error 1:", err)
	// }
	//value := iter.Value()

	return transactions
}

type Stock struct {
	Date  []byte
	Value float64
}

func check(e error) {
	if e != nil {
		fmt.Println("transaction error", e)
	}
}
