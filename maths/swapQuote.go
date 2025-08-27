package maths

import (
	"dbcGoSDK/constants"
	"dbcGoSDK/generated/dbc"
	"dbcGoSDK/types"
	"errors"
	"fmt"
	"math/big"
)

// SwapQuote V1 //

// GetSwapResult gets swap result.
func GetSwapResult(
	poolState *dbc.VirtualPoolAccount,
	configState *dbc.PoolConfigAccount,
	amountIn *big.Int,
	feeMode types.FeeMode,
	tradeDirection types.TradeDirection,
	currentPoint *big.Int,
) (dbc.SwapResult, error) {

	actualProtocolFee, actualTradingFee, actualReferralFee :=
		big.NewInt(0), big.NewInt(0), big.NewInt(0)

	tradeFeeNumerator, err := GetTotalFeeNumeratorFromIncludedFeeAmount(
		configState.PoolFees,
		poolState.VolatilityTracker,
		currentPoint,
		new(big.Int).SetUint64(poolState.ActivationPoint),
		amountIn,
		tradeDirection,
	)
	if err != nil {
		return dbc.SwapResult{}, err
	}

	actualAmountIn := new(big.Int).Set(amountIn)
	if feeMode.FeesOnInput {
		feeResult, err := GetFeeOnAmount(
			tradeFeeNumerator,
			amountIn,
			configState.PoolFees,
			feeMode.HasReferral,
		)
		if err != nil {
			return dbc.SwapResult{}, err
		}

		actualProtocolFee, actualTradingFee, actualReferralFee =
			feeResult.ProtocolFee, feeResult.TradingFee, feeResult.ReferralFee

		actualAmountIn = feeResult.Amount
	}

	var swapAmountFromInput types.SwapAmount
	if tradeDirection == types.TradeDirectionBaseToQuote {
		if swapAmountFromInput, err = CalculateBaseToQuoteFromAmountIn(
			configState.Curve[:],
			poolState.SqrtPrice.BigInt(),
			actualAmountIn,
		); err != nil {
			return dbc.SwapResult{}, err
		}
	} else {
		if swapAmountFromInput, err = CalculateQuoteToBaseFromAmountIn(
			configState.Curve[:],
			poolState.SqrtPrice.BigInt(),
			actualAmountIn,
			constants.U128MaxBigInt,
		); err != nil {
			return dbc.SwapResult{}, err
		}
	}

	var actualAmountOut = swapAmountFromInput.OutputAmount
	if !feeMode.FeesOnInput {
		feeResult, err := GetFeeOnAmount(
			tradeFeeNumerator,
			swapAmountFromInput.OutputAmount,
			configState.PoolFees,
			feeMode.HasReferral,
		)
		if err != nil {
			return dbc.SwapResult{}, err
		}

		actualProtocolFee = feeResult.ProtocolFee
		actualTradingFee = feeResult.TradingFee
		actualReferralFee = feeResult.ReferralFee

		actualAmountOut = feeResult.Amount
	}

	if !actualAmountIn.IsUint64() ||
		!actualAmountOut.IsUint64() ||
		!actualTradingFee.IsUint64() ||
		!actualProtocolFee.IsUint64() ||
		!actualReferralFee.IsUint64() {
		return dbc.SwapResult{},
			fmt.Errorf(
				"one of the values cannot fit into uint64: "+
					"ActualInputAmount(%s), OutputAmount(%s), "+
					"TradingFee(%s), ProtocolFee(%s), ReferralFee(%s)",
				actualAmountIn,
				actualAmountOut,
				actualTradingFee,
				actualProtocolFee,
				actualReferralFee,
			)
	}

	return dbc.SwapResult{
		ActualInputAmount: actualAmountIn.Uint64(),
		OutputAmount:      actualAmountOut.Uint64(),
		NextSqrtPrice:     MustBigIntToUint128(swapAmountFromInput.NextSqrtPrice),
		TradingFee:        actualTradingFee.Uint64(),
		ProtocolFee:       actualProtocolFee.Uint64(),
		ReferralFee:       actualReferralFee.Uint64(),
	}, nil
}

// SwapQuote calculates quote for a swap with exact input amount (for swapQuote v1).
func SwapQuote(
	virtualPool *dbc.VirtualPoolAccount,
	config *dbc.PoolConfigAccount,
	swapBaseForQuote bool,
	amountIn *big.Int,
	slippageBps uint64,
	hasReferral bool,
	currentPoint *big.Int,
) (types.SwapQuoteResult, error) {

	if virtualPool.QuoteReserve >= config.MigrationQuoteThreshold {
		return types.SwapQuoteResult{}, errors.New("virtual pool is completed")
	}

	if amountIn.Sign() == 0 {
		return types.SwapQuoteResult{}, errors.New("amountIn cannot be zero")
	}

	tradeDirection := types.TradeDirectionQuoteToBase
	if swapBaseForQuote {
		tradeDirection = types.TradeDirectionBaseToQuote
	}

	feeMode := GetFeeMode(
		types.CollectFeeMode(config.CollectFeeMode),
		tradeDirection,
		hasReferral,
	)

	result, err := GetSwapResult(
		virtualPool,
		config,
		amountIn,
		feeMode,
		tradeDirection,
		currentPoint,
	)
	if err != nil {
		return types.SwapQuoteResult{}, err
	}

	minimumAmountOut := result.OutputAmount
	if slippageBps > 0 {
		// slippage factor: (10000 - slippageBps) / 10000
		slippageFactor, denominator := 10_000-slippageBps, big.NewInt(10_000)

		// minimum amount out: amountOut * (10000 - slippageBps) / 10000
		minimumAmountOutBigInt := new(big.Int).Quo(
			new(big.Int).Mul(new(big.Int).SetUint64(result.OutputAmount), new(big.Int).SetUint64(slippageFactor)),
			denominator,
		)

		if !minimumAmountOutBigInt.IsUint64() {
			return types.SwapQuoteResult{}, fmt.Errorf("cannot fit minimumAmountOutBigInt(%s) into uint64", minimumAmountOutBigInt)
		}

		minimumAmountOut = minimumAmountOutBigInt.Uint64()
	}

	return types.SwapQuoteResult{
		SwapResult:       result,
		MinimumAmountOut: minimumAmountOut,
	}, nil
}

