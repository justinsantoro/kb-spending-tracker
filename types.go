package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func toJsonString(x interface{}) (string, error) {
	var b = new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(x)
	return b.String(), err
}

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

type TxnKey struct {
	table     byte     //table identifier prefix
	month     [6]byte  //month identifier yyyymm
	tag       [32]byte //transaction type tag ie utils:electricity
	timestamp int64    //transaction timestamp
	user      byte     //user identifier
}

//Txn represents a single transaction
type Txn struct {
	Date   Timestamp //the unix timestamp of the transaction
	Amount USD       //the amount of the transaction in cents
	Tag    string    //tags for the transaction
	User   string    //name of user who submitted the tx
}

//String returns the default string representation of a Txn
func (t *Txn) String() string {
	return fmt.Sprintf("%s %s %s on %s", t.User, t.Action(), t.Amount.Abs(), t.Tag)
}

//Returns spent or received depending on whether the txn amnt is positive or negative
func (t *Txn) Action() string {
	if t.Amount < 0 {
		return "spent"
	}
	return "received"
}

//Json returns the Txn as a Json Encoded string
func (t *Txn) Json() (string, error) {
	return toJsonString(t)
}

func ActionString(amt USD) string {
	if amt < 0 {
		return "spent " + amt.Abs().String() + " on"
	}
	return "received " + amt.Abs().String() + " from"
}

type Tags struct {
	persister
	Tagmap map[string]struct{}
}

func LoadTags() (*Tags, error) {
	tags := new(Tags)
	err := tags.Load(tags)
	if tags.Tagmap == nil {
		tags.Tagmap = make(map[string]struct{})
	}
	return tags, err
}

func (t *Tags) IsTag(tag string) bool {
	_, ok := t.Tagmap[tag]
	return ok
}

func (t *Tags) AddTag(tag string) error {
	if !t.IsTag(tag) {
		t.Tagmap[tag] = struct{}{}
	}
	return t.Persist(t)
}

//Tags returns a lexicographically sorted slice of tag strings
func (t *Tags) Tags() []string {
	tags := make([]string, len(t.Tagmap))
	for key, _ := range t.Tagmap {
		tags = append(tags, key)
	}
	sort.Slice(tags, func(i, j int) bool { return tags[i] < tags[j] })
	return tags
}

type TagBalance struct {
	usrs  map[string]USD
	total USD
	tag   string
}

func NewTagBalance(tag string) *TagBalance {
	return &TagBalance{
		make(map[string]USD),
		USD(0),
		tag,
	}
}

func (tb *TagBalance) Add(usr string, bal USD) {
	tb.usrs[usr] = bal
	tb.total += bal
}

func (tb TagBalance) String() string {
	str := fmt.Sprintln(ActionString(tb.total), tb.tag)
	for usr, bal := range tb.usrs {
		percent := bal.InDollars() / tb.total.InDollars() * 100
		str += ">@" + usr + ": " + bal.Abs().String() + fmt.Sprintf(" (%.1f", percent) + "%%)\n"
	}
	return str
}

type persister interface {
	Load(interface{}) error
	Persist(interface{}) error
}

type Users struct {
	persister
	IdToUsername map[byte]string
	UsernameToId map[string]byte
	Admin        byte
}

func LoadUsers() (*Users, error) {
	users := new(Users)
	err := users.Load(users)
	return users, err
}

func (u *Users) Count() int {
	return len(u.UsernameToId)
}

func (u *Users) Username(id byte) string {
	username, ok := u.IdToUsername[id]
	if !ok {
		return ""
	}
	return username
}

func (u *Users) Userid(username string) byte {
	userid, ok := u.UsernameToId[username]
	if !ok {
		return 0
	}
	return userid
}

func (u *Users) IsUser(username string) bool {
	_, ok := u.UsernameToId[username]
	return ok
}

func (u *Users) AddUser(username string) error {
	//keybase usernames cannot be more than 15 chars long or contain spaces
	if len(username) > 15 || strings.Contains(username, " ") {
		return errors.New(fmt.Sprint("Invalid authorized users string:", username))
	}
	//1 based - 0 is reserved for server actions like summary txns
	userid := byte(u.Count() + 1)
	u.UsernameToId[username] = userid
	u.IdToUsername[userid] = username
	return u.Persist(u)
}

func (u *Users) IsAdmin(username string) bool {
	return u.IdToUsername[u.Admin] == username
}

func (u *Users) IsAdminId(id byte) bool {
	return id == u.Admin
}
