package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) CreateValidatorSetPreference(ctx sdk.Context, valOperatorAddress string, weight sdk.Dec) (err error) {
	// check if the validator address provided is correct format
	// check to see if the pubkey or sender has been registered before
	if _, found := k.GetValidator(ctx, valAddr); found {
		return fmt.Errorf("validator address doesnot exist")
	}

	// convert weight to float point and get the raw percentage
	// make sure the weights sum up to 1

}
