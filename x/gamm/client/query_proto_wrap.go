package client

import (
	"context"
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/osmosis-labs/osmosis/v11/x/gamm/client/queryproto"
	gammKeeper "github.com/osmosis-labs/osmosis/v11/x/gamm/keeper"
	"github.com/osmosis-labs/osmosis/v11/x/gamm/pool-models/balancer"

	"github.com/osmosis-labs/osmosis/v11/x/gamm/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// This file should evolve to being code gen'd, off of `proto/gamm/v1beta/query.yml`

var sdkIntMaxValue = sdk.NewInt(0)

// Querier defines a wrapper around the x/gamm keeper providing gRPC method
// handlers.
type Querier struct {
	Keeper gammKeeper.Keeper
}

// Pool checks if a pool exists and their respective poolWeights.
func (q Querier) Pool(ctx context.Context,
	req queryproto.QueryPoolRequest,
) (*queryproto.QueryPoolResponse, error) {

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := q.Keeper.GetPoolAndPoke(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	any, err := codectypes.NewAnyWithValue(pool)
	if err != nil {
		return nil, err
	}

	return &queryproto.QueryPoolResponse{Pool: any}, nil
}

// Pools checks existence of multiple pools and their poolWeights
func (q Querier) Pools(
	ctx context.Context,
	req *queryproto.QueryPoolsRequest,
) (*queryproto.QueryPoolsResponse, error) {

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	store := sdkCtx.KVStore(q.Keeper.StoreKey)
	poolStore := prefix.NewStore(store, types.KeyPrefixPools)

	var anys []*codectypes.Any
	pageRes, err := query.Paginate(poolStore, req.Pagination, func(_, value []byte) error {
		poolI, err := q.Keeper.UnmarshalPool(value)
		if err != nil {
			return err
		}

		// Use GetPoolAndPoke function because it runs PokeWeights
		poolI, err = q.Keeper.GetPoolAndPoke(sdkCtx, poolI.GetId())
		if err != nil {
			return err
		}

		// TODO: pools query should not be balancer specific
		pool, ok := poolI.(*balancer.Pool)
		if !ok {
			return fmt.Errorf("pool (%d) is not basic pool", pool.GetId())
		}

		any, err := codectypes.NewAnyWithValue(pool)
		if err != nil {
			return err
		}

		anys = append(anys, any)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.QueryPoolsResponse{
		Pools:      anys,
		Pagination: pageRes,
	}, nil
}

// NumPools returns total number of pools.
func (q Querier) NumPools(ctx context.Context, _ *queryproto.QueryNumPoolsRequest) (*queryproto.QueryNumPoolsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &queryproto.QueryNumPoolsResponse{
		NumPools: q.Keeper.GetNextPoolId(sdkCtx) - 1,
	}, nil
}

// PoolParams queries a specified pool for its params.
func (q Querier) PoolParams(ctx context.Context, req *queryproto.QueryPoolParamsRequest) (*queryproto.QueryPoolParamsResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := q.Keeper.GetPoolAndPoke(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	switch pool := pool.(type) {
	case *balancer.Pool:
		any, err := codectypes.NewAnyWithValue(&pool.PoolParams)
		if err != nil {
			return nil, err
		}

		return &queryproto.QueryPoolParamsResponse{
			Params: any,
		}, nil

	default:
		errMsg := fmt.Sprintf("unrecognized %s pool type: %T", types.ModuleName, pool)
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnpackAny, errMsg)
	}
}

// TotalPoolLiquidity returns total liquidity in pool.
func (q Querier) TotalPoolLiquidity(ctx context.Context, req *queryproto.QueryTotalPoolLiquidityRequest) (*queryproto.QueryTotalPoolLiquidityResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := q.Keeper.GetPoolAndPoke(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.QueryTotalPoolLiquidityResponse{
		Liquidity: pool.GetTotalPoolLiquidity(sdkCtx),
	}, nil
}

// TotalShares returns total pool shares.
func (q Querier) TotalShares(ctx context.Context, req *queryproto.QueryTotalSharesRequest) (*queryproto.QueryTotalSharesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := q.Keeper.GetPoolAndPoke(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.QueryTotalSharesResponse{
		TotalShares: sdk.NewCoin(
			types.GetPoolShareDenom(req.PoolId),
			pool.GetTotalShares()),
	}, nil
}

// SpotPrice returns target pool asset prices on base and quote assets.
func (q Querier) SpotPrice(ctx context.Context, req *queryproto.QuerySpotPriceRequest) (*queryproto.QuerySpotPriceResponse, error) {
	if req.BaseAssetDenom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid base asset denom")
	}

	if req.QuoteAssetDenom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid quote asset denom")
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	pool, err := q.Keeper.GetPoolAndPoke(sdkCtx, req.PoolId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get pool by ID: %s", err)
	}

	sp, err := pool.SpotPrice(sdkCtx, req.BaseAssetDenom, req.QuoteAssetDenom)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.QuerySpotPriceResponse{
		SpotPrice: sp.String(),
	}, nil
}

// TotalLiquidity returns total liquidity across all pools.
func (q Querier) TotalLiquidity(ctx context.Context, _ *queryproto.QueryTotalLiquidityRequest) (*queryproto.QueryTotalLiquidityResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	return &queryproto.QueryTotalLiquidityResponse{
		Liquidity: q.Keeper.GetTotalLiquidity(sdkCtx),
	}, nil
}

// EstimateSwapExactAmountIn estimates input token amount for a swap.
func (q Querier) EstimateSwapExactAmountIn(ctx context.Context, req *queryproto.QuerySwapExactAmountInRequest) (*queryproto.QuerySwapExactAmountInResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Sender == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	if req.TokenIn == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	if err := types.SwapAmountInRoutes(req.Routes).Validate(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	tokenIn, err := sdk.ParseCoinNormalized(req.TokenIn)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid token: %s", err.Error())
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	tokenOutAmount, err := q.Keeper.MultihopSwapExactAmountIn(sdkCtx, sender, req.Routes, tokenIn, sdk.NewInt(1))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.QuerySwapExactAmountInResponse{
		TokenOutAmount: tokenOutAmount,
	}, nil
}

// EstimateSwapExactAmountOut estimates token output amount for a swap.
func (q Querier) EstimateSwapExactAmountOut(ctx context.Context, req *queryproto.QuerySwapExactAmountOutRequest) (*queryproto.QuerySwapExactAmountOutResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Sender == "" {
		return nil, status.Error(codes.InvalidArgument, "address cannot be empty")
	}

	if req.TokenOut == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid token")
	}

	if err := types.SwapAmountOutRoutes(req.Routes).Validate(); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	sender, err := sdk.AccAddressFromBech32(req.Sender)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid address: %s", err.Error())
	}

	tokenOut, err := sdk.ParseCoinNormalized(req.TokenOut)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid token: %s", err.Error())
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)

	tokenInAmount, err := q.Keeper.MultihopSwapExactAmountOut(sdkCtx, sender, req.Routes, sdkIntMaxValue, tokenOut)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &queryproto.QuerySwapExactAmountOutResponse{
		TokenInAmount: tokenInAmount,
	}, nil
}
