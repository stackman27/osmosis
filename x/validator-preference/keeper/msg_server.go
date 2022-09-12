package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v12/x/validator-preference/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (server msgServer) CreateValidatorSetPreference(goCtx context.Context, msg *types.MsgValidatorSetPreference) (*types.MsgValidatorSetPreferenceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	total_weight := sdk.Int(0)
	for _, validator := range msg.Preferences { 
		// For multiple validators we just run a loop here based on how many validators the user chooses
		err := server.Keeper.CreateValidatorSetPreference(ctx, validator.ValOperAddress, validator.Weight)
		if err != nil {
			return nil, err
		}
	}

	return &types.MsgValidatorSetPreferenceResponse{
		success: true
	}, nil
}