// SwapQuote V2 //

func GetSwapResultFromExactInput(
	virtualPool *dbc.VirtualPoolAccount,
	config *dbc.PoolConfigAccount,
	amountIn *big.Int,
	feeMode types.FeeMode,
	tradeDirection types.TradeDirection,
	currentPoint *big.Int,
) (dbc.SwapResult2, error) {

	if amountIn == nil || currentPoint == nil {
		return dbc.SwapResult2{},
			fmt.Errorf("GetSwapResultFromExactInput:nil value: amountIn(%+v) or currentPoint(%+v) cannot be nil",
				amountIn, currentPoint)
	}

	if !amountIn.IsUint64() {
		return dbc.SwapResult2{},
			fmt.Errorf("cannot fit amountIn(%s) into uint64", amountIn)
	}

	actualProtocolFee, actualTradingFee, actualReferralFee :=
		big.NewInt(0), big.NewInt(0), big.NewInt(0)

	tradeFeeNumerator, err := GetTotalFeeNumeratorFromIncludedFeeAmount(
		config.PoolFees,
		virtualPool.VolatilityTracker,
		currentPoint,
		new(big.Int).SetUint64(virtualPool.ActivationPoint),
		amountIn,
		tradeDirection,
	)
	if err != nil {
		return dbc.SwapResult2{}, err
	}

	actualAmountIn := new(big.Int).Set(amountIn)
	if feeMode.FeesOnInput {
		feeResult, err := GetFeeOnAmount(
			tradeFeeNumerator,
			amountIn,
			config.PoolFees,
			feeMode.HasReferral,
		)
		if err != nil {
			return dbc.SwapResult2{}, err
		}

		actualProtocolFee = feeResult.ProtocolFee
		actualTradingFee = feeResult.TradingFee
		actualReferralFee = feeResult.ReferralFee

		actualAmountIn = feeResult.Amount
	}

	var swapAmountFromInput types.SwapAmount
	if tradeDirection == types.TradeDirectionBaseToQuote {
		if swapAmountFromInput, err = CalculateBaseToQuoteFromAmountIn(
			config.Curve[:],
			virtualPool.SqrtPrice.BigInt(),
			actualAmountIn,
		); err != nil {
			return dbc.SwapResult2{}, err
		}
	} else {
		if swapAmountFromInput, err = CalculateQuoteToBaseFromAmountIn(
			config.Curve[:],
			virtualPool.SqrtPrice.BigInt(),
			actualAmountIn,
			constants.U128MaxBigInt,
		); err != nil {
			return dbc.SwapResult2{}, err
		}
	}

	var actualAmountOut = swapAmountFromInput.OutputAmount
	if !feeMode.FeesOnInput {
		feeResult, err := GetFeeOnAmount(
			tradeFeeNumerator,
			swapAmountFromInput.OutputAmount,
			config.PoolFees,
			feeMode.HasReferral,
		)
		if err != nil {
			return dbc.SwapResult2{}, err
		}

		actualProtocolFee = feeResult.ProtocolFee
		actualTradingFee = feeResult.TradingFee
		actualReferralFee = feeResult.ReferralFee

		actualAmountOut = feeResult.Amount
	}

	if !swapAmountFromInput.AmountLeft.IsUint64() ||
		!actualAmountIn.IsUint64() ||
		!actualAmountOut.IsUint64() ||
		!actualTradingFee.IsUint64() ||
		!actualProtocolFee.IsUint64() ||
		!actualReferralFee.IsUint64() {
		return dbc.SwapResult2{},
			fmt.Errorf(
				"one of these values cannot fit into uint64: "+
					"AmountLeft(%s), ExcludedFeeInputAmount(%s), "+
					"OutputAmount(%s), TradingFee(%s), ProtocolFee(%s), ReferralFee(%s)",
				swapAmountFromInput.AmountLeft,
				actualAmountIn,
				actualAmountOut,
				actualTradingFee,
				actualProtocolFee,
				actualReferralFee,
			)
	}

	return dbc.SwapResult2{
		AmountLeft:             swapAmountFromInput.AmountLeft.Uint64(),
		IncludedFeeInputAmount: amountIn.Uint64(),
		ExcludedFeeInputAmount: actualAmountIn.Uint64(),
		OutputAmount:           actualAmountOut.Uint64(),
		NextSqrtPrice:          MustBigIntToUint128(swapAmountFromInput.NextSqrtPrice),
		TradingFee:             actualTradingFee.Uint64(),
		ProtocolFee:            actualProtocolFee.Uint64(),
		ReferralFee:            actualReferralFee.Uint64(),
	}, nil
}

