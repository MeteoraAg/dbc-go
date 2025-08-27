package poolfees

import (
	"dbcGoSDK/constants"
	"dbcGoSDK/types"
	"errors"
	"fmt"
	"math/big"
)

// IsRateLimiterApplied checks if rate limiter is applied based on current conditions.
func IsRateLimiterApplied(
	currentPoint,
	activationPoint *big.Int,
	tradeDirection types.TradeDirection,
	maxLimiterDuration,
	referenceAmount,
	feeIncrementBps *big.Int,
) bool {
	if referenceAmount == nil || activationPoint == nil ||
		maxLimiterDuration == nil || feeIncrementBps == nil || currentPoint == nil {
		return false
	}

	// Only handle for the case quote to base and collect fee mode in quote token
	if tradeDirection == types.TradeDirectionBaseToQuote {
		return false
	}

	lastEffectiveRateLimiterPoint := new(big.Int).Add(activationPoint, maxLimiterDuration)
	return currentPoint.Cmp(lastEffectiveRateLimiterPoint) <= 0
}

// GetMaxIndex calculates the max index for rate limiter.
func GetMaxIndex(cliffFeeNumerator, feeIncrementBps *big.Int) (*big.Int, error) {
	deltaNumerator := new(big.Int).Sub(
		new(big.Int).SetUint64(constants.MaxFeeNumerator), cliffFeeNumerator)

	if deltaNumerator.Sign() <= 0 {
		return nil, fmt.Errorf("GetMaxIndex: cliffFeeNumerator(%s) exceeds MaxFeeNumerator(%d)", cliffFeeNumerator, constants.MaxFeeNumerator)
	}

	feeIncrementNumerator := toNumerators(
		feeIncrementBps,
		new(big.Int).SetInt64(constants.FeeDenominator),
	)

	if feeIncrementNumerator.Sign() == 0 {
		return nil, errors.New("feeIncrementNumerator cannot be zero")
	}

	return deltaNumerator.Div(deltaNumerator, feeIncrementNumerator), nil
}

// GetMaxOutAmountWithMinBaseFee gets max out amount with min base fee.
func GetMaxOutAmountWithMinBaseFee(
	cliffFeeNumerator,
	referenceAmount,
	feeIncrementBps *big.Int,
) (*big.Int, error) {
	return GetRateLimiterExcludedFeeAmount(
		cliffFeeNumerator,
		referenceAmount,
		feeIncrementBps,
		referenceAmount,
	)
}

// GetCheckedAmounts gets checked amounts for rate limiter.
func GetCheckedAmounts(
	cliffFeeNumerator,
	referenceAmount,
	feeIncrementBps *big.Int,
) (struct {
	CheckedExcludedFeeAmount, CheckedIncludedFeeAmount *big.Int
	IsOverflow                                         bool
}, error) {

	maxIndex, err := GetMaxIndex(cliffFeeNumerator, feeIncrementBps)
	if err != nil {
		return struct {
			CheckedExcludedFeeAmount *big.Int
			CheckedIncludedFeeAmount *big.Int
			IsOverflow               bool
		}{}, err
	}

	maxIndexInputAmount := new(big.Int).Mul(
		new(big.Int).Add(maxIndex, big.NewInt(1)),
		referenceAmount,
	)

	if maxIndexInputAmount.Cmp(constants.U64MaxBigInt) <= 0 {
		checkedIncludedFeeAmount := maxIndexInputAmount
		checkedOutputAmount, err := GetRateLimiterExcludedFeeAmount(
			cliffFeeNumerator,
			referenceAmount,
			feeIncrementBps,
			checkedIncludedFeeAmount,
		)
		if err != nil {
			return struct {
				CheckedExcludedFeeAmount *big.Int
				CheckedIncludedFeeAmount *big.Int
				IsOverflow               bool
			}{}, err
		}
		return struct {
			CheckedExcludedFeeAmount *big.Int
			CheckedIncludedFeeAmount *big.Int
			IsOverflow               bool
		}{
			CheckedExcludedFeeAmount: checkedOutputAmount,
			CheckedIncludedFeeAmount: checkedIncludedFeeAmount,
		}, nil
	}

	checkedOutputAmount, err := GetRateLimiterExcludedFeeAmount(
		cliffFeeNumerator,
		referenceAmount,
		feeIncrementBps,
		constants.U64MaxBigInt,
	)
	if err != nil {
		return struct {
			CheckedExcludedFeeAmount *big.Int
			CheckedIncludedFeeAmount *big.Int
			IsOverflow               bool
		}{}, err
	}
	return struct {
		CheckedExcludedFeeAmount *big.Int
		CheckedIncludedFeeAmount *big.Int
		IsOverflow               bool
	}{
		CheckedExcludedFeeAmount: checkedOutputAmount,
		CheckedIncludedFeeAmount: constants.U64MaxBigInt,
		IsOverflow:               true,
	}, nil
}

