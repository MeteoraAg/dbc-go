package helpers

import (
	"dbcGoSDK/constants"
	"dbcGoSDK/generated/dbc"
	"dbcGoSDK/types"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/ALTree/bigfloat"
)

// BuildCurve builds a custom constant product curve.
func BuildCurve(param types.BuildCurveParam) (dbc.ConfigParameters, error) {

	migrationBaseSupply := new(big.Float).Quo(
		new(big.Float).Mul(
			new(big.Float).SetUint64(param.TotalTokenSupply),
			big.NewFloat(param.PercentageSupplyOnMigration),
		),
		constants.HundredInBigFloat,
	)

	migrationQuoteAmount := GetMigrationQuoteAmountFromMigrationQuoteThreshold(
		big.NewFloat(param.MigrationQuoteThreshold),
		param.MigrationFee.FeePercentage,
	)

	migrationPrice := new(big.Float).Quo(
		migrationQuoteAmount, migrationBaseSupply,
	)

	migrateSqrtPrice := GetSqrtPriceFromPrice(
		migrationPrice,
		param.TokenBaseDecimal,
		param.TokenQuoteDecimal,
	)

	migrationQuoteAmountInLamportBigFloat := new(big.Float).Mul(
		migrationQuoteAmount,
		new(big.Float).SetFloat64(math.Pow10(int(param.TokenQuoteDecimal))),
	)
	migrationQuoteAmountInLamport, _ := migrationQuoteAmountInLamportBigFloat.Int(nil)

	migrationBaseAmount, err := GetMigrationBaseToken(
		migrationQuoteAmountInLamport,
		migrateSqrtPrice,
		param.MigrationOption,
	)
	if err != nil {
		return dbc.ConfigParameters{}, err
	}

	totalSupply := ConvertToLamports(float64(param.TotalTokenSupply), param.TokenBaseDecimal)
	if !totalSupply.IsUint64() {
		return dbc.ConfigParameters{},
			fmt.Errorf("cannot fit totalSupply(%s) into uint64", totalSupply)
	}

	totalLeftover := ConvertToLamports(
		float64(param.Leftover), param.TokenBaseDecimal,
	)

	lockedVesting, err := GetLockedVestingParams(
		param.LockedVestingParam.TotalLockedVestingAmount,
		param.LockedVestingParam.NumberOfVestingPeriod,
		param.LockedVestingParam.CliffUnlockAmount,
		param.LockedVestingParam.TotalVestingDuration,
		param.LockedVestingParam.CliffDurationFromMigrationTime,
		param.TokenBaseDecimal,
	)
	if err != nil {
		return dbc.ConfigParameters{}, err
	}

	totalVestingAmount := GetTotalVestingAmount(
		dbc.LockedVestingParams{
			AmountPerPeriod:                lockedVesting.AmountPerPeriod,
			CliffDurationFromMigrationTime: lockedVesting.CliffDurationFromMigrationTime,
			Frequency:                      lockedVesting.Frequency,
			NumberOfPeriod:                 lockedVesting.NumberOfPeriod,
			CliffUnlockAmount:              lockedVesting.CliffUnlockAmount,
		},
	)

	swapAmount := new(big.Int).Sub(
		new(big.Int).Sub(
			new(big.Int).Sub(totalSupply, migrationBaseAmount),
			totalVestingAmount,
		),
		totalLeftover,
	)

	migrationQuoteThresholdInLamport := ConvertToLamports(
		param.MigrationQuoteThreshold, param.TokenQuoteDecimal,
	)

	if !migrationQuoteThresholdInLamport.IsUint64() {
		return dbc.ConfigParameters{},
			fmt.Errorf("cannot fit migrationQuoteThresholdInLamport(%s) as uint64", migrationQuoteThresholdInLamport)
	}

	firstCurve, err := GetFirstCurve(
		migrateSqrtPrice,
		migrationBaseAmount,
		swapAmount,
		migrationQuoteThresholdInLamport,
		param.MigrationFee.FeePercentage,
	)
	if err != nil {
		return dbc.ConfigParameters{}, err
	}

	totalDynamicSupply, err := GetTotalSupplyFromCurve(
		migrationQuoteThresholdInLamport,
		firstCurve.SqrtStartPrice,
		firstCurve.Curve,
		lockedVesting,
		param.MigrationOption,
		totalLeftover,
		param.MigrationFee.FeePercentage,
	)
	if err != nil {
		return dbc.ConfigParameters{}, err
	}

	remainingAmount := new(big.Int).Sub(totalSupply, totalDynamicSupply)

	lastLiquidity, err := GetInitialLiquidityFromDeltaBase(
		remainingAmount,
		constants.MaxSqrtPrice,
		migrateSqrtPrice,
	)
	if err != nil {
		return dbc.ConfigParameters{}, err
	}

	if lastLiquidity.Sign() != 0 {
		firstCurve.Curve = append(firstCurve.Curve, dbc.LiquidityDistributionParameters{
			SqrtPrice: MustBigIntToUint128(constants.MaxSqrtPrice),
			Liquidity: MustBigIntToUint128(lastLiquidity),
		})
	}

	baseFee, err := GetBaseFeeParams(
		param.BaseFeeParams,
		param.TokenQuoteDecimal,
		param.ActivationType,
	)
	if err != nil {
		return dbc.ConfigParameters{}, err
	}

	var dynamicField *dbc.DynamicFeeParameters
	if param.DynamicFeeEnabled {
		baseFeeBp := param.BaseFeeParams.FeeSchedulerParam.EndingFeeBps
		if param.BaseFeeParams.BaseFeeMode == types.BaseFeeModeRateLimiter {
			baseFeeBp = param.BaseFeeParams.RateLimiterParam.BaseFeeBps
		}
		d, err := GetDynamicFeeParams(baseFeeBp, 0)
		if err != nil {
			return dbc.ConfigParameters{}, err
		}
		dynamicField = &d
	}

	return dbc.ConfigParameters{
		PoolFees: dbc.PoolFeeParameters{
			BaseFee: dbc.BaseFeeParameters{
				CliffFeeNumerator: baseFee.CliffFeeNumerator,
				FirstFactor:       baseFee.FirstFactor,
				SecondFactor:      baseFee.SecondFactor,
				ThirdFactor:       baseFee.ThirdFactor,
				BaseFeeMode:       uint8(baseFee.BaseFeeMode),
			},
			DynamicFee: dynamicField,
		},
		ActivationType:            uint8(param.ActivationType),
		CollectFeeMode:            uint8(param.CollectFeeMode),
		MigrationOption:           uint8(param.MigrationFeeOption),
		TokenType:                 uint8(param.TokenType),
		TokenDecimal:              uint8(param.TokenBaseDecimal),
		MigrationQuoteThreshold:   migrationQuoteThresholdInLamport.Uint64(),
		PartnerLpPercentage:       param.PartnerLpPercentage,
		CreatorLpPercentage:       param.CreatorLpPercentage,
		PartnerLockedLpPercentage: param.PartnerLockedLpPercentage,
		CreatorLockedLpPercentage: param.CreatorLockedLpPercentage,
		SqrtStartPrice:            MustBigIntToUint128(firstCurve.SqrtStartPrice),
		LockedVesting:             lockedVesting,
		MigrationFeeOption:        uint8(param.MigrationFeeOption),
		TokenSupply: &dbc.TokenSupplyParams{
			PreMigrationTokenSupply:  totalSupply.Uint64(),
			PostMigrationTokenSupply: totalSupply.Uint64(),
		},
		CreatorTradingFeePercentage: param.CreatorTradingFeePercentage,
		TokenUpdateAuthority:        param.TokenUpdateAuthority,
		MigrationFee: dbc.MigrationFee{
			FeePercentage:        uint8(param.MigrationFee.FeePercentage),
			CreatorFeePercentage: uint8(param.MigrationFee.CreatorFeePercentage),
		},
		MigratedPoolFee: GetMigratedPoolFeeParams(
			param.MigrationOption,
			param.MigrationFeeOption,
			param.MigratedPoolFee,
		),
		Curve: firstCurve.Curve,
	}, nil

}

