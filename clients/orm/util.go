package orm

import (
	"errors"

	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
)

func ErrNotFound(err error) bool {
	if err == gorm.ErrRecordNotFound {
		return true
	}

	return errors.Is(err, gorm.ErrRecordNotFound)
}

func Columns(cols ...field.Expr) gen.Columns { return cols }
