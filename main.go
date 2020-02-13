package main

import (
	"fmt"
	"os"
)

func main() {
	homedir := os.Getenv("KFT_KBHOME")
	keybase := os.Getenv("KFT_KBLOC")
	dbloc := os.Getenv("KFT_DBLOC")
	errConvID := os.Getenv("KFT_DBGCONV")
	users, err := NewAuthorizedUsers(os.Getenv("KFT_USERS"))
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

