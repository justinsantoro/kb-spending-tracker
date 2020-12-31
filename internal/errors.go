package internal

import "errors"

var ErrBreakIter = errors.New("done iterating")
var ErrReservedTag = errors.New("tag is reserved")

type withMessage struct {
	msg string
	err error
}

func (wm *withMessage) Error() string {
	return wm.msg + ": " + wm.err.Error()
}

func ErrWithMessage(err error, msg string) error {
	return &withMessage{
		msg,
		err,
	}
}
