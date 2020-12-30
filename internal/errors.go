package internal

import "errors"

var ErrBreakIter = errors.New("done iterating")

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
