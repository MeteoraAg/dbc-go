package poolfees

import (
	"dbcGoSDK/constants"
	"dbcGoSDK/types"
	"errors"
	"math/big"
)

func MulDiv(x, y, denominator *big.Int, rounding types.Rounding) (*big.Int, error) {
	if denominator.Sign() == 0 {
		return nil, errors.New("MulDiv: division by zero")
	}

	prod := new(big.Int).Mul(x, y)
	if denominator.Cmp(big.NewInt(1)) == 0 || x.Sign() == 0 || y.Sign() == 0 {
		return prod, nil
	}

	if rounding == types.RoundingUp {
		// Calculate ceiling division: (prod + denominator - 1) / denominator
		return new(big.Int).Quo(
			new(big.Int).Add(
				prod,
				new(big.Int).Sub(denominator, big.NewInt(1)),
			),
			denominator,
		), nil
	}

	return new(big.Int).Quo(prod, denominator), nil
}

// toNumerators converts basis points to fee numerator.
func toNumerators(bps, feeDenominator *big.Int) *big.Int {

	r, _ := MulDiv(
		bps,
		feeDenominator,
		big.NewInt(constants.BasisPointMax),
		types.RoundingDown,
	)

	return r
}

// PowQ64 is a custom power function for [big.Int] with scaling.
func PowQ64(base, exponent *big.Int, scaling bool) *big.Int {

	// special cases
	if exponent.Sign() == 0 {
		return constants.OneQ64
	}
	if base.Sign() == 0 {
		return big.NewInt(0)
	}
	if base.Cmp(constants.OneQ64) == 0 {
		return constants.OneQ64
	}

	isNegative, absExponent := exponent.Sign() < 0, new(big.Int).Abs(exponent)

	result, currentBase, exp := constants.OneQ64, base, absExponent
	for exp.Sign() != 0 {
		if one := big.NewInt(1); new(big.Int).And(exp, one).Cmp(one) == 0 {
			result = new(big.Int).Quo(
				new(big.Int).Mul(result, currentBase),
				constants.OneQ64,
			)
		}
		currentBase = new(big.Int).Quo(
			new(big.Int).Mul(currentBase, currentBase),
			constants.OneQ64,
		)
		exp = new(big.Int).Rsh(exp, 1)
	}

	// handle negative exponent
	if isNegative {
		result = new(big.Int).Quo(
			new(big.Int).Mul(constants.OneQ64, constants.OneQ64),
			result,
		)
	}

	if !scaling {
		return new(big.Int).Quo(result, constants.OneQ64)
	}

	return result
}

// Sqrt calculates square root of a BN number using Newton's method.
func Sqrt(value *big.Int) *big.Int {
	if value.Sign() == 0 {
		return big.NewInt(0)
	}

	hold := big.NewInt(1)
	if value.Cmp(hold) == 0 {
		return hold
	}

	hold = big.NewInt(2)
	x, y := value, new(big.Int).Quo(
		new(big.Int).Add(value, big.NewInt(1)),
		hold,
	)

	for y.Cmp(x) < 0 {
		x = y
		y = new(big.Int).Quo(
			new(big.Int).Add(x, new(big.Int).Quo(value, x)),
			hold,
		)
	}
	return x
}