func BuildCurveWithMarketCap(
	param types.BuildCurveWithMarketCapParam,
) (dbc.ConfigParameters, error) {

	lockedVesting, err := GetLockedVestingParams(
		param.LockedVestingParam.TotalLockedVestingAmount,
		param.LockedVestingParam.NumberOfVestingPeriod,
		param.LockedVestingParam.CliffUnlockAmount,
		param.LockedVestingParam.TotalVestingDuration,
		param.LockedVestingParam.CliffDurationFromMigrationTime,
		param.TokenBaseDecimal,
	)
	if err != nil {
		return dbc.ConfigParameters{}, err
	}

	totalLeftover := ConvertToLamports(
		float64(param.Leftover), param.TokenBaseDecimal,
	)

	totalSupply := ConvertToLamports(float64(param.TotalTokenSupply), param.TokenBaseDecimal)
	if !totalSupply.IsUint64() {
		return dbc.ConfigParameters{},
			fmt.Errorf("cannot fit totalSupply(%s) into uint64", totalSupply)
	}

	percentageSupplyOnMigration := GetPercentageSupplyOnMigration(
		big.NewFloat(param.InitialMarketCap),
		big.NewFloat(param.MigrationMarketCap),
		lockedVesting,
		totalLeftover,
		totalSupply,
	)

	migrationQuoteAmount := GetMigrationQuoteAmount(
		big.NewFloat(param.MigrationMarketCap),
		big.NewFloat(percentageSupplyOnMigration),
	)

	migrationQuoteThreshold := GetMigrationQuoteThresholdFromMigrationQuoteAmount(
		migrationQuoteAmount,
		param.MigrationFee.FeePercentage,
	)

	migrationQuoteThresholdFloat64, _ := migrationQuoteThreshold.Float64()
	return BuildCurve(
		types.BuildCurveParam{
			BuildCurveBaseParam:         param.BuildCurveBaseParam,
			PercentageSupplyOnMigration: percentageSupplyOnMigration,
			MigrationQuoteThreshold:     migrationQuoteThresholdFloat64,
		},
	)
}

