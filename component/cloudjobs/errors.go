package eventjob

import "github.com/pubgo/funk/errors"

var errReject = errors.New("asyncjob: reject retry and discard msg")

func Reject(errs ...error) error {
	var reason = "reject"
	if len(errs) > 0 {
		reason = errs[0].Error()
	}
	return errors.Wrap(errReject, reason)
}

func isRejectErr(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, errReject)
}
