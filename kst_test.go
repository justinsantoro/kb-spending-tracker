package main

import (
	"fmt"
	"golang.org/x/sync/errgroup"
	"os"
	"testing"
	"time"
)

func TestDb(t *testing.T) {
	db := DB("test.db")

	conn, err := db.conn()
	if err != nil {
		t.Error(err)
	}
	handleClose(conn)

	if err = db.Init(); err != nil {
		t.Error(err)
	}

	txn := Txn{
		TimestampNow(),
		-10 * 100,
		[]string{"nugget", "cat-food", "cat-toys"},
		"Catfood and nip",
		"Sarah",
		false,
	}
	AmntTotal := USD(0)
	var FirstTs time.Time
	for i := 1; i < 6; i++ {
		txn.User = "Sarah"
		if i%2 == 0 {
			txn.User = "Justin"
		}
		txn.Date = TimestampNow()
		if i == 1 {
			FirstTs = txn.Date.Time()
		}
		txn.Amount = txn.Amount.Times(float64(i))
		AmntTotal += txn.Amount
		if err := db.PutTransaction(txn); err != nil {
			t.Error(err)
		}
	}

	bal, err := db.GetBalance(FirstTs)
	if err != nil {
		t.Error(err)
	}
	if bal != AmntTotal {
		t.Errorf("incorrect balance. expected %s, got %s", AmntTotal, bal)
	}

	txs, err := db.GetTransactions(FirstTs, time.Now())
	if err != nil {
		t.Error(err)
	}
	if l := len(txs); l != 5 {
		t.Error("not enough txs from GetTransactions: Expected 5 got", l)
	}

	txs, err = db.GetTransactionsSince(txs[1].Date.Time())
	if err != nil {
		t.Error(err)
	}
	if l := len(txs); l != 4 {
		t.Error("not correct amount of txs returned. Expected 4 got", l)
	}

	tags, err := db.GetTags()
	if err != nil {
		t.Error(err)
	}

	if l := len(tags); l != 3 {
		t.Error("not correct amount of tags returned. Expected 3 got", l)
	}

	tb, err := db.GetTagBalance("nugget", FirstTs, time.Now())
	if err != nil {
		t.Error(err)
	}

	if l := len(tb.usrs); l != 2 {
		t.Error("GetTagBalance: not correct amount of users returned. Expected 2 got", l)
	}
	fmt.Println(tb)

	//test Balancer
	ts := time.Now()
	shutdownCh := make(chan struct{})
	heartbeat := make(chan struct{})
	handler := NewHandler(nil, &db, "1234")
	s := new(Server)
	s.Output = NewDebugOutput("test", nil, "")
	var eg errgroup.Group
	eg.Go(func() error { return s.waitToBalance(shutdownCh, handler, ts.Add(2*time.Second), heartbeat) })

	select {
	case <-heartbeat:
		close(shutdownCh)
	case <-time.After(3 * time.Second):
		t.Error("Balancer timed out")
	}

	if err := eg.Wait(); err != nil {
		t.Error(err)
	}

	nbal, err := db.GetBalance(ts)
	if err != nil {
		t.Error(err)
	}
	if nbal != bal {
		t.Error(fmt.Sprintf("Incorrect balance. Got %s expected %s", nbal, bal))
	}
}

func TestAuthorizedUsers(t *testing.T) {
	usr1 := "username1"
	usr2 := "username2"
	userstr := usr1 + "," + usr2

	auth, err := NewAuthorizedUsers(userstr)
	if err != nil {
		t.Error("error creating Autorized users:", err)
	}
	s := new(Server)
	s.SetUsers(auth)
	if !s.IsUser(usr1) {
		t.Error("User Auth error: " + usr1 + "not authorized")
	}
	if !s.IsUser(usr2) {
		t.Error("User Auth error: " + usr2 + "not authorized")
	}
	fmt.Print(auth)

}

func TestMain(m *testing.M) {
	x := m.Run()
	err := os.Remove("test.db")
	if err != nil {
		fmt.Print(err)
	}
	os.Exit(x)
}
