package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// USD represents US dollar amount in terms of cents
type USD int64

// ToUSD converts a float64 to USD
// e.g. 1.23 to $1.23, 1.345 to $1.35
func ToUSD(f float64) USD {
	return USD((f * 100) + 0.5)
}

//StringToUSD converts a string representation of a currency amount
//to a USD
func StringToUSD(s string) (USD, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return -1, err
	}
	x := ToUSD(f)
	return x, nil
}

// InDollars converts a USD to float64
func (m USD) InDollars() float64 {
	x := float64(m)
	x = x / 100
	return x
}

// Multiply safely multiplies a USD value by a float64, rounding
// to the nearest cent.
func (m USD) Times(x float64) USD {
	y := (float64(m) * x) + 0.5
	return USD(y)
}

// String returns a formatted USD value
func (m USD) String() string {
	return fmt.Sprintf("$%.2f", m.InDollars())
}

//Abs returns the absolute value of the USD
func (m USD) Abs() USD {
	x := int64(m)
	if x < 0 {
		return USD(-x)
	}
	return USD(x)
}

//MarshallJson marshals a USD amount in cents to a Json number
func (m USD) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(m))
}

//UnmarshalJson un-marshals a json number in USD amount in cents
func (m *USD) UnmarshalJSON(b []byte) error {
	var x int64
	if err := json.Unmarshal(b, &x); err != nil {
		return err
	}
	*m = USD(x)
	return nil
}

const txnTable = byte(1)

type txnKey struct {
	table          byte               //table identifier prefix
	month          [6]byte            //month identifier yyyymm
	tag            [MaxTagLength]byte //transaction type tag ie utils:electricity
	timestamp      int64              //transaction timestamp
	timestampBytes []byte             //transaction timestamp encoded as json number
}

func padTag(t string) string {
	if l := len(t); l < MaxTagLength {
		for i := l; i < 32; i++ {
			t += " "
		}
	}
	return t
}

func timeToMonthPrefix(t *time.Time) string {
	return t.Format("200601")
}

func newTxnKey(txn *Txn) (*txnKey, error) {
	var month [6]byte
	copy(month[:], timeToMonthPrefix(txn.Date))

	var tag [MaxTagLength]byte
	copy(tag[:], padTag(txn.Tag))

	//encode timestamp as bytes
	timestamp := new(bytes.Buffer)
	err := json.NewEncoder(timestamp).Encode(txn.Date)
	if err != nil {
		return nil, err
	}

	return &txnKey{
		table:          txnTable,
		month:          month,
		tag:            [32]byte{},
		timestampBytes: timestamp.Bytes(),
	}, nil
}

func txnKeyFromBytes(b []byte) (*txnKey, error) {
	var (
		month     [6]byte
		tag       [32]byte
		timestamp int64
	)
	copy(month[:], b[1:7])
	copy(tag[:], b[8:MaxTagLength-8])
	tbytes := b[MaxTagLength+8:]
	tbuff := bytes.NewBuffer(tbytes)
	err := json.NewDecoder(tbuff).Decode(timestamp)
	if err != nil {
		return nil, err
	}
	return &txnKey{
		table:          b[0],
		month:          month,
		tag:            tag,
		timestamp:      timestamp,
		timestampBytes: b[MaxTagLength+8:],
	}, nil
}

func (k *txnKey) Bytes() []byte {
	key := []byte{txnTable}
	key = append(key, k.month[:]...)
	key = append(key, k.tag[:]...)
	key = append(key, k.timestampBytes...)
	return key
}

//Txn represents a single transaction
type Txn struct {
	Date   *time.Time //the unix timestamp of the transaction
	Amount USD        //the amount of the transaction in cents
	Tag    string     //tags for the transaction
	Id     string     //txn unique id
	User   byte       //id of user who submitted the tx
}

type txnStore struct {
	db *db
}

func txnPrefix(prefix string) []byte {
	p := []byte{txnTable}
	return append(p, []byte(prefix)...)
}

func (s *txnStore) AddTxn(txn *Txn) error {
	k, err := newTxnKey(txn)
	if err != nil {
		return ErrWithMessage(err, "newTxnKey")
	}
	b, err := json.Marshal(txn)
	if err != nil {
		return ErrWithMessage(err, "jsonMarshal")
	}
	return s.db.SetWithMetadata(k.Bytes(), b, txn.User)
}

func (s *txnStore) IterateTxnValues(prefix string, f func(txn *Txn) error) error {
	p := txnPrefix(prefix)
	return s.db.IterateValues(p, func(b []byte) error {
		var txn *Txn
		err := json.Unmarshal(b, txn)
		if err != nil {
			return ErrWithMessage(err, "jsonUnmarshal")
		}
		return f(txn)
	})
}

func (s *txnStore) IterateTxnKeys(prefix string, f func(key *txnKey) error) error {
	p := txnPrefix(prefix)
	return s.db.IterateKeys(p, func(k []byte) error {
		key, err := txnKeyFromBytes(k)
		if err != nil {
			return ErrWithMessage(err, "txnKeyFromBytes")
		}
		return f(key)
	})
}

func (s *txnStore) CountTxns(prefix string) (int, error) {
	p := txnPrefix(prefix)
	var i int
	err := s.db.IterateKeys(p, func(k []byte) error {
		i++
		return nil
	})
	return i, err
}
