package main

import (
	"fmt"
	"os"
)

func main() {
	homedir := os.Getenv("KST_KBHOME")
	keybase := os.Getenv("KST_KBLOC")
	dbloc := os.Getenv("KST_DBLOC")
	errConvID := os.Getenv("KST_DBGCONV")
	users, err := NewAuthorizedUsers(os.Getenv("KST_USERS"))
	if err != nil {
		panic(err)
	}

	s := new(Server)
	s.SetUsers(users)

	kbc, err := s.Start(keybase, homedir, errConvID)
	if err != nil {
		fmt.Println("error starting server:", err)
		os.Exit(1)
	}

	db := DB(dbloc)
	h := NewHandler(kbc, &db, errConvID)
	err = s.Listen(h)
	if err != nil {
		fmt.Println("error starting listeners", err)
		os.Exit(2)
	}
}
