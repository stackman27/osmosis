package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// constants
const (
	TypeMsgCreateValSetPreference = "create_validator_set_preference"
)

var _ sdk.Msg = &MsgValidatorSetPreference{}

// NewMsgCreateValidatorSetPreference creates a msg to create a new denom
func NewMsgCreateValidatorSetPreference(valOperatorAddress string, weight sdk.Dec) *MsgValidatorSetPreference {
	return &MsgValidatorSetPreference{
		valOperatorAddress: valOperatorAddress,
		weight:             weight,
	}
}

func (m MsgValidatorSetPreference) Route() string { return RouterKey }
func (m MsgValidatorSetPreference) Type() string  { return TypeMsgCreateValSetPreference }
func (m MsgValidatorSetPreference) ValidateBasic() error {
	for _, validator := range m.Preferences {
		_, err := sdk.AccAddressFromBech32(validator.ValOperAddress)
		if err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid validator address (%s)", err)
		}
	}

	return nil
}

func (m MsgValidatorSetPreference) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgValidatorSetPreference) GetSigners() []sdk.AccAddress {
	var validators []sdk.AccAddress
	for _, validator := range m.Preferences {
		valAddr, _ := sdk.AccAddressFromBech32(validator.ValOperAddress)
		validators = append(validators, valAddr)
	}
	
	return validators
}
