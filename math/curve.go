package math

import (
	"math/big"

	"github.com/MeteoraAg/dbc-go/common"
)

// gets the delta amount_quote for given liquidity and price range
// Formula: Δb = L (√P_upper - √P_lower)
func GetDeltaAmountQuoteUnsigned(
	lowerSqrtPrice *big.Int,
	upperSqrtPrice *big.Int,
	liquidity *big.Int,
	round common.Rounding,
) (*big.Int, error) {
	if liquidity.Sign() == 0 {
		return big.NewInt(0), nil
	}

	// delta sqrt price: (√P_upper - √P_lower)
	deltaSqrtPrice, err := Sub(upperSqrtPrice, lowerSqrtPrice)
	if err != nil {
		return nil, err
	}

	// L * (√P_upper - √P_lower)
	prod := Mul(liquidity, deltaSqrtPrice)

	if round == common.Up {
		denominator := new(big.Int).Lsh(big.NewInt(1), uint(common.Resolution*2))

		// ceiling division: (a + b - 1) / b
		denominatorMinusOne, err := Sub(denominator, big.NewInt(1))
		if err != nil {
			return nil, err
		}
		numerator := Add(prod, denominatorMinusOne)
		return Div(numerator, denominator)
	} else { // common.Down
		return Shr(prod, uint(common.Resolution*2)), nil
	}
}

func GetQuoteReserveFromNextSqrtPrice(nextSqrtPrice *big.Int, config *common.PoolConfig) (*big.Int, error) {
	totalAmount := big.NewInt(0)

	for i := 0; i < common.MaxCurvePoint; i++ {
		var lowerSqrtPrice *big.Int
		if i == 0 {
			lowerSqrtPrice = u128ToBig(config.SqrtStartPrice)
		} else {
			lowerSqrtPrice = u128ToBig(config.Curve[i-1].SqrtPrice)
		}

		if nextSqrtPrice.Cmp(lowerSqrtPrice) > 0 {
			curveUpperSqrtPrice := u128ToBig(config.Curve[i].SqrtPrice)

			var upperSqrtPrice *big.Int
			if nextSqrtPrice.Cmp(curveUpperSqrtPrice) < 0 {
				upperSqrtPrice = nextSqrtPrice
			} else {
				upperSqrtPrice = curveUpperSqrtPrice
			}

			liquidity := u128ToBig(config.Curve[i].Liquidity)

			maxAmountIn, err := GetDeltaAmountQuoteUnsigned(
				lowerSqrtPrice,
				upperSqrtPrice,
				liquidity,
				common.Up,
			)
			if err != nil {
				return nil, err
			}

			totalAmount.Add(totalAmount, maxAmountIn)
		}
	}

	return totalAmount, nil
}