// GetSwapResultFromPartialInput gets swap result from partial input.
func GetSwapResultFromPartialInput(
	virtualPool *dbc.VirtualPoolAccount,
	config *dbc.PoolConfigAccount,
	amountIn *big.Int,
	feeMode types.FeeMode,
	tradeDirection types.TradeDirection,
	currentPoint *big.Int,
) (dbc.SwapResult2, error) {

	actualProtocolFee, actualTradingFee, actualReferralFee :=
		big.NewInt(0), big.NewInt(0), big.NewInt(0)

	tradeFeeNumerator, err := GetTotalFeeNumeratorFromIncludedFeeAmount(
		config.PoolFees,
		virtualPool.VolatilityTracker,
		currentPoint,
		new(big.Int).SetUint64(virtualPool.ActivationPoint),
		amountIn,
		tradeDirection,
	)
	if err != nil {
		return dbc.SwapResult2{}, err
	}

	actualAmountIn := new(big.Int).Set(amountIn)
	if feeMode.FeesOnInput {
		feeResult, err := GetFeeOnAmount(
			tradeFeeNumerator,
			amountIn,
			config.PoolFees,
			feeMode.HasReferral,
		)
		if err != nil {
			return dbc.SwapResult2{}, err
		}

		actualProtocolFee = feeResult.ProtocolFee
		actualTradingFee = feeResult.TradingFee
		actualReferralFee = feeResult.ReferralFee

		actualAmountIn = feeResult.Amount
	}

	var swapAmountFromInput types.SwapAmount
	if tradeDirection == types.TradeDirectionBaseToQuote {
		if swapAmountFromInput, err = CalculateBaseToQuoteFromAmountIn(
			config.Curve[:],
			virtualPool.SqrtPrice.BigInt(),
			actualAmountIn,
		); err != nil {
			return dbc.SwapResult2{}, err
		}
	} else {
		if swapAmountFromInput, err = CalculateQuoteToBaseFromAmountIn(
			config.Curve[:],
			virtualPool.SqrtPrice.BigInt(),
			actualAmountIn,
			config.MigrationSqrtPrice.BigInt(),
		); err != nil {
			return dbc.SwapResult2{}, err
		}
	}

	var includedFeeInputAmount = amountIn
	if swapAmountFromInput.AmountLeft.Sign() != 0 {

		actualAmountIn = new(big.Int).Sub(
			actualAmountIn, swapAmountFromInput.AmountLeft,
		)
		if actualAmountIn.Sign() < 0 {
			return dbc.SwapResult2{},
				fmt.Errorf("GetSwapResultFromPartialInput:safeMath requires value not negative: value is %s", actualAmountIn)
		}

		if feeMode.FeesOnInput {
			tradeFeeNumeratorPartial, err := GetTotalFeeNumeratorFromExcludedFeeAmount(
				config.PoolFees,
				virtualPool.VolatilityTracker,
				currentPoint,
				new(big.Int).SetUint64(virtualPool.ActivationPoint),
				actualAmountIn,
				tradeDirection,
			)
			if err != nil {
				return dbc.SwapResult2{}, err
			}

			out, err := GetIncludedFeeAmount(
				tradeFeeNumeratorPartial, actualAmountIn,
			)
			if err != nil {
				return dbc.SwapResult2{}, err
			}

			out2, err := SplitFees(
				config.PoolFees,
				out.FeeAmount,
				feeMode.HasReferral,
			)
			if err != nil {
				return dbc.SwapResult2{}, err
			}

			actualProtocolFee = out2.ProtocolFee
			actualTradingFee = out2.TradingFee
			actualReferralFee = out2.ReferralFee

			includedFeeInputAmount = out.IncludedFeeAmount
		}
	}

	actualAmountOut := new(big.Int).Set(swapAmountFromInput.OutputAmount)

	if !feeMode.FeesOnInput {
		feeResult, err := GetFeeOnAmount(
			tradeFeeNumerator,
			swapAmountFromInput.OutputAmount,
			config.PoolFees,
			feeMode.HasReferral,
		)
		if err != nil {
			return dbc.SwapResult2{}, err
		}

		actualProtocolFee = feeResult.ProtocolFee
		actualTradingFee = feeResult.TradingFee
		actualReferralFee = feeResult.ReferralFee

		actualAmountOut = feeResult.Amount
	}

	if !swapAmountFromInput.AmountLeft.IsUint64() ||
		!includedFeeInputAmount.IsUint64() ||
		!actualAmountIn.IsUint64() ||
		!actualAmountOut.IsUint64() ||
		!actualTradingFee.IsUint64() ||
		!actualProtocolFee.IsUint64() ||
		!actualReferralFee.IsUint64() {
		return dbc.SwapResult2{},
			fmt.Errorf(
				"one of the values cannot fit into uint64: "+
					"AmountLeft(%s), IncludedFeeInputAmount(%s), ExcludedFeeInputAmount(%s), "+
					"OutputAmount(%s), TradingFee(%s), ProtocolFee(%s), ReferralFee(%s)",
				swapAmountFromInput.AmountLeft,
				includedFeeInputAmount,
				actualAmountIn,
				actualAmountOut,
				actualTradingFee,
				actualProtocolFee,
				actualReferralFee,
			)
	}

	return dbc.SwapResult2{
		AmountLeft:             swapAmountFromInput.AmountLeft.Uint64(),
		IncludedFeeInputAmount: includedFeeInputAmount.Uint64(),
		ExcludedFeeInputAmount: actualAmountIn.Uint64(),
		OutputAmount:           actualAmountOut.Uint64(),
		NextSqrtPrice:          MustBigIntToUint128(swapAmountFromInput.NextSqrtPrice),
		TradingFee:             actualTradingFee.Uint64(),
		ProtocolFee:            actualProtocolFee.Uint64(),
		ReferralFee:            actualReferralFee.Uint64(),
	}, nil
}

