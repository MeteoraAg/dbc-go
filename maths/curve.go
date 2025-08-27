package maths

import (
	"dbcGoSDK/constants"
	"dbcGoSDK/types"
	"fmt"
	"math/big"
)

// GetDeltaAmountBaseUnsigned gets the delta amount_base for given liquidity and price range.
//
//	Formula: Δa = L * (1 / √P_lower - 1 / √P_upper)
//	i.e. L * (√P_upper - √P_lower) / (√P_upper * √P_lower)
func GetDeltaAmountBaseUnsigned(
	lowerSqrtPrice, upperSqrtPrice, liquidity *big.Int,
	round types.Rounding,
) (*big.Int, error) {

	result, err := GetDeltaAmountBaseUnsigned256(
		lowerSqrtPrice,
		upperSqrtPrice,
		liquidity,
		round,
	)
	if err != nil {
		return nil, err
	}

	if result.Cmp(constants.U64MaxBigInt) > 0 {
		return nil, fmt.Errorf("GetDeltaAmountBaseUnsigned:result(%s) exceeded %s", result, constants.U64MaxBigInt)
	}

	return result, nil
}

func GetDeltaAmountBaseUnsigned256(
	lowerSqrtPrice, upperSqrtPrice, liquidity *big.Int,
	round types.Rounding,
) (*big.Int, error) {
	return GetDeltaAmountBaseUnsignedUnchecked(
		lowerSqrtPrice,
		upperSqrtPrice,
		liquidity,
		round,
	)
}

func GetDeltaAmountBaseUnsignedUnchecked(
	lowerSqrtPrice, upperSqrtPrice, liquidity *big.Int,
	round types.Rounding,
) (*big.Int, error) {
	// numerator1 := new(big.Int).Set(liquidity)

	// numerator: (√P_upper - √P_lower)
	numerator2 := new(big.Int).Sub(upperSqrtPrice, lowerSqrtPrice)
	if numerator2.Sign() < 0 {
		return nil, fmt.Errorf("GetDeltaAmountBaseUnsignedUnchecked:safeMath requires value not negative: value is %s", numerator2)
	}

	// denominator: (√P_upper * √P_lower)
	denominator := new(big.Int).Mul(lowerSqrtPrice, upperSqrtPrice)
	if denominator.Sign() == 0 {
		return nil, fmt.Errorf("GetDeltaAmountBaseUnsignedUnchecked:denominator(%s) cannot be zero", denominator)
	}

	// L * (√P_upper - √P_lower) / (√P_upper * √P_lower)
	return MulDiv(
		liquidity,
		numerator2,
		denominator,
		round,
	)
}

// GetDeltaAmountQuoteUnsigned gets the delta amount_quote for given liquidity and price range.
//
//	Formula: Δb = L (√P_upper - √P_lower)
func GetDeltaAmountQuoteUnsigned(
	lowerSqrtPrice, upperSqrtPrice, liquidity *big.Int,
	round types.Rounding,
) (*big.Int, error) {

	result, err := GetDeltaAmountQuoteUnsigned256(
		lowerSqrtPrice,
		upperSqrtPrice,
		liquidity,
		round,
	)
	if err != nil {
		return nil, err
	}

	if result.Cmp(constants.U64MaxBigInt) > 0 {
		return nil, fmt.Errorf("GetDeltaAmountQuoteUnsigned256:result(%s) exceeded %s", result, constants.U64MaxBigInt)
	}

	return result, nil
}

func GetDeltaAmountQuoteUnsigned256(
	lowerSqrtPrice, upperSqrtPrice, liquidity *big.Int,
	round types.Rounding,
) (*big.Int, error) {
	return GetDeltaAmountQuoteUnsignedUnchecked(
		lowerSqrtPrice,
		upperSqrtPrice,
		liquidity,
		round,
	)
}

