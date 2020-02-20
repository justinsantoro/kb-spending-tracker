// +build !nodb

package main

import (
	"encoding/json"
	"fmt"
	"github.com/bvinc/go-sqlite-lite/sqlite3"
)

const date string = `json_extract(txs.tx, '$.Date')`

type closer interface {
	Close() error
}

//DB describes a low level sqlite3 database implementation
type DB string

func betweenTimes() string {
	return fmt.Sprintf("%s >= (?) AND %s <= (?)", date, date)
}

func handleClose(c closer) {
	err := c.Close()
	if err != nil {
		fmt.Print("error closing :", err)
	}
}

func txRowsToSlice(stmt *sqlite3.Stmt) ([]Txn, error) {
	var txs []Txn
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return nil, err
		}
		if !hasRow {
			break
		}

		var tx string
		err = stmt.Scan(&tx)
		if err != nil {
			return nil, err
		}

		var t Txn
		if err := json.Unmarshal([]byte(tx), &t); err != nil {
			return nil, err
		}
		txs = append(txs, t)
	}
	return txs, nil
}

func (db DB) conn() (*sqlite3.Conn, error) {
	return sqlite3.Open(string(db))
}

//String casts db as a string
func (db DB) String() string {
	return string(db)
}

func (db *DB) PutTransaction(t Txn) error {
	conn, err := db.conn()
	if err != nil {
		return err
	}
	defer handleClose(conn)

	tjson, err := t.Json()
	if err != nil {
		return err
	}
	stmt, err := conn.Prepare(`Insert INTO txs VALUES (?)`, tjson)
	if err != nil {
		return err
	}
	defer handleClose(stmt)
	return stmt.Exec()
}

//GetTransactions returns a slice of Txns within the given time range.
//Ignores Summary transactions
func (db *DB) GetTransactions(t1 Timestamp, t2 Timestamp) ([]Txn, error) {

	sql := `SELECT tx FROM txs
WHERE %s AND NOT json_extract(txs.tx, '$.Summary')`

	conn, err := db.conn()
	if err != nil {
		return nil, err
	}
	defer handleClose(conn)

	stmt, err := conn.Prepare(fmt.Sprintf(sql, betweenTimes()), t1.UnixNano(), t2.UnixNano())
	if err != nil {
		return nil, err
	}
	defer handleClose(stmt)

	return txRowsToSlice(stmt)
}

func (db *DB) GetTransactionsSince(t Timestamp) ([]Txn, error) {

	sql := `SELECT tx FROM txs
WHERE %s >= (?) AND NOT json_extract(txs.tx, '$.Summary')`

	conn, err := db.conn()
	if err != nil {
		return nil, err
	}
	defer handleClose(conn)

	stmt, err := conn.Prepare(fmt.Sprintf(sql, date), t.UnixNano())
	if err != nil {
		return nil, err
	}
	defer handleClose(stmt)

	return txRowsToSlice(stmt)
}

//GetBalance returns the sum of transaction amounts since a given time.
func (db DB) GetBalance(t Timestamp) (USD, error) {
	sql := `SELECT SUM(json_extract(txs.tx, '$.Amount')) AS amt FROM txs WHERE %s >= (?)`

	conn, err := db.conn()
	if err != nil {
		return -1, err
	}
	defer handleClose(conn)

	stmt, err := conn.Prepare(fmt.Sprintf(sql, date), t.UnixNano())
	if err != nil {
		return -1, err
	}
	defer handleClose(stmt)

	hasRow, err := stmt.Step()
	if !hasRow {
		return 0, nil
	}

	var amt int64
	err = stmt.Scan(&amt)
	if err != nil {
		return -1, err
	}
	return USD(amt), nil
}

//GetBalance returns the sum of transaction amounts grouped by username between two timestamps
func (db DB) GetTagBalance(tag string, t1 Timestamp, t2 Timestamp) (*TagBalance, error){
	sql := `Select json_extract(txs.tx, '$.User'), SUM(json_extract(txs.tx, '$.Amount')) as amt 
From txs, json_each(json_extract(txs.tx, '$.Tags'))
WHERE %s AND json_each.value = (?)
GROUP BY json_extract(txs.tx, '$.User')
ORDER BY amt`

	conn, err := db.conn()
	if err != nil {
		return nil, err
	}
	defer handleClose(conn)

	stmt, err := conn.Prepare(fmt.Sprintf(sql, betweenTimes()), t1.UnixNano(), t2.UnixNano(), tag)
	if err != nil {
		return nil, err
	}
	defer handleClose(stmt)

	tb := NewTagBalance(tag)
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return nil, err
		}
		if !hasRow {
			break
		}

		var (
			usr string
			bal int64
		)
		err = stmt.Scan(&usr, &bal)
		if err != nil {
			return nil, err
		}

		tb.Add(usr, USD(bal))
	}
	return tb, nil
}

//GetTags returns a list of distinct tags
func (db DB) GetTags() ([]string, error) {
	sql := `SELECT DISTINCT json_each.value FROM txs, json_each(json_extract(txs.tx, '$.Tags'))`

	conn, err := db.conn()
	if err != nil {
		return nil, err
	}
	defer handleClose(conn)

	stmt, err := conn.Prepare(sql)
	if err != nil {
		return nil, err
	}
	defer handleClose(stmt)

	var tags []string
	for {
		hasRow, err := stmt.Step()
		if !hasRow {
			break
		}

		var tag string
		err = stmt.Scan(&tag)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}