// CalculateBaseToQuoteFromAmountIn calculates output amount from base to quote from amount in.
func CalculateBaseToQuoteFromAmountIn(
	configStateCurve []dbc.LiquidityDistributionConfig,
	currentSqrtPrice, amountIn *big.Int,
) (out types.SwapAmount, err error) {

	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	if amountIn.Sign() == 0 {
		return types.SwapAmount{
			OutputAmount:  big.NewInt(0),
			NextSqrtPrice: currentSqrtPrice,
			AmountLeft:    big.NewInt(0),
		}, nil
	}

	totalOutputAmount, currentSqrtPriceLocal, amountLeft :=
		big.NewInt(0), new(big.Int).Set(currentSqrtPrice), new(big.Int).Set(amountIn)

		// Use curve.length for backward compatibility for existing pools with 20 points
	for i := len(configStateCurve) - 2; i >= 0; i-- {
		if configStateCurve[i].SqrtPrice.BigInt().Sign() == 0 ||
			configStateCurve[i].Liquidity.BigInt().Sign() == 0 {
			continue
		}

		if configStateCurve[i].SqrtPrice.BigInt().Cmp(currentSqrtPriceLocal) < 0 {
			maxAmountIn, err := GetDeltaAmountBaseUnsigned(
				configStateCurve[i].SqrtPrice.BigInt(),
				currentSqrtPriceLocal,
				configStateCurve[i+1].Liquidity.BigInt(),
				types.RoundingUp,
			)
			if err != nil {
				return types.SwapAmount{}, err
			}

			if amountLeft.Cmp(maxAmountIn) < 0 {
				nextSqrtPrice, err := GetNextSqrtPriceFromInput(
					currentSqrtPriceLocal,
					configStateCurve[i+1].Liquidity.BigInt(),
					amountLeft,
					true,
				)
				if err != nil {
					return types.SwapAmount{}, err
				}

				outputAmount, err := GetDeltaAmountQuoteUnsigned(
					nextSqrtPrice,
					currentSqrtPriceLocal,
					configStateCurve[i+1].Liquidity.BigInt(),
					types.RoundingDown,
				)
				if err != nil {
					return types.SwapAmount{}, err
				}

				totalOutputAmount = new(big.Int).Add(totalOutputAmount, outputAmount)
				currentSqrtPriceLocal = nextSqrtPrice
				amountLeft = big.NewInt(0)
				break
			}

			nextSqrtPrice := new(big.Int).Set(configStateCurve[i].SqrtPrice.BigInt())
			outputAmount, err := GetDeltaAmountQuoteUnsigned(
				nextSqrtPrice,
				currentSqrtPriceLocal,
				configStateCurve[i+1].Liquidity.BigInt(),
				types.RoundingDown,
			)
			if err != nil {
				return types.SwapAmount{}, err
			}

			totalOutputAmount = new(big.Int).Add(totalOutputAmount, outputAmount)
			currentSqrtPriceLocal = nextSqrtPrice
			amountLeft = new(big.Int).Sub(amountLeft, maxAmountIn)
		}
	}

	if amountLeft.Sign() != 0 {
		nextSqrtPrice, err := GetNextSqrtPriceFromInput(
			currentSqrtPriceLocal,
			configStateCurve[0].Liquidity.BigInt(),
			amountLeft,
			true,
		)
		if err != nil {
			return types.SwapAmount{}, err
		}
		outputAmount, err := GetDeltaAmountQuoteUnsigned(
			nextSqrtPrice,
			currentSqrtPriceLocal,
			configStateCurve[0].Liquidity.BigInt(),
			types.RoundingDown,
		)
		if err != nil {
			return types.SwapAmount{}, err
		}

		totalOutputAmount.Add(totalOutputAmount, outputAmount)
		currentSqrtPriceLocal = nextSqrtPrice
	}

	// no need to validate amount_left because if user sell more than what has in quote reserve,
	// then it will be failed when deduct pool.quote_reserve
	return types.SwapAmount{
		OutputAmount:  totalOutputAmount,
		NextSqrtPrice: currentSqrtPriceLocal,
		AmountLeft:    big.NewInt(0),
	}, nil
}

