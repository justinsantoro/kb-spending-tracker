// +build !nodb

package internal

import (
	badger "github.com/dgraph-io/badger/v2"
	"time"
)

type db struct {
	internal *badger.DB
}

//OpenDb opens a badgerdb in dir returning
//and wraps it in a db type
func OpenDb(dir string) (d *db, err error) {
	d = new(db)
	opts := badger.DefaultOptions(dir)
	opts.Logger = nil
	d.internal, err = badger.Open(opts)
	return
}

//Close closes the db
func (db *db) Close() error {
	return db.internal.Close()
}

//IsClosed returns true if the underlying db is closed
func (db *db) IsClosed() bool {
	return db.internal.IsClosed()
}

//Get returns a copy of the value with the given key
func (db *db) Get(key []byte) ([]byte, error) {
	var v []byte
	err := db.internal.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			v = append([]byte{}, val...)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if err != badger.ErrKeyNotFound {
			return nil, err
		}
		return nil, nil
	}
	return v, nil
}

//Sets the value at the given key
func (db *db) Set(key []byte, value []byte) error {
	return db.internal.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, value)
		return err
	})
}

//SetWithMetadata sets a value with a metadata byte
func (db *db) SetWithMetadata(key []byte, value []byte, meta byte) error {
	return db.internal.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(key, value).WithMeta(meta)
		return txn.SetEntry(e)
	})
}

//SetWithTTL sets a value with a Time To Live attribute
func (db *db) SetWithTTL(key []byte, value []byte, ttl time.Duration) error {
	return db.internal.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry(key, value).WithTTL(ttl)
		return txn.SetEntry(e)
	})
}

//Iterate values iterates over keys mathing the giving prefix, running f func
//on each value. f may return ErrBreakIter to break out of iteration early
func (db *db) IterateValues(prefix []byte, f func(v []byte) error) error {
	return db.internal.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			err := item.Value(f)
			if err != nil {
				if err != ErrBreakIter {
					return err
				}
				break
			}
		}
		return nil
	})
}

//IterateKeys iterates over keys only. The underlying iterator
//does not prefetch any values. The value of each key is passed to f. Iteration can be stopped
//early if f returns ErrBreakIter
func (db *db) IterateKeys(prefix []byte, f func(k []byte) error) error {
	return db.internal.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			err := f(it.Item().Key())
			if err != nil {
				if err != ErrBreakIter {
					return err
				}
				break
			}
		}
		return nil
	})
}

//SparseRead allows doing a sparse read of values. The underlying iterator does not prefect any values.
//Each key is passed into kfunc. if kfunc returns true, nil than the value of the key will be passed
//into vfunc. If either functions return ErrBreakIter than iteration will be stopped early.
func (db *db) SparseRead(prefix []byte, kfunc func (k []byte) (bool, error), vfunc func(v []byte) error) error {
	return db.internal.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			ok, err := kfunc(it.Item().Key())
			if err != nil {
				if err != ErrBreakIter {
					return err
				}
				break
			}
			if ok {
				err = it.Item().Value(vfunc)
				if err != nil {
					if err != ErrBreakIter {
						return err
					}
					break
				}
			}
		}
		return nil
	})
}
