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

	// TODO: might want to check if a user already have a validator-set created


	total_weight := sdk.NewDec(0)
	for _, validator := range msg.Preferences { 
		// validation checks making sure the weights add up to 1 and also the validator given is correct
		vals, err := sdk.AccAddressFromBech32(validator)
		if err != nil {
			return fmt.Errorf("validator not formatted")
		}

		if _, found := k.GetValidator(ctx, vals) found {
			return fmt.Errorf("validator address doesnot exist")
		}

		total_weight = total_weight.Add(validator.Weight)
	}

	if total_weight != sdk.NewDec(1) {
		return fmt.Errorf("The weights allocated to the validators do not add up")
	}

	server.Keeper.SetValidatorSetPreferences(ctx, msg.Preferences)
	
	return &types.MsgValidatorSetPreferenceResponse{
		success: true
	}, nil
}
