package handler

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/mattheath/kala/bigflake"
	"github.com/mattheath/kala/snowflake"
	"github.com/teris-io/shortid"

	"github.com/pubgo/lug/errors"
	"github.com/pubgo/lug/example/gid/proto/gid"
	"github.com/pubgo/lug/logger"
)

var _ gid.IdServer = (*Id)(nil)

type Id struct {
	Snowflake *snowflake.Snowflake
	Bigflake  *bigflake.Bigflake
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