// BuildCurveWithTwoSegments builds a custom constant product curve by market cap.
func BuildCurveWithTwoSegments(
	param types.BuildCurveWithTwoSegmentsParam,
) (dbc.ConfigParameters, error) {

	totalSupply := ConvertToLamports(float64(param.TotalTokenSupply), param.TokenBaseDecimal)
	if !totalSupply.IsUint64() {
		return dbc.ConfigParameters{},
			fmt.Errorf("BuildCurveWithTwoSegments:cannot fit totalSupply(%s) into uint64", totalSupply)
	}

	migrationBaseSupply := new(big.Int).Quo(
		new(big.Int).Mul(
			new(big.Int).SetUint64(uint64(param.PercentageSupplyOnMigration)),
			new(big.Int).SetUint64(param.TotalTokenSupply),
		),
		big.NewInt(100),
	)

	migrationQuoteAmount := GetMigrationQuoteAmount(
		new(big.Float).SetUint64(param.MigrationMarketCap),
		new(big.Float).SetUint64(uint64(param.PercentageSupplyOnMigration)),
	)

	migrationQuoteThreshold := GetMigrationQuoteThresholdFromMigrationQuoteAmount(
		migrationQuoteAmount,
		param.MigrationFee.FeePercentage,
	)

	migrationPrice := new(big.Float).Quo(
		migrationQuoteAmount, new(big.Float).SetInt(migrationBaseSupply))

	migrationQuoteThresholdInLamport := new(big.Int)
	new(big.Float).Mul(migrationQuoteThreshold,
		big.NewFloat(math.Pow10(int(param.TokenQuoteDecimal)))).Int(migrationQuoteThresholdInLamport)

	migrationQuoteAmountInLamport := new(big.Int)
	new(big.Float).Mul(migrationQuoteAmount,
		big.NewFloat(math.Pow10(int(param.TokenQuoteDecimal)))).Int(migrationQuoteAmountInLamport)

	migrateSqrtPrice := GetSqrtPriceFromPrice(
		migrationPrice,
		param.TokenBaseDecimal,
		param.TokenQuoteDecimal,
	)

	migrationBaseAmount, err := GetMigrationBaseToken(
		migrationQuoteAmountInLamport,
		migrateSqrtPrice,
		param.MigrationOption,
	)
	if err != nil {
		return dbc.ConfigParameters{}, err
	}

	lockedVesting, err := GetLockedVestingParams(
		param.LockedVestingParam.TotalLockedVestingAmount,
		param.LockedVestingParam.NumberOfVestingPeriod,
		param.LockedVestingParam.CliffUnlockAmount,
		param.LockedVestingParam.TotalVestingDuration,
		param.LockedVestingParam.CliffDurationFromMigrationTime,
		param.TokenBaseDecimal,
	)
	if err != nil {
		return dbc.ConfigParameters{}, err
	}

	totalVestingAmount := GetTotalVestingAmount(lockedVesting)
	totalLeftover := ConvertToLamports(float64(param.Leftover), param.TokenBaseDecimal)

	initialSqrtPrice := GetSqrtPriceFromMarketCap(
		param.InitialMarketCap,
		param.TotalTokenSupply,
		param.TokenBaseDecimal,
		param.TokenQuoteDecimal,
	)

	// instantiate midSqrtPriceDecimal1
	midSqrtPriceDecimal1 := new(big.Float).Sqrt(
		new(big.Float).SetInt(
			new(big.Int).Mul(migrateSqrtPrice, initialSqrtPrice),
		),
	)

	midSqrtPrice1 := new(big.Int)
	midSqrtPriceDecimal1.Int(midSqrtPrice1)

	// mid_price2 = (p1 * p2^3)^(1/4)
	numerator1, numerator2 := new(big.Float).SetInt(initialSqrtPrice),
		new(big.Float).SetInt(new(big.Int).Exp(migrateSqrtPrice, big.NewInt(3), nil))
	product1 := numerator1.Mul(numerator1, numerator2) // stills points to numerator1

	// midSqrtPriceDecimal2, err := decimal.NewFromString(product1.String())
	// if err != nil {
	// 	return dbc.ConfigParameters{}, err
	// }
	// midSqrtPriceDecimal2 = midSqrtPriceDecimal2.Pow(decimal.NewFromFloat32(0.25))
	// midSqrtPrice2 := midSqrtPriceDecimal2.BigInt()
	f64, _ := product1.Float64()
	midSqrtPrice2 := new(big.Int).SetInt64(int64(math.Floor(math.Pow(f64, 0.25))))

	// mid_price3 = (p1^3 * p2)^(1/4)
	// numerator3, err := decimal.NewFromString(initialSqrtPrice.String())
	// if err != nil {
	// 	return dbc.ConfigParameters{}, err
	// }
	// numerator3 = numerator3.Pow(decimal.NewFromInt(3))
	numerator3 := new(big.Int).Exp(initialSqrtPrice, big.NewInt(3), nil)
	// numerator4, err := decimal.NewFromString(migrateSqrtPrice.String())
	// if err != nil {
	// 	return dbc.ConfigParameters{}, err
	// }
	numerator4 := new(big.Int).Set(migrateSqrtPrice)
	product2 := numerator3.Mul(numerator3, numerator4) // still points to numerator3

	// midSqrtPriceDecimal3 := product2.Pow(decimal.NewFromFloat32(0.25))
	// midSqrtPrice3 := midSqrtPriceDecimal3.BigInt()
	f64, _ = product2.Float64()
	midSqrtPrice3 := new(big.Int).SetInt64(int64(math.Floor(math.Pow(f64, 0.25))))

	swapAmount := new(big.Int).Sub(
		new(big.Int).Sub(
			new(big.Int).Sub(
				totalSupply, migrationBaseAmount),
			totalVestingAmount,
		),
		totalLeftover,
	)

	var (
		curve          []dbc.LiquidityDistributionParameters
		midPrices      = [3]*big.Int{midSqrtPrice1, midSqrtPrice2, midSqrtPrice3}
		sqrtStartPrice = big.NewInt(0)
	)

	for i := range len(midPrices) {
		result := GetTwoCurve(
			migrateSqrtPrice,
			midPrices[i],
			initialSqrtPrice,
			swapAmount,
			migrationQuoteThresholdInLamport,
		)
		if result.IsoK {
			curve = result.TwoCurve.Curve
			sqrtStartPrice = result.TwoCurve.SqrtStartPrice
			break
		}
	}

	totalDynamicSupply, err := GetTotalSupplyFromCurve(
		migrationQuoteThresholdInLamport,
		sqrtStartPrice,
		curve,
		lockedVesting,
		param.MigrationOption,
		totalLeftover,
		param.MigrationFee.FeePercentage,
	)
	if err != nil {
		return dbc.ConfigParameters{}, err
	}

	if totalDynamicSupply.Cmp(totalSupply) > 0 {
		// precision loss is used for leftover
		if leftOverDelta := new(big.Int).Sub(
			totalDynamicSupply, totalSupply); !(leftOverDelta.Cmp(totalLeftover) < 0) {
			return dbc.ConfigParameters{}, errors.New(
				"leftOverDelta must be less than totalLeftover",
			)
		}
	}

	baseFee, err := GetBaseFeeParams(
		param.BaseFeeParams,
		param.TokenQuoteDecimal,
		param.ActivationType,
	)
	if err != nil {
		return dbc.ConfigParameters{}, err
	}

	var dynamicField *dbc.DynamicFeeParameters
	if param.DynamicFeeEnabled {
		baseFeeBp := param.BaseFeeParams.FeeSchedulerParam.EndingFeeBps
		if param.BaseFeeParams.BaseFeeMode == types.BaseFeeModeRateLimiter {
			baseFeeBp = param.BaseFeeParams.RateLimiterParam.BaseFeeBps
		}

		d, err := GetDynamicFeeParams(baseFeeBp, 0)
		if err != nil {
			return dbc.ConfigParameters{}, err
		}

		dynamicField = &d
	}

	return dbc.ConfigParameters{
		PoolFees: dbc.PoolFeeParameters{
			BaseFee: dbc.BaseFeeParameters{
				CliffFeeNumerator: baseFee.CliffFeeNumerator,
				FirstFactor:       baseFee.FirstFactor,
				SecondFactor:      baseFee.SecondFactor,
				ThirdFactor:       baseFee.ThirdFactor,
				BaseFeeMode:       uint8(baseFee.BaseFeeMode),
			},
			DynamicFee: dynamicField,
		},
		ActivationType:            uint8(param.ActivationType),
		CollectFeeMode:            uint8(param.CollectFeeMode),
		MigrationOption:           uint8(param.MigrationFeeOption),
		TokenType:                 uint8(param.TokenType),
		TokenDecimal:              uint8(param.TokenBaseDecimal),
		MigrationQuoteThreshold:   migrationQuoteThresholdInLamport.Uint64(),
		PartnerLpPercentage:       param.PartnerLpPercentage,
		CreatorLpPercentage:       param.CreatorLpPercentage,
		PartnerLockedLpPercentage: param.PartnerLockedLpPercentage,
		CreatorLockedLpPercentage: param.CreatorLockedLpPercentage,
		SqrtStartPrice:            MustBigIntToUint128(sqrtStartPrice),
		LockedVesting:             lockedVesting,
		MigrationFeeOption:        uint8(param.MigrationFeeOption),
		TokenSupply: &dbc.TokenSupplyParams{
			PreMigrationTokenSupply:  totalSupply.Uint64(),
			PostMigrationTokenSupply: totalSupply.Uint64(),
		},
		CreatorTradingFeePercentage: param.CreatorTradingFeePercentage,
		MigratedPoolFee: GetMigratedPoolFeeParams(
			param.MigrationOption,
			param.MigrationFeeOption,
			param.MigratedPoolFee,
		),
		TokenUpdateAuthority: param.TokenUpdateAuthority,
		MigrationFee: dbc.MigrationFee{
			FeePercentage:        uint8(param.MigrationFee.FeePercentage),
			CreatorFeePercentage: uint8(param.MigrationFee.CreatorFeePercentage),
		},
		Curve: curve,
	}, nil
}

