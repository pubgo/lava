package sqlmock

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm/schema"
)

func TestAnyTime(t *testing.T) {
	o := AnyTime{}
	assert.True(t, o.Match(time.Now()))
	assert.False(t, o.Match(1))
	assert.False(t, o.Match("S"))
	assert.False(t, o.Match(true))
}

func TestMock(t *testing.T) {
	db, m := NewMockPG(t)
	assert.NotNil(t, db)
	assert.NotNil(t, m)

	_, err := GetColumns("string")
	assert.ErrorIs(t, err, schema.ErrUnsupportedDataType)

	_, err = GetValues("string", []string{})
	assert.ErrorIs(t, err, schema.ErrUnsupportedDataType)

	cols, err := GetColumns(&User{})
	assert.NoError(t, err)
	assert.Equal(t, []string{"id", "name", "active"}, cols)

	rows := MockRows(
		&User{ID: 1, Name: "Alice", Active: true},
		&User{ID: 2, Name: "Bob", Active: false},
		&User{ID: 3, Name: "Charlie", Active: true},
	)
	expected := sqlmock.NewRows(cols).
		AddRow(1, "Alice", true).
		AddRow(2, "Bob", false).
		AddRow(3, "Charlie", true)
	assert.Equal(t, expected, rows)
}

type User struct {
	ID     int
	Name   string
	Active bool
}
