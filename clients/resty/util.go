package resty

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/url"

	"github.com/goccy/go-json"
	"github.com/pubgo/funk/convert"
	"github.com/pubgo/funk/strutil"
	"github.com/valyala/bytebufferpool"
)

func getBodyReader(rawBody interface{}) ([]byte, error) {
	switch body := rawBody.(type) {
	case nil:
		return nil, nil
	case []byte:
		return body, nil
	case string:
		return convert.StoB(body), nil
	case *bytes.Buffer:
		return body.Bytes(), nil

	// We prioritize *bytes.Reader here because we don't really want to
	// deal with it seeking so want it to match here instead of the
	// io.ReadSeeker case.
	case *bytes.Reader:
		buf, err := ioutil.ReadAll(body)
		if err != nil {
			return nil, err
		}
		return buf, nil

	// Compat case
	case io.ReadSeeker:
		_, err := body.Seek(0, 0)
		if err != nil {
			return nil, err
		}

		buf, err := ioutil.ReadAll(body)
		if err != nil {
			return nil, err
		}
		return buf, nil

	// Read all in so we can reset
	case io.Reader:
		buf, err := ioutil.ReadAll(body)
		if err != nil {
			return nil, err
		}
		return buf, nil

	case url.Values:
		return strutil.ToBytes(body.Encode()), nil

	case json.Marshaler:
		return body.MarshalJSON()

	default:
		bb := bytebufferpool.Get()
		defer bytebufferpool.Put(bb)

		if err := json.NewEncoder(bb).Encode(rawBody); err != nil {
			return nil, err
		}

		return bb.Bytes(), nil
	}
}
