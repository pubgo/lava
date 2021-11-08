package handler

import (
	"context"
	"fmt"
	"github.com/pubgo/lava/entry/ginEntry"
	"github.com/pubgo/lava/logger"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/igm/sockjs-go/v3/sockjs"
	"github.com/mattheath/kala/bigflake"
	"github.com/mattheath/kala/snowflake"
	"github.com/pubgo/xerror"
	"github.com/teris-io/shortid"

	"github.com/pubgo/lava/errors"
	"github.com/pubgo/lava/internal/example/services/protopb/proto/gid"
	"github.com/pubgo/lava/plugins/metric"
	"github.com/pubgo/lava/plugins/scheduler"
)

var _ gid.IdServer = (*Id)(nil)
var _ ginEntry.Router = (*Id)(nil)

type Id struct {
	Snowflake *snowflake.Snowflake
	Bigflake  *bigflake.Bigflake
	Cron      *scheduler.Scheduler `dix:""`
	Metric    *metric.Resource     `dix:""`
}

func (id *Id) Router(r gin.IRouter) {
	r.GET("/hello1", func(c *gin.Context) {
		c.JSON(200, gin.H{"msg": "ok"})
	})

	r.StaticFS("/web", http.Dir("./web"))
	r.GET("/echo/*action", func(c *gin.Context) {
		sockjs.NewHandler("/echo", sockjs.DefaultOptions, func(session sockjs.Session) {
			for i := 0; i < 100; i++ {
				xerror.Panic(session.Send(time.Now().String()))
			}

			for {
				var ss, err = session.Recv()
				if err == sockjs.ErrSessionNotOpen {
					break
				}
				xerror.Panic(err)
				xerror.Panic(session.Send(ss + "=>" + time.Now().String()))
			}
		}).ServeHTTP(c.Writer, c.Request)
	})
}

func (id *Id) Init() {
	//id.Cron.Every("test gid", time.Second*2, func(name string) {
	//	//id.Metric.Tagged(metric.Tags{"name": name, "time": time.Now().Format("15:04")}).Counter(name).Inc(1)
	//	//id.Metric.Tagged(metric.Tags{"name": name, "time": time.Now().Format("15:04")}).Gauge(name).Update(1)
	//	//"time": time.Now().Format("15:04:05")
	//	id.Metric.Tagged(metric.Tags{"module": "scheduler"}).Gauge(name).Update(1)
	//	fmt.Println("test cron every")
	//})
	fmt.Println("init ok")
}

func NewId() *Id {
	id := rand.Intn(100)

	sf, err := snowflake.New(uint32(id))
	if err != nil {
		panic(err.Error())
	}
	bg, err := bigflake.New(uint64(id))
	if err != nil {
		panic(err.Error())
	}

	return &Id{
		Snowflake: sf,
		Bigflake:  bg,
	}
}

func (id *Id) Generate(ctx context.Context, req *gid.GenerateRequest) (*gid.GenerateResponse, error) {
	var rsp = new(gid.GenerateResponse)
	var log = logger.GetLog(ctx)

	if len(req.Type) == 0 {
		req.Type = "uuid"
	}

	switch req.Type {
	case "uuid":
		rsp.Type = "uuid"
		rsp.Id = uuid.New().String()
	case "snowflake":
		id, err := id.Snowflake.Mint()
		if err != nil {
			log.Sugar().Errorf("Failed to generate snowflake id: %v", err)
			return nil, errors.InternalServerError("id.generate", "failed to mint snowflake id")
		}
		rsp.Type = "snowflake"
		rsp.Id = fmt.Sprintf("%v", id)
	case "bigflake":
		id, err := id.Bigflake.Mint()
		if err != nil {
			log.Sugar().Errorf("Failed to generate bigflake id: %v", err)
			return nil, errors.InternalServerError("id.generate", "failed to mint bigflake id")
		}
		rsp.Type = "bigflake"
		rsp.Id = fmt.Sprintf("%v", id)
	case "shortid":
		id, err := shortid.Generate()
		if err != nil {
			log.Sugar().Errorf("Failed to generate shortid id: %v", err)
			return nil, errors.InternalServerError("id.generate", "failed to generate short id")
		}
		rsp.Type = "shortid"
		rsp.Id = id
	default:
		return nil, errors.BadRequest("id.generate", "unsupported id type")
	}

	return rsp, nil
}

func (id *Id) Types(ctx context.Context, req *gid.TypesRequest) (*gid.TypesResponse, error) {
	var rsp = new(gid.TypesResponse)
	rsp.Types = []string{
		"uuid",
		"shortid",
		"snowflake",
		"bigflake",
	}
	return rsp, nil
}