// CalculateQuoteToBaseFromAmountIn calculates output amount from quote to base from amount in.
func CalculateQuoteToBaseFromAmountIn(
	configStateCurve []dbc.LiquidityDistributionConfig,
	currentSqrtPrice, amountIn, stopSqrtPrice *big.Int,
) (types.SwapAmount, error) {

	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	if amountIn.Sign() == 0 {
		return types.SwapAmount{
			OutputAmount:  big.NewInt(0),
			NextSqrtPrice: currentSqrtPrice,
			AmountLeft:    big.NewInt(0),
		}, nil
	}

	totalOutputAmount, currentSqrtPriceLocal, amountLeft :=
		big.NewInt(0), new(big.Int).Set(currentSqrtPrice), new(big.Int).Set(amountIn)

		// Use curve.len() for backward compatibility for existing pools with 20 points
	for i := range len(configStateCurve) {
		if configStateCurve[i].SqrtPrice.BigInt().Sign() == 0 ||
			configStateCurve[i].Liquidity.BigInt().Sign() == 0 {
			break
		}

		referenceSqrtPrice := new(big.Int).Set(configStateCurve[i].SqrtPrice.BigInt())
		if stopSqrtPrice.Cmp(configStateCurve[i].SqrtPrice.BigInt()) < 0 {
			referenceSqrtPrice = new(big.Int).Set(stopSqrtPrice)
		}

		if referenceSqrtPrice.Cmp(currentSqrtPriceLocal) > 0 {
			maxAmountIn, err := GetDeltaAmountQuoteUnsigned(
				currentSqrtPriceLocal,
				referenceSqrtPrice,
				configStateCurve[i].Liquidity.BigInt(),
				types.RoundingUp,
			)
			if err != nil {
				return types.SwapAmount{}, err
			}

			if amountLeft.Cmp(maxAmountIn) < 0 {
				nextSqrtPrice, err := GetNextSqrtPriceFromInput(
					currentSqrtPriceLocal,
					configStateCurve[i].Liquidity.BigInt(),
					amountLeft,
					false,
				)
				if err != nil {
					return types.SwapAmount{}, err
				}

				outputAmount, err := GetDeltaAmountBaseUnsigned(
					currentSqrtPriceLocal,
					nextSqrtPrice,
					configStateCurve[i].Liquidity.BigInt(),
					types.RoundingDown,
				)

				if err != nil {
					return types.SwapAmount{}, err
				}

				totalOutputAmount = new(big.Int).Add(totalOutputAmount, outputAmount)
				currentSqrtPriceLocal = nextSqrtPrice
				amountLeft = big.NewInt(0)
				break
			}

			nextSqrtPrice := new(big.Int).Set(referenceSqrtPrice)
			outputAmount, err := GetDeltaAmountBaseUnsigned(
				currentSqrtPriceLocal,
				nextSqrtPrice,
				configStateCurve[i].Liquidity.BigInt(),
				types.RoundingDown,
			)
			if err != nil {
				return types.SwapAmount{}, err
			}

			totalOutputAmount = new(big.Int).Add(totalOutputAmount, outputAmount)
			currentSqrtPriceLocal = nextSqrtPrice
			amountLeft = new(big.Int).Sub(amountLeft, maxAmountIn)
			if amountLeft.Sign() < 0 {
				return types.SwapAmount{},
					fmt.Errorf("CalculateQuoteToBaseFromAmountIn:safeMath requires value not negative: value is %s", amountLeft)
			}

			if nextSqrtPrice.Cmp(stopSqrtPrice) == 0 {
				break
			}
		}
	}

	return types.SwapAmount{
		OutputAmount:  totalOutputAmount,
		NextSqrtPrice: currentSqrtPriceLocal,
		AmountLeft:    amountLeft,
	}, nil
}

func GetSwapResultFromExactOutput(
	virtualPool *dbc.VirtualPoolAccount,
	config *dbc.PoolConfigAccount,
	amountOut *big.Int,
	feeMode types.FeeMode,
	tradeDirection types.TradeDirection,
	currentPoint *big.Int,
) (dbc.SwapResult2, error) {

	actualProtocolFee, actualTradingFee, actualReferralFee :=
		big.NewInt(0), big.NewInt(0), big.NewInt(0)

	includedFeeOutAmount := new(big.Int).Set(amountOut)
	if !feeMode.FeesOnInput {
		tradeFeeNumerator, err := GetTotalFeeNumeratorFromExcludedFeeAmount(
			config.PoolFees,
			virtualPool.VolatilityTracker,
			currentPoint,
			new(big.Int).SetUint64(virtualPool.ActivationPoint),
			amountOut,
			tradeDirection,
		)
		if err != nil {
			return dbc.SwapResult2{}, err
		}

		out, err := GetIncludedFeeAmount(tradeFeeNumerator, amountOut)
		if err != nil {
			return dbc.SwapResult2{}, err
		}

		//   that ensure includedFeeOutAmount = amountOut + tradingFee + protocolFee + referralFee
		out2, err := SplitFees(
			config.PoolFees,
			out.FeeAmount,
			feeMode.HasReferral,
		)
		if err != nil {
			return dbc.SwapResult2{}, err
		}

		actualProtocolFee = out2.ProtocolFee
		actualTradingFee = out2.TradingFee
		actualReferralFee = out2.ReferralFee

		includedFeeOutAmount = out.IncludedFeeAmount
	}

	var (
		swapAmountFromOutput types.SwapAmount
		err                  error
	)
	if tradeDirection == types.TradeDirectionBaseToQuote {
		if swapAmountFromOutput, err = CalculateBaseToQuoteFromAmountOut(
			config,
			virtualPool.SqrtPrice.BigInt(),
			includedFeeOutAmount,
		); err != nil {
			return dbc.SwapResult2{}, err
		}
	} else {
		if swapAmountFromOutput, err = CalculateQuoteToBaseFromAmountOut(
			config,
			virtualPool.SqrtPrice.BigInt(),
			includedFeeOutAmount,
		); err != nil {
			return dbc.SwapResult2{}, err
		}
	}
	var (
		excludedFeeInputAmount = swapAmountFromOutput.OutputAmount
		includedFeeInputAmount = swapAmountFromOutput.OutputAmount
	)
	if feeMode.FeesOnInput {
		tradeFeeNumerator, err := GetTotalFeeNumeratorFromExcludedFeeAmount(
			config.PoolFees,
			virtualPool.VolatilityTracker,
			currentPoint,
			new(big.Int).SetUint64(virtualPool.ActivationPoint),
			swapAmountFromOutput.OutputAmount,
			tradeDirection,
		)
		if err != nil {
			return dbc.SwapResult2{}, err
		}

		out, err := GetIncludedFeeAmount(
			tradeFeeNumerator, swapAmountFromOutput.OutputAmount,
		)
		if err != nil {
			return dbc.SwapResult2{}, err
		}

		// that ensure includedFeeInAmount = excludedFeeInputAmount + tradingFee + protocolFee + referralFee
		out2, err := SplitFees(
			config.PoolFees,
			out.FeeAmount,
			feeMode.HasReferral,
		)
		if err != nil {
			return dbc.SwapResult2{}, err
		}

		actualProtocolFee = out2.ProtocolFee
		actualTradingFee = out2.TradingFee
		actualReferralFee = out2.ReferralFee

		// excludedFeeInputAmount = swapAmountFromOutput.OutputAmount
		includedFeeInputAmount = out.IncludedFeeAmount
	}

	if !includedFeeInputAmount.IsUint64() ||
		!excludedFeeInputAmount.IsUint64() ||
		!amountOut.IsUint64() ||
		!actualTradingFee.IsUint64() ||
		!actualProtocolFee.IsUint64() ||
		!actualReferralFee.IsUint64() {
		return dbc.SwapResult2{},
			fmt.Errorf(
				"one of the values cannot fit into uint64: "+
					"includedFeeInputAmount(%s), ExcludedFeeInputAmount(%s), "+
					"amount0ut(%s), TradingFee(%s), ProtocolFee(%s), ReferralFee(%s)",
				includedFeeInputAmount,
				excludedFeeInputAmount,
				amountOut,
				actualTradingFee,
				actualProtocolFee,
				actualReferralFee,
			)
	}

	return dbc.SwapResult2{
		AmountLeft:             0,
		IncludedFeeInputAmount: includedFeeInputAmount.Uint64(),
		ExcludedFeeInputAmount: excludedFeeInputAmount.Uint64(),
		OutputAmount:           amountOut.Uint64(),
		NextSqrtPrice:          MustBigIntToUint128(swapAmountFromOutput.NextSqrtPrice),
		TradingFee:             actualTradingFee.Uint64(),
		ProtocolFee:            actualProtocolFee.Uint64(),
		ReferralFee:            actualReferralFee.Uint64(),
	}, nil
}

