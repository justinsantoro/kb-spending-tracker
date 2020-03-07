// +build nodb

package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

type closer interface {
	Close() error
}

//DB describes a low level sqlite3 database implementation
type DB string

func handleClose(c closer) {
	fmt.Print("mockdb: handle close")
}

func (db DB) conn() {
	f, err := os.Create(db.String())
	defer f.Close()
	if err != nil {
		log.Println("Failed to create db file "+db.String(), err)
	}
}

func (db DB) Init() error {
	log.Println("mockdb: initialize")
	db.conn()
	return nil
}

//String casts db as a string
func (db DB) String() string {
	return string(db)
}

func (db *DB) PutTransaction(t Txn) error {
	db.conn()
	log.Println("MockDb: Put Txn:", t)
	return nil
}

//GetTransactions returns a slice of Txns within the given time range.
//Ignores Summary transactions
func (db *DB) GetTransactions(t1 time.Time, t2 time.Time) ([]Txn, error) {
	log.Println(fmt.Sprintf("MockDb: GetTxns: tx1: %v, tx2: %v", t1, t2))
	return []Txn{
		Txn{
			Timestamp(t1),
			USD(1000),
			[]string{"mock_db"},
			"mock_db GetTxn",
			"mocker",
			false,
		},
		Txn{
			Timestamp(t2),
			USD(-2000),
			[]string{"mock_db"},
			"mock_db GetTxn",
			"mocker",
			false,
		},
	}, nil
}

func (db *DB) GetTransactionsSince(t time.Time) ([]Txn, error) {
	log.Println("MockDb: GetTxnsSince: tx1:", t)
	return []Txn{
		Txn{
			Timestamp(t.Add(1 * time.Hour)),
			USD(1000),
			[]string{"mock_db"},
			"mock_db GetTxn",
			"mocker",
			false,
		},
		Txn{
			Timestamp(t),
			USD(-2000),
			[]string{"mock_db"},
			"mock_db GetTxn",
			"mocker",
			false,
		},
	}, nil
}

//GetBalance returns the sum of transaction amounts since a given time.
func (db DB) GetBalance(t time.Time) (USD, error) {
	log.Println("MockDb: GetBalance:", t)
	return USD(300), nil
}

//GetBalance returns the sum of transaction amounts grouped by username between two timestamps
func (db DB) GetTagBalance(tag string, t1 time.Time, t2 time.Time) (*TagBalance, error) {
	log.Printf("MockDb: GetTagBalance: tag:%s, t1-%v, t2-%v", tag, t1, t2)
	tb := NewTagBalance(tag)
	tb.Add("user1", 100)
	tb.Add("user2", 200)
	return tb, nil
}

//GetTags returns a list of distinct tags
func (db DB) GetTags() ([]string, error) {
	log.Print("mockDb: GetTags")
	return []string{
		"tag1",
		"tag2",
		"tag3",
	}, nil
}
