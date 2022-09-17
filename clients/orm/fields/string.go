package fields

import (
	"database/sql"
	"database/sql/driver"
)

var _ Field = (*String)(nil)

type String struct {
	fieldImpl
}

func (s2 *String) handler(s string) (driver.Valuer, error) {
	var vv sql.NullString
	if err := vv.Scan(s); err != nil {
		return nil, err
	}
	return vv, nil
}
