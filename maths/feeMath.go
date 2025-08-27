package maths

import (
	"dbcGoSDK/constants"
	"dbcGoSDK/generated/dbc"
	mathsPoolfees "dbcGoSDK/maths/poolFees"
	"dbcGoSDK/types"
	"fmt"
	"math/big"
)

// GetMaxSwallowQuoteAmount gets maximum swallow quote amount.
func GetMaxSwallowQuoteAmount(config *dbc.PoolConfigAccount) *big.Int {
	maxSwallowAmount, _ := mathsPoolfees.MulDiv(
		new(big.Int).SetUint64(config.MigrationQuoteThreshold),
		big.NewInt(constants.MaxSwallowPercentage),
		big.NewInt(100),
		types.RoundingDown,
	)
	return maxSwallowAmount
}

// GetFeeMode gets fee mode.
func GetFeeMode(
	collectFeeMode types.CollectFeeMode,
	tradeDirection types.TradeDirection,
	hasReferral bool,
) types.FeeMode {
	// (CollectFeeMode::OutputToken, TradeDirection::BaseToQuote) => (false, false),
	// (CollectFeeMode::OutputToken, TradeDirection::QuoteToBase) => (false, true),
	// (CollectFeeMode::QuoteToken, TradeDirection::BaseToQuote) => (false, false),
	// (CollectFeeMode::QuoteToken, TradeDirection::QuoteToBase) => (true, false),

	if collectFeeMode == types.CollectFeeModeOutputToken {
		if tradeDirection == types.TradeDirectionBaseToQuote {
			return types.FeeMode{HasReferral: hasReferral}
		}

		// TradeDirection.QuoteToBase
		return types.FeeMode{
			FeesOnBaseToken: true,
			HasReferral:     hasReferral,
		}
	}

	if tradeDirection == types.TradeDirectionBaseToQuote {
		return types.FeeMode{HasReferral: hasReferral}
	}

	return types.FeeMode{
		FeesOnInput: true,
		HasReferral: hasReferral,
	}
}

// GetTotalFeeNumeratorFromIncludedFeeAmount gets total fee numerator from included fee amount.
func GetTotalFeeNumeratorFromIncludedFeeAmount(
	poolFees dbc.PoolFeesConfig,
	volatilityTracker dbc.VolatilityTracker,
	currentPoint, activationPoint, includedFeeAmount *big.Int,
	tradeDirection types.TradeDirection,
) (*big.Int, error) {
	baseFeeHandler, err := mathsPoolfees.GetBaseFeeHandler(
		new(big.Int).SetUint64(poolFees.BaseFee.CliffFeeNumerator),
		poolFees.BaseFee.FirstFactor,
		new(big.Int).SetUint64(poolFees.BaseFee.SecondFactor),
		new(big.Int).SetUint64(poolFees.BaseFee.ThirdFactor),
		types.BaseFeeMode(poolFees.BaseFee.BaseFeeMode),
	)
	if err != nil {
		return nil, err
	}

	baseFeeNumerator, err := baseFeeHandler.GetBaseFeeNumeratorFromIncludedFeeAmount(
		currentPoint,
		activationPoint,
		tradeDirection,
		includedFeeAmount,
	)
	if err != nil {
		return nil, err
	}

	return GetTotalFeeNumerator(
		baseFeeNumerator,
		poolFees.DynamicFee,
		volatilityTracker,
	), nil
}

