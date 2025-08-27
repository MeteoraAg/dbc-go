package maths

import (
	"dbcGoSDK/types"
	"errors"
	"fmt"
	"math/big"

	ag_binary "github.com/gagliardetto/binary"
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

func Q64(n float64) *big.Int {
	f := new(big.Float).Mul(
		new(big.Float).SetFloat64(n),
		new(big.Float).SetFloat64(1<<64),
	)
	i := new(big.Int)
	f.Int(i) // truncates
	return i
}

func BigIntToUint128(b *big.Int) (ag_binary.Uint128, error) {
	if b.Sign() < 0 {
		return ag_binary.Uint128{}, fmt.Errorf("value must be unsigned")
	}

	if b.BitLen() > 128 {
		return ag_binary.Uint128{}, fmt.Errorf("value %s exceeds 128 bits", b.String())
	}

	var buf [16]byte
	b.FillBytes(buf[:]) // zero-pads on the left

	ag_binary.ReverseBytes(buf[:])

	var u ag_binary.Uint128
	if err := u.UnmarshalWithDecoder(ag_binary.NewBinDecoder(buf[:])); err != nil {
		return ag_binary.Uint128{}, err
	}
	return u, nil
}

// Must helper
func MustBigIntToUint128(b *big.Int) ag_binary.Uint128 {
	v, err := BigIntToUint128(b)
	if err != nil {
		panic(fmt.Errorf("cannot fit big.Int into Uint128: %s", err.Error()))
	}
	return v
}
