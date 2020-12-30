package internal

import (
	"errors"
	"os"
	"testing"
)

func TestDb(t *testing.T) {
	//delete old database if exists
	dir := os.TempDir() + "/badgerdbtest"
	err := os.RemoveAll(os.TempDir() + "/badgerdbtest")
	if err != nil {
		t.Fatal(err)
	}

	//test db init
	db, err := OpenDb(dir)
	if err != nil {
		t.Fatal(db)
	}

	//test set
	err = db.Set([]byte{1}, []byte{1})
	if err != nil {
		t.Fatal(db)
	}

	//test get
	val, err := db.Get([]byte{1})
	if err != nil {
		t.Fatal(err)
	}
	if len(val) != 1 || val[0] != byte(1) {
		t.Fatal("got unexpected value: ", string(val[0]))
	}

	//test iterate values
	for i := 0; i < 3; i++ {
		err = db.Set([]byte{0, byte(i)}, []byte{byte(i)})
	}
	//should iterate over all three values
	i := 0
	err = db.IterateValues([]byte{0}, func(v []byte) error {
		if int(v[0]) != i {
			return errors.New("unexpected value: " + string(v[0]))
		}
		i++
		return nil
	})
	if err != nil {
		t.Fatal("error iterating values: ", err)
	}
	if i != 3 {
		t.Fatalf("iterated over unexpected number of values: expected %d got %d", 3, i)
	}
	//should iterate over one key and stop
	i = 0
	err = db.IterateValues([]byte{0}, func(v []byte) error {
		i++
		return ErrBreakIter
	})
	if err != nil {
		t.Fatal("error testing iterating values with ErrBreakIter: ", err)
	}
	if i != 1 {
		t.Fatalf("iterated over unexpected number of values: expected %d got %d", 1, i)
	}
}