// GetFeeNumeratorFromExcludedAmount calculates the fee numerator on rate limiter from excluded fee amount.
func GetFeeNumeratorFromExcludedAmount(
	cliffFeeNumerator,
	referenceAmount,
	feeIncrementBps,
	excludedFeeAmount *big.Int,
) (*big.Int, error) {
	// Need to categorize in 3 cases:
	// - excluded_fee_amount <= get_excluded_fee_amount(reference_amount)
	// - excluded_fee_amount > get_excluded_fee_amount(reference_amount) && excluded_fee_amount < get_excluded_fee_amount(reference_amount * (max_index+1))
	// - excluded_fee_amount >= get_excluded_fee_amount(reference_amount * (max_index+1))
	// Note: because excluded_fee_amount = included_fee_amount - fee_numerator * included_fee_amount / fee_denominator
	// It is very difficult to calculate exactly fee_numerator from excluded_fee_amount,
	// With any precision difference, even 1 unit, the excluded_fee_amount will be changed a lot when value of included_fee_amount is high
	// Then a sanity check here is we just ensure fee_numerator >= cliff_fee_numerator
	// Note: That also exclude the dynamic fee in calculation, so in rate limiter fee mode, fees can be different for different swap modes

	excludedFeeReferenceAmount, err := GetRateLimiterExcludedFeeAmount(
		cliffFeeNumerator,
		referenceAmount,
		feeIncrementBps,
		referenceAmount,
	)
	if err != nil {
		return nil, err
	}

	if excludedFeeAmount.Cmp(excludedFeeReferenceAmount) <= 0 {
		return cliffFeeNumerator, nil
	}

	out, err := GetCheckedAmounts(cliffFeeNumerator, referenceAmount, feeIncrementBps)
	if err != nil {
		return nil, err
	}

	// Add the early check
	if excludedFeeAmount.Cmp(out.CheckedExcludedFeeAmount) == 0 {
		return GetFeeNumeratorFromIncludedAmount(
			cliffFeeNumerator,
			referenceAmount,
			feeIncrementBps,
			out.CheckedIncludedFeeAmount,
		)
	}

	var includedFeeAmount *big.Int
	if excludedFeeAmount.Cmp(out.CheckedExcludedFeeAmount) < 0 {
		two, four := big.NewInt(2), big.NewInt(4)

		// d: fee denominator
		// ex: excluded_fee_amount
		// input_amount = x0 + (a * x0)
		// fee = x0 * (c + c*a + i*a*(a+1)/2) / d
		// fee = x0 * (a+1) * (c + i*a/2) / d
		// fee = input_amount * (c + i * (input_amount/x0-1)/2) / d
		// ex = input_amount - fee
		// ex = input_amount - input_amount * (c + i * (input_amount/x0-1)/2) / d
		// ex * d * 2 = input_amount * d * 2 - input_amount * (2 * c + i * (input_amount/x0-1))
		// ex * d * 2 * x0 = input_amount * d * 2 * x0 - input_amount * (2 * c * x0 + i * (input_amount-x0))
		// ex * d * 2 * x0 = input_amount * d * 2 * x0 - input_amount * (2 * c * x0 + i * input_amount- i*x0)
		// ex * d * 2 * x0 = input_amount * d * 2 * x0 - input_amount * 2 * c * x0 - i * input_amount ^ 2 + input_amount * i*x0
		// i * input_amount ^ 2 - input_amount * (-2 * c * x0 + i*x0 + d * 2 * x0) + ex * d * 2 * x0 = 0
		// equation: x * input_amount ^ 2  - y * input_amount + z = 0
		// x = i, y =  (-2 * c * x0 + i*x0 + d * 2 * x0), z = ex * d * 2 * x0
		// input_amount = (y +(-) sqrt(y^2 - 4xz)) / 2x

		i := toNumerators(feeIncrementBps, constants.FeeDenominatorBigInt)
		x, x0, d, c, ex := i, referenceAmount, constants.FeeDenominatorBigInt,
			cliffFeeNumerator, excludedFeeAmount

		y := new(big.Int).Sub(
			new(big.Int).Add(
				new(big.Int).Mul(
					new(big.Int).Mul(two, d), x0),
				new(big.Int).Mul(i, x0),
			),
			new(big.Int).Mul(
				new(big.Int).Mul(two, c), x0),
		)

		z := new(big.Int).Mul(
			new(big.Int).Mul(
				new(big.Int).Mul(two, ex), d),
			x0,
		)

		// solve quadratic equation
		// check it again, why sub, not add
		discriminant := new(big.Int).Sub(
			new(big.Int).Mul(y, y),
			new(big.Int).Mul(
				new(big.Int).Mul(four, x), z,
			),
		)
		sqrtDiscriminant := Sqrt(discriminant)

		includedFeeAmount = new(big.Int).Quo(
			new(big.Int).Sub(y, sqrtDiscriminant),
			new(big.Int).Mul(two, x),
		)

		firstExcludedFeeAmount, err := GetRateLimiterExcludedFeeAmount(
			cliffFeeNumerator,
			referenceAmount,
			feeIncrementBps,
			includedFeeAmount,
		)
		if err != nil {
			return nil, err
		}

		excludedFeeRemainingAmount := new(big.Int).Sub(excludedFeeAmount, firstExcludedFeeAmount)
		aPlusOne := new(big.Int).Quo(includedFeeAmount, x0)

		remainingAmountFeeNumerator := new(big.Int).Add(c, new(big.Int).Mul(i, aPlusOne))

		includedFeeRemainingAmount, _ := MulDiv(
			excludedFeeRemainingAmount,
			constants.FeeDenominatorBigInt,
			new(big.Int).Sub(constants.FeeDenominatorBigInt, remainingAmountFeeNumerator),
			types.RoundingUp,
		)
		totalInAmount := new(big.Int).Add(includedFeeAmount, includedFeeRemainingAmount)
		includedFeeAmount = totalInAmount

	} else {
		// excluded_fee_amount > checked_excluded_fee_amount
		if out.IsOverflow {
			return nil, errors.New("math overflow")
		}

		excludedFeeRemainingAmount := new(big.Int).Sub(excludedFeeAmount, out.CheckedExcludedFeeAmount)

		// remaining_amount should take the max fee
		includedFeeRemainingAmount, _ := MulDiv(
			excludedFeeRemainingAmount,
			constants.FeeDenominatorBigInt,
			new(big.Int).Sub(constants.FeeDenominatorBigInt, big.NewInt(constants.MaxFeeNumerator)),
			types.RoundingUp,
		)
		totalInAmount := new(big.Int).Add(includedFeeRemainingAmount, out.CheckedIncludedFeeAmount)
		includedFeeAmount = totalInAmount
	}

	tradingFee := new(big.Int).Sub(includedFeeAmount, excludedFeeAmount)

	feeNumerator, _ := MulDiv(
		tradingFee,
		constants.FeeDenominatorBigInt,
		includedFeeAmount,
		types.RoundingUp,
	)

	// sanity check
	if feeNumerator.Cmp(cliffFeeNumerator) < 0 {
		return nil,
			fmt.Errorf("undetermined error: feeNumerator(%s) less than cliffFeeNumerator(%s)",
				feeNumerator, cliffFeeNumerator)
	}

	return feeNumerator, nil
}

