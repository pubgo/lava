package fields

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	var is = assert.New(t)
	var req = new(struct {
		Ordering  string  `query:"ordering,omitempty"`
		StartedAt *Int    `field:"started_at"`
		EndedAt   *String `field:"ended_at"`
		Offset    int64   `query:"offset,omitempty"`
		Limit     int64   `query:"limit,omitempty"`
	})

	var err = bindBuild(req, map[string][]string{
		"started_at__gt": {"1637205860"},
		"ended_at__gt":   {"1637205860"},
		"limit":          {"123"}})
	if err != nil {
		t.Fatal(err)
	}

	is.Equal(req.StartedAt.name, "started_at")
	is.Equal(req.EndedAt.name, "ended_at")
	is.NotNil(req.StartedAt.value["gt"])
	is.NotNil(req.EndedAt.value["gt"])
}