// GetDeltaAmountQuoteUnsignedUnchecked
//
//	Formula: Δb = L (√P_upper - √P_lower)
func GetDeltaAmountQuoteUnsignedUnchecked(
	lowerSqrtPrice, upperSqrtPrice, liquidity *big.Int,
	round types.Rounding,
) (*big.Int, error) {

	// delta sqrt price: (√P_upper - √P_lower)
	deltaSqrtPrice := new(big.Int).Sub(upperSqrtPrice, lowerSqrtPrice)
	if deltaSqrtPrice.Sign() < 0 {
		return nil, fmt.Errorf("GetDeltaAmountQuoteUnsignedUnchecked:safeMath requires value not negative: value is %s", deltaSqrtPrice.String())
	}

	// L * (√P_upper - √P_lower)
	prod := new(big.Int).Mul(liquidity, deltaSqrtPrice)

	if round == types.RoundingUp {
		denominator := new(big.Int).Lsh(big.NewInt(1), constants.RESOLUTION*2)
		// ceiling division: (a + b - 1) / b
		numerator := new(big.Int).Add(
			prod,
			new(big.Int).Sub(denominator, big.NewInt(1)),
		)
		return new(big.Int).Quo(numerator, denominator), nil
	}

	return new(big.Int).Rsh(prod, constants.RESOLUTION*2), nil
}

// GetInitialLiquidityFromDeltaQuote gets the initial liquidity from delta quote.
//
//	Formula: Δb = L (√P_upper - √P_lower) => L = Δb / (√P_upper - √P_lower)
func GetInitialLiquidityFromDeltaQuote(
	quoteAmount, sqrtMinPrice, sqrtPrice *big.Int,
) (*big.Int, error) {
	priceDelta := new(big.Int).Sub(sqrtPrice, sqrtMinPrice)
	if priceDelta.Sign() < 0 {
		return nil, fmt.Errorf("GetInitialLiquidityFromDeltaQuote:safeMath requires value not negative: value is %s", priceDelta.String())
	}
	return new(big.Int).Quo(
		new(big.Int).Rsh(quoteAmount, constants.RESOLUTION*2),
		priceDelta,
	), nil
}

// GetInitialLiquidityFromDeltaBase gets the initial liquidity from delta base.
//
//	Formula: Δa = L * (1 / √P_lower - 1 / √P_upper) => L = Δa / (1 / √P_lower - 1 / √P_upper)
func GetInitialLiquidityFromDeltaBase(
	baseAmount, sqrtMaxPrice, sqrtPrice *big.Int,
) (*big.Int, error) {
	priceDelta := new(big.Int).Sub(sqrtMaxPrice, sqrtPrice)
	if priceDelta.Sign() < 0 {
		return nil, fmt.Errorf("GetInitialLiquidityFromDeltaQuote:safeMath requires value not negative: value is %s", priceDelta.String())
	}
	return new(big.Int).Quo( // rounds down
		new(big.Int).Mul(
			new(big.Int).Mul(baseAmount, sqrtPrice),
			sqrtMaxPrice,
		),
		priceDelta,
	), nil
}

// GetNextSqrtPriceFromInput gets the next sqrt price given an input amount of token_a or token_b.
func GetNextSqrtPriceFromInput(
	sqrtPrice, liquidity, amountIn *big.Int,
	baseForQuote bool,
) (*big.Int, error) {

	if sqrtPrice.Sign() == 0 || liquidity.Sign() == 0 {
		return nil, fmt.Errorf("GetNextSqrtPriceFromInput:sqrtPrice(%s) or liquidity(%s) cannot be zero", sqrtPrice, liquidity)
	}

	// round to make sure that we don't pass the target price
	if baseForQuote {
		return GetNextSqrtPriceFromBaseAmountInRoundingUp(
			sqrtPrice, liquidity, amountIn,
		), nil
	}

	return GetNextSqrtPriceFromQuoteAmountInRoundingDown(
		sqrtPrice, liquidity, amountIn,
	), nil
}

// GetNextSqrtPriceFromOutput gets the next sqrt price from output amount.
func GetNextSqrtPriceFromOutput(
	sqrtPrice, liquidity, outAmount *big.Int,
	baseForQuote bool,
) (*big.Int, error) {

	if sqrtPrice.Sign() == 0 || liquidity.Sign() == 0 {
		return nil, fmt.Errorf("GetNextSqrtPriceFromOutput:sqrtPrice(%s) or liquidity(%s) cannot be zero", sqrtPrice, liquidity)
	}

	if baseForQuote {
		return GetNextSqrtPriceFromQuoteAmountOutRoundingDown(
			sqrtPrice, liquidity, outAmount,
		)
	}

	return GetNextSqrtPriceFromBaseAmountOutRoundingUp(
		sqrtPrice, liquidity, outAmount,
	)
}

