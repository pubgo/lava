package zrpc

import (
	"fmt"
	"strconv"

	"github.com/nats-io/nats.go"
)

const (
	HeaderStatusCode    = "Zrpc-Status-Code"
	HeaderStatusMessage = "Zrpc-Status-Message"
)

// Code is a zrpc status code.
type Code int

const (
	CodeOK              Code = 0
	CodeInvalidArgument Code = 3
	CodeInternal        Code = 13
)

// Status is an application error returned from handlers.
type Status struct {
	Code    Code
	Message string
}

func (s *Status) Error() string {
	if s == nil {
		return "zrpc: nil status"
	}

	return fmt.Sprintf("zrpc %d: %s", s.Code, s.Message)
}

// Errorf builds a *Status for handler return values.
func Errorf(code Code, format string, args ...any) error {
	return &Status{Code: code, Message: fmt.Sprintf(format, args...)}
}

// ReplyError sends a zrpc error reply on NATS.
func ReplyError(msg *nats.Msg, code Code, text string) {
	if msg == nil || msg.Reply == "" {
		return
	}

	h := nats.Header{}
	h.Set(HeaderStatusCode, strconv.Itoa(int(code)))
	h.Set(HeaderStatusMessage, text)
	_ = msg.RespondMsg(&nats.Msg{Header: h, Data: []byte(text)})
}

// StatusFromError maps handler errors to a zrpc code.
func StatusFromError(err error) (Code, string) {
	if err == nil {
		return CodeOK, ""
	}

	if st, ok := err.(*Status); ok && st != nil {
		return st.Code, st.Message
	}

	return CodeInternal, err.Error()
}
