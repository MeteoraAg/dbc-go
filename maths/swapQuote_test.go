package maths_test

// import (
// 	"dbcGoSDK/generated/dbc"
// 	"dbcGoSDK/helpers"
// 	"dbcGoSDK/maths"
// 	"dbcGoSDK/types"
// 	"math/big"

// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func TestFeeMode(t *testing.T) {
// 	tests := []struct {
// 		Name  string
// 		Param struct {
// 			collectFeeMode types.CollectFeeMode
// 			tradeDirection types.TradeDirection
// 			hasReferral    bool
// 		}
// 		expected types.FeeMode
// 	}{
// 		{
// 			Name: "fee mode output token base to quote",
// 			Param: struct {
// 				collectFeeMode types.CollectFeeMode
// 				tradeDirection types.TradeDirection
// 				hasReferral    bool
// 			}{
// 				collectFeeMode: types.CollectFeeModeOutputToken,
// 				tradeDirection: types.TradeDirectionBaseToQuote,
// 			},
// 			expected: types.FeeMode{},
// 		},
// 		{
// 			Name: "fee mode output token quote to base",
// 			Param: struct {
// 				collectFeeMode types.CollectFeeMode
// 				tradeDirection types.TradeDirection
// 				hasReferral    bool
// 			}{
// 				collectFeeMode: types.CollectFeeModeOutputToken,
// 				tradeDirection: types.TradeDirectionQuoteToBase,
// 				hasReferral:    true,
// 			},
// 			expected: types.FeeMode{
// 				FeesOnBaseToken: true,
// 				HasReferral:     true,
// 			},
// 		},
// 		{
// 			Name: "fee mode quote token base to quote",
// 			Param: struct {
// 				collectFeeMode types.CollectFeeMode
// 				tradeDirection types.TradeDirection
// 				hasReferral    bool
// 			}{
// 				collectFeeMode: types.CollectFeeModeQuoteToken,
// 				tradeDirection: types.TradeDirectionBaseToQuote,
// 			},
// 			expected: types.FeeMode{},
// 		},
// 		{
// 			Name: "fee mode quote token quote to base",
// 			Param: struct {
// 				collectFeeMode types.CollectFeeMode
// 				tradeDirection types.TradeDirection
// 				hasReferral    bool
// 			}{
// 				collectFeeMode: types.CollectFeeModeQuoteToken,
// 				tradeDirection: types.TradeDirectionQuoteToBase,
// 				hasReferral:    true,
// 			},
// 			expected: types.FeeMode{
// 				FeesOnInput:  true,
// 				HasReferral: true,
// 			},
// 		},
// 		{
// 			// Test default values by passing default collect fee mode
// 			Name: "fee mode default values",
// 			Param: struct {
// 				collectFeeMode types.CollectFeeMode
// 				tradeDirection types.TradeDirection
// 				hasReferral    bool
// 			}{
// 				collectFeeMode: types.CollectFeeModeQuoteToken,
// 				tradeDirection: types.TradeDirectionBaseToQuote,
// 			},
// 			expected: types.FeeMode{},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.Name, func(t *testing.T) {
// 			got := maths.GetFeeMode(
// 				tt.Param.collectFeeMode,
// 				tt.Param.tradeDirection,
// 				tt.Param.hasReferral,
// 			)

// 			assert.Equal(t, tt.expected, got)
// 		})

// 	}

// 	t.Run("fee mode properties", func(t *testing.T) {
// 		got := maths.GetFeeMode(
// 			types.CollectFeeModeQuoteToken,
// 			types.TradeDirectionBaseToQuote,
// 			true,
// 		)

// 		assert.Equal(t, false, got.FeesOnInput)
// 	})

// 	t.Run("fee mode properties", func(t *testing.T) {
// 		t.Run(" trading BaseToQuote, fees should never be on input", func(t *testing.T) {
// 			got := maths.GetFeeMode(
// 				types.CollectFeeModeQuoteToken,
// 				types.TradeDirectionBaseToQuote,
// 				true,
// 			)

// 			assert.Equal(t, false, got.FeesOnInput)
// 		})
// 		t.Run(" QuoteToken mode, base_token should always be false", func(t *testing.T) {
// 			got := maths.GetFeeMode(
// 				types.CollectFeeModeQuoteToken,
// 				types.TradeDirectionQuoteToBase,
// 				false,
// 			)

// 			assert.Equal(t, false, got.FeesOnBaseToken)
// 		})
// 	})
// }

// func TestSwapQuote(t *testing.T) {
// 	sqrtStartPrice := maths.Q64(1.0)
// 	t.Run("getSwapAmountFromBaseToQuote zero amount", func(t *testing.T) {

// 		// Test with zero amount
// 		result, err := maths.GetSwapAmountFromBaseToQuote(
// 			[]dbc.LiquidityDistributionConfig{ // Create a simple config with one curve point
// 				{
// 					SqrtPrice: helpers.MustBigIntToUint128(sqrtStartPrice),
// 					Liquidity: helpers.MustBigIntToUint128(big.NewInt(1_000_000_000)),
// 				},
// 			},
// 			sqrtStartPrice,
// 			big.NewInt(0),
// 		)
// 		if err != nil {
// 			t.Fatalf("GetSwapAmountFromBaseToQuote errored: %s", err.Error())
// 		}

// 		assert.True(t, result.OutputAmount.Sign() == 0)
// 		assert.True(t, result.NextSqrtPrice.Cmp(sqrtStartPrice) == 0)
// 	})

// 	t.Run("getSwapAmountFromQuoteToBase not enough liquidity", func(t *testing.T) {

// 		amountIn, _ := new(big.Int).SetString("10000000000000000000000", 10)

// 		// Test with extremely large amount that exceeds available liquidity
// 		_, err := maths.GetSwapAmountFromQuoteToBase(
// 			[]dbc.LiquidityDistributionConfig{ // Create a simple config with one curve point
// 				{
// 					SqrtPrice: helpers.MustBigIntToUint128(
// 						new(big.Int).Mul(sqrtStartPrice, big.NewInt(2)),
// 					),
// 					Liquidity: helpers.MustBigIntToUint128(big.NewInt(1_000_000_000)),
// 				},
// 			},
// 			sqrtStartPrice,
// 			amountIn,
// 		)

// 		assert.NotNil(t, err)
// 		assert.Contains(t, "not enough liquidity to process the entire amount", err.Error())
// 	})
// }
