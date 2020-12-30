package internal

import (
	"encoding/json"
)

const (
	cacheTable = byte(0)
)

type cacheStore struct {
	*db
	key []byte
}

func (c *cacheStore) load(v interface{}) error {
	b, err := c.db.Get(c.key)
	if err != nil {
		return ErrWithMessage(err, "db get")
	}
	if v != nil {
		err := json.Unmarshal(b, v)
		if err != nil {
			return ErrWithMessage(err, "error unmarshaling")
		}
	}
	return nil
}

func (c *cacheStore) save(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return ErrWithMessage(err, "error marshaling")
	}
	err = c.db.Set(c.key, b)
	if err != nil {
		return ErrWithMessage(err, "db set")
	}
	return nil
}

func cacheKey(s string) []byte {
	k := []byte{cacheTable}
	k = append(k, []byte(s)...)
	return k
}