// CalculateBaseToQuoteFromAmountOut calculates input amount from base to quote from amount out.
func CalculateBaseToQuoteFromAmountOut(
	configState *dbc.PoolConfigAccount,
	currentSqrtPrice, outAmount *big.Int,
) (types.SwapAmount, error) {

	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	totalAmountIn, currentSqrtPriceLocal, amountLeft :=
		big.NewInt(0), new(big.Int).Set(currentSqrtPrice), new(big.Int).Set(outAmount)

	configStateCurve := configState.Curve
	// Use curve.length for backward compatibility for existing pools with 20 points
	for i := len(configStateCurve) - 2; i >= 0; i-- {
		if configStateCurve[i].SqrtPrice.BigInt().Sign() == 0 ||
			configStateCurve[i].Liquidity.BigInt().Sign() == 0 {
			continue
		}

		if configStateCurve[i].SqrtPrice.BigInt().Cmp(currentSqrtPriceLocal) < 0 {
			maxAmountIn, err := GetDeltaAmountQuoteUnsigned(
				configStateCurve[i].SqrtPrice.BigInt(),
				currentSqrtPriceLocal,
				configStateCurve[i+1].Liquidity.BigInt(),
				types.RoundingDown,
			)
			if err != nil {
				return types.SwapAmount{}, err
			}

			if amountLeft.Cmp(maxAmountIn) < 0 {
				nextSqrtPrice, err := GetNextSqrtPriceFromOutput(
					currentSqrtPriceLocal,
					configStateCurve[i+1].Liquidity.BigInt(),
					amountLeft,
					true,
				)
				if err != nil {
					return types.SwapAmount{}, err
				}

				inAmount, err := GetDeltaAmountBaseUnsigned(
					nextSqrtPrice,
					currentSqrtPriceLocal,
					configStateCurve[i+1].Liquidity.BigInt(),
					types.RoundingUp,
				)
				if err != nil {
					return types.SwapAmount{}, err
				}

				totalAmountIn = new(big.Int).Add(totalAmountIn, inAmount)
				currentSqrtPriceLocal = nextSqrtPrice
				amountLeft = big.NewInt(0)
				break
			}

			nextSqrtPrice := new(big.Int).Set(configStateCurve[i].SqrtPrice.BigInt())
			inAmount, err := GetDeltaAmountBaseUnsigned(
				nextSqrtPrice,
				currentSqrtPriceLocal,
				configStateCurve[i+1].Liquidity.BigInt(),
				types.RoundingUp,
			)
			if err != nil {
				return types.SwapAmount{}, err
			}

			totalAmountIn = new(big.Int).Add(totalAmountIn, inAmount)
			currentSqrtPriceLocal = nextSqrtPrice
			amountLeft = new(big.Int).Sub(amountLeft, maxAmountIn)
		}
	}

	if amountLeft.Sign() != 0 {
		nextSqrtPrice, err := GetNextSqrtPriceFromOutput(
			currentSqrtPriceLocal,
			configStateCurve[0].Liquidity.BigInt(),
			amountLeft,
			true,
		)
		if err != nil {
			return types.SwapAmount{}, err
		}

		if nextSqrtPrice.Cmp(configState.SqrtStartPrice.BigInt()) < 0 {
			return types.SwapAmount{}, errors.New("CalculateBaseToQuoteFromAmountOut:not enough liquidity")
		}

		inAmount, err := GetDeltaAmountBaseUnsigned(
			nextSqrtPrice,
			currentSqrtPriceLocal,
			configStateCurve[0].Liquidity.BigInt(),
			types.RoundingUp,
		)
		if err != nil {
			return types.SwapAmount{}, err
		}

		totalAmountIn.Add(totalAmountIn, inAmount)
		currentSqrtPriceLocal = nextSqrtPrice
	}

	return types.SwapAmount{
		OutputAmount:  totalAmountIn,
		NextSqrtPrice: currentSqrtPriceLocal,
		AmountLeft:    big.NewInt(0),
	}, nil
}

