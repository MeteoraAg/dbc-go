package poolfees

import (
	"dbcGoSDK/constants"
	"dbcGoSDK/generated/dbc"
	"math/big"
)

func IsDynamicFeeEnabled(dynamicFee dbc.DynamicFeeConfig) bool {
	return dynamicFee.Initialized != 0
}

// GetVariableFeeNumerator gets variable fee numerator from dynamic fee.
func GetVariableFeeNumerator(
	dynamicFee dbc.DynamicFeeConfig,
	volatilityTracker dbc.VolatilityTracker,
) *big.Int {
	if !IsDynamicFeeEnabled(dynamicFee) {
		return big.NewInt(0)
	}

	// 1. Computing the squared price movement (volatility_accumulator * bin_step)^2
	volatilityTimesBinStep := new(big.Int).Mul(
		volatilityTracker.VolatilityReference.BigInt(),
		new(big.Int).SetUint64(uint64(dynamicFee.BinStep)),
	)
	squareVfaBin := new(big.Int).Mul(volatilityTimesBinStep, volatilityTimesBinStep)

	// 2. Multiplying by the fee control factor
	vFee := new(big.Int).Mul(
		squareVfaBin,
		new(big.Int).SetUint64(uint64(dynamicFee.VariableFeeControl)),
	)

	// 3. Scaling down the result to fit within u64 range (dividing by 1e11 and rounding up)
	return new(big.Int).Quo(
		new(big.Int).Add(vFee, constants.DynamicFeeRoundingOffset),
		constants.DynamicFeeScalingFactor,
	)
}
