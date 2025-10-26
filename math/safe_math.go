package math

import (
	"errors"
	"math/big"

	"github.com/MeteoraAg/dbc-go/common"
	"lukechampine.com/uint128"
)

// safe addition
func Add(a, b *big.Int) *big.Int {
	return new(big.Int).Add(a, b)
}

// safe subtraction - return error if b > a
func Sub(a, b *big.Int) (*big.Int, error) {
	if b.Cmp(a) > 0 {
		return nil, errors.New("SafeMath: subtraction overflow")
	}
	return new(big.Int).Sub(a, b), nil
}

// safe multiplication
func Mul(a, b *big.Int) *big.Int {
	return new(big.Int).Mul(a, b)
}

// safe division - return error if b is zero
func Div(a, b *big.Int) (*big.Int, error) {
	if b.Sign() == 0 {
		return nil, errors.New("SafeMath: division by zero")
	}
	return new(big.Int).Div(a, b), nil
}

// safe modulo - return error if b is zero
func Mod(a, b *big.Int) (*big.Int, error) {
	if b.Sign() == 0 {
		return nil, errors.New("SafeMath: modulo by zero")
	}
	return new(big.Int).Mod(a, b), nil
}

// safe left shift
func Shl(a *big.Int, b uint) *big.Int {
	return new(big.Int).Lsh(a, b)
}

// safe right shift.
func Shr(a *big.Int, b uint) *big.Int {
	return new(big.Int).Rsh(a, b)
}

// base^exponent with scaling
func Pow(base, exponent *big.Int, scaling bool) (*big.Int, error) {
	one := new(big.Int).Lsh(big.NewInt(1), uint(common.Resolution))

	if exponent.Sign() == 0 {
		return new(big.Int).Set(one), nil
	}
	if base.Sign() == 0 {
		return big.NewInt(0), nil
	}
	if base.Cmp(one) == 0 {
		return new(big.Int).Set(one), nil
	}

	isNegative := exponent.Sign() < 0
	absExponent := new(big.Int)
	if isNegative {
		absExponent.Neg(exponent)
	} else {
		absExponent.Set(exponent)
	}

	result := new(big.Int).Set(one)
	currentBase := new(big.Int).Set(base)
	exp := new(big.Int).Set(absExponent)

	for exp.Sign() != 0 {
		if exp.Bit(0) == 1 {
			mulResult := Mul(result, currentBase)
			divResult, err := Div(mulResult, one)
			if err != nil {
				return nil, err
			}
			result = divResult
		}
		mulResult := Mul(currentBase, currentBase)
		divResult, err := Div(mulResult, one)
		if err != nil {
			return nil, err
		}
		currentBase = divResult
		exp.Rsh(exp, 1)
	}

	if isNegative {
		oneSquared := Mul(one, one)
		divResult, err := Div(oneSquared, result)
		if err != nil {
			return nil, err
		}
		result = divResult
	}

	if scaling {
		return result, nil
	}

	divResult, err := Div(result, one)
	if err != nil {
		return nil, err
	}
	return divResult, nil
}

// converts uint128.Uint128 to *big.Int
func u128ToBig(val uint128.Uint128) *big.Int {
	hi := new(big.Int).SetUint64(val.Hi)
	lo := new(big.Int).SetUint64(val.Lo)
	hi.Lsh(hi, 64)
	return hi.Or(hi, lo)
}
