package orm

import (
	"errors"
	"gorm.io/gorm/clause"

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

var _ clause.Expression = (*expr)(nil)
var _ gen.Condition = (*expr)(nil)

type expr struct {
	*clause.Expr
}

func Cond(sql string, args ...interface{}) gen.Condition {
	return expr{&clause.Expr{SQL: sql, Vars: args}}
}

func (s expr) BeCond() interface{} { return s.Expr }
func (s expr) CondError() error    { return nil }
