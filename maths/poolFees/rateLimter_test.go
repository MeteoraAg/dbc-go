package poolfees_test

import (
	"dbcGoSDK/helpers"
	poolfees "dbcGoSDK/maths/poolFees"
	"dbcGoSDK/types"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
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

	assert.Equal(t, types.BaseFee{
		CliffFeeNumerator: 10000000,
		FirstFactor:       10,
		SecondFactor:      100000,
		ThirdFactor:       200000,
		BaseFeeMode:       types.BaseFeeModeRateLimiter,
	}, params)

	tests := []struct {
		name  string
		param struct {
			includedFeeAmount uint64
		}
		wanted uint64
	}{
		{
			name:   "0.4 SOL tx fee:",
			param:  struct{ includedFeeAmount uint64 }{includedFeeAmount: 0.4 * 1e9},
			wanted: 10_500_000,
		},
		{
			name:   "0.2 SOL tx fee:",
			param:  struct{ includedFeeAmount uint64 }{includedFeeAmount: 0.2 * 1e9},
			wanted: 10_000_000,
		},
		{
			name:   "0.1 SOL tx fee:",
			param:  struct{ includedFeeAmount uint64 }{includedFeeAmount: 0.1 * 1e9},
			wanted: 10_000_000,
		},
		{
			name:   "1 SOL tx fee:",
			param:  struct{ includedFeeAmount uint64 }{includedFeeAmount: 1 * 1e9},
			wanted: 12_000_000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := poolfees.GetFeeNumeratorFromIncludedAmount(
				new(big.Int).SetUint64(params.CliffFeeNumerator),
				new(big.Int).SetUint64(referenceAmount*1e9),
				new(big.Int).SetUint64(feeIncrementBps),
				new(big.Int).SetUint64(tt.param.includedFeeAmount),
			)
			if err != nil {
				t.Fatalf("GetFeeNumeratorFromIncludedAmount errored: %s", err.Error())
			}

			if !got.IsUint64() {
				t.Fatalf("cannot fit got(%s) into uint64", got)
			}

			assert.Equal(t, tt.wanted, got.Uint64())
		})
	}
}
