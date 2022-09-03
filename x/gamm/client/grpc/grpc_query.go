package grpc 

// THIS FILE IS GENERATED CODE, DO NOT EDIT
// SOURCE AT `proto/osmosis/gamm/v1beta1/query.yml`

import (
	context "context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v11/x/gamm/client"
	"github.com/osmosis-labs/osmosis/v11/x/gamm/client/queryproto"
)

type Querier struct {
	Q client.Querier
}

var _ queryproto.QueryServer = Querier{}

func (q Querier) Pool(grpcCtx context.Context,
	req *queryproto.PoolRequest,
) (*queryproto.PoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.Pool(ctx, *req)
}
