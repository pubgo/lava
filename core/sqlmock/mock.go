package sqlmock

import (
	"context"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/tidwall/match"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type TestingTB interface {
	// Name Returns current test name.
	Name() string
	Cleanup(f func())
	Logf(fmt string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Errorf(message string, args ...interface{})
}

func NewResult(lastInsertID, rowsAffected int64) driver.Result {
	return sqlmock.NewResult(lastInsertID, rowsAffected)
}

func AnyArg() sqlmock.Argument {
	return sqlmock.AnyArg()
}

func AnyArgs(n int) (args []driver.Value) {
	for i := 0; i < n; i++ {
		args = append(args, sqlmock.AnyArg())
	}
	return
}

type AnyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func NewMockPG(tb TestingTB) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(func(expectedSQL, actualSQL string) error {
		expectedSQL = strings.TrimSpace(strings.ReplaceAll(expectedSQL, "**", "*"))
		actualSQL = strings.TrimSpace(strings.ReplaceAll(actualSQL, "  ", " "))

		actualSQLUpper := strings.ToUpper(actualSQL)
		expectedSQLUpper := strings.ToUpper(expectedSQL)
		if actualSQLUpper == expectedSQLUpper || match.Match(actualSQLUpper, expectedSQLUpper) {
			return nil
		}

		tb.Logf("sql not match\n expectedSQL => %s \n actualSQL   => %s \n matchSQL    => %v",
			expectedSQL, actualSQL, match.Match(actualSQLUpper, expectedSQLUpper))

		return fmt.Errorf(`could not match actual sql: "%s" with expected regexp "%s"`, actualSQL, expectedSQL)
	})))
	if err != nil {
		tb.Fatalf("failed to create sql mock, err=%w", err)
		return nil, nil
	}

	tb.Cleanup(func() {
		err = mock.ExpectationsWereMet()
		if err != nil {
			tb.Fatalf("failed to ExpectationsWereMet, err=%w", err)
		}
	})

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  "sqlmock_db_0",
		DriverName:           "postgres",
		Conn:                 db,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		tb.Fatalf("failed to create gorm, err=%w", err)
		return nil, nil
	}

	return gormDB, mock
}

func NewRows(columns ...string) *sqlmock.Rows {
	return sqlmock.NewRows(columns)
}

func MockRows[T any](objs ...*T) *sqlmock.Rows {
	var t T
	columns, _ := GetColumns(&t)
	rows := sqlmock.NewRows(columns)
	for _, w := range objs {
		vals, _ := GetValues(w, columns)
		rows.AddRow(vals...)
	}
	return rows
}

func GetValues(dest any, columns []string) ([]driver.Value, error) {
	s, err := schema.Parse(dest, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		return nil, err
	}

	rv := reflect.ValueOf(dest)
	values := make([]driver.Value, 0, len(columns))
	for _, col := range columns {
		fv, _ := s.FieldsByDBName[col].ValueOf(context.Background(), rv)
		values = append(values, fv)
	}
	return values, nil
}

func GetColumns(dest any) ([]string, error) {
	s, err := schema.Parse(dest, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		return nil, err
	}

	columns := make([]string, 0, len(s.Fields))
	for _, v := range s.Fields {
		if len(v.DBName) != 0 {
			columns = append(columns, v.DBName)
		}
	}
	return columns, nil
}
