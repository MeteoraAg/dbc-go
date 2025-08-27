package poolfees

import (
	"dbcGoSDK/constants"
	"dbcGoSDK/types"
	"errors"
	"fmt"
	"math"
	"math/big"
)

// GetMaxBaseFeeNumerator gets max base fee numerator.
func GetMaxBaseFeeNumerator(cliffFeeNumerator *big.Int) *big.Int {
	return cliffFeeNumerator
}

// GetMinBaseFeeNumerator gets min base fee numerator.
func GetMinBaseFeeNumerator(
	cliffFeeNumerator *big.Int,
	numberOfPeriod uint16,
	periodFrequency,
	reductionFactor *big.Int,
	feeSchedulerMode types.BaseFeeMode,
) (*big.Int, error) {
	return GetBaseFeeNumeratorByPeriod(
		cliffFeeNumerator,
		numberOfPeriod,
		periodFrequency,
		reductionFactor,
		feeSchedulerMode,
	)
}

// GetBaseFeeNumerator gets base fee numerator.
func GetBaseFeeNumerator(
	cliffFeeNumerator *big.Int,
	numberOfPeriod uint16,
	periodFrequency,
	reductionFactor *big.Int,
	feeSchedulerMode types.BaseFeeMode,
	currentPoint,
	activationPoint *big.Int,
) (*big.Int, error) {
	if periodFrequency.Sign() == 0 {
		return cliffFeeNumerator, nil
	}
	period := new(big.Int).Quo(
		new(big.Int).Sub(currentPoint, activationPoint),
		periodFrequency,
	)

	return GetBaseFeeNumeratorByPeriod(
		cliffFeeNumerator,
		numberOfPeriod,
		period,
		reductionFactor,
		feeSchedulerMode,
	)
}

// GetBaseFeeNumeratorByPeriod gets base fee numerator by period.
func GetBaseFeeNumeratorByPeriod(
	cliffFeeNumerator *big.Int,
	numberOfPeriod uint16,
	period,
	reductionFactor *big.Int,
	feeSchedulerMode types.BaseFeeMode,
) (*big.Int, error) {

	periodValue := new(big.Int).SetUint64(uint64(numberOfPeriod))
	if period.Cmp(periodValue) <= 0 {
		if !period.IsUint64() {
			return nil, fmt.Errorf("cannot fit period(%s) as uint64", period)
		}
		periodValue = period
	}

	periodNumber := periodValue.Uint64()
	if periodNumber > math.MaxUint16 {
		return nil, fmt.Errorf("periodNumber(%d) cannot be greater than %d", periodNumber, math.MaxUint16)
	}

	switch feeSchedulerMode {
	case types.BaseFeeModeFeeSchedulerLinear:
		return GetFeeNumeratorOnLinearFeeScheduler(
			cliffFeeNumerator,
			reductionFactor,
			periodNumber,
		)

	case types.BaseFeeModeFeeSchedulerExponential:
		return GetFeeNumeratorOnExponentialFeeScheduler(
			cliffFeeNumerator,
			reductionFactor,
			periodNumber,
		)
	}

	return nil, errors.New("invalid feeSchedulerMode option")
}

// GetFeeNumeratorOnLinearFeeScheduler gets fee in period for linear fee scheduler.
func GetFeeNumeratorOnLinearFeeScheduler(
	cliffFeeNumerator, reductionFactor *big.Int,
	period uint64,
) (*big.Int, error) {
	reduction := new(big.Int).Mul(
		reductionFactor, new(big.Int).SetUint64(period))

	if reduction.Cmp(cliffFeeNumerator) > 0 {
		return big.NewInt(0), nil
	}

	v := new(big.Int).Sub(cliffFeeNumerator, reduction)
	if v.Sign() < 0 {
		return nil, fmt.Errorf("safeMath requires value not negative: value is %s", v.String())
	}

	return v, nil
}

func GetFeeNumeratorOnExponentialFeeScheduler(
	cliffFeeNumerator, reductionFactor *big.Int,
	period uint64,
) (*big.Int, error) {
	if period == 0 {
		return cliffFeeNumerator, nil
	}

	// Match Rust implementation exactly
	// Make reduction_factor into Q64x64, and divided by BASIS_POINT_MAX
	basisPointMax := big.NewInt(constants.BasisPointMax)

	bps := new(big.Int).Quo(
		new(big.Int).Lsh(reductionFactor, 64),
		basisPointMax,
	)

	// base = ONE_Q64 - bps (equivalent to 1 - reduction_factor/10_000 in Q64.64)
	base := new(big.Int).Sub(constants.OneQ64, bps)
	if base.Sign() < 0 {
		return nil, fmt.Errorf("safeMath requires value not negative: value is %s", base.String())
	}

	result := PowQ64(base, new(big.Int).SetUint64(period), true)

	// final fee: cliffFeeNumerator * result >> 64
	return new(big.Int).Quo(
		new(big.Int).Mul(cliffFeeNumerator, result),
		constants.OneQ64,
	), nil
}
