package ormutil

import (
	"errors"

	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// https://github.com/acmestack/gorm-plus/tree/main/gplus

func ErrNotFound(err error) bool {
	if err == gorm.ErrRecordNotFound {
		return true
	}

	return errors.Is(err, gorm.ErrRecordNotFound)
}

func SubQuery(cols ...field.Expr) gen.Columns { return cols }

var _ gen.Condition = (*expr)(nil)

type expr struct {
	field.Expr
	expr *clause.Expr
}

func RawCond(sql string, args ...interface{}) gen.Condition {
	return expr{expr: &clause.Expr{SQL: sql, Vars: args}}
}

func (s expr) BeCond() interface{} { return s.expr }
func (s expr) CondError() error    { return nil }
