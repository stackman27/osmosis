package types

const (
	// ModuleName defines the module name
	ModuleName = "validator-set-preference"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// KeyPrefixSuperfluidAsset defines prefix key for validator set.
	KeyPrefixValidatorSet = []byte{0x01}
)
