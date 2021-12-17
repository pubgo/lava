package restc

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/goccy/go-json"
	"github.com/valyala/bytebufferpool"
)

func getBodyReader(rawBody interface{}) ([]byte, error) {
	switch body := rawBody.(type) {
	// If a regular byte slice, we can read it over and over via new
	// readers
	case []byte:
		return body, nil

	// If a bytes.Buffer we can read the underlying byte slice over and
	// over
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

	default:
		bb := bytebufferpool.Get()
		defer bytebufferpool.Put(bb)

		if err := json.NewEncoder(bb).Encode(rawBody); err != nil {
			return nil, err
		}

		return bb.Bytes(), nil
	}
}
