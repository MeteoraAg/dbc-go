package helpers

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"lukechampine.com/uint128"

	"github.com/MeteoraAg/dbc-go/common"
)

// Deserializes the pool config account data
func DeserializePoolConfig(data []byte) (*common.PoolConfig, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("data too short to deserialize")
	}

	// Skip the 8-byte discriminator
	data = data[8:]

	config := &common.PoolConfig{}
	reader := bytes.NewReader(data)

	// Read QuoteMint
	if err := binary.Read(reader, binary.LittleEndian, &config.QuoteMint); err != nil {
		return nil, fmt.Errorf("failed to read QuoteMint: %w", err)
	}

	// Read FeeClaimer
	if err := binary.Read(reader, binary.LittleEndian, &config.FeeClaimer); err != nil {
		return nil, fmt.Errorf("failed to read FeeClaimer: %w", err)
	}

	// Read LeftoverReceiver
	if err := binary.Read(reader, binary.LittleEndian, &config.LeftoverReceiver); err != nil {
		return nil, fmt.Errorf("failed to read LeftoverReceiver: %w", err)
	}

	// Read PoolFees
	// BaseFee
	if err := binary.Read(reader, binary.LittleEndian, &config.PoolFees.BaseFee.CliffFeeNumerator); err != nil {
		return nil, fmt.Errorf("failed to read CliffFeeNumerator: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.PoolFees.BaseFee.PeriodFrequency); err != nil {
		return nil, fmt.Errorf("failed to read PeriodFrequency: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.PoolFees.BaseFee.ReductionFactor); err != nil {
		return nil, fmt.Errorf("failed to read ReductionFactor: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.PoolFees.BaseFee.NumberOfPeriod); err != nil {
		return nil, fmt.Errorf("failed to read NumberOfPeriod: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.PoolFees.BaseFee.FeeSchedulerMode); err != nil {
		return nil, fmt.Errorf("failed to read FeeSchedulerMode: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.PoolFees.BaseFee.Padding0); err != nil {
		return nil, fmt.Errorf("failed to read BaseFee.Padding0: %w", err)
	}

	// DynamicFee
	if err := binary.Read(reader, binary.LittleEndian, &config.PoolFees.DynamicFee.Initialized); err != nil {
		return nil, fmt.Errorf("failed to read Initialized: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.PoolFees.DynamicFee.Padding); err != nil {
		return nil, fmt.Errorf("failed to read DynamicFee.Padding: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.PoolFees.DynamicFee.MaxVolatilityAccumulator); err != nil {
		return nil, fmt.Errorf("failed to read MaxVolatilityAccumulator: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.PoolFees.DynamicFee.VariableFeeControl); err != nil {
		return nil, fmt.Errorf("failed to read VariableFeeControl: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.PoolFees.DynamicFee.BinStep); err != nil {
		return nil, fmt.Errorf("failed to read BinStep: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.PoolFees.DynamicFee.FilterPeriod); err != nil {
		return nil, fmt.Errorf("failed to read FilterPeriod: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.PoolFees.DynamicFee.DecayPeriod); err != nil {
		return nil, fmt.Errorf("failed to read DecayPeriod: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.PoolFees.DynamicFee.ReductionFactor); err != nil {
		return nil, fmt.Errorf("failed to read DynamicFee.ReductionFactor: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.PoolFees.DynamicFee.Padding2); err != nil {
		return nil, fmt.Errorf("failed to read DynamicFee.Padding2: %w", err)
	}

	// Read BinStepU128
	var binstepU128Lo, binstepU128Hi uint64
	if err := binary.Read(reader, binary.LittleEndian, &binstepU128Lo); err != nil {
		return nil, fmt.Errorf("failed to read BinStepU128.Lo: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &binstepU128Hi); err != nil {
		return nil, fmt.Errorf("failed to read BinStepU128.Hi: %w", err)
	}
	config.PoolFees.DynamicFee.BinStepU128 = uint128.Uint128{Lo: binstepU128Lo, Hi: binstepU128Hi}

	// PoolFees remaining fields
	if err := binary.Read(reader, binary.LittleEndian, &config.PoolFees.Padding0); err != nil {
		return nil, fmt.Errorf("failed to read PoolFees.Padding0: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.PoolFees.Padding1); err != nil {
		return nil, fmt.Errorf("failed to read PoolFees.Padding1: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.PoolFees.ProtocolFeePercent); err != nil {
		return nil, fmt.Errorf("failed to read ProtocolFeePercent: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.PoolFees.ReferralFeePercent); err != nil {
		return nil, fmt.Errorf("failed to read ReferralFeePercent: %w", err)
	}

	// Read the uint8 fields
	if err := binary.Read(reader, binary.LittleEndian, &config.CollectFeeMode); err != nil {
		return nil, fmt.Errorf("failed to read CollectFeeMode: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.MigrationOption); err != nil {
		return nil, fmt.Errorf("failed to read MigrationOption: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.ActivationType); err != nil {
		return nil, fmt.Errorf("failed to read ActivationType: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.TokenDecimal); err != nil {
		return nil, fmt.Errorf("failed to read TokenDecimal: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.Version); err != nil {
		return nil, fmt.Errorf("failed to read Version: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.TokenType); err != nil {
		return nil, fmt.Errorf("failed to read TokenType: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.QuoteTokenFlag); err != nil {
		return nil, fmt.Errorf("failed to read QuoteTokenFlag: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.PartnerLockedLpPercentage); err != nil {
		return nil, fmt.Errorf("failed to read PartnerLockedLpPercentage: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.PartnerLpPercentage); err != nil {
		return nil, fmt.Errorf("failed to read PartnerLpPercentage: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.CreatorLockedLpPercentage); err != nil {
		return nil, fmt.Errorf("failed to read CreatorLockedLpPercentage: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.CreatorLpPercentage); err != nil {
		return nil, fmt.Errorf("failed to read CreatorLpPercentage: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.MigrationFeeOption); err != nil {
		return nil, fmt.Errorf("failed to read MigrationFeeOption: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.FixedTokenSupplyFlag); err != nil {
		return nil, fmt.Errorf("failed to read FixedTokenSupplyFlag: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.CreatorTradingFeePercentage); err != nil {
		return nil, fmt.Errorf("failed to read CreatorTradingFeePercentage: %w", err)
	}

	// Read padding fields
	if err := binary.Read(reader, binary.LittleEndian, &config.Padding0); err != nil {
		return nil, fmt.Errorf("failed to read Padding0: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.Padding1); err != nil {
		return nil, fmt.Errorf("failed to read Padding1: %w", err)
	}

	// Read uint64 fields
	if err := binary.Read(reader, binary.LittleEndian, &config.SwapBaseAmount); err != nil {
		return nil, fmt.Errorf("failed to read SwapBaseAmount: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.MigrationQuoteThreshold); err != nil {
		return nil, fmt.Errorf("failed to read MigrationQuoteThreshold: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.MigrationBaseThreshold); err != nil {
		return nil, fmt.Errorf("failed to read MigrationBaseThreshold: %w", err)
	}

	// Read MigrationSqrtPrice using the helper function
	var migrationSqrtPriceLo, migrationSqrtPriceHi uint64
	if err := binary.Read(reader, binary.LittleEndian, &migrationSqrtPriceLo); err != nil {
		return nil, fmt.Errorf("failed to read MigrationSqrtPrice.Lo: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &migrationSqrtPriceHi); err != nil {
		return nil, fmt.Errorf("failed to read MigrationSqrtPrice.Hi: %w", err)
	}
	config.MigrationSqrtPrice = uint128.Uint128{Lo: migrationSqrtPriceLo, Hi: migrationSqrtPriceHi}

	// Read LockedVestingConfig
	if err := binary.Read(reader, binary.LittleEndian, &config.LockedVestingConfig.AmountPerPeriod); err != nil {
		return nil, fmt.Errorf("failed to read AmountPerPeriod: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.LockedVestingConfig.CliffDurationFromMigrationTime); err != nil {
		return nil, fmt.Errorf("failed to read CliffDurationFromMigrationTime: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.LockedVestingConfig.Frequency); err != nil {
		return nil, fmt.Errorf("failed to read Frequency: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.LockedVestingConfig.NumberOfPeriod); err != nil {
		return nil, fmt.Errorf("failed to read NumberOfPeriod: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.LockedVestingConfig.CliffUnlockAmount); err != nil {
		return nil, fmt.Errorf("failed to read CliffUnlockAmount: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.LockedVestingConfig.Padding); err != nil {
		return nil, fmt.Errorf("failed to read LockedVestingConfig.Padding: %w", err)
	}

	// Read PreMigrationTokenSupply and PostMigrationTokenSupply
	if err := binary.Read(reader, binary.LittleEndian, &config.PreMigrationTokenSupply); err != nil {
		return nil, fmt.Errorf("failed to read PreMigrationTokenSupply: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &config.PostMigrationTokenSupply); err != nil {
		return nil, fmt.Errorf("failed to read PostMigrationTokenSupply: %w", err)
	}

	// Read Padding2 using the helper function
	for i := 0; i < len(config.Padding2); i++ {
		if err := binary.Read(reader, binary.LittleEndian, &config.Padding2[i]); err != nil {
			return nil, fmt.Errorf("failed to read Padding2[%d]: %w", i, err)
		}
	}

	// Read SqrtStartPrice using the helper function
	var sqrtStartPriceLo, sqrtStartPriceHi uint64
	if err := binary.Read(reader, binary.LittleEndian, &sqrtStartPriceLo); err != nil {
		return nil, fmt.Errorf("failed to read SqrtStartPrice.Lo: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &sqrtStartPriceHi); err != nil {
		return nil, fmt.Errorf("failed to read SqrtStartPrice.Hi: %w", err)
	}
	config.SqrtStartPrice = uint128.Uint128{Lo: sqrtStartPriceLo, Hi: sqrtStartPriceHi}

	// Read Curve
	for i := 0; i < len(config.Curve); i++ {
		// Read SqrtPrice using the helper function
		var sqrtPriceLo, sqrtPriceHi uint64
		if err := binary.Read(reader, binary.LittleEndian, &sqrtPriceLo); err != nil {
			return nil, fmt.Errorf("failed to read Curve[%d].SqrtPrice.Lo: %w", i, err)
		}
		if err := binary.Read(reader, binary.LittleEndian, &sqrtPriceHi); err != nil {
			return nil, fmt.Errorf("failed to read Curve[%d].SqrtPrice.Hi: %w", i, err)
		}
		config.Curve[i].SqrtPrice = uint128.Uint128{Lo: sqrtPriceLo, Hi: sqrtPriceHi}

		// Read Liquidity using the helper function
		var liquidityLo, liquidityHi uint64
		if err := binary.Read(reader, binary.LittleEndian, &liquidityLo); err != nil {
			return nil, fmt.Errorf("failed to read Curve[%d].Liquidity.Lo: %w", i, err)
		}
		if err := binary.Read(reader, binary.LittleEndian, &liquidityHi); err != nil {
			return nil, fmt.Errorf("failed to read Curve[%d].Liquidity.Hi: %w", i, err)
		}
		config.Curve[i].Liquidity = uint128.Uint128{Lo: liquidityLo, Hi: liquidityHi}
	}

	return config, nil
}

// Deserializes the pool account data
func DeserializePool(data []byte) (*common.Pool, error) {
	if len(data) < 8 {
		return nil, fmt.Errorf("data too short to deserialize")
	}

	// Skip the 8-byte discriminator
	data = data[8:]

	pool := &common.Pool{}
	reader := bytes.NewReader(data)

	// Read VolatilityTracker
	if err := binary.Read(reader, binary.LittleEndian, &pool.VolatilityTracker.LastUpdateTimestamp); err != nil {
		return nil, fmt.Errorf("failed to read LastUpdateTimestamp: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.VolatilityTracker.Padding); err != nil {
		return nil, fmt.Errorf("failed to read VolatilityTracker.Padding: %w", err)
	}

	// Read SqrtPriceReference (u128)
	var sqrtPriceRefLo, sqrtPriceRefHi uint64
	if err := binary.Read(reader, binary.LittleEndian, &sqrtPriceRefLo); err != nil {
		return nil, fmt.Errorf("failed to read SqrtPriceReference.Lo: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &sqrtPriceRefHi); err != nil {
		return nil, fmt.Errorf("failed to read SqrtPriceReference.Hi: %w", err)
	}
	pool.VolatilityTracker.SqrtPriceReference = uint128.Uint128{Lo: sqrtPriceRefLo, Hi: sqrtPriceRefHi}

	// Read VolatilityAccumulator (u128)
	var volAccLo, volAccHi uint64
	if err := binary.Read(reader, binary.LittleEndian, &volAccLo); err != nil {
		return nil, fmt.Errorf("failed to read VolatilityAccumulator.Lo: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &volAccHi); err != nil {
		return nil, fmt.Errorf("failed to read VolatilityAccumulator.Hi: %w", err)
	}
	pool.VolatilityTracker.VolatilityAccumulator = uint128.Uint128{Lo: volAccLo, Hi: volAccHi}

	// Read VolatilityReference (u128)
	var volRefLo, volRefHi uint64
	if err := binary.Read(reader, binary.LittleEndian, &volRefLo); err != nil {
		return nil, fmt.Errorf("failed to read VolatilityReference.Lo: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &volRefHi); err != nil {
		return nil, fmt.Errorf("failed to read VolatilityReference.Hi: %w", err)
	}
	pool.VolatilityTracker.VolatilityReference = uint128.Uint128{Lo: volRefLo, Hi: volRefHi}

	// Read PublicKeys
	if err := binary.Read(reader, binary.LittleEndian, &pool.Config); err != nil {
		return nil, fmt.Errorf("failed to read Config: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.Creator); err != nil {
		return nil, fmt.Errorf("failed to read Creator: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.BaseMint); err != nil {
		return nil, fmt.Errorf("failed to read BaseMint: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.BaseVault); err != nil {
		return nil, fmt.Errorf("failed to read BaseVault: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.QuoteVault); err != nil {
		return nil, fmt.Errorf("failed to read QuoteVault: %w", err)
	}

	// Read uint64 fields
	if err := binary.Read(reader, binary.LittleEndian, &pool.BaseReserve); err != nil {
		return nil, fmt.Errorf("failed to read BaseReserve: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.QuoteReserve); err != nil {
		return nil, fmt.Errorf("failed to read QuoteReserve: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.ProtocolBaseFee); err != nil {
		return nil, fmt.Errorf("failed to read ProtocolBaseFee: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.ProtocolQuoteFee); err != nil {
		return nil, fmt.Errorf("failed to read ProtocolQuoteFee: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.PartnerBaseFee); err != nil {
		return nil, fmt.Errorf("failed to read PartnerBaseFee: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.PartnerQuoteFee); err != nil {
		return nil, fmt.Errorf("failed to read PartnerQuoteFee: %w", err)
	}

	// Read SqrtPrice (u128)
	var sqrtPriceLo, sqrtPriceHi uint64
	if err := binary.Read(reader, binary.LittleEndian, &sqrtPriceLo); err != nil {
		return nil, fmt.Errorf("failed to read SqrtPrice.Lo: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &sqrtPriceHi); err != nil {
		return nil, fmt.Errorf("failed to read SqrtPrice.Hi: %w", err)
	}
	pool.SqrtPrice = uint128.Uint128{Lo: sqrtPriceLo, Hi: sqrtPriceHi}

	// Read uint64 fields
	if err := binary.Read(reader, binary.LittleEndian, &pool.ActivationPoint); err != nil {
		return nil, fmt.Errorf("failed to read ActivationPoint: %w", err)
	}

	// Read uint8 fields
	if err := binary.Read(reader, binary.LittleEndian, &pool.PoolType); err != nil {
		return nil, fmt.Errorf("failed to read PoolType: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.IsMigrated); err != nil {
		return nil, fmt.Errorf("failed to read IsMigrated: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.IsPartnerWithdrawSurplus); err != nil {
		return nil, fmt.Errorf("failed to read IsPartnerWithdrawSurplus: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.IsProtocolWithdrawSurplus); err != nil {
		return nil, fmt.Errorf("failed to read IsProtocolWithdrawSurplus: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.MigrationProgress); err != nil {
		return nil, fmt.Errorf("failed to read MigrationProgress: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.IsWithdrawLeftover); err != nil {
		return nil, fmt.Errorf("failed to read IsWithdrawLeftover: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.IsCreatorWithdrawSurplus); err != nil {
		return nil, fmt.Errorf("failed to read IsCreatorWithdrawSurplus: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.MigrationFeeWithdrawStatus); err != nil {
		return nil, fmt.Errorf("failed to read MigrationFeeWithdrawStatus: %w", err)
	}

	// Read Metrics
	if err := binary.Read(reader, binary.LittleEndian, &pool.Metrics.TotalProtocolBaseFee); err != nil {
		return nil, fmt.Errorf("failed to read TotalProtocolBaseFee: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.Metrics.TotalProtocolQuoteFee); err != nil {
		return nil, fmt.Errorf("failed to read TotalProtocolQuoteFee: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.Metrics.TotalTradingBaseFee); err != nil {
		return nil, fmt.Errorf("failed to read TotalTradingBaseFee: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.Metrics.TotalTradingQuoteFee); err != nil {
		return nil, fmt.Errorf("failed to read TotalTradingQuoteFee: %w", err)
	}

	// Read uint64 fields
	if err := binary.Read(reader, binary.LittleEndian, &pool.FinishCurveTimestamp); err != nil {
		return nil, fmt.Errorf("failed to read FinishCurveTimestamp: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.CreatorBaseFee); err != nil {
		return nil, fmt.Errorf("failed to read CreatorBaseFee: %w", err)
	}
	if err := binary.Read(reader, binary.LittleEndian, &pool.CreatorQuoteFee); err != nil {
		return nil, fmt.Errorf("failed to read CreatorQuoteFee: %w", err)
	}

	// Read padding
	if err := binary.Read(reader, binary.LittleEndian, &pool.Padding1); err != nil {
		return nil, fmt.Errorf("failed to read Padding1: %w", err)
	}

	return pool, nil
}
