package types

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/pubgo/xerror"
)

var _ json.Unmarshaler = (*Duration)(nil)
var _ json.Marshaler = (*Duration)(nil)

func Dur(dur time.Duration) Duration {
	return Duration{Duration: dur}
}

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(d.String()), nil
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	data = bytes.Trim(data, "\"")

	if len(data) == 0 {
		return nil
	}

	dur, err := time.ParseDuration(string(data))
	if err != nil {
		return xerror.WrapF(err, "data: %s", data)
	}

	d.Duration = dur

	return nil
}
