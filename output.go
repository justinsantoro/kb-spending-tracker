package main

import (
	"fmt"
	"github.com/keybase/go-keybase-chat-bot/kbchat"
	"github.com/keybase/go-keybase-chat-bot/kbchat/types/chat1"
)

type Output struct {
	name          string
	KBC           *kbchat.API
	ErrReportConv string
}

func NewDebugOutput(name string, kbc *kbchat.API, errConv string) *Output {
	return &Output{
		name:          name,
		KBC:           kbc,
		ErrReportConv: errConv,
	}
}

func (d *Output) Debug(msg string, args ...interface{}) {
	fmt.Printf(d.name+": "+msg+"\n", args...)
}

func (d *Output) ChatDebug(convID chat1.ConvIDStr, msg string, args ...interface{}) {
	d.Debug(msg, args...)
	if _, err := d.KBC.SendMessageByConvID(convID, "Something went wrong!"); err != nil {
		d.Debug("ChatDebug: failed to send error message: %s", err)
	}
}

func (d *Output) ReactSuccess(msg chat1.MsgSummary) {
	d.react(msg.ConvID, msg.Id, "✔")
}

func (d *Output) ReactError(msg chat1.MsgSummary) {
	d.react(msg.ConvID, msg.Id, "❗")
}

func (d *Output) ReactDollar(msg chat1.MsgSummary) {
	d.react(msg.ConvID, msg.Id, "💲")

}

func (d *Output) ReactQuestion(msg chat1.MsgSummary) {
	d.react(msg.ConvID, msg.Id, "❓")

}

func (d *Output) react(convID chat1.ConvIDStr, msgID chat1.MessageID, reaction string) {
	if _, err := d.KBC.ReactByConvID(convID, msgID, reaction); err != nil {
		d.Debug("ChatConfirm: failed to react to message", err)
	}
}

func (d *Output) ChatEcho(convID chat1.ConvIDStr, msg string, args ...interface{}) {
	if _, err := d.KBC.SendMessageByConvID(convID, msg, args...); err != nil {
		d.Debug("ChatEcho: failed to send echo message", err)
	}
}

//Notify broadcasts the given message
func (d *Output) Notify(args ...interface{}) {
	if _, err := d.KBC.Broadcast(fmt.Sprint(args...)); err != nil {
		d.Debug("Notify: failed to broadcast message", err)
	}
}
