package main

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/keybase/go-keybase-chat-bot/kbchat"
	"github.com/keybase/go-keybase-chat-bot/kbchat/types/chat1"
)

const (
	//SPACE is a whitespace
	SPACE = `\s`
	//WORD is a space followed by a word
	WORD = `\s\w+`
	//MONEY is a currency decimal ie 100.00
	MONEY = SPACE + `\d*\.?\d{2}` + SPACE
	//Tags matches either a single tag or a comma-space separated list of tags
	//ie: tag1, tag2, tag3
	TAGS = SPACE + `((\w+,\s)*)?\w+`
)

type command struct {
	Name       string
	Pattern    *regexp.Regexp
	EntryPoint func(cmd []string, msg chat1.MsgSummary) error
}

func (c *command) PatternMatches(cmd string) bool {
	return c.Pattern.MatchString(cmd)
}

type cmdMap map[string]command

func (m cmdMap) add(entryPoint func(cmd []string, msg chat1.MsgSummary) error, pattern ...string) {
	cmd := new(command)
	cmd.Name = pattern[0]
	cmd.EntryPoint = entryPoint
	expr := `(?is)^` + strings.Join(pattern, "") + `(:?\s|$)`
	cmd.Pattern = regexp.MustCompile(expr)
	m[cmd.Name] = *cmd
}

type Handler struct {
	*Output
	db   *DB
	cmds cmdMap
}

func NewHandler(kbc *kbchat.API, db *DB, ErrConvID string) Handler {
	h := Handler{
		Output: NewDebugOutput("handler", kbc, ErrConvID),
		db:     db,
	}
	cmds := make(cmdMap)
	cmds.add(h.HandleStart, "start", MONEY, "?")
	cmds.add(h.HandleSpent, "spent", MONEY, "on", TAGS)
	cmds.add(h.HandleReceived, "received", MONEY, "from", TAGS)
	cmds.add(h.HandleBalance, "balance")
	cmds.add(h.HandleListTags, "list", WORD)
	cmds.add(h.HandleHowMuch, "howmuch", SPACE, "on|from", WORD)
	h.cmds = cmds
	return h
}

func (h *Handler) commandExists(cmdName string) *command {
	cmd := h.cmds[cmdName]
	if len(cmd.Name) > 0 {
		return &cmd
	}
	return nil
}

func (h *Handler) HandleReceived(cmd []string, msg chat1.MsgSummary) error {
	ts := TimestampNow()
	amt, err := StringToUSD(cmd[1])

	if err != nil {
		h.ReactError(msg)
		h.ReactDollar(msg)
		return err
	}
	tags, note := parseTagsAndNote(cmd[3:])
	if tags == nil {
		h.ReactQuestion(msg)
		return errors.New("HandleReceived: couldn't parse tag(s)")
	}
	txn := Txn{
		ts,
		amt,
		tags,
		note,
		msg.Sender.Username,
		false,
	}
	if err := h.db.PutTransaction(txn); err != nil {
		h.ReactError(msg)
		return err
	}
	h.ReactSuccess(msg)
	return nil
}

func (h *Handler) HandleStart(cmd []string, msg chat1.MsgSummary) error {
	ts := TimestampNow()
	amt, err := StringToUSD(cmd[1])
	if err != nil {
		h.ReactError(msg)
		h.ReactDollar(msg)
		return err
	}
	txn := Txn{
		ts,
		amt,
		[]string{},
		"Starting transaction",
		msg.Sender.Username,
		true,
	}
	err = h.db.PutTransaction(txn)
	if err != nil {
		h.ReactError(msg)
	}
	h.ReactSuccess(msg)
	return nil
}

func (h *Handler) HandleSpent(cmd []string, msg chat1.MsgSummary) error {
	ts := TimestampNow()
	amt, err := StringToUSD(cmd[1])

	if err != nil {
		h.ReactError(msg)
		h.ReactDollar(msg)
		return err
	}
	tags, note := parseTagsAndNote(cmd[3:])
	if tags == nil {
		h.ReactQuestion(msg)
		return errors.New("HandleSpent: couldn't parse tag(s)")
	}
	txn := Txn{
		ts,
		-amt,
		tags,
		note,
		msg.Sender.Username,
		false,
	}
	if err := h.db.PutTransaction(txn); err != nil {
		h.ReactError(msg)
		return err
	}
	h.ReactSuccess(msg)
	return nil
}

func (h *Handler) HandleBalance(cmd []string, msg chat1.MsgSummary) error {
	bal, err := h.db.GetBalance(StartOfMonth())
	if err != nil {
		return err
	}
	h.ChatEcho(msg.ConvID, fmt.Sprintf("current balance is **%s**", bal))
	return nil
}

func (h *Handler) HandleListTags(cmd []string, msg chat1.MsgSummary) error {
	if cmd[1] == "tags" {
		tags, err := h.db.GetTags()
		if err != nil {
			return err
		}
		var ts string
		for _, t := range tags {
			ts += fmt.Sprintln(t)
		}
		h.ReactSuccess(msg)
		h.ChatEcho(msg.ConvID, ts)
	}
	return nil
}

func (h *Handler) HandleHowMuch(cmd []string, msg chat1.MsgSummary) error {
	var m *[2]time.Time
	m = CurrentMonthRange()
	if len(cmd) > 4 {
		var ok bool
		m, ok = MonthRangeFromString(cmd[4])
		if !ok {
			h.ReactQuestion(msg)
			h.Debug("HandleHowMuch: invalid month given")
			return nil
		}
	}
	tb, err := h.db.GetTagBalance(cmd[2], m[0], m[1])
	if err != nil {
		return err
	}
	h.ChatEcho(msg.ConvID, tb.String())
	return nil
}

func (h *Handler) HandleMonthSummary(m time.Month) error {
	bal, err := h.db.GetBalance(MonthStart(m))
	if err != nil {
		return err
	}
	txn := Txn{
		TimestampNow(),
		bal,
		[]string{},
		"summary txn",
		"Server",
		true,
	}
	err = h.db.PutTransaction(txn)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) HandleNewConv(conv chat1.ConvSummary) error {
	h.ChatEcho("Ciao! This convID is:", string(conv.Id))
	return nil
}

func (h *Handler) HandleCommand(msg chat1.MsgSummary) error {
	if msg.Content.Text == nil {
		h.Debug("skipping non-text message")
		return nil
	}
	cmdstring := strings.TrimSpace(msg.Content.Text.Body)
	parts := strings.Split(cmdstring, " ")
	name := parts[0]
	//if first word is a command trigger word
	if cmd := h.commandExists(strings.ToLower(name)); cmd != nil {
		// check if required data was given
		if cmd.PatternMatches(cmdstring) {
			//execute command
			return cmd.EntryPoint(parts, msg)
		}
		//command pattern did not match
		h.ReactQuestion(msg)
		h.Debug("cmd %v pattern did not match: %s", name, cmd.Pattern)
		return nil
	}
	return nil
}

func parseTagsAndNote(s []string) ([]string, string) {
	var note string
	tags, ntags := parseTagInput(s)
	if ntags < len(s) {
		note = strings.Join(s[ntags:], " ")
	}
	return tags, note
}

func parseTagInput(tags []string) ([]string, int) {
	l := len(tags)
	switch l {
	case 0:
		return nil, -1
	case 1:
		return []string{tags[0]}, 1
	default:
		tagList := make([]string, 0)
		for i, val := range tags {
			tagList = append(tagList, val[:len(val)-1])
			if !strings.HasSuffix(val, ",") {
				tagList[len(tagList)-1] = val
				return tagList, i + 1
			}
		}
		return nil, -2
	}
}
