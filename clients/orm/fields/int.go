package fields

import (
	"database/sql"
	"database/sql/driver"
)

var _ Field = (*Int)(nil)

type Int struct {
	fieldImpl
}

func (i *Int) handler(s string) (driver.Valuer, error) {
	var vv sql.NullInt64
	if err := vv.Scan(s); err != nil {
		return nil, err
	}
	return vv, nil
}
