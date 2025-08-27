package helpers_test

import (
	"dbcGoSDK/constants"
	"dbcGoSDK/generated/dbc"
	"dbcGoSDK/helpers"
	"dbcGoSDK/types"
	"math"
	"math/big"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildCurve(t *testing.T) {

	param := types.BuildCurveBaseParam{
		TotalTokenSupply:   1_000_000_000,
		MigrationOption:    types.MigrationOptionMET_DAMM_V2,
		TokenBaseDecimal:   types.TokenDecimalSIX,
		TokenQuoteDecimal:  types.TokenDecimalNINE,
		LockedVestingParam: types.LockedVestingParams{},
		BaseFeeParams: types.BaseFeeParams{
			BaseFeeMode: types.BaseFeeModeFeeSchedulerLinear,
			FeeSchedulerParam: &types.FeeSchedulerParams{
				StartingFeeBps: 100,
				EndingFeeBps:   100,
			},
		},
		DynamicFeeEnabled:         true,
		ActivationType:            types.ActivationTypeSlot,
		CollectFeeMode:            types.CollectFeeModeQuoteToken,
		MigrationFeeOption:        types.MigrationFeeOptionFixedBps100,
		TokenType:                 types.TokenTypeSPL,
		PartnerLockedLpPercentage: 100,
		Leftover:                  10_000,
	}

	t.Run("build curve with percentage and threshold parameters", func(t *testing.T) {

		config, err := helpers.BuildCurve(types.BuildCurveParam{
			BuildCurveBaseParam:         param,
			PercentageSupplyOnMigration: 2.983257229832572,
			MigrationQuoteThreshold:     95.07640791476408,
		})
		if err != nil {
			t.Fatalf("BuildCurve errored: %s", err.Error())
		}

		migrationQuoteThreshold := new(big.Int)
		c := new(big.Float).SetUint64(config.MigrationQuoteThreshold)
		c.Quo(c, big.NewFloat(math.Pow10(int(types.TokenDecimalNINE))))
		c.Int(migrationQuoteThreshold)

		const threshold = 0.00001 //  0.001%

		assert.Equal(t, uint64(95), migrationQuoteThreshold.Uint64())

		assert.True(t,
			ApproxBigInt(config.SqrtStartPrice.BigInt(), big.NewInt(32022795711993578), threshold))

		assert.Equal(t, 2, len(config.Curve))

		assert.True(t,
			ApproxBigInt(config.Curve[0].SqrtPrice.BigInt(), big.NewInt(1041383648506654343), threshold))

		liquidity0, _ := new(big.Int).SetString("32052783733131623276178198534722", 10)

		assert.True(t,
			ApproxBigInt(config.Curve[0].Liquidity.BigInt(), liquidity0, threshold))

		sqrtPrice1, _ := new(big.Int).SetString("79226673521066979257578248091", 10)
		liquidity1, _ := new(big.Int).SetString("4153615958224055322353231", 10)

		assert.True(t,
			ApproxBigInt(config.Curve[1].SqrtPrice.BigInt(), sqrtPrice1, threshold))

		assert.True(t,
			ApproxBigInt(config.Curve[1].Liquidity.BigInt(), liquidity1, threshold))
	})
}

func TestBuildCurveWithMarketCap(t *testing.T) {

	param := types.BuildCurveBaseParam{
		TotalTokenSupply:   1_000_000_000,
		MigrationOption:    types.MigrationOptionMET_DAMM_V2,
		TokenBaseDecimal:   types.TokenDecimalSIX,
		TokenQuoteDecimal:  types.TokenDecimalNINE,
		LockedVestingParam: types.LockedVestingParams{},
		BaseFeeParams: types.BaseFeeParams{
			BaseFeeMode: types.BaseFeeModeFeeSchedulerLinear,
			FeeSchedulerParam: &types.FeeSchedulerParams{
				StartingFeeBps: 100,
				EndingFeeBps:   100,
			},
		},
		DynamicFeeEnabled:         true,
		ActivationType:            types.ActivationTypeSlot,
		CollectFeeMode:            types.CollectFeeModeQuoteToken,
		MigrationFeeOption:        types.MigrationFeeOptionFixedBps100,
		TokenType:                 types.TokenTypeSPL,
		PartnerLockedLpPercentage: 100,
		Leftover:                  10_000,
		MigrationFee: types.MigrationFee{
			FeePercentage:        10,
			CreatorFeePercentage: 50,
		},
	}

	t.Run("build curve by market cap 1", func(t *testing.T) {
		config, err := helpers.BuildCurveWithMarketCap(types.BuildCurveWithMarketCapParam{
			BuildCurveBaseParam: param,
			InitialMarketCap:    23.5,
			MigrationMarketCap:  405.882352941,
		})
		if err != nil {
			t.Fatalf("BuildCurveWithMarketCap errored: %s", err.Error())
		}

		migrationQuoteThreshold := new(big.Int)
		c := new(big.Float).SetUint64(config.MigrationQuoteThreshold)
		c.Quo(c, big.NewFloat(math.Pow10(int(types.TokenDecimalNINE))))
		c.Int(migrationQuoteThreshold)

		const threshold = 0.00001 //  0.001%

		assert.Equal(t, uint64(87), migrationQuoteThreshold.Uint64())

		assert.True(t,
			ApproxBigInt(config.SqrtStartPrice.BigInt(), big.NewInt(99359901068392311), threshold))

		assert.Equal(t, 2, len(config.Curve))

		assert.True(t,
			ApproxBigInt(config.Curve[0].SqrtPrice.BigInt(), big.NewInt(371637737252560534), threshold))

		liquidity0, _ := new(big.Int).SetString("109313834256123014321586447219617", 10)

		assert.True(t,
			ApproxBigInt(config.Curve[0].Liquidity.BigInt(), liquidity0, threshold))

		sqrtPrice1, _ := new(big.Int).SetString("79226673521066979257578248091", 10)
		liquidity1, _ := new(big.Int).SetString("3458292059355837878186760", 10)

		assert.True(t,
			ApproxBigInt(config.Curve[1].SqrtPrice.BigInt(), sqrtPrice1, threshold))

		assert.True(t,
			ApproxBigInt(config.Curve[1].Liquidity.BigInt(), liquidity1, threshold))
	})

	t.Run("build curve by market cap 2", func(t *testing.T) {
		config, err := helpers.BuildCurveWithMarketCap(types.BuildCurveWithMarketCapParam{
			BuildCurveBaseParam: param,
			InitialMarketCap:    0.1,
			MigrationMarketCap:  0.5,
		})
		if err != nil {
			t.Fatalf("BuildCurveWithMarketCap errored: %s", err.Error())
		}

		const threshold = 0.00001 //  0.001%

		assert.Equal(t, uint64(171674391), config.MigrationQuoteThreshold)

		assert.True(t,
			ApproxBigInt(config.SqrtStartPrice.BigInt(), big.NewInt(6481528269918120), threshold))

		assert.Equal(t, 2, len(config.Curve))

		assert.True(t,
			ApproxBigInt(config.Curve[0].SqrtPrice.BigInt(), big.NewInt(13043817825332782), threshold))

		liquidity0, _ := new(big.Int).SetString("8902040608828227467724510683754", 10)

		assert.True(t,
			ApproxBigInt(config.Curve[0].Liquidity.BigInt(), liquidity0, threshold))

		sqrtPrice1, _ := new(big.Int).SetString("79226673521066979257578248091", 10)
		liquidity1, _ := new(big.Int).SetString("2999694611582968943593834", 10)

		assert.True(t,
			ApproxBigInt(config.Curve[1].SqrtPrice.BigInt(), sqrtPrice1, threshold))

		assert.True(t,
			ApproxBigInt(config.Curve[1].Liquidity.BigInt(), liquidity1, threshold))
	})

	t.Run("build curve by market cap with locked vesting", func(t *testing.T) {
		param.LockedVestingParam = types.LockedVestingParams{
			TotalLockedVestingAmount: 10_000_000,
			NumberOfVestingPeriod:    1_000,
			TotalVestingDuration:     365 * 24 * 60 * 60,
		}
		config, err := helpers.BuildCurveWithMarketCap(types.BuildCurveWithMarketCapParam{
			BuildCurveBaseParam: param,
			InitialMarketCap:    99.1669972233,
			MigrationMarketCap:  462.779320376,
		})
		if err != nil {
			t.Fatalf("BuildCurveWithMarketCap errored: %s", err.Error())
		}

		migrationQuoteThreshold := new(big.Int)
		c := new(big.Float).SetUint64(config.MigrationQuoteThreshold)
		c.Quo(c, big.NewFloat(math.Pow10(int(types.TokenDecimalNINE))))
		c.Int(migrationQuoteThreshold)

		const threshold = 0.001 //  0.1%

		assert.Equal(t, uint64(161), migrationQuoteThreshold.Uint64())
		assert.True(t,
			ApproxBigInt(config.SqrtStartPrice.BigInt(), big.NewInt(204108363870746018), threshold))

		assert.Equal(t, 2, len(config.Curve))

		assert.True(t,
			ApproxBigInt(config.Curve[0].SqrtPrice.BigInt(), big.NewInt(396832007907135132), threshold))

		liquidity0, _ := new(big.Int).SetString("284410531934842851466931437257040", 10)

		assert.True(t,
			ApproxBigInt(config.Curve[0].Liquidity.BigInt(), liquidity0, threshold))

		sqrtPrice1, _ := new(big.Int).SetString("79226673521066979257578248091", 10)
		liquidity1, _ := new(big.Int).SetString("2907011525042650737185350", 10)

		assert.True(t,
			ApproxBigInt(config.Curve[1].SqrtPrice.BigInt(), sqrtPrice1, threshold))

		assert.True(t,
			ApproxBigInt(config.Curve[1].Liquidity.BigInt(), liquidity1, threshold))

		lockedVesting, err := helpers.GetLockedVestingParams(
			param.LockedVestingParam.TotalLockedVestingAmount,
			param.LockedVestingParam.NumberOfVestingPeriod,
			param.LockedVestingParam.CliffUnlockAmount,
			param.LockedVestingParam.TotalVestingDuration,
			param.LockedVestingParam.CliffDurationFromMigrationTime,
			param.TokenBaseDecimal,
		)
		if err != nil {
			t.Fatalf("GetLockedVestingParams errored: %s", err.Error())
		}

		assert.Equal(t, dbc.LockedVestingParams{
			AmountPerPeriod: 10_000_000_000,
			Frequency:       31_536,
			NumberOfPeriod:  1_000,
		},
			lockedVesting,
		)

		totalVestingAmount := helpers.GetTotalVestingAmount(lockedVesting)
		if err != nil {
			t.Fatalf("GetTotalVestingAmount errored: %s", err.Error())
		}

		totalVestingAmountFloat64, _ := totalVestingAmount.Float64()
		assert.InEpsilon(t,
			float64(param.LockedVestingParam.TotalLockedVestingAmount)*math.Pow10(int(param.TokenBaseDecimal)),
			totalVestingAmountFloat64,
			1e-13,
		)

		assert.NotNil(t, config.TokenSupply)

		vestingPercentage := new(big.Int).Quo(
			new(big.Int).Mul(totalVestingAmount, constants.HundredInBigInt),
			new(big.Int).Mul(
				new(big.Int).SetUint64(param.TotalTokenSupply),
				new(big.Int).Exp(big.NewInt(10), new(big.Int).SetUint64(uint64(param.TokenBaseDecimal)), nil),
			),
		)

		migrationPercentage := new(big.Int).Quo(
			new(big.Int).Mul(new(big.Int).SetUint64(config.MigrationQuoteThreshold), constants.HundredInBigInt),
			new(big.Int).SetUint64(config.TokenSupply.PreMigrationTokenSupply),
		)

		assert.True(t, migrationPercentage.Cmp(new(big.Int).Sub(
			constants.HundredInBigInt, vestingPercentage,
		)) < 0)
	})

	t.Run("build curve by market cap 3", func(t *testing.T) {
		config, err := helpers.BuildCurveWithMarketCap(types.BuildCurveWithMarketCapParam{
			BuildCurveBaseParam: types.BuildCurveBaseParam{
				TotalTokenSupply:  100_000_000,
				MigrationOption:   types.MigrationOptionMET_DAMM_V2,
				TokenBaseDecimal:  types.TokenDecimalSIX,
				TokenQuoteDecimal: types.TokenDecimalSIX,
				LockedVestingParam: types.LockedVestingParams{
					TotalLockedVestingAmount: 50_000_000,
					NumberOfVestingPeriod:    1,
					CliffUnlockAmount:        50_000_000,
					TotalVestingDuration:     1,
				},
				BaseFeeParams: types.BaseFeeParams{
					BaseFeeMode: types.BaseFeeModeFeeSchedulerLinear,
					FeeSchedulerParam: &types.FeeSchedulerParams{
						StartingFeeBps: 100,
						EndingFeeBps:   100,
					},
				},
				DynamicFeeEnabled:           true,
				ActivationType:              types.ActivationTypeSlot,
				CollectFeeMode:              types.CollectFeeModeQuoteToken,
				MigrationFeeOption:          types.MigrationFeeOptionFixedBps100,
				TokenType:                   types.TokenTypeSPL,
				PartnerLockedLpPercentage:   100,
				CreatorTradingFeePercentage: 50,
				TokenUpdateAuthority:        uint8(types.TokenUpdateAuthorityOptionImmutable),
				MigrationFee: types.MigrationFee{
					FeePercentage:        1.5,
					CreatorFeePercentage: 50,
				},
			},
			InitialMarketCap:   1_000,
			MigrationMarketCap: 3_000,
		})
		if err != nil {
			t.Fatalf("BuildCurveWithMarketCap errored: %s", err.Error())
		}

		const threshold = 0.00001 //  0.001%

		assert.Equal(t, uint64(557399092), config.MigrationQuoteThreshold)

		assert.True(t,
			ApproxBigInt(config.SqrtStartPrice.BigInt(), big.NewInt(59222061406143610), threshold))

		assert.Equal(t, 2, len(config.Curve))

		assert.True(t,
			ApproxBigInt(config.Curve[0].SqrtPrice.BigInt(), big.NewInt(101036978416954629), threshold))

		liquidity0, _ := new(big.Int).SetString("4536014798171260106890790269146", 10)

		assert.True(t,
			ApproxBigInt(config.Curve[0].Liquidity.BigInt(), liquidity0, threshold))

		sqrtPrice1, _ := new(big.Int).SetString("79226673521066979257578248091", 10)
		liquidity1, _ := new(big.Int).SetString("132526870369400637538923", 10)

		assert.True(t,
			ApproxBigInt(config.Curve[1].SqrtPrice.BigInt(), sqrtPrice1, threshold))

		assert.True(t,
			ApproxBigInt(config.Curve[1].Liquidity.BigInt(), liquidity1, threshold))
	})
}

func TestBuildCurveWithTwoSegments(t *testing.T) {

	config, err := helpers.BuildCurveWithTwoSegments(types.BuildCurveWithTwoSegmentsParam{
		BuildCurveBaseParam: types.BuildCurveBaseParam{
			TotalTokenSupply:   1_000_000_000,
			MigrationOption:    types.MigrationOptionMET_DAMM_V2,
			TokenBaseDecimal:   types.TokenDecimalNINE,
			TokenQuoteDecimal:  types.TokenDecimalNINE,
			LockedVestingParam: types.LockedVestingParams{},
			BaseFeeParams: types.BaseFeeParams{
				BaseFeeMode: types.BaseFeeModeFeeSchedulerExponential,
				FeeSchedulerParam: &types.FeeSchedulerParams{
					StartingFeeBps: 5_000,
					EndingFeeBps:   100,
					NumberOfPeriod: 120,
					TotalDuration:  120,
				},
			},
			DynamicFeeEnabled:         true,
			ActivationType:            types.ActivationTypeSlot,
			CollectFeeMode:            types.CollectFeeModeQuoteToken,
			MigrationFeeOption:        types.MigrationFeeOptionFixedBps100,
			TokenType:                 types.TokenTypeSPL,
			PartnerLockedLpPercentage: 100,
			Leftover:                  350_000_000,
			MigrationFee: types.MigrationFee{
				FeePercentage:        10,
				CreatorFeePercentage: 50,
			},
		},
		InitialMarketCap:            20_000,
		MigrationMarketCap:          1_000_000,
		PercentageSupplyOnMigration: 20,
	})

	if err != nil {
		t.Fatalf("BuildCurveWithTwoSegments errored: %s", err.Error())
	}

	migrationQuoteThreshold := new(big.Int)
	c := new(big.Float).SetUint64(config.MigrationQuoteThreshold)
	c.Quo(c, big.NewFloat(math.Pow10(int(types.TokenDecimalNINE))))
	c.Int(migrationQuoteThreshold)

	const threshold = 0.00001 //  0.001%

	assert.Equal(t, uint64(222222), migrationQuoteThreshold.Uint64())

	assert.Equal(t, dbc.BaseFeeParameters{
		CliffFeeNumerator: 500_000_000,
		FirstFactor:       120,
		SecondFactor:      1,
		ThirdFactor:       320,
		BaseFeeMode:       uint8(types.BaseFeeModeFeeSchedulerExponential),
	}, config.PoolFees.BaseFee)

	assert.True(t, reflect.ValueOf(config.LockedVesting).IsZero())

	assert.True(t,
		ApproxBigInt(config.SqrtStartPrice.BigInt(), big.NewInt(82496347424711897), threshold))

	assert.Equal(t, 2, len(config.Curve))

	assert.True(t,
		ApproxBigInt(config.Curve[0].SqrtPrice.BigInt(), big.NewInt(357724324245030935), threshold))

	liquidity0, _ := new(big.Int).SetString("10943001382869720509000000000000000", 10)

	assert.True(t,
		ApproxBigInt(config.Curve[0].Liquidity.BigInt(), liquidity0, threshold))

	sqrtPrice1, _ := new(big.Int).SetString("583337266871351588", 10)
	liquidity1, _ := new(big.Int).SetString("321818787451021015000000000000000000", 10)

	assert.True(t,
		ApproxBigInt(config.Curve[1].SqrtPrice.BigInt(), sqrtPrice1, threshold))

	assert.True(t,
		ApproxBigInt(config.Curve[1].Liquidity.BigInt(), liquidity1, threshold))
}

func TestBBuildCurveWithLiquidityWeights(t *testing.T) {

	param := types.BuildCurveBaseParam{
		TotalTokenSupply:   1_000_000_000,
		MigrationOption:    types.MigrationOptionMET_DAMM_V2,
		TokenBaseDecimal:   types.TokenDecimalSIX,
		TokenQuoteDecimal:  types.TokenDecimalNINE,
		LockedVestingParam: types.LockedVestingParams{},
		BaseFeeParams: types.BaseFeeParams{
			BaseFeeMode: types.BaseFeeModeFeeSchedulerLinear,
			FeeSchedulerParam: &types.FeeSchedulerParams{
				StartingFeeBps: 100,
				EndingFeeBps:   100,
			},
		},
		DynamicFeeEnabled:         true,
		ActivationType:            types.ActivationTypeSlot,
		CollectFeeMode:            types.CollectFeeModeQuoteToken,
		MigrationFeeOption:        types.MigrationFeeOptionFixedBps100,
		TokenType:                 types.TokenTypeSPL,
		PartnerLockedLpPercentage: 100,
		Leftover:                  10_000,
		MigrationFee: types.MigrationFee{
			FeePercentage:        10,
			CreatorFeePercentage: 50,
		},
	}

	t.Run("build curve with liquidity weights 1.2^n", func(t *testing.T) {
		liquidityWeights := make([]float64, 0, 16)
		for i := range 16 {
			n := math.Pow(1.2, float64(i))
			liquidityWeights = append(liquidityWeights, n)
		}

		config, err := helpers.BuildCurveWithLiquidityWeights(types.BuildCurveWithLiquidityWeightsParam{
			BuildCurveBaseParam: param,
			InitialMarketCap:    30,
			MigrationMarketCap:  300,
			LiquidityWeights:    liquidityWeights,
		})

		if err != nil {
			t.Fatalf("BuildCurveWithLiquidityWeights errored: %s", err.Error())
		}

		migrationQuoteThreshold := new(big.Int)
		c := new(big.Float).SetUint64(config.MigrationQuoteThreshold)
		c.Quo(c, big.NewFloat(math.Pow10(int(types.TokenDecimalNINE))))
		c.Int(migrationQuoteThreshold)

		const threshold = 0.00001 //  0.001%

		assert.Equal(t, uint64(105), migrationQuoteThreshold.Uint64())

		assert.True(t,
			ApproxBigInt(config.SqrtStartPrice.BigInt(), big.NewInt(101036978416954620), threshold))

		assert.Equal(t, 16, len(config.Curve))

		testConfigCurve := []struct {
			sqrtPrice string
			liquidity string
		}{
			{"108575127956791431", "23962568427735005198000000000000"},
			{"116675682463356805", "28755082113282006238000000000000"},
			{"125380601749855484", "34506098535938407485000000000000"},
			{"134734976160032182", "41407318243126088982000000000000"},
			{"144787260130256668", "49688781891751306779000000000000"},
			{"155589523177168795", "59626538270101568134000000000000"},
			{"167197719610966654", "71551845924121881761000000000000"},
			{"179671978371417433", "85862215108946258113000000000000"},
			{"193076914487903615", "103034658130735509740000000000000"},
			{"207481963776826827", "123641589756882611680000000000000"},
			{"222961742510058138", "148369907708259134020000000000000"},
			{"239596433917470353", "178043889249910960820000000000000"},
			{"257472203525584873", "213652667099893152990000000000000"},
			{"276681645483733026", "256383200519871783590000000000000"},
			{"297324262189643010", "307659840623846140300000000000000"},
			{"319506979698850302", "369191808748615368360000000000000"},
		}
		var (
			sqrtPrice, liquidity *big.Int
			ok                   bool
		)

		for i := range testConfigCurve {
			if sqrtPrice, ok = new(big.Int).SetString(testConfigCurve[i].sqrtPrice, 10); !ok {
				t.Fatalf("cannot set value of sqrtPrice at index %d", i)
			}
			if liquidity, ok = new(big.Int).SetString(testConfigCurve[i].liquidity, 10); !ok {
				t.Fatalf("cannot set value of liquidity at index %d", i)
			}

			assert.True(t,
				ApproxBigInt(config.Curve[i].SqrtPrice.BigInt(), sqrtPrice, threshold))

			assert.True(t,
				ApproxBigInt(config.Curve[i].Liquidity.BigInt(), liquidity, threshold))
		}
	})

	t.Run("build curve with liquidity weights 0.6^n", func(t *testing.T) {
		liquidityWeights := make([]float64, 0, 16)
		for i := range 16 {
			n := math.Pow(0.6, float64(i))
			liquidityWeights = append(liquidityWeights, n)
		}

		config, err := helpers.BuildCurveWithLiquidityWeights(types.BuildCurveWithLiquidityWeightsParam{
			BuildCurveBaseParam: param,
			InitialMarketCap:    30,
			MigrationMarketCap:  300,
			LiquidityWeights:    liquidityWeights,
		})
		if err != nil {
			t.Fatalf("BuildCurveWithLiquidityWeights errored: %s", err.Error())
		}

		migrationQuoteThreshold := new(big.Int)
		c := new(big.Float).SetUint64(config.MigrationQuoteThreshold)
		c.Quo(c, big.NewFloat(math.Pow10(int(types.TokenDecimalNINE))))
		c.Int(migrationQuoteThreshold)

		const threshold = 0.00001 //  0.001%

		assert.Equal(t, uint64(35), migrationQuoteThreshold.Uint64())

		assert.True(t,
			ApproxBigInt(config.SqrtStartPrice.BigInt(), big.NewInt(101036978416954620), threshold))

		assert.Equal(t, 16, len(config.Curve))

		testConfigCurve := []struct {
			sqrtPrice string
			liquidity string
		}{
			{"108575127956791431", "573839297089455533520000000000000"},
			{"116675682463356805", "344303578253673320110000000000000"},
			{"125380601749855484", "206582146952203992070000000000000"},
			{"134734976160032182", "123949288171322395240000000000000"},
			{"144787260130256668", "74369572902793437144000000000000"},
			{"155589523177168795", "44621743741676062287000000000000"},
			{"167197719610966654", "26773046245005637372000000000000"},
			{"179671978371417433", "16063827747003382423000000000000"},
			{"193076914487903615", "9638296648202029453900000000000"},
			{"207481963776826827", "5782977988921217672300000000000"},
			{"222961742510058138", "3469786793352730603400000000000"},
			{"239596433917470353", "2081872076011638362000000000000"},
			{"257472203525584873", "1249123245606983017200000000000"},
			{"276681645483733026", "749473947364189810330000000000"},
			{"297324262189643010", "449684368418513886200000000000"},
			{"319506979698850302", "269810621051108331720000000000"},
		}
		var (
			sqrtPrice, liquidity *big.Int
			ok                   bool
		)

		for i := range testConfigCurve {
			if sqrtPrice, ok = new(big.Int).SetString(testConfigCurve[i].sqrtPrice, 10); !ok {
				t.Fatalf("cannot set value of sqrtPrice at index %d", i)
			}
			if liquidity, ok = new(big.Int).SetString(testConfigCurve[i].liquidity, 10); !ok {
				t.Fatalf("cannot set value of liquidity at index %d", i)
			}

			assert.True(t,
				ApproxBigInt(config.Curve[i].SqrtPrice.BigInt(), sqrtPrice, threshold))

			assert.True(t,
				ApproxBigInt(config.Curve[i].Liquidity.BigInt(), liquidity, threshold))
		}
	})

	t.Run("build curve with liquidity weights v1", func(t *testing.T) {
		liquidityWeights := make([]float64, 0, 16)
		for i := range 16 {
			n := math.Pow(1.2, float64(i))
			liquidityWeights = append(liquidityWeights, n)
		}
		liquidityWeights[15] = 80

		param := param // copy
		param.TokenBaseDecimal = types.TokenDecimalNINE
		param.TokenQuoteDecimal = types.TokenDecimalSIX
		param.LockedVestingParam = types.LockedVestingParams{
			TotalLockedVestingAmount: 10_000_000,
			NumberOfVestingPeriod:    1,
			TotalVestingDuration:     1,
		}
		param.Leftover = 200_000_000
		param.MigrationOption = types.MigrationOptionMET_DAMM

		config, err := helpers.BuildCurveWithLiquidityWeights(types.BuildCurveWithLiquidityWeightsParam{
			BuildCurveBaseParam: param,
			InitialMarketCap:    15,
			MigrationMarketCap:  255,
			LiquidityWeights:    liquidityWeights,
		})

		if err != nil {
			t.Fatalf("BuildCurveWithLiquidityWeights errored: %s", err.Error())
		}

		migrationQuoteThreshold := new(big.Int)
		c := new(big.Float).SetUint64(config.MigrationQuoteThreshold)
		c.Quo(c, big.NewFloat(math.Pow10(int(types.TokenDecimalSIX))))
		c.Int(migrationQuoteThreshold)

		const threshold = 0.00001 //  0.001%

		assert.Equal(t, uint64(77), migrationQuoteThreshold.Uint64())

		assert.True(t,
			ApproxBigInt(config.SqrtStartPrice.BigInt(), big.NewInt(71443932589227), threshold))

		assert.Equal(t, 16, len(config.Curve))

		testConfigCurve := []struct {
			sqrtPrice string
			liquidity string
		}{
			{"78057903160464", "8260393602760946085000000000000"},
			{"85284166548345", "9912472323313135302000000000000"},
			{"93179406176130", "11894966787975762362000000000000"},
			{"101805552973475", "14273960145570914835000000000000"},
			{"111230271167902", "17128752174685097802000000000000"},
			{"121527489048741", "20554502609622117362000000000000"},
			{"132777978866906", "24665403131546540835000000000000"},
			{"145069990419284", "29598483757855849002000000000000"},
			{"158499943287633", "35518180509427018802000000000000"},
			{"173173183161963", "42621816611312422562000000000000"},
			{"189204808181069", "51146179933574907075000000000000"},
			{"206720571772097", "61375415920289888490000000000000"},
			{"225857869071101", "73650499104347866188000000000000"},
			{"246766814662150", "88380598925217439425000000000000"},
			{"269611420088862", "106056718710260927310000000000000"},
			{"294570880374892", "660831488220875686800000000000000"},
		}
		var (
			sqrtPrice, liquidity *big.Int
			ok                   bool
		)

		for i := range testConfigCurve {
			if sqrtPrice, ok = new(big.Int).SetString(testConfigCurve[i].sqrtPrice, 10); !ok {
				t.Fatalf("cannot set value of sqrtPrice at index %d", i)
			}
			if liquidity, ok = new(big.Int).SetString(testConfigCurve[i].liquidity, 10); !ok {
				t.Fatalf("cannot set value of liquidity at index %d", i)
			}

			assert.True(t,
				ApproxBigInt(config.Curve[i].SqrtPrice.BigInt(), sqrtPrice, threshold))

			assert.True(t,
				ApproxBigInt(config.Curve[i].Liquidity.BigInt(), liquidity, threshold))
		}
	})

	t.Run("build curve with liquidity weights v2", func(t *testing.T) {
		liquidityWeights := []float64{0.01, 0.02, 0.04, 0.08, 0.16, 0.32, 0.64, 1.28, 2.56, 5.12, 10.24,
			20.48, 40.96, 81.92, 163.84, 327.68}

		param := param // copy
		param.TotalTokenSupply = 100_000_000
		param.TokenBaseDecimal = types.TokenDecimalSIX
		param.TokenQuoteDecimal = types.TokenDecimalSIX
		param.Leftover = 50_000_000
		param.MigrationOption = types.MigrationOptionMET_DAMM

		config, err := helpers.BuildCurveWithLiquidityWeights(types.BuildCurveWithLiquidityWeightsParam{
			BuildCurveBaseParam: param,
			InitialMarketCap:    50,
			MigrationMarketCap:  100_000,
			LiquidityWeights:    liquidityWeights,
		})
		if err != nil {
			t.Fatalf("BuildCurveWithLiquidityWeights errored: %s", err.Error())
		}
		migrationQuoteThreshold := new(big.Int)
		c := new(big.Float).SetUint64(config.MigrationQuoteThreshold)
		c.Quo(c, big.NewFloat(math.Pow10(int(types.TokenDecimalSIX))))
		c.Int(migrationQuoteThreshold)

		const threshold = 0.00001 //  0.001%

		assert.Equal(t, uint64(16680), migrationQuoteThreshold.Uint64())

		assert.True(t,
			ApproxBigInt(config.SqrtStartPrice.BigInt(), big.NewInt(13043817825332782), threshold))

		assert.Equal(t, 16, len(config.Curve))

		testConfigCurve := []struct {
			sqrtPrice string
			liquidity string
		}{
			{"16541005727558956", "850713289535804874010000000"},
			{"20975827333908492", "1701426579071609748000000000"},
			{"26599672328804254", "3402853158143219496000000000"},
			{"33731330675857355", "6805706316286438992100000000"},
			{"42775063357902031", "13611412632572877984000000000"},
			{"54243518076863625", "27222825265145755968000000000"},
			{"68786788899319668", "54445650530291511937000000000"},
			{"87229267105699335", "108891301060583023870000000000"},
			{"110616372148645188", "217782602121166047750000000000"},
			{"140273811684107333", "435565204242332095490000000000"},
			{"177882729854374496", "871130408484664190990000000000"},
			{"225575003634333864", "1742260816969328382000000000000"},
			{"286054089153491780", "3484521633938656763900000000000"},
			{"362748268217158410", "6969043267877313527900000000000"},
			{"460004981868797949", "13938086535754627056000000000000"},
			{"583337266871351588", "27876173071509254112000000000000"},
		}
		var (
			sqrtPrice, liquidity *big.Int
			ok                   bool
		)

		for i := range testConfigCurve {
			if sqrtPrice, ok = new(big.Int).SetString(testConfigCurve[i].sqrtPrice, 10); !ok {
				t.Fatalf("cannot set value of sqrtPrice at index %d", i)
			}
			if liquidity, ok = new(big.Int).SetString(testConfigCurve[i].liquidity, 10); !ok {
				t.Fatalf("cannot set value of liquidity at index %d", i)
			}

			assert.True(t,
				ApproxBigInt(config.Curve[i].SqrtPrice.BigInt(), sqrtPrice, threshold))

			assert.True(t,
				ApproxBigInt(config.Curve[i].Liquidity.BigInt(), liquidity, threshold))
		}
	})
}

func ApproxBigFloat(a, b *big.Float, threshold float64) bool {
	zero := big.NewFloat(0)

	if a.Sign() == 0 && b.Sign() == 0 {
		return true
	}
	if a.Cmp(zero) == 0 || b.Cmp(zero) == 0 {
		return false
	}

	diff := new(big.Float).Sub(a, b)
	diff.Abs(diff)

	max := new(big.Float)
	if a.Cmp(b) >= 0 {
		max.Copy(a)
	} else {
		max.Copy(b)
	}

	ratio := new(big.Float).Quo(diff, max)
	thresholdF := big.NewFloat(threshold)

	return ratio.Cmp(thresholdF) <= 0
}

func ApproxBigInt(a, b *big.Int, threshold float64) bool {
	if a.Sign() == 0 && b.Sign() == 0 {
		return true
	}

	if a.Sign() == 0 || b.Sign() == 0 {
		return false
	}

	diff := new(big.Int).Sub(a, b).Abs(new(big.Int).Sub(a, b))

	max := new(big.Int)
	if a.Cmp(b) >= 0 {
		max.Set(a)
	} else {
		max.Set(b)
	}

	diffF := new(big.Float).SetInt(diff)
	maxF := new(big.Float).SetInt(max)
	thresholdF := big.NewFloat(threshold)

	// ratio = |a - b| / max(a, b)
	ratio := new(big.Float).Quo(diffF, maxF)
	return ratio.Cmp(thresholdF) <= 0
}

// approxOrdered handles ordered numeric types (int, float64, etc.)
func ApproxOrdered(a, b any, threshold float64) bool {
	aF := toFloat64(a)
	bF := toFloat64(b)

	if aF == 0 && bF == 0 {
		return true
	}
	if aF == 0 || bF == 0 {
		return false
	}

	diff := absFloat64(aF - bF)
	max := aF
	if bF > max {
		max = bF
	}
	ratio := diff / max

	return ratio <= threshold
}

func absFloat64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func toFloat64(x any) float64 {
	switch v := x.(type) {
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	case uint:
		return float64(v)
	case uint64:
		return float64(v)
	default:
		return 0
	}
}