func PowBigFloat(x, y *big.Float) *big.Float {
	// Uses ln/exp for power calc
	lx := new(big.Float).SetPrec(256)
	lx, _ = lx.SetString(fmt.Sprintf("%g", math.Log(float64FromBigFloat(x))))
	result := math.Exp(float64FromBigFloat(y) * float64FromBigFloat(lx))
	return new(big.Float).SetPrec(256).SetFloat64(result)
}

func float64FromBigFloat(b *big.Float) float64 {
	f, _ := b.Float64()
	return f
}

func BuildCurveWithLiquidityWeights(
	param types.BuildCurveWithLiquidityWeightsParam,
) (dbc.ConfigParameters, error) {

	// 1. finding Pmax and Pmin
	pMin := GetSqrtPriceFromMarketCap(
		param.InitialMarketCap,
		param.TotalTokenSupply,
		param.TokenBaseDecimal,
		param.TokenQuoteDecimal,
	)

	pMax := GetSqrtPriceFromMarketCap(
		param.MigrationMarketCap,
		param.TotalTokenSupply,
		param.TokenBaseDecimal,
		param.TokenQuoteDecimal,
	)

	// find q^16 = pMax / pMin
	priceRatio := new(big.Float).Quo(
		new(big.Float).SetInt(pMax),
		new(big.Float).SetInt(pMin),
	)

	qDecimal := bigfloat.Pow(priceRatio, big.NewFloat(float64(1)/float64(16)))
	sqrtPrices, currentPrice := make([]*big.Float, 0, 17), new(big.Float).SetInt(pMin)

	// finding all prices
	for range 17 {
		sqrtPrices = append(sqrtPrices, currentPrice)
		currentPrice = new(big.Float).Mul(qDecimal, currentPrice)
	}

	lockedVesting, err := GetLockedVestingParams(
		param.LockedVestingParam.TotalLockedVestingAmount,
		param.LockedVestingParam.NumberOfVestingPeriod,
		param.LockedVestingParam.CliffUnlockAmount,
		param.LockedVestingParam.TotalVestingDuration,
		param.LockedVestingParam.CliffDurationFromMigrationTime,
		param.TokenBaseDecimal,
	)
	if err != nil {
		return dbc.ConfigParameters{}, err
	}

	totalSupply, totalLeftover, totalVestingAmount :=
		ConvertToLamports(float64(param.TotalTokenSupply), param.TokenBaseDecimal),
		ConvertToLamports(float64(param.Leftover), param.TokenBaseDecimal),
		GetTotalVestingAmount(lockedVesting)

	if !totalSupply.IsUint64() || !totalLeftover.IsUint64() {
		return dbc.ConfigParameters{}, fmt.Errorf(
			"either totalSupply(%s) or totalLeftover(%s) cannot fit into uint64",
			totalSupply, totalLeftover,
		)
	}

	// Swap_Amount = sum(li * (1/p(i-1) - 1/pi))
	// Quote_Amount = sum(li * (pi-p(i-1)))
	// Quote_Amount * (1-migrationFee/100) / Base_Amount = Pmax ^ 2

	// -> Base_Amount = Quote_Amount * (1-migrationFee) / Pmax ^ 2
	// -> Swap_Amount + Base_Amount = sum(li * (1/p(i-1) - 1/pi)) + sum(li * (pi-p(i-1))) * (1-migrationFee/100) / Pmax ^ 2
	// l0 * sum_factor = Swap_Amount + Base_Amount
	// => l0 * sum_factor = sum(li * (1/p(i-1) - 1/pi)) + sum(li * (pi-p(i-1))) * (1-migrationFee/100) / Pmax ^ 2
	// => l0 = (Swap_Amount + Base_Amount ) / sum_factor

	sumFactor, pmaxWeight, migrationFeeFactor := big.NewFloat(0), new(big.Float).SetInt(pMax),
		new(big.Float).Quo(
			new(big.Float).SetPrec(20).Sub(big.NewFloat(100), big.NewFloat(param.MigrationFee.FeePercentage)),
			big.NewFloat(100),
		)

	if l := len(param.LiquidityWeights); l < 16 {
		return dbc.ConfigParameters{},
			fmt.Errorf("len of param.LiquidityWeights is expected to be >= 16, len is %d", l)
	}

	for i := 1; i < 17; i++ {
		pi, piMinus, k := sqrtPrices[i], sqrtPrices[i-1],
			big.NewFloat(param.LiquidityWeights[i-1])
		w1 := new(big.Float).Quo(
			new(big.Float).Sub(pi, piMinus),
			new(big.Float).Mul(pi, piMinus),
		)
		w2 := new(big.Float).Quo(
			new(big.Float).Mul(migrationFeeFactor, new(big.Float).Sub(pi, piMinus)),
			new(big.Float).Mul(pmaxWeight, pmaxWeight),
		)

		weight := new(big.Float).Mul(k, new(big.Float).Add(w1, w2))
		sumFactor = new(big.Float).Add(sumFactor, weight)
	}

	totalSwapAndMigrationAmount := new(big.Int).Sub(
		new(big.Int).Sub(totalSupply, totalVestingAmount),
		totalLeftover,
	)

	if sumFactor.Sign() == 0 {
		return dbc.ConfigParameters{}, errors.New("sumFactor cannot be zero")
	}

	l1 := new(big.Float).Quo(
		new(big.Float).SetInt(totalSwapAndMigrationAmount),
		sumFactor,
	)

	// construct curve
	curve := make([]dbc.LiquidityDistributionParameters, 0, 16)
	for i := range 16 {
		k := big.NewFloat(param.LiquidityWeights[i])
		liquidity, _ := new(big.Float).Mul(l1, k).Int(nil)
		sqrtPrice := new(big.Int).Set(pMax)
		if i < 15 {
			sqrtPrice, _ = sqrtPrices[i+1].Int(nil)
		}
		curve = append(curve, dbc.LiquidityDistributionParameters{
			SqrtPrice: MustBigIntToUint128(sqrtPrice),
			Liquidity: MustBigIntToUint128(liquidity),
		})
	}

	// reverse to calculate swap amount and migration amount
	swapBaseAmount, err := GetBaseTokenForSwap(pMin, pMax, curve)
	if err != nil {
		return dbc.ConfigParameters{}, err
	}

	swapBaseAmountBuffer, err := GetSwapAmountWithBuffer(swapBaseAmount, pMin, curve)
	if err != nil {
		return dbc.ConfigParameters{}, err
	}

	migrationAmount := new(big.Int).Sub(totalSwapAndMigrationAmount, swapBaseAmountBuffer)

	migrationQuoteAmount := new(big.Int).Rsh(
		new(big.Int).Mul(
			new(big.Int).Mul(migrationAmount, pMax),
			pMax,
		),
		128,
	)

	migrationQuoteThreshold := GetMigrationQuoteThresholdFromMigrationQuoteAmount(
		new(big.Float).SetInt(migrationQuoteAmount),
		param.MigrationFee.FeePercentage,
	)

	migrationQuoteThresholdInLamport, _ := migrationQuoteThreshold.Int(nil)
	if !migrationQuoteThresholdInLamport.IsUint64() {
		return dbc.ConfigParameters{}, fmt.Errorf("cannot fit migrationQuoteThresholdInLamport(%s) into uint64", migrationQuoteThresholdInLamport)
	}

	// sanity check
	{
		totalDynamicSupply, err := GetTotalSupplyFromCurve(
			migrationQuoteThresholdInLamport,
			pMin,
			curve,
			lockedVesting,
			param.MigrationOption,
			totalLeftover,
			param.MigrationFee.FeePercentage,
		)
		if err != nil {
			return dbc.ConfigParameters{}, err
		}

		if totalDynamicSupply.Cmp(totalSupply) > 0 {
			// precision loss is used for leftover
			leftOverDelta := new(big.Int).Sub(totalDynamicSupply, totalSupply)
			if leftOverDelta.Cmp(totalLeftover) >= 0 {
				return dbc.ConfigParameters{}, errors.New("leftOverDelta must be less than totalLeftover")
			}

		}
	}

	baseFee, err := GetBaseFeeParams(
		param.BaseFeeParams,
		param.TokenQuoteDecimal,
		param.ActivationType,
	)
	if err != nil {
		return dbc.ConfigParameters{}, err
	}

	var dynamicField *dbc.DynamicFeeParameters
	if param.DynamicFeeEnabled {
		baseFeeBp := param.BaseFeeParams.FeeSchedulerParam.EndingFeeBps
		if param.BaseFeeParams.BaseFeeMode == types.BaseFeeModeRateLimiter {
			baseFeeBp = param.BaseFeeParams.RateLimiterParam.BaseFeeBps
		}

		d, err := GetDynamicFeeParams(baseFeeBp, 0)
		if err != nil {
			return dbc.ConfigParameters{}, err
		}

		dynamicField = &d
	}

	return dbc.ConfigParameters{
		PoolFees: dbc.PoolFeeParameters{
			BaseFee: dbc.BaseFeeParameters{
				CliffFeeNumerator: baseFee.CliffFeeNumerator,
				FirstFactor:       baseFee.FirstFactor,
				SecondFactor:      baseFee.SecondFactor,
				ThirdFactor:       baseFee.ThirdFactor,
				BaseFeeMode:       uint8(baseFee.BaseFeeMode),
			},
			DynamicFee: dynamicField,
		},
		ActivationType:            uint8(param.ActivationType),
		CollectFeeMode:            uint8(param.CollectFeeMode),
		MigrationOption:           uint8(param.MigrationFeeOption),
		TokenType:                 uint8(param.TokenType),
		TokenDecimal:              uint8(param.TokenBaseDecimal),
		MigrationQuoteThreshold:   migrationQuoteThresholdInLamport.Uint64(),
		PartnerLpPercentage:       param.PartnerLpPercentage,
		CreatorLpPercentage:       param.CreatorLpPercentage,
		PartnerLockedLpPercentage: param.PartnerLockedLpPercentage,
		CreatorLockedLpPercentage: param.CreatorLockedLpPercentage,
		SqrtStartPrice:            MustBigIntToUint128(pMin),
		LockedVesting:             lockedVesting,
		MigrationFeeOption:        uint8(param.MigrationFeeOption),
		TokenSupply: &dbc.TokenSupplyParams{
			PreMigrationTokenSupply:  totalSupply.Uint64(),
			PostMigrationTokenSupply: totalSupply.Uint64(),
		},
		CreatorTradingFeePercentage: param.CreatorTradingFeePercentage,
		MigratedPoolFee: GetMigratedPoolFeeParams(
			param.MigrationOption,
			param.MigrationFeeOption,
			param.MigratedPoolFee,
		),
		TokenUpdateAuthority: param.TokenUpdateAuthority,
		MigrationFee: dbc.MigrationFee{
			FeePercentage:        uint8(param.MigrationFee.FeePercentage),
			CreatorFeePercentage: uint8(param.MigrationFee.CreatorFeePercentage),
		},
		Curve: curve,
	}, nil

}

// GetSqrtPriceFromMarketCap gets the sqrt price from the market cap.
func GetSqrtPriceFromMarketCap(
	marketCap, totalSupply uint64, tokenBaseDecimal, tokenQuoteDecimal types.TokenDecimal,
) *big.Int {
	return GetSqrtPriceFromPrice(
		new(big.Float).SetPrec(256).Quo(
			new(big.Float).SetUint64(marketCap),
			new(big.Float).SetUint64(totalSupply),
		),
		tokenBaseDecimal,
		tokenQuoteDecimal,
	)
}