// GetNextSqrtPriceFromQuoteAmountOutRoundingDown gets the next sqrt price from amount quote rounding up.
//
//	Formula: √P' = √P - Δy / L
func GetNextSqrtPriceFromQuoteAmountOutRoundingDown(
	sqrtPrice, liquidity, amount *big.Int,
) (*big.Int, error) {

	// q_amount = amount << 128
	qAmount := new(big.Int).Lsh(amount, 128)

	// quotient = q_amount.div_ceil(liquidity)
	// div_ceil is equivalent to (a + b - 1) / b
	numerator := new(big.Int).Add(
		qAmount,
		new(big.Int).Sub(liquidity, big.NewInt(1)),
	)
	quotient := new(big.Int).Quo(numerator, liquidity)

	// √P - quotient
	r := new(big.Int).Sub(sqrtPrice, quotient)
	if r.Sign() < 0 {
		return nil, fmt.Errorf("GetNextSqrtPriceFromQuoteAmountOutRoundingDown:safeMath requires value non-zero: value is %s", r.String())
	}

	return r, nil
}

// GetNextSqrtPriceFromBaseAmountOutRoundingUp gets the next sqrt price from amount base rounding down.
//
//	Formula: √P' = √P * L / (L - Δx * √P)
func GetNextSqrtPriceFromBaseAmountOutRoundingUp(
	sqrtPrice, liquidity, amount *big.Int,
) (*big.Int, error) {

	if amount.Sign() == 0 {
		return sqrtPrice, nil
	}

	// Δx * √P
	product := new(big.Int).Mul(amount, sqrtPrice)

	// L - Δx * √P
	denominator := new(big.Int).Sub(liquidity, product)
	if denominator.Sign() <= 0 {
		return nil,
			fmt.Errorf("invalid denominator(%s): liquidity must be greater than amount * sqrt_price", denominator.String())
	}

	// √P * L / (L - Δx * √P) with rounding down
	return MulDiv(
		liquidity, sqrtPrice, denominator, types.RoundingUp,
	)
}

// GetNextSqrtPriceFromBaseAmountInRoundingUp gets the next sqrt price from amount base rounding up.
//
// Always round up because:
//
//  1. In the exact output case, token 0 supply decreases leading to price increase.
//     Move price up so that exact output is met.
//
//  2. In the exact input case, token 0 supply increases leading to price decrease.
//     Do not round down to minimize price impact. We only need to meet input
//     change and not guarantee exact output.
//
//     Formula: √P' = √P * L / (L + Δx * √P)
//
//     If Δx * √P overflows, use alternate form √P' = L / (L/√P + Δx)
func GetNextSqrtPriceFromBaseAmountInRoundingUp(
	sqrtPrice, liquidity, amount *big.Int,
) *big.Int {

	if amount.Sign() == 0 {
		return sqrtPrice
	}

	// Check for potential overflow in Δx * √P
	product := new(big.Int).Mul(amount, sqrtPrice)

	// Check if product would overflow - if so, use alternate form
	if product.Cmp(constants.U128MaxBigInt) > 0 {
		// Alternate form: √P' = L / (L/√P + Δx)
		return new(big.Int).Quo(
			liquidity,
			new(big.Int).Add(
				new(big.Int).Quo(liquidity, sqrtPrice),
				amount,
			),
		)
	}

	// Standard form: √P' = √P * L / (L + Δx * √P)
	r, _ := MulDiv(
		liquidity,
		sqrtPrice,
		new(big.Int).Add(liquidity, product),
		types.RoundingUp,
	)
	return r
}

// GetNextSqrtPriceFromQuoteAmountInRoundingDown gets the next sqrt price given a delta of token_quote.
//
// Always round down because:
//
//  1. In the exact output case, token 1 supply decreases leading to price decrease.
//     Move price down by rounding down so that exact output of token 0 is met.
//
//  2. In the exact input case, token 1 supply increases leading to price increase.
//     Do not round down to minimize price impact. We only need to meet input
//     change and not guarantee exact output for token 0.
//
//     Formula: √P' = √P + Δy / L
func GetNextSqrtPriceFromQuoteAmountInRoundingDown(
	sqrtPrice, liquidity, amount *big.Int,
) *big.Int {

	// quotient: Δy << (RESOLUTION * 2) / L
	quotient := new(big.Int).Quo(
		new(big.Int).Lsh(amount, constants.RESOLUTION*2),
		liquidity,
	)

	// √P' = √P + Δy / L
	return new(big.Int).Add(sqrtPrice, quotient)
}