// GetRateLimiterExcludedFeeAmount gets excluded fee amount from included fee amount using rate limiter.
func GetRateLimiterExcludedFeeAmount(
	cliffFeeNumerator,
	referenceAmount,
	feeIncrementBps,
	includedFeeAmount *big.Int,
) (*big.Int, error) {
	feeNumerator, err := GetFeeNumeratorFromIncludedAmount(
		cliffFeeNumerator,
		referenceAmount,
		feeIncrementBps,
		includedFeeAmount,
	)
	if err != nil {
		return nil, err
	}

	tradingFee, _ := MulDiv(
		includedFeeAmount,
		feeNumerator,
		constants.FeeDenominatorBigInt,
		types.RoundingUp,
	)

	return new(big.Int).Sub(includedFeeAmount, tradingFee), nil
}

// GetFeeNumeratorOnRateLimiter calculates the fee numerator on rate limiter from included fee amount.
func GetFeeNumeratorFromIncludedAmount(
	cliffFeeNumerator, referenceAmount, feeIncrementBps, includedFeeAmount *big.Int,
) (*big.Int, error) {

	if includedFeeAmount.Cmp(referenceAmount) <= 0 {
		return cliffFeeNumerator, nil
	}

	if includedFeeAmount.Sign() == 0 {
		return nil, errors.New("input amount cannot be zero")
	}

	diff := new(big.Int).Sub(includedFeeAmount, referenceAmount)
	a, b := new(big.Int).QuoRem(diff, referenceAmount, new(big.Int))
	maxIndex, err := GetMaxIndex(cliffFeeNumerator, feeIncrementBps)

	if err != nil {
		return nil, err
	}

	i := toNumerators(
		feeIncrementBps,
		new(big.Int).SetInt64(constants.FeeDenominator),
	)

	one, two := big.NewInt(1), big.NewInt(2)

	var tradingFeeNumerator *big.Int
	if a.Cmp(maxIndex) < 0 {
		// c + c * a
		partOne := new(big.Int).Add(
			cliffFeeNumerator,
			new(big.Int).Mul(cliffFeeNumerator, a),
		)
		// i * a * (a + 1) / 2
		partTwo := new(big.Int).Quo(
			new(big.Int).Mul(
				new(big.Int).Mul(i, a),
				new(big.Int).Add(a, one),
			),
			two,
		)
		numerator1 := new(big.Int).Add(partOne, partTwo)

		// c + i * (a + 1)
		numerator2 := new(big.Int).Add(
			cliffFeeNumerator,
			new(big.Int).Mul(
				i,
				new(big.Int).Add(a, one),
			),
		)

		firstFee, secondFee := new(big.Int).Mul(referenceAmount, numerator1),
			new(big.Int).Mul(b, numerator2)

		tradingFeeNumerator = new(big.Int).Add(firstFee, secondFee)
	} else {
		// c + (c * maxIndex)
		partOne := new(big.Int).Add(
			cliffFeeNumerator,
			new(big.Int).Mul(cliffFeeNumerator, maxIndex),
		)
		// (i * maxIndex * (maxIndex + 1)) / 2
		partTwo := new(big.Int).Quo(
			new(big.Int).Mul(
				new(big.Int).Mul(i, maxIndex),
				new(big.Int).Add(maxIndex, one),
			),
			two,
		)
		numerator1, numerator2 := new(big.Int).Add(partOne, partTwo), new(big.Int).SetUint64(constants.MaxFeeNumerator)

		firstFee, d := new(big.Int).Mul(referenceAmount, numerator1),
			new(big.Int).Sub(a, maxIndex)
		leftAmount := new(big.Int).Add(new(big.Int).Mul(d, referenceAmount), b)
		secondFee := new(big.Int).Mul(leftAmount, numerator2)

		tradingFeeNumerator = new(big.Int).Add(firstFee, secondFee)
	}

	denominator := new(big.Int).SetUint64(constants.FeeDenominator)
	tradingFee := new(big.Int).Div(
		new(big.Int).Sub(
			new(big.Int).Add(tradingFeeNumerator, denominator),
			one,
		),
		denominator,
	)

	// reverse to fee numerator:
	// input_amount * numerator / FEE_DENOMINATOR = trading_fee
	// => numerator = trading_fee * FEE_DENOMINATOR / input_amount
	feeNumerator, _ := MulDiv(
		tradingFee,
		denominator,
		includedFeeAmount,
		types.RoundingUp,
	)

	return feeNumerator, nil
}
