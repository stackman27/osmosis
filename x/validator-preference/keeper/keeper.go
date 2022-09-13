package keeper

import (
	"fmt"
	"log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/osmosis-labs/osmosis/v12/x/validator-preference/types"
)

type Keeper struct {
	storeKey      sdk.StoreKey
	paramSpace    paramtypes.Subspace
	cdc           codec.BinaryCodec
	stakingKeeper types.StakingInterface
}

func NewKeeper(storeKey sdk.StoreKey, paramSpace types.Subspace, stakingKeeper types.StakingInterface) keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:      storeKey,
		paramSpace:    paramSpace,
		stakingKeeper: stakingKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) SetValidatorSetPreferences(ctx sdk.Context, validators types.MsgValidatorSetPreference) {
	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixValidatorSet)
	bz, err := proto.Marshal(&validators)
	if err != nil {
		panic err) 
	}
	store.Set(types.KeyPrefixValidatorSet, bz)
}

func (k Keeper) GetValidatorSetPreference(ctx sdk.Context) (*types.MsgvalidatorSetPreference) []types.ValidatorPreference{
	validatorSet := []types.ValidatorPreference{}

	store := ctx.KVStore(k.storeKey)
	prefixStore := prefix.NewStore(store, types.KeyPrefixValidatorSet)
	bz := store.Get(types.KeyPrefixValidatorSet) 
	err := proto.Unmarshal(bz, validatorSet)

	return validatorSet
}