// GetTotalFeeNumeratorFromExcludedFeeAmount gets total fee numerator from excluded fee amount.
func GetTotalFeeNumeratorFromExcludedFeeAmount(
	poolFees dbc.PoolFeesConfig,
	volatilityTracker dbc.VolatilityTracker,
	currentPoint, activationPoint, excludedFeeAmount *big.Int,
	tradeDirection types.TradeDirection,
) (*big.Int, error) {
	baseFeeHandler, err := mathsPoolfees.GetBaseFeeHandler(
		new(big.Int).SetUint64(poolFees.BaseFee.CliffFeeNumerator),
		poolFees.BaseFee.FirstFactor,
		new(big.Int).SetUint64(poolFees.BaseFee.SecondFactor),
		new(big.Int).SetUint64(poolFees.BaseFee.ThirdFactor),
		types.BaseFeeMode(poolFees.BaseFee.BaseFeeMode),
	)
	if err != nil {
		return nil, err
	}

	baseFeeNumerator, err := baseFeeHandler.GetBaseFeeNumeratorFromExcludedFeeAmount(
		currentPoint,
		activationPoint,
		tradeDirection,
		excludedFeeAmount,
	)
	if err != nil {
		return nil, err
	}

	return GetTotalFeeNumerator(
		baseFeeNumerator,
		poolFees.DynamicFee,
		volatilityTracker,
	), nil
}

// GetTotalFeeNumerator gets total fee numerator from excluded fee amount.
func GetTotalFeeNumerator(
	baseFeeNumerator *big.Int,
	dynamicFee dbc.DynamicFeeConfig,
	volatilityTracker dbc.VolatilityTracker,
) *big.Int {
	variableFeeNumerator := mathsPoolfees.GetVariableFeeNumerator(
		dynamicFee,
		volatilityTracker,
	)

	totalFeeNumerator := new(big.Int).Add(variableFeeNumerator, baseFeeNumerator)

	// Cap the total fee at MAX_FEE_NUMERATOR
	var cappedTotalFeeNumerator = totalFeeNumerator
	if maxFeeNumeratorBN := big.NewInt(constants.MaxFeeNumerator); totalFeeNumerator.Cmp(maxFeeNumeratorBN) > 0 {
		cappedTotalFeeNumerator = maxFeeNumeratorBN
	}

	return cappedTotalFeeNumerator
}

// GetFeeOnAmount gets fee on amount with trade fee numerator.
func GetFeeOnAmount(
	tradeFeeNumerator, amount *big.Int,
	poolFees dbc.PoolFeesConfig,
	hasReferral bool,
) (types.FeeOnAmountResult, error) {
	out, err := GetExcludedFeeAmount(tradeFeeNumerator, amount)
	if err != nil {
		return types.FeeOnAmountResult{}, nil
	}

	protocolFee, _ := mathsPoolfees.MulDiv(
		out.TradingFee,
		new(big.Int).SetUint64(uint64(poolFees.ProtocolFeePercent)),
		big.NewInt(100),
		types.RoundingDown,
	)

	updatedTradingFee := new(big.Int).Sub(out.TradingFee, protocolFee)
	if updatedTradingFee.Sign() < 0 {
		return types.FeeOnAmountResult{}, fmt.Errorf("GetFeeOnAmount:safeMath requires value not negative: value is %s", updatedTradingFee)
	}

	referralFee := big.NewInt(0)
	if hasReferral {
		referralFee, _ = mathsPoolfees.MulDiv(
			protocolFee,
			new(big.Int).SetUint64(uint64(poolFees.ReferralFeePercent)),
			big.NewInt(100),
			types.RoundingDown,
		)
	}

	updatedProtocolFee := new(big.Int).Sub(protocolFee, referralFee)
	if updatedProtocolFee.Sign() < 0 {
		return types.FeeOnAmountResult{}, fmt.Errorf("GetFeeOnAmount:safeMath requires value not negative: value is %s", updatedProtocolFee)
	}

	return types.FeeOnAmountResult{
		Amount:      out.ExcludedFeeAmount,
		ProtocolFee: protocolFee,
		ReferralFee: referralFee,
		TradingFee:  updatedTradingFee,
	}, nil
}