// CalculateQuoteToBaseFromAmountOut calculates input amount from base to quote from amount out.
func CalculateQuoteToBaseFromAmountOut(
	configState *dbc.PoolConfigAccount,
	currentSqrtPrice, outAmount *big.Int,
) (out types.SwapAmount, err error) {

	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	totalAmountIn, currentSqrtPriceLocal, amountLeft :=
		big.NewInt(0), new(big.Int).Set(currentSqrtPrice), new(big.Int).Set(outAmount)

	configStateCurve := configState.Curve
	// iterate through curve points
	for i := range configStateCurve {
		if configStateCurve[i].SqrtPrice.BigInt().Sign() == 0 ||
			configStateCurve[i].Liquidity.BigInt().Sign() == 0 {
			break
		}

		if configStateCurve[i].SqrtPrice.BigInt().Cmp(currentSqrtPriceLocal) > 0 {
			maxAmountOut, err := GetDeltaAmountBaseUnsigned(
				currentSqrtPriceLocal,
				configStateCurve[i].SqrtPrice.BigInt(),
				configStateCurve[i].Liquidity.BigInt(),
				types.RoundingDown,
			)
			if err != nil {
				return types.SwapAmount{}, err
			}

			if amountLeft.Cmp(maxAmountOut) < 0 {
				nextSqrtPrice, err := GetNextSqrtPriceFromOutput(
					currentSqrtPriceLocal,
					configStateCurve[i].Liquidity.BigInt(),
					amountLeft,
					false,
				)
				if err != nil {
					return types.SwapAmount{}, err
				}

				inAmount, err := GetDeltaAmountQuoteUnsigned(
					currentSqrtPriceLocal,
					nextSqrtPrice,
					configStateCurve[i].Liquidity.BigInt(),
					types.RoundingUp,
				)
				if err != nil {
					return types.SwapAmount{}, err
				}

				totalAmountIn = new(big.Int).Add(totalAmountIn, inAmount)
				currentSqrtPriceLocal = new(big.Int).Set(nextSqrtPrice)
				amountLeft = big.NewInt(0)
				break
			}

			nextSqrtPrice := new(big.Int).Set(configStateCurve[i].SqrtPrice.BigInt())
			inAmount, err := GetDeltaAmountQuoteUnsigned(
				currentSqrtPriceLocal,
				nextSqrtPrice,
				configStateCurve[i].Liquidity.BigInt(),
				types.RoundingUp,
			)
			if err != nil {
				return types.SwapAmount{}, err
			}

			totalAmountIn = new(big.Int).Add(totalAmountIn, inAmount)
			currentSqrtPriceLocal = new(big.Int).Set(nextSqrtPrice)
			amountLeft = new(big.Int).Sub(amountLeft, maxAmountOut)
		}
	}

	if amountLeft.Sign() != 0 {
		return types.SwapAmount{}, errors.New("CalculateQuoteToBaseFromAmountOut:not enough liquidity")
	}

	return types.SwapAmount{
		OutputAmount:  totalAmountIn,
		NextSqrtPrice: currentSqrtPriceLocal,
		AmountLeft:    big.NewInt(0),
	}, nil
}

// SwapQuoteExactIn calculates quote for a swap with exact input amount.
func SwapQuoteExactIn(
	virtualPool *dbc.VirtualPoolAccount,
	config *dbc.PoolConfigAccount,
	swapBaseForQuote bool,
	amountIn *big.Int,
	slippageBps uint64,
	hasReferral bool,
	currentPoint *big.Int,
) (types.SwapQuote2Result, error) {

	if virtualPool.QuoteReserve >= config.MigrationQuoteThreshold {
		return types.SwapQuote2Result{}, errors.New("virtual pool is completed")
	}

	if amountIn.Sign() == 0 {
		return types.SwapQuote2Result{}, errors.New("amountIn cannot be zero")
	}

	tradeDirection := types.TradeDirectionQuoteToBase
	if swapBaseForQuote {
		tradeDirection = types.TradeDirectionBaseToQuote
	}

	feeMode := GetFeeMode(
		types.CollectFeeMode(config.CollectFeeMode),
		tradeDirection,
		hasReferral,
	)

	result, err := GetSwapResultFromExactInput(
		virtualPool,
		config,
		amountIn,
		feeMode,
		tradeDirection,
		currentPoint,
	)
	if err != nil {
		return types.SwapQuote2Result{}, err
	}

	// check amount left threshold for exact in
	maxSwallowQuoteAmount := GetMaxSwallowQuoteAmount(config)
	if new(big.Int).SetUint64(result.AmountLeft).Cmp(maxSwallowQuoteAmount) > 0 {
		return types.SwapQuote2Result{}, fmt.Errorf(
			"amountLeft(%d) cannot be over maxSwallowQuoteAmount(%s)",
			result.AmountLeft, maxSwallowQuoteAmount,
		)
	}

	minimumAmountOut := result.OutputAmount
	if slippageBps > 0 {
		// slippage factor: (10000 + slippageBps) / 10000
		slippageFactor, denominator :=
			new(big.Int).SetUint64(10_000-slippageBps), big.NewInt(10_000)

			// minimum amount out: amountOut * (10000 - slippageBps) / 10000
		minimumAmountOutBigInt := new(big.Int).Div(
			new(big.Int).Mul(new(big.Int).SetUint64(result.OutputAmount), slippageFactor),
			denominator,
		)

		if !minimumAmountOutBigInt.IsUint64() {
			return types.SwapQuote2Result{}, fmt.Errorf("cannot fit minimumAmountOutBigInt(%s) into uint64", minimumAmountOutBigInt)
		}

		minimumAmountOut = minimumAmountOutBigInt.Uint64()
	}

	return types.SwapQuote2Result{
		SwapResult2:      result,
		MinimumAmountOut: minimumAmountOut,
	}, nil
}

