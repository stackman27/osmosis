package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// StakingKeeper expected staking keeper.
type StakingInterface interface {
	BondDenom(ctx sdk.Context) string
	GetAllValidators(ctx sdk.Context) (validators []stakingtypes.Validator)
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (validator stakingtypes.Validator, found bool)
	ValidateUnbondAmount(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, amt sdk.Int) (shares sdk.Dec, err error)
	Delegate(ctx sdk.Context, delAddr sdk.AccAddress, bondAmt sdk.Int, tokenSrc stakingtypes.BondStatus, validator stakingtypes.Validator, subtractAccount bool) (newShares sdk.Dec, err error)
	InstantUndelegate(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress, sharesAmount sdk.Dec) (sdk.Coins, error)
	GetDelegation(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) (delegation stakingtypes.Delegation, found bool)
	UnbondingTime(ctx sdk.Context) time.Duration
	GetParams(ctx sdk.Context) stakingtypes.Params

	IterateBondedValidatorsByPower(ctx sdk.Context, fn func(int64, stakingtypes.ValidatorI) bool)
	TotalBondedTokens(ctx sdk.Context) sdk.Int
	IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress, fn func(int64, stakingtypes.DelegationI) bool)
}
