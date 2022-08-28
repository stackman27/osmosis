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

func (q Querier) NumPools(grpcCtx context.Context,
	req *queryproto.NumPoolsRequest,
) (*queryproto.NumPoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.NumPools(ctx, *req)
}
func (q Querier) TotalShares(grpcCtx context.Context,
	req *queryproto.TotalSharesRequest,
) (*queryproto.TotalSharesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.TotalShares(ctx, *req)
}
func (q Querier) SpotPrice(grpcCtx context.Context,
	req *queryproto.SpotPriceRequest,
) (*queryproto.SpotPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.SpotPrice(ctx, *req)
}
func (q Querier) TotalLiquidity(grpcCtx context.Context,
	req *queryproto.TotalLiquidityRequest,
) (*queryproto.TotalLiquidityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.TotalLiquidity(ctx, *req)
}
func (q Querier) EstimateSwapExactAmountIn(grpcCtx context.Context,
	req *queryproto.EstimateSwapExactAmountInRequest,
) (*queryproto.EstimateSwapExactAmountInResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.EstimateSwapExactAmountIn(ctx, *req)
}
func (q Querier) EstimateSwapExactAmountOut(grpcCtx context.Context,
	req *queryproto.EstimateSwapExactAmountOutRequest,
) (*queryproto.EstimateSwapExactAmountOutResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.EstimateSwapExactAmountOut(ctx, *req)
}
func (q Querier) Pool(grpcCtx context.Context,
	req *queryproto.PoolRequest,
) (*queryproto.PoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.Pool(ctx, *req)
}
func (q Querier) Pools(grpcCtx context.Context,
	req *queryproto.PoolsRequest,
) (*queryproto.PoolsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.Pools(ctx, *req)
}
func (q Querier) PoolParams(grpcCtx context.Context,
	req *queryproto.PoolParamsRequest,
) (*queryproto.PoolParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.PoolParams(ctx, *req)
}
func (q Querier) TotalPoolLiquidity(grpcCtx context.Context,
	req *queryproto.TotalPoolLiquidityRequest,
) (*queryproto.TotalPoolLiquidityResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(grpcCtx)
	return q.Q.TotalPoolLiquidity(ctx, *req)
}
