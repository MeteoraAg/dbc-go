package types

import "math/big"

type BaseFeeHandler interface {
	Validate(
		collectFeeMode CollectFeeMode,
		activationType ActivationType,
	) bool

	GetBaseFeeNumeratorFromExcludedFeeAmount(
		currentPoint,
		activationPoint *big.Int,
		tradeDirection TradeDirection,
		excludedFeeAmount *big.Int,
	) (*big.Int, error)

	GetBaseFeeNumeratorFromIncludedFeeAmount(
		currentPoint,
		activationPoint *big.Int,
		tradeDirection TradeDirection,
		includedFeeAmount *big.Int,
	) (*big.Int, error)
}
