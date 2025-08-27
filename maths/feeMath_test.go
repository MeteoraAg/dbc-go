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

// func TestFeeMath(t *testing.T) {
// 	t.Run("getFeeInPeriod calculation", func(t *testing.T) {
// 		t.Run("Test case 1: No reduction", func(t *testing.T) {
// 			got, err := maths.GetFeeNumeratorOnExponentialFeeScheduler(
// 				big.NewInt(1_000),
// 				big.NewInt(0),
// 				0,
// 			)
// 			if err != nil {
// 				t.Fatalf("GetFeeNumeratorOnExponentialFeeScheduler errored: %s", err.Error())
// 			}

// 			assert.True(t, got.Cmp(big.NewInt(1000)) == 0)
// 		})

// 		t.Run("Test case 2: With reduction", func(t *testing.T) {
// 			got, err := maths.GetFeeNumeratorOnExponentialFeeScheduler(
// 				big.NewInt(1_000),
// 				big.NewInt(100),
// 				1,
// 			)
// 			if err != nil {
// 				t.Fatalf("GetFeeNumeratorOnExponentialFeeScheduler errored: %s", err.Error())
// 			}

// 			assert.True(t, got.Cmp(big.NewInt(989)) > 0)
// 			assert.True(t, got.Cmp(big.NewInt(991)) < 0)
// 		})
// 	})

// 	t.Run("getFeeInPeriod with higher periods", func(t *testing.T) {

// 		got, err := maths.GetFeeNumeratorOnExponentialFeeScheduler(
// 			big.NewInt(1_000),
// 			big.NewInt(100),
// 			5,
// 		)
// 		if err != nil {
// 			t.Fatalf("GetFeeNumeratorOnExponentialFeeScheduler errored: %s", err.Error())
// 		}

// 		assert.True(t, got.Cmp(big.NewInt(1_000)) < 0)
// 		assert.True(t, got.Sign() >= 0)
// 	})

// 	t.Run("getBaseFeeNumerator with linear mode", func(t *testing.T) {
// 		baseFee := dbc.BaseFeeConfig{
// 			CliffFeeNumerator: 1_000,
// 			BaseFeeMode:       uint8(types.BaseFeeModeFeeSchedulerLinear),
// 			FirstFactor:       10,
// 			SecondFactor:      50, // 50 per period
// 			ThirdFactor:       100,
// 		}
// 		t.Run("before activation point", func(t *testing.T) {
// 			got, err := maths.GetBaseFeeNumerator(
// 				baseFee,
// 				types.TradeDirectionQuoteToBase,
// 				big.NewInt(50),
// 				big.NewInt(100),
// 				big.NewInt(0),
// 			)
// 			if err != nil {
// 				t.Fatalf("GetBaseFeeNumerator errored: %s", err.Error())
// 			}

// 			// Use max period (min fee)
// 			assert.True(t, got.Sign() == 0)
// 		})
// 		t.Run("after activation point, 2 periods elapsed", func(t *testing.T) {
// 			got, err := maths.GetBaseFeeNumerator(
// 				baseFee,
// 				types.TradeDirectionQuoteToBase,
// 				big.NewInt(300),
// 				big.NewInt(100),
// 				big.NewInt(0),
// 			)
// 			if err != nil {
// 				t.Fatalf("GetBaseFeeNumerator errored: %s", err.Error())
// 			}

// 			// Use max period (min fee)
// 			assert.True(t, got.Cmp(big.NewInt(600)) == 0)
// 		})
// 	})
// 	t.Run("getBaseFeeNumerator with exponential mode", func(t *testing.T) {
// 		baseFee := dbc.BaseFeeConfig{
// 			CliffFeeNumerator: 1_000,
// 			BaseFeeMode:       uint8(types.BaseFeeModeFeeSchedulerExponential),
// 			FirstFactor:       5,
// 			SecondFactor:      100, // 50 per period
// 			ThirdFactor:       100,
// 		}
// 		t.Run("before activation point", func(t *testing.T) {
// 			got, err := maths.GetBaseFeeNumerator(
// 				baseFee,
// 				types.TradeDirectionQuoteToBase,
// 				big.NewInt(350),
// 				big.NewInt(100),
// 				big.NewInt(0),
// 			)
// 			if err != nil {
// 				t.Fatalf("GetBaseFeeNumerator errored: %s", err.Error())
// 			}

// 			// Use exponential reduction
// 			assert.True(t, got.Cmp(big.NewInt(1_000)) < 0)
// 			assert.True(t, got.Cmp(big.NewInt(950)) > 0)
// 		})
// 	})

// 	t.Run("getVariableFee calculation", func(t *testing.T) {
// 		t.Run("with non-zero volatility", func(t *testing.T) {
// 			got := maths.GetVariableFee(
// 				dbc.DynamicFeeConfig{
// 					Initialized:              1,
// 					MaxVolatilityAccumulator: 1_000,
// 					VariableFeeControl:       10,
// 					BinStep:                  100,
// 					BinStepU128:              helpers.MustBigIntToUint128(big.NewInt(100)),
// 				},
// 				dbc.VolatilityTracker{
// 					VolatilityAccumulator: helpers.MustBigIntToUint128(big.NewInt(1_000)),
// 				},
// 			)

// 			// Return a non-zero fee
// 			assert.True(t, got.Sign() > 0)
// 		})
// 		t.Run("with zero volatility", func(t *testing.T) {
// 			got := maths.GetVariableFee(
// 				dbc.DynamicFeeConfig{
// 					Initialized:              1,
// 					MaxVolatilityAccumulator: 1_000,
// 					VariableFeeControl:       10,
// 					BinStep:                  100,
// 					BinStepU128:              helpers.MustBigIntToUint128(big.NewInt(100)),
// 				},
// 				dbc.VolatilityTracker{},
// 			)

// 			// Return zero fee
// 			assert.True(t, got.Sign() == 0)
// 		})

// 		t.Run("with uninitialized dynamic fee", func(t *testing.T) {
// 			got := maths.GetVariableFee(
// 				dbc.DynamicFeeConfig{
// 					Initialized:              0, // disabled
// 					MaxVolatilityAccumulator: 1_000,
// 					VariableFeeControl:       10,
// 					BinStep:                  100,
// 					BinStepU128:              helpers.MustBigIntToUint128(big.NewInt(100)),
// 				},
// 				dbc.VolatilityTracker{
// 					VolatilityAccumulator: helpers.MustBigIntToUint128(big.NewInt(1_000)),
// 				},
// 			)

// 			// Return zero fee
// 			assert.True(t, got.Sign() == 0)
// 		})
// 	})
// }
