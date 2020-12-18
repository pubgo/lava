package task

import (
	"github.com/pubgo/golug/example/tickrun/server/db"
	"github.com/pubgo/golug/example/tickrun/server/models"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/xerror"
)

func Create(view *fiber.Ctx) error {
	var task models.Task

	xerror.Panic(view.BodyParser(&task))

	xerror.Panic(models.TaskInsert(db.GetDb(), &task))
	return xerror.Wrap(view.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": task,
	}))
}

func Delete(view *fiber.Ctx) error {
	return nil
}

func Update(view *fiber.Ctx) error {
	var task map[string]interface{}
	xerror.Panic(view.BodyParser(&task))
	xerror.Panic(models.TaskPut(db.GetDb(), view.Params("id"), task))
	return xerror.Wrap(view.JSON(fiber.Map{
		"code": http.StatusOK,
	}))
}

func Find(view *fiber.Ctx) error {
	idP := view.Params("id")
	id := xerror.PanicErr(strconv.Atoi(idP)).(int)
	task, err := models.TaskGet(db.GetDb(), id)
	xerror.Panic(err)
	return view.JSON(fiber.Map{
		"code": http.StatusOK,
		"data": task,
	})
}

func List(view *fiber.Ctx) error {
	// 随机获取
	random := view.Query("random")
	if random != "" {
		rd, _ := strconv.Atoi(random)
		if rd == 0 {
			rd = 10
		}

		tasks, err := models.TaskRandom(db.GetDb(), rd)
		xerror.Panic(err)
		return view.JSON(fiber.Map{
			"total": len(tasks),
			"data":  tasks,
		})
	}

	idP := view.Query("id")
	id, _ := strconv.Atoi(idP)
	statusP := view.Query("status")
	status, _ := strconv.Atoi(statusP)
	pageP := view.Query("page")
	page, _ := strconv.Atoi(pageP)
	perPageP := view.Query("per_page")
	perPage, _ := strconv.Atoi(perPageP)

	_ = id
	page, perPage = models.Pagination(page, perPage)
	tasks, total, err := models.TaskRange(db.GetDb(), status, page, perPage)
	xerror.Panic(err)

	next, total := models.NextPage(int64(page), int64(perPage), total)
	return view.JSON(fiber.Map{
		"total": total,
		"data":  tasks,
		"next":  next,
	})
}
