package models

import (
	"strconv"
	"strings"

	"github.com/pubgo/xerror"
	"xorm.io/xorm"
)

type TaskStatus uint8

func (t TaskStatus) String() string {
	switch t {
	case 0:
		return "Created"
	case 1:
		return "Pending"
	case 2:
		return "Success"
	case 3:
		return "Failed"
	default:
		return "Created"
	}
}

const (
	TaskStatusCreated TaskStatus = iota
	TaskStatusPending
	TaskStatusSuccess
	TaskStatusFailed
)

type Task struct {
	Id        int64             `json:"id"`
	CreatedAt JsonTime          `json:"created_at" xorm:"created"`
	UpdatedAt JsonTime          `json:"updated_at" xorm:"updated"`
	TaskID    string            `json:"task_id" xorm:"varchar(100) notnull unique 'task_id'"`
	Service   string            `json:"service"`
	Attempts  uint16            `json:"attempts"`
	Priority  uint8             `json:"priority"`
	Timestamp int64             `json:"timestamp"`
	Status    TaskStatus        `json:"status" xorm:"int notnull index 'status'"`
	Method    string            `json:"method"`
	Body      []byte            `json:"body"`
	Header    map[string]string `json:"header"`
}

func TaskRandom(db *xorm.Engine, n int) (tasks []Task, err error) {
	defer xerror.RespErr(&err)
	xerror.Panic(db.SQL(
		"select * from task where id>=(select floor(rand() * (select max(id) from task))) order by id limit ?", n,
	).Find(&tasks))
	return
}

func TaskInsert(db *xorm.Engine, task *Task) error {
	_, err := db.InsertOne(task)
	return xerror.Wrap(err)
}

func TaskGet(db *xorm.Engine, id int) (Task, error) {
	var sess = db.Table(&Task{})

	var task Task
	_, err := sess.Where("id=?", id).Get(&task)
	return task, xerror.Wrap(err)
}

func TaskPut(db *xorm.Engine, id string, task map[string]interface{}) error {
	var sess = db.Table(&Task{})
	if strings.Contains(id, "-") {
		sess = sess.Where("task_id=?", id)
	} else {
		sess = sess.ID(xerror.PanicErr(strconv.Atoi(id)).(int))
	}

	_, err := sess.Update(task)
	return xerror.Wrap(err)
}

func TaskRange(db *xorm.Engine, status, page, perPage int) ([]Task, int64, error) {
	var start int
	var tasks []Task

	var sess = db.Table(&Task{})
	sess = sess.Where("status=?", status)
	_, perPage, start = pagination(page, perPage)
	sess = sess.Limit(perPage, start)

	count := xerror.PanicErr(sess.Where("status=?", status).Count()).(int64)
	return tasks, count, xerror.Wrap(sess.Find(&tasks))
}