// SwapQuotePartialFill calculates quote for a swap with partial fill.
func SwapQuotePartialFill(
	virtualPool *dbc.VirtualPoolAccount,
	config *dbc.PoolConfigAccount,
	swapBaseForQuote bool,
	amountIn *big.Int,
	slippageBps uint64,
	hasReferral bool,
	currentPoint *big.Int,
) (types.SwapQuote2Result, error) {
	if virtualPool == nil || config == nil {
		return types.SwapQuote2Result{},
			fmt.Errorf("SwapQuotePartialFill: nil value:virtualPool(%+v) or config(%+v)",
				virtualPool, config)
	}

	if virtualPool.QuoteReserve >= config.MigrationQuoteThreshold {
		return types.SwapQuote2Result{}, errors.New("virtual pool is completed")
	}

	if amountIn.Sign() == 0 {
		return types.SwapQuote2Result{}, errors.New("amountIn cannot be zero")
	}

	tradeDirection := types.TradeDirectionQuoteToBase
	if swapBaseForQuote {
		tradeDirection = types.TradeDirectionBaseToQuote
	}

	feeMode := GetFeeMode(
		types.CollectFeeMode(config.CollectFeeMode),
		tradeDirection,
		hasReferral,
	)

	result, err := GetSwapResultFromPartialInput(
		virtualPool,
		config,
		amountIn,
		feeMode,
		tradeDirection,
		currentPoint,
	)
	if err != nil {
		return types.SwapQuote2Result{}, err
	}

	// calculate minimum amount out
	minimumAmountOut := result.OutputAmount
	if slippageBps > 0 {
		// slippage factor: (10000 + slippageBps) / 10000
		slippageFactor, denominator :=
			new(big.Int).SetUint64(10_000-slippageBps), big.NewInt(10_000)

			// minimum amount out: amountOut * (10000 - slippageBps) / 10000
		minimumAmountOutBigInt := new(big.Int).Div(
			new(big.Int).Mul(new(big.Int).SetUint64(result.OutputAmount), slippageFactor),
			denominator,
		)

		if !minimumAmountOutBigInt.IsUint64() {
			return types.SwapQuote2Result{}, fmt.Errorf("cannot fit minimumAmountOutBigInt(%s) into uint64", minimumAmountOutBigInt)
		}

		minimumAmountOut = minimumAmountOutBigInt.Uint64()
	}

	return types.SwapQuote2Result{
		SwapResult2:      result,
		MinimumAmountOut: minimumAmountOut,
	}, nil
}

// SwapQuoteExactOut calculates quote for a swap with exact output amount.
func SwapQuoteExactOut(
	virtualPool *dbc.VirtualPoolAccount,
	config *dbc.PoolConfigAccount,
	swapBaseForQuote bool,
	amountIn *big.Int,
	slippageBps uint64,
	hasReferral bool,
	currentPoint *big.Int,
) (types.SwapQuote2Result, error) {

	if virtualPool.QuoteReserve >= config.MigrationQuoteThreshold {
		return types.SwapQuote2Result{}, errors.New("virtual pool is completed")
	}

	if amountIn.Sign() == 0 {
		return types.SwapQuote2Result{}, errors.New("amountIn cannot be zero")
	}

	tradeDirection := types.TradeDirectionQuoteToBase
	if swapBaseForQuote {
		tradeDirection = types.TradeDirectionBaseToQuote
	}

	feeMode := GetFeeMode(
		types.CollectFeeMode(config.CollectFeeMode),
		tradeDirection,
		hasReferral,
	)

	result, err := GetSwapResultFromExactOutput(
		virtualPool,
		config,
		amountIn,
		feeMode,
		tradeDirection,
		currentPoint,
	)
	if err != nil {
		return types.SwapQuote2Result{}, err
	}

	// calculate maximum amount in (for slippage protection)
	maximumAmountIn := result.IncludedFeeInputAmount
	if slippageBps > 0 {
		// slippage factor: (10000 + slippageBps) / 10000
		slippageFactor, denominator :=
			new(big.Int).SetUint64(10_000+slippageBps), big.NewInt(10_000)

			// minimum amount out: amountOut * (10000 - slippageBps) / 10000
		maximumAmountInBigInt := new(big.Int).Div(
			new(big.Int).Mul(new(big.Int).SetUint64(result.IncludedFeeInputAmount), slippageFactor),
			denominator,
		)

		if !maximumAmountInBigInt.IsUint64() {
			return types.SwapQuote2Result{}, fmt.Errorf("cannot fit maximumAmountInBigInt(%s) into uint64", maximumAmountInBigInt)
		}

		maximumAmountIn = maximumAmountInBigInt.Uint64()
	}

	return types.SwapQuote2Result{
		SwapResult2:     result,
		MaximumAmountIn: maximumAmountIn,
	}, nil
}
