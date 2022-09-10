package gidrpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pubgo/lava/example/gen/proto/gidpb"
	"github.com/pubgo/lava/logging"
	"github.com/teris-io/shortid"
)

func (id *Id) Generate(ctx context.Context, req *gidpb.GenerateRequest) (*gidpb.GenerateResponse, error) {
	var rsp = new(gidpb.GenerateResponse)
	var log = logging.GetLog(ctx)

	if len(req.Type) == 0 {
		req.Type = "uuid"
	}

	switch req.Type {
	case "uuid":
		rsp.Type = "uuid"
		rsp.Id = uuid.New().String()
	case "snowflake":
		id, err := id.snowflake.Mint()
		if err != nil {
			log.Sugar().Errorf("Failed to generate snowflake id: %v", err)
			return nil, err1.Msg("id.generate", "failed to mint snowflake id").StatusBadRequest()
		}
		rsp.Type = "snowflake"
		rsp.Id = fmt.Sprintf("%v", id)
	case "bigflake":
		id, err := id.bigflake.Mint()
		if err != nil {
			log.Sugar().Errorf("Failed to generate bigflake id: %v", err)
			return nil, err1.Msg("id.generate", "failed to mint bigflake id").StatusBadRequest()
		}
		rsp.Type = "bigflake"
		rsp.Id = fmt.Sprintf("%v", id)
	case "shortid":
		id, err := shortid.Generate()
		if err != nil {
			log.Sugar().Errorf("Failed to generate shortid id: %v", err)
			return nil, err1.Msg("id.generate", "failed to generate short id").StatusBadRequest()
		}
		rsp.Type = "shortid"
		rsp.Id = id
	default:
		return nil, err1.Msg("id.generate", "unsupported id type").StatusBadRequest()
	}

	return rsp, nil
}

func (id *Id) Types(ctx context.Context, req *gidpb.TypesRequest) (*gidpb.TypesResponse, error) {
	var rsp = new(gidpb.TypesResponse)
	rsp.Types = []string{
		"uuid",
		"shortid",
		"snowflake",
		"bigflake",
	}
	return rsp, nil
}
