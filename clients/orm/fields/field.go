package fields

import (
	"database/sql/driver"

	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm/schema"
)

type Field interface {
	setName(name string)
	handler(s string) (driver.Valuer, error)
	setValue(value map[string]driver.Valuer)
	Expr(table schema.Tabler) []gen.Condition
}

type fieldImpl struct {
	name  string
	value map[string]driver.Valuer
}

func (t *fieldImpl) setName(name string) {
	t.name = name
}

func (t *fieldImpl) setValue(value map[string]driver.Valuer) {
	t.value = value
}

func (t *fieldImpl) Expr(table schema.Tabler) []gen.Condition {
	var f = field.NewField(table.TableName(), t.name)
	var expr = make([]gen.Condition, 0, len(t.value))
	for k, v := range t.value {
		expr = append(expr, handle(f, k, v))
	}
	return expr
}

func handle(t field.Field, k string, vv driver.Valuer) field.Expr {
	switch k {
	case "eq", "exact", "":
		return t.Eq(vv)
	case "gt":
		return t.Gt(vv)
	case "lt":
		return t.Lt(vv)
	case "gte":
		return t.Gte(vv)
	case "lte":
		return t.Lte(vv)
	case "neq":
		return t.Neq(vv)
	}
	return nil
}
