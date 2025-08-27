package helpers_test

import (
	"dbcGoSDK/helpers"
	"dbcGoSDK/types"
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

// func TestGetMinBaseFeeBps(t *testing.T) {
// 	t.Run("linear fee scheduler - should calculate minimum fee correctly", func(t *testing.T) {

// 		baseFeeBps := float64(5_000)
// 		cliffFeeNumerator := baseFeeBps * float64(constants.FeeDenominator) / float64(constants.BasisPointMax)
// 		numberOfPeriod, reductionFactor := float64(144), float64(3_333_333)

// 		minBaseFeeBps := helpers.CalculateFeeSchedulerEndingBaseFeeBps(
// 			cliffFeeNumerator,
// 			numberOfPeriod,
// 			reductionFactor,
// 			types.BaseFeeModeFeeSchedulerLinear,
// 		)

// 		// linear mode: cliffFeeNumerator - (numberOfPeriod * reductionFactor)
// 		expectedMinFeeNumerator := cliffFeeNumerator - numberOfPeriod*reductionFactor
// 		expectedMinFeeBps := math.Max(0, expectedMinFeeNumerator/float64(constants.FeeDenominator)*float64(constants.BasisPointMax))

// 		t.Log("minBaseFeeBps:", minBaseFeeBps)
// 		t.Log("expectedMinFeeBps:", expectedMinFeeBps)

// 		assert.Less(t, minBaseFeeBps, baseFeeBps)
// 		assert.Equal(t, minBaseFeeBps, expectedMinFeeBps)
// 	})

// 	t.Run("exponential fee scheduler - should calculate minimum fee correctly", func(t *testing.T) {

// 		baseFeeBps := float64(5_000)
// 		cliffFeeNumerator := baseFeeBps * float64(constants.FeeDenominator) / float64(constants.BasisPointMax)
// 		numberOfPeriod, reductionFactor := 37.5, 822.5

// 		minBaseFeeBps := helpers.CalculateFeeSchedulerEndingBaseFeeBps(
// 			cliffFeeNumerator,
// 			numberOfPeriod,
// 			reductionFactor,
// 			types.BaseFeeModeFeeSchedulerExponential,
// 		)

// 		// exponential mode: cliffFeeNumerator * (1 - reductionFactor/BASIS_POINT_MAX)^numberOfPeriod
// 		decayRate := 1 - reductionFactor/float64(constants.BasisPointMax)
// 		expectedMinFeeNumerator := cliffFeeNumerator * math.Pow(decayRate, numberOfPeriod)
// 		expectedMinFeeBps := math.Max(0, expectedMinFeeNumerator/float64(constants.FeeDenominator)*float64(constants.BasisPointMax))

// 		t.Log("minBaseFeeBps:", minBaseFeeBps)
// 		t.Log("expectedMinFeeBps:", expectedMinFeeBps)

// 		assert.Less(t, minBaseFeeBps, baseFeeBps)
// 		assert.Equal(t, minBaseFeeBps, expectedMinFeeBps)
// 	})
// }

// func TestCalculateFeeScheduler(t *testing.T) {
// 	t.Run("linear fee scheduler - should calculate parameters correctly", func(t *testing.T) {
// 		const (
// 			startingFeeBps   = 5_000
// 			endingFeeBps     = 1_000
// 			numberOfPeriod   = 144
// 			feeSchedulerMode = types.BaseFeeModeFeeSchedulerLinear
// 			totalDuration    = 60
// 		)

// 		result, err := helpers.GetFeeSchedulerParams(
// 			startingFeeBps,
// 			endingFeeBps,
// 			feeSchedulerMode,
// 			numberOfPeriod,
// 			totalDuration,
// 		)
// 		if err != nil {
// 			t.Fatalf("GetFeeSchedulerParams errored: %s", err.Error())
// 		}

// 		t.Logf("%+v\n", result)

// 		// linear mode: cliffFeeNumerator - (numberOfPeriod * reductionFactor)
// 		assert.Equal(t, uint64(2777777), result.ThirdFactor)

// 	})
// 	t.Run("exponential fee scheduler - should calculate parameters correctly", func(t *testing.T) {
// 		const (
// 			startingFeeBps   = 5_000
// 			endingFeeBps     = 100
// 			numberOfPeriod   = 100
// 			feeSchedulerMode = types.BaseFeeModeFeeSchedulerExponential
// 			totalDuration    = 10 * 60 * 60
// 		)

// 		result, err := helpers.GetFeeSchedulerParams(
// 			startingFeeBps,
// 			endingFeeBps,
// 			feeSchedulerMode,
// 			numberOfPeriod,
// 			totalDuration,
// 		)
// 		if err != nil {
// 			t.Fatalf("GetFeeSchedulerParams errored: %s", err.Error())
// 		}

// 		t.Logf("%+v\n", result)

// 		// linear mode: cliffFeeNumerator - (numberOfPeriod * reductionFactor)
// 		assert.Equal(t, uint64(383), result.ThirdFactor)
// 	})

// 	t.Run("build curve with market cap - should calculate parameters correctly", func(t *testing.T) {
// 		// helpers.BuildCurveWithMarketCap()
// 	})
// }

// func TestCalculateLockedVesting(t *testing.T) {
// 	t.Run("calculate locked vesting parameters 1", func(t *testing.T) {
// 		const (
// 			totalLockedVestingAmount       = 7_777_777
// 			numberOfVestingPeriod          = 13
// 			cliffUnlockAmount              = 8
// 			totalVestingDuration           = 365 * 24 * 60 * 60
// 			cliffDurationFromMigrationTime = 0
// 		)

// 		result, err := helpers.GetLockedVestingParams(
// 			totalLockedVestingAmount,
// 			numberOfVestingPeriod,
// 			cliffUnlockAmount,
// 			totalVestingDuration,
// 			cliffDurationFromMigrationTime,
// 			types.TokenDecimalSIX,
// 		)
// 		if err != nil {
// 			t.Fatalf("GetLockedVestingParams errored: %s", err.Error())
// 		}

// 		assert.Equal(t, dbc.LockedVestingParams{
// 			AmountPerPeriod:   598289000000,
// 			Frequency:         2425846,
// 			NumberOfPeriod:    13,
// 			CliffUnlockAmount: 20000000,
// 		}, result)

// 		totalCalculatedVestingAmount := helpers.GetTotalVestingAmount(result)

// 		hold := new(big.Int).Mul(
// 			big.NewInt(totalLockedVestingAmount),
// 			new(big.Int).SetUint64(uint64(math.Pow10(int(types.TokenDecimalSIX)))),
// 		)

// 		assert.True(t, totalCalculatedVestingAmount.Cmp(hold) == 0)
// 	})

// 	t.Run("calculate locked vesting parameters 2", func(t *testing.T) {
// 		const (
// 			totalLockedVestingAmount       = 10_000_000
// 			numberOfVestingPeriod          = 365
// 			cliffUnlockAmount              = 0
// 			totalVestingDuration           = 365 * 24 * 60 * 60
// 			cliffDurationFromMigrationTime = 0
// 		)

// 		result, err := helpers.GetLockedVestingParams(
// 			totalLockedVestingAmount,
// 			numberOfVestingPeriod,
// 			cliffUnlockAmount,
// 			totalVestingDuration,
// 			cliffDurationFromMigrationTime,
// 			types.TokenDecimalSIX,
// 		)
// 		if err != nil {
// 			t.Fatalf("GetLockedVestingParams errored: %s", err.Error())
// 		}

// 		assert.Equal(t, dbc.LockedVestingParams{
// 			AmountPerPeriod:   27397000000,
// 			Frequency:         86400,
// 			NumberOfPeriod:    365,
// 			CliffUnlockAmount: 95000000,
// 		}, result)

// 		totalCalculatedVestingAmount := helpers.GetTotalVestingAmount(result)

// 		hold := new(big.Int).Mul(
// 			big.NewInt(totalLockedVestingAmount),
// 			new(big.Int).SetUint64(uint64(math.Pow10(int(types.TokenDecimalSIX)))),
// 		)

// 		assert.True(t, totalCalculatedVestingAmount.Cmp(hold) == 0)
// 	})

// 	t.Run("calculate locked vesting parameters 3", func(t *testing.T) {
// 		const (
// 			totalLockedVestingAmount       = 20_000_000
// 			numberOfVestingPeriod          = 1
// 			cliffUnlockAmount              = 20_000_000
// 			totalVestingDuration           = 1
// 			cliffDurationFromMigrationTime = 1000 * 365 * 24 * 60 * 60
// 		)

// 		result, err := helpers.GetLockedVestingParams(
// 			totalLockedVestingAmount,
// 			numberOfVestingPeriod,
// 			cliffUnlockAmount,
// 			totalVestingDuration,
// 			cliffDurationFromMigrationTime,
// 			types.TokenDecimalSIX,
// 		)
// 		if err != nil {
// 			t.Fatalf("GetLockedVestingParams errored: %s", err.Error())
// 		}

// 		assert.Equal(t, dbc.LockedVestingParams{
// 			AmountPerPeriod:                1000000,
// 			CliffDurationFromMigrationTime: 31536000000,
// 			Frequency:                      1,
// 			NumberOfPeriod:                 1,
// 			CliffUnlockAmount:              19999999000000,
// 		}, result)

// 		totalCalculatedVestingAmount := helpers.GetTotalVestingAmount(result)

// 		hold := new(big.Int).Mul(
// 			big.NewInt(totalLockedVestingAmount),
// 			new(big.Int).SetUint64(uint64(math.Pow10(int(types.TokenDecimalSIX)))),
// 		)

// 		assert.True(t, totalCalculatedVestingAmount.Cmp(hold) == 0)
// 	})

// 	t.Run("calculate locked vesting parameters 4", func(t *testing.T) {
// 		const (
// 			totalLockedVestingAmount       = 8_888_888
// 			numberOfVestingPeriod          = 9
// 			cliffUnlockAmount              = 9_999
// 			totalVestingDuration           = 365 * 24 * 60 * 60
// 			cliffDurationFromMigrationTime = 0
// 		)

// 		result, err := helpers.GetLockedVestingParams(
// 			totalLockedVestingAmount,
// 			numberOfVestingPeriod,
// 			cliffUnlockAmount,
// 			totalVestingDuration,
// 			cliffDurationFromMigrationTime,
// 			types.TokenDecimalSIX,
// 		)
// 		if err != nil {
// 			t.Fatalf("GetLockedVestingParams errored: %s", err.Error())
// 		}

// 		assert.Equal(t, dbc.LockedVestingParams{
// 			AmountPerPeriod:   986543000000,
// 			Frequency:         3504000,
// 			NumberOfPeriod:    9,
// 			CliffUnlockAmount: 10001000000,
// 		}, result)

// 		totalCalculatedVestingAmount := helpers.GetTotalVestingAmount(result)

// 		hold := new(big.Int).Mul(
// 			big.NewInt(totalLockedVestingAmount),
// 			new(big.Int).SetUint64(uint64(math.Pow10(int(types.TokenDecimalSIX)))),
// 		)

// 		assert.True(t, totalCalculatedVestingAmount.Cmp(hold) == 0)
// 	})

// 	t.Run("calculate locked vesting parameters 5", func(t *testing.T) {
// 		const (
// 			totalLockedVestingAmount       = 1_000_000
// 			numberOfVestingPeriod          = 1
// 			cliffUnlockAmount              = 1_000_000
// 			totalVestingDuration           = 0
// 			cliffDurationFromMigrationTime = 365 * 24 * 60 * 60
// 		)

// 		result, err := helpers.GetLockedVestingParams(
// 			totalLockedVestingAmount,
// 			numberOfVestingPeriod,
// 			cliffUnlockAmount,
// 			totalVestingDuration,
// 			cliffDurationFromMigrationTime,
// 			types.TokenDecimalSIX,
// 		)
// 		if err != nil {
// 			t.Fatalf("GetLockedVestingParams errored: %s", err.Error())
// 		}

// 		assert.Equal(t, dbc.LockedVestingParams{
// 			AmountPerPeriod:                1000000,
// 			CliffDurationFromMigrationTime: 31536000,
// 			Frequency:                      1,
// 			NumberOfPeriod:                 1,
// 			CliffUnlockAmount:              999999000000,
// 		}, result)

// 		totalCalculatedVestingAmount := helpers.GetTotalVestingAmount(result)

// 		hold := new(big.Int).Mul(
// 			big.NewInt(totalLockedVestingAmount),
// 			new(big.Int).SetUint64(uint64(math.Pow10(int(types.TokenDecimalSIX)))),
// 		)

// 		assert.True(t, totalCalculatedVestingAmount.Cmp(hold) == 0)
// 	})
// }

func TestRateLimiter(t *testing.T) {
	t.Run("getRateLimiterParams", func(t *testing.T) {
		const (
			baseFeeBps         = 100 // 1%
			feeIncrementBps    = 10  // 10 bps
			referenceAmount    = 0.2
			maxLimiterDuration = 100_000 // slots
			tokenQuoteDecimal  = 6
			activationType     = types.ActivationTypeSlot
		)

		params, err := helpers.GetRateLimiterParams(
			baseFeeBps,
			feeIncrementBps,
			referenceAmount,
			maxLimiterDuration,
			tokenQuoteDecimal,
			activationType,
		)
		if err != nil {
			t.Fatalf("GetRateLimiterParams errored: %s", err.Error())
		}

		assert.Equal(t, types.BaseFeeModeRateLimiter, params.BaseFeeMode)
		assert.True(t, new(big.Int).SetUint64(params.CliffFeeNumerator).Cmp(helpers.BpsToFeeNumerator(baseFeeBps)) == 0)

		assert.Greater(t, params.FirstFactor, uint16(0)) // feeIncrementBps
		assert.Equal(t, uint64(maxLimiterDuration), params.SecondFactor)
		assert.Equal(t, float64(referenceAmount)*math.Pow10(tokenQuoteDecimal), float64(params.ThirdFactor))
	})
}