// GetExcludedFeeAmount gets excluded fee amount from included fee amount.
func GetExcludedFeeAmount(
	tradeFeeNumerator, includedFeeAmount *big.Int,
) (struct{ ExcludedFeeAmount, TradingFee *big.Int }, error) {

	tradingFee, _ := mathsPoolfees.MulDiv(
		includedFeeAmount,
		tradeFeeNumerator,
		new(big.Int).SetUint64(constants.FeeDenominator),
		types.RoundingUp,
	)

	// update amount
	excludedFeeAmount := new(big.Int).Sub(includedFeeAmount, tradingFee)
	if excludedFeeAmount.Sign() < 0 {
		return struct {
			ExcludedFeeAmount *big.Int
			TradingFee        *big.Int
		}{}, fmt.Errorf("GetExcludedFeeAmount:safeMath requires value not negative: value is %s", excludedFeeAmount.String())
	}

	return struct {
		ExcludedFeeAmount *big.Int
		TradingFee        *big.Int
	}{
		ExcludedFeeAmount: excludedFeeAmount,
		TradingFee:        tradingFee,
	}, nil
}

// GetIncludedFeeAmount gets included fee amount from excluded fee amount.
func GetIncludedFeeAmount(
	tradeFeeNumerator, excludedFeeAmount *big.Int,
) (struct{ IncludedFeeAmount, FeeAmount *big.Int }, error) {

	includedFeeAmount, err := mathsPoolfees.MulDiv(
		excludedFeeAmount,
		constants.FeeDenominatorBigInt,
		new(big.Int).Sub(constants.FeeDenominatorBigInt, tradeFeeNumerator),
		types.RoundingUp,
	)
	if err != nil {
		return struct {
			IncludedFeeAmount *big.Int
			FeeAmount         *big.Int
		}{}, err
	}
	feeAmount := new(big.Int).Sub(includedFeeAmount, excludedFeeAmount)
	if feeAmount.Sign() < 0 {
		return struct{ IncludedFeeAmount, FeeAmount *big.Int }{},
			fmt.Errorf("GetIncludedFeeAmount:safeMath requires value not negative: value is %s", feeAmount)
	}

	return struct {
		IncludedFeeAmount *big.Int
		FeeAmount         *big.Int
	}{
		IncludedFeeAmount: includedFeeAmount,
		FeeAmount:         feeAmount,
	}, nil
}

// SplitFees splits fees into trading, protocol, and referral fees.
func SplitFees(
	poolFees dbc.PoolFeesConfig,
	feeAmount *big.Int,
	hasReferral bool,
) (struct{ TradingFee, ProtocolFee, ReferralFee *big.Int }, error) {
	protocolFee, _ := mathsPoolfees.MulDiv(
		feeAmount,
		new(big.Int).SetUint64(uint64(poolFees.ProtocolFeePercent)),
		big.NewInt(100),
		types.RoundingDown,
	)

	// update trading fee
	tradingFee := new(big.Int).Sub(feeAmount, protocolFee)
	if tradingFee.Sign() < 0 {
		return struct{ TradingFee, ProtocolFee, ReferralFee *big.Int }{},
			fmt.Errorf("SplitFees:safeMath requires value not negative: value is %s", tradingFee)
	}

	referralFee := big.NewInt(0)
	if hasReferral {
		referralFee, _ = mathsPoolfees.MulDiv(
			protocolFee,
			new(big.Int).SetUint64(uint64(poolFees.ProtocolFeePercent)),
			big.NewInt(100),
			types.RoundingDown,
		)
	}

	protocolFeeAfterReferral := new(big.Int).Sub(protocolFee, referralFee)
	if protocolFeeAfterReferral.Sign() < 0 {
		return struct{ TradingFee, ProtocolFee, ReferralFee *big.Int }{},
			fmt.Errorf("SplitFees:safeMath requires value not negative: value is %s", protocolFeeAfterReferral)
	}

	return struct {
		TradingFee  *big.Int
		ProtocolFee *big.Int
		ReferralFee *big.Int
	}{
		TradingFee:  tradingFee,
		ProtocolFee: protocolFeeAfterReferral,
		ReferralFee: referralFee,
	}, nil

}
