package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const usersKey = "users"

type UsersCache struct {
	*cacheStore
	idToUsername map[byte]string
	usernameToId map[string]byte
	admin        byte
}

type usersCacheJSON struct {
	IdToUsername map[byte]string `json:"id_to_username"`
	UsernameToId map[string]byte `json:"username_to_id"`
	Admin        byte            `json:"admin"`
}

func LoadUsersCache(db *db) (*UsersCache, error) {
	var users UsersCache
	store := &cacheStore{
		db:  db,
		key: cacheKey(usersKey),
	}
	err := store.load(users)
	if err != nil {
		return nil, err
	}
	if users.usernameToId == nil {
		users.idToUsername = make(map[byte]string)
		users.usernameToId = make(map[string]byte)
	}
	return &users, nil
}

func (u *UsersCache) Count() int {
	return len(u.usernameToId)
}

func (u *UsersCache) Username(id byte) string {
	username, ok := u.idToUsername[id]
	if !ok {
		return ""
	}
	return username
}

func (u *UsersCache) Userid(username string) byte {
	userid, ok := u.usernameToId[username]
	if !ok {
		return 0
	}
	return userid
}

func (u *UsersCache) IsUser(username string) bool {
	_, ok := u.usernameToId[username]
	return ok
}

func (u *UsersCache) AddUser(username string) error {
	//keybase usernames cannot be more than 15 chars long or contain spaces
	if len(username) > 15 || strings.Contains(username, " ") {
		return errors.New(fmt.Sprint("Invalid username:", username))
	}
	//1 based - 0 is reserved for server actions like summary txns
	userid := byte(u.Count() + 1)
	u.usernameToId[username] = userid
	u.idToUsername[userid] = username
	return u.save(u)
}

func (u *UsersCache) IsAdmin(username string) bool {
	return u.idToUsername[u.admin] == username
}

func (u *UsersCache) IsAdminId(id byte) bool {
	return id == u.admin
}

func (u *UsersCache) MarshallJSON() ([]byte, error) {
	return json.Marshal(&usersCacheJSON{
		IdToUsername: u.idToUsername,
		UsernameToId: u.usernameToId,
		Admin:        u.admin,
	})
}

func (u *UsersCache) UnmarshalJSON(data []byte) error {
	var temp *usersCacheJSON
	if err := json.Unmarshal(data, &temp); err != nil {
		return nil
	}

	u.usernameToId = temp.UsernameToId
	u.idToUsername = temp.IdToUsername
	u.admin = temp.Admin
}
