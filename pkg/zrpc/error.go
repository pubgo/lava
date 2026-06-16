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
// Used by server handlers when decode/validation fails before middleware runs.
func ReplyError(msg *nats.Msg, code Code, text string) {
	ReplyErrorWithHeader(msg, nil, code, text)
}

// ReplyErrorWithHeader sends a zrpc error reply on NATS and preserves custom headers.
func ReplyErrorWithHeader(msg *nats.Msg, header nats.Header, code Code, text string) {
	if msg == nil || msg.Reply == "" {
		return
	}

	h := cloneHeader(header)
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

// StatusFromMessage maps a zrpc reply message to its embedded status code.
func StatusFromMessage(msg *nats.Msg) (Code, string) {
	if msg == nil {
		return CodeInternal, "nil message"
	}

	if msg.Header == nil {
		return CodeOK, ""
	}

	codeStr := msg.Header.Get(HeaderStatusCode)
	if codeStr == "" || codeStr == "0" {
		return CodeOK, ""
	}

	codeInt, err := strconv.Atoi(codeStr)
	if err != nil {
		return CodeInternal, fmt.Sprintf("invalid zrpc status code: %s", codeStr)
	}

	return Code(codeInt), msg.Header.Get(HeaderStatusMessage)
}

// ErrorFromMessage converts a zrpc reply message to a typed error when needed.
func ErrorFromMessage(msg *nats.Msg) error {
	code, text := StatusFromMessage(msg)
	if code == CodeOK {
		return nil
	}

	return &Status{Code: code, Message: text}
}
