package poolfees

import (
	"dbcGoSDK/types"
	"errors"
	"math/big"
)

type FeeRateLimiter struct {
	CliffFeeNumerator  *big.Int
	FeeIncrementBps    uint16
	MaxLimiterDuration *big.Int
	ReferenceAmount    *big.Int
}

func (fr *FeeRateLimiter) Validate(
	collectFeeMode types.CollectFeeMode,
	activationType types.ActivationType,
) bool {
	return true
}

func (fr *FeeRateLimiter) GetBaseFeeNumeratorFromIncludedFeeAmount(
	currentPoint,
	activationPoint *big.Int,
	tradeDirection types.TradeDirection,
	includedFeeAmount *big.Int,
) (*big.Int, error) {
	if IsRateLimiterApplied(
		currentPoint,
		activationPoint,
		tradeDirection,
		fr.MaxLimiterDuration,
		fr.ReferenceAmount,
		new(big.Int).SetUint64(uint64(fr.FeeIncrementBps)),
	) {
		return GetFeeNumeratorFromIncludedAmount(
			fr.CliffFeeNumerator,
			fr.ReferenceAmount,
			new(big.Int).SetUint64(uint64(fr.FeeIncrementBps)),
			includedFeeAmount,
		)
	}

	return fr.CliffFeeNumerator, nil
}

func (fr *FeeRateLimiter) GetBaseFeeNumeratorFromExcludedFeeAmount(
	currentPoint,
	activationPoint *big.Int,
	tradeDirection types.TradeDirection,
	excludedFeeAmount *big.Int,
) (*big.Int, error) {
	if IsRateLimiterApplied(
		currentPoint,
		activationPoint,
		tradeDirection,
		fr.MaxLimiterDuration,
		fr.ReferenceAmount,
		new(big.Int).SetUint64(uint64(fr.FeeIncrementBps)),
	) {
		return GetFeeNumeratorFromExcludedAmount(
			fr.CliffFeeNumerator,
			fr.ReferenceAmount,
			new(big.Int).SetUint64(uint64(fr.FeeIncrementBps)),
			excludedFeeAmount,
		)
	}

	return fr.CliffFeeNumerator, nil
}

type FeeScheduler struct {
	CliffFeeNumerator *big.Int
	NumberOfPeriod    uint16
	PeriodFrequency   *big.Int
	ReductionFactor   *big.Int
	FeeSchedulerMode  types.BaseFeeMode
}

func (fs *FeeScheduler) Validate(
	collectFeeMode types.CollectFeeMode,
	activationType types.ActivationType,
) bool {
	return true
}

func (fs *FeeScheduler) GetBaseFeeNumeratorFromIncludedFeeAmount(
	currentPoint,
	activationPoint *big.Int,
	_ types.TradeDirection, // ignored
	_ *big.Int,
) (*big.Int, error) {
	return GetBaseFeeNumerator(
		fs.CliffFeeNumerator,
		fs.NumberOfPeriod,
		fs.PeriodFrequency,
		fs.ReductionFactor,
		fs.FeeSchedulerMode,
		currentPoint,
		activationPoint,
	)
}

func (fs *FeeScheduler) GetBaseFeeNumeratorFromExcludedFeeAmount(
	currentPoint,
	activationPoint *big.Int,
	_ types.TradeDirection, // ignored
	_ *big.Int,
) (*big.Int, error) {
	return GetBaseFeeNumerator(
		fs.CliffFeeNumerator,
		fs.NumberOfPeriod,
		fs.PeriodFrequency,
		fs.ReductionFactor,
		fs.FeeSchedulerMode,
		currentPoint,
		activationPoint,
	)
}

// GetBaseFeeHandler gets base fee handler based on base fee mode.
func GetBaseFeeHandler(
	cliffFeeNumerator *big.Int,
	firstFactor uint16,
	secondFactor, thirdFactor *big.Int,
	baseFeeMode types.BaseFeeMode,
) (types.BaseFeeHandler, error) {

	switch baseFeeMode {
	case types.BaseFeeModeFeeSchedulerExponential,
		types.BaseFeeModeFeeSchedulerLinear:
		return &FeeScheduler{
			CliffFeeNumerator: cliffFeeNumerator,
			NumberOfPeriod:    firstFactor,
			PeriodFrequency:   thirdFactor,
			ReductionFactor:   thirdFactor,
			FeeSchedulerMode:  baseFeeMode,
		}, nil

	case types.BaseFeeModeRateLimiter:
		return &FeeRateLimiter{
			CliffFeeNumerator:  cliffFeeNumerator,
			FeeIncrementBps:    firstFactor,
			MaxLimiterDuration: secondFactor,
			ReferenceAmount:    thirdFactor,
		}, nil
	}

	return nil, errors.New("invalid baseFeeMode")
}
