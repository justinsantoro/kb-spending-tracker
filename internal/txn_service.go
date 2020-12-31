package internal

import "time"

type TxnService struct {
	store *txnStore
	tags *TagsCache
}

const summaryTxnTag = "summary"

func (txs *TxnService) AddTransaction(amount USD, tag string, date *time.Time, txid string, uid byte) error {
	if tag == summaryTxnTag {
		return ErrReservedTag
	}
	err := txs.store.AddTxn(&Txn{
		Date:   date,
		Amount: amount,
		Tag:    tag,
		User:   uid,
	})
	if err != nil {
		return err
	}
	return txs.tags.AddTag(tag)
}

func (txs *TxnService) GetBalance(month *time.Time) (USD, error) {
	var bal USD
	err := txs.store.IterateTxnValues(timeToMonthPrefix(month), func(txn *Txn) error {
		bal += txn.Amount
		return nil
	})
	return bal, err
}

func (txs *TxnService) GetTags() []string {
	return txs.tags.Tags()
}

func (txs *TxnService) GetTagBalance(month *time.Time, tag string) (USD, error) {
	var bal USD
	pre := timeToMonthPrefix(month) + tag
	err := txs.store.IterateTxnValues(pre, func(txn *Txn) error {
		bal += txn.Amount
		return nil
	})
	return bal, err
}
