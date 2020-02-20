package main

import (
	"errors"
	"fmt"
	"github.com/keybase/go-keybase-chat-bot/kbchat"
	"github.com/keybase/go-keybase-chat-bot/kbchat/types/chat1"
	"golang.org/x/sync/errgroup"
	"os"
	"sync"
)

type CmdHandler interface {
	HandleCommand(chat1.MsgSummary) error
	HandleNewConv(chat1.ConvSummary) error
	BuildCommandMap()
}

type Server struct {
	*Output
	sync.Mutex
	shutdownCh chan struct{}
	kbc        *kbchat.API
	Users AuthorizedUsers
}

func (s *Server) SetUsers(users AuthorizedUsers) {
	s.Users = users
}

func (s *Server) IsUser(username string) bool {
	_, ok := s.Users[username]
	return ok
}

func (s *Server) Start(keybaseLoc, home string, ErrorConvId string) (kbc *kbchat.API, err error) {
	if s.kbc, err = kbchat.Start(kbchat.RunOptions{
		KeybaseLocation: keybaseLoc,
		HomeDir:         home,
	}); err != nil {
		return s.kbc, err
	}
	s.Output = NewDebugOutput("server", s.kbc, ErrorConvId)
	return s.kbc, nil
}

func (s *Server) Listen(handler Handler) error {
	sub, err := s.kbc.Listen(kbchat.ListenOptions{Convs: true})
	if err != nil {
		s.Debug("Listen: failed to listen: %s", err)
		return err
	}
	s.Debug("startup success, listening for messages and convs...")
	s.Lock()
	shutdownCh := s.shutdownCh
	s.Unlock()
	var eg errgroup.Group
	eg.Go(func() error { return s.listenForMsgs(shutdownCh, sub, handler) })
	eg.Go(func() error { return s.listenForConvs(shutdownCh, sub, handler) })
	eg.Go(func() error { return s.waitToBalance(shutdownCh, handler, EndOfMonth(), nil)})
	if err := eg.Wait(); err != nil {
		s.Debug("wait error: %s", err)
		return err
	}
	s.Debug("Listen: shut down")
	return nil
}

func (s *Server) listenForMsgs(shutdownCh chan struct{}, sub *kbchat.NewSubscription, handler Handler) error {
	for {
		select {
		case <-shutdownCh:
			s.Debug("listenForMsgs: shutting down")
			return nil
		default:
		}

		m, err := sub.Read()
		if err != nil {
			s.Debug("listenForMsgs: Read() error: %s", err)
			continue
		}
		usr := m.Message.Sender.Username
		if !s.IsUser(usr) {
			if usr != os.Getenv("KEYBASE_USERNAME") {
				s.Debug("Ignoring message from %s", usr)
			}
			continue
		}

		msg := m.Message
		s.Debug("convid = %v", m.Conversation.Id)
		if err := handler.HandleCommand(msg); err != nil {
			s.ChatDebug(msg.ConvID, "listenForMsgs: unable to HandleCommand: %v", err)
		}
	}
}

func (s *Server) listenForConvs(shutdownCh chan struct{}, sub *kbchat.NewSubscription, handler Handler) error {
	for {
		select {
		case <-shutdownCh:
			s.Debug("listenForConvs: shutting down")
			return nil
		default:
		}

		c, err := sub.ReadNewConvs()
		if err != nil {
			s.Debug("listenForConvs: ReadNewConvs() error: %s", err)
			continue
		}

		if !s.IsUser(c.Conversation.CreatorInfo.Username) {
			s.Debug("Ignored new conversation created by %s", c.Conversation.CreatorInfo.Username)
			return nil
		}

		if err := handler.HandleNewConv(c.Conversation); err != nil {
			s.Debug("listenForConvs: unable to HandleNewConv: %v", err)
		}
	}
}

func (s *Server) waitToBalance(shutdownCh chan struct{}, handler Handler, startTrigger Timestamp, heartbeat chan struct{}) error {
	triggerTime := startTrigger
	for {
		select {
		case <-shutdownCh:
			s.Debug("listenForConvs: shutting down")
			return nil
		default:
		}
		if TimestampNow().After(triggerTime.Time) {
			err := handler.HandleMonthSummary(triggerTime.Month())
			if err != nil {
				return errors.New(fmt.Sprint("error handling month summary:", err))
			}
			triggerTime = EndOfMonth()
			if heartbeat != nil {
				heartbeat<-struct{}{}
			}
		}
	}
}

