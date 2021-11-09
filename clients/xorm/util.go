package xorm

import (
	"time"

	"github.com/pubgo/xerror"
	"xorm.io/xorm"
)

type JsonTime struct {
	time.Time
}

func (j JsonTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + j.Time.Format("2006-01-02 15:04:05") + `"`), nil
}

func NextPage(page, perPage, total int64) (int64, int64) {
	if total > perPage {
		return page + 1, total/perPage + 1
	}
	return page, total/perPage + 1
}

// 从1开始
func pagination(page, perPage int) (int, int, int) {
	if perPage < 1 {
		perPage = 20
	}

	if perPage > 100 {
		perPage = 20
	}

	if page < 2 {
		page = 1
	}

	return page, perPage, (page - 1) * perPage
}

func Random(db *xorm.Session, data interface{}, n int, table string) (err error) {
	defer xerror.RespErr(&err)

	sql := "select * from ? where id>=(select floor(rand() * (select max(id) from ?))) order by id limit ?"
	return xerror.Wrap(db.SQL(sql, table, table, n).Find(data))
}

func Range(db *xorm.Session, data interface{}, page, perPage int, where string, a ...interface{}) (_ int64, err error) {
	defer xerror.RespErr(&err)

	var start int

	ses := db.Where(where, a...)
	_, perPage, start = pagination(page, perPage)
	count := xerror.PanicErr(ses.Count()).(int64)
	return count, xerror.Wrap(ses.Limit(perPage, start).Find(data))
}

func GetOne(db *xorm.Session, task interface{}) {
}

func InsertOne(db *xorm.Session, task interface{}) (err error) {
	_, err = db.InsertOne(task)
	return xerror.Wrap(err)
}

func Insert(db *xorm.Session, beans ...interface{}) (err error) {
	_, err = db.Insert(beans...)
	return xerror.Wrap(err)
}

func UpdateById(db *xorm.Session, task map[string]interface{}, query interface{}, args ...interface{}) error {
	switch len(args) {
	case 0:
		db = db.ID(query)
	default:
		db = db.Where(query, args...)
	}

	_, err := db.Update(task)
	return xerror.Wrap(err)
}
