package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/maticnetwork/heimdall/types"
)

//
// Send
//

// MsgSend - high level transaction of the coin module
type MsgSend struct {
	FromAddress types.HeimdallAddress `json:"from_address"`
	ToAddress   types.HeimdallAddress `json:"to_address"`
	Amount      types.Coins           `json:"amount"`
}

var _ sdk.Msg = MsgSend{}

// NewMsgSend - construct arbitrary multi-in, multi-out send msg.
func NewMsgSend(fromAddr, toAddr types.HeimdallAddress, amount types.Coins) MsgSend {
	return MsgSend{FromAddress: fromAddr, ToAddress: toAddr, Amount: amount}
}

// Route Implements Msg.
func (msg MsgSend) Route() string { return RouterKey }

// Type Implements Msg.
func (msg MsgSend) Type() string { return "send" }

// ValidateBasic Implements Msg.
func (msg MsgSend) ValidateBasic() sdk.Error {
	if msg.FromAddress.Empty() {
		return sdk.ErrInvalidAddress("missing sender address")
	}
	if msg.ToAddress.Empty() {
		return sdk.ErrInvalidAddress("missing recipient address")
	}
	if !msg.Amount.IsValid() {
		return sdk.ErrInvalidCoins("send amount is invalid: " + msg.Amount.String())
	}
	if !msg.Amount.IsAllPositive() {
		return sdk.ErrInsufficientCoins("send amount must be positive")
	}
	return nil
}

// GetSignBytes Implements Msg.
func (msg MsgSend) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgSend) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{types.HeimdallAddressToAccAddress(msg.FromAddress)}
}

//
// Multi send
//

// MsgMultiSend - high level transaction of the coin module
type MsgMultiSend struct {
	Inputs  []Input  `json:"inputs"`
	Outputs []Output `json:"outputs"`
}

var _ sdk.Msg = MsgMultiSend{}

// NewMsgMultiSend - construct arbitrary multi-in, multi-out send msg.
func NewMsgMultiSend(in []Input, out []Output) MsgMultiSend {
	return MsgMultiSend{Inputs: in, Outputs: out}
}

// Route Implements Msg
func (msg MsgMultiSend) Route() string { return RouterKey }

// Type Implements Msg
func (msg MsgMultiSend) Type() string { return "multisend" }

// ValidateBasic Implements Msg.
func (msg MsgMultiSend) ValidateBasic() sdk.Error {
	// this just makes sure all the inputs and outputs are properly formatted,
	// not that they actually have the money inside
	if len(msg.Inputs) == 0 {
		return ErrNoInputs(DefaultCodespace).TraceSDK("")
	}
	if len(msg.Outputs) == 0 {
		return ErrNoOutputs(DefaultCodespace).TraceSDK("")
	}

	return ValidateInputsOutputs(msg.Inputs, msg.Outputs)
}

// GetSignBytes Implements Msg.
func (msg MsgMultiSend) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// GetSigners Implements Msg.
func (msg MsgMultiSend) GetSigners() []sdk.AccAddress {
	addrs := make([]sdk.AccAddress, len(msg.Inputs))
	for i, in := range msg.Inputs {
		addrs[i] = types.HeimdallAddressToAccAddress(in.Address)
	}
	return addrs
}

// Input models transaction input
type Input struct {
	Address types.HeimdallAddress `json:"address"`
	Coins   types.Coins           `json:"coins"`
}

// ValidateBasic - validate transaction input
func (in Input) ValidateBasic() sdk.Error {
	if len(in.Address) == 0 {
		return sdk.ErrInvalidAddress(in.Address.String())
	}
	if !in.Coins.IsValid() {
		return sdk.ErrInvalidCoins(in.Coins.String())
	}
	if !in.Coins.IsAllPositive() {
		return sdk.ErrInvalidCoins(in.Coins.String())
	}
	return nil
}

// NewInput - create a transaction input, used with MsgMultiSend
func NewInput(addr types.HeimdallAddress, coins types.Coins) Input {
	return Input{
		Address: addr,
		Coins:   coins,
	}
}

// Output models transaction outputs
type Output struct {
	Address types.HeimdallAddress `json:"address"`
	Coins   types.Coins           `json:"coins"`
}

// ValidateBasic - validate transaction output
func (out Output) ValidateBasic() sdk.Error {
	if len(out.Address) == 0 {
		return sdk.ErrInvalidAddress(out.Address.String())
	}
	if !out.Coins.IsValid() {
		return sdk.ErrInvalidCoins(out.Coins.String())
	}
	if !out.Coins.IsAllPositive() {
		return sdk.ErrInvalidCoins(out.Coins.String())
	}
	return nil
}

// NewOutput - create a transaction output, used with MsgMultiSend
func NewOutput(addr types.HeimdallAddress, coins types.Coins) Output {
	return Output{
		Address: addr,
		Coins:   coins,
	}
}

// ValidateInputsOutputs validates that each respective input and output is
// valid and that the sum of inputs is equal to the sum of outputs.
func ValidateInputsOutputs(inputs []Input, outputs []Output) sdk.Error {
	var totalIn, totalOut types.Coins

	for _, in := range inputs {
		if err := in.ValidateBasic(); err != nil {
			return err.TraceSDK("")
		}
		totalIn = totalIn.Add(in.Coins)
	}

	for _, out := range outputs {
		if err := out.ValidateBasic(); err != nil {
			return err.TraceSDK("")
		}
		totalOut = totalOut.Add(out.Coins)
	}

	// make sure inputs and outputs match
	if !totalIn.IsEqual(totalOut) {
		return ErrInputOutputMismatch(DefaultCodespace)
	}

	return nil
}
