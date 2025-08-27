package helpers

import (
	"context"
	"dbcGoSDK/constants"
	"dbcGoSDK/generated/dbc"
	"dbcGoSDK/maths"
	"dbcGoSDK/types"
	"errors"
	"fmt"
	"math"
	"math/big"
	"slices"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func GetFirstkey(key1, key2 solana.PublicKey) []byte {
	if slices.Compare(key1.Bytes(), key2.Bytes()) == 1 {
		return key1.Bytes()
	}
	return key2.Bytes()
}

func GetSecondkey(key1, key2 solana.PublicKey) []byte {
	if slices.Compare(key1.Bytes(), key2.Bytes()) == 1 {
		return key2.Bytes()
	}
	return key1.Bytes()
}

func GetTotalVestingAmount(
	lockedVesting dbc.LockedVestingParams,
) *big.Int {
	return new(big.Int).Add(
		new(big.Int).SetUint64(lockedVesting.CliffUnlockAmount),
		new(big.Int).Mul(
			new(big.Int).SetUint64(lockedVesting.AmountPerPeriod),
			new(big.Int).SetUint64(lockedVesting.NumberOfPeriod),
		),
	)
}

// GetMigrationQuoteAmountFromMigrationQuoteThreshold gets migrationQuoteAmount from migrationQuoteThreshold and migrationFeePercent.
func GetMigrationQuoteAmountFromMigrationQuoteThreshold(
	migrationQuoteThreshold *big.Float,
	migrationFeePercent float64,
) *big.Float {
	// (migrationQuoteThreshold * (100 - feePercent)) / 100
	hold := new(big.Float).Mul(
		migrationQuoteThreshold,
		new(big.Float).Sub(constants.HundredInBigFloat, big.NewFloat(migrationFeePercent)),
	)
	return new(big.Float).Quo(hold, constants.HundredInBigFloat)
}

// GetMigrationQuoteThresholdFromMigrationQuoteAmount gets migrationQuoteThreshold from migrationQuoteAmount and migrationFeePercent.
func GetMigrationQuoteThresholdFromMigrationQuoteAmount(
	migrationQuoteAmount *big.Float, migrationFeePercent float64,
) *big.Float {
	return new(big.Float).SetPrec(256).Quo(
		new(big.Float).SetPrec(256).Mul(migrationQuoteAmount, constants.HundredInBigFloat),
		new(big.Float).SetPrec(256).Sub(
			constants.HundredInBigFloat, big.NewFloat(migrationFeePercent),
		),
	)
}

func GetBaseFeeParams(
	baseFeeParams types.BaseFeeParams,
	tokenQuoteDecimal types.TokenDecimal,
	activationType types.ActivationType,
) (types.BaseFee, error) {

	if baseFeeParams.BaseFeeMode == types.BaseFeeModeRateLimiter {
		if baseFeeParams.RateLimiterParam == nil {
			return types.BaseFee{}, errors.New("rate limiter parameters are required for RateLimiter mode")
		}

		return GetRateLimiterParams(
			baseFeeParams.RateLimiterParam.BaseFeeBps,
			baseFeeParams.RateLimiterParam.FeeIncrementBps,
			baseFeeParams.RateLimiterParam.ReferenceAmount,
			baseFeeParams.RateLimiterParam.MaxLimiterDuration,
			tokenQuoteDecimal,
			activationType,
		)
	}

	if baseFeeParams.FeeSchedulerParam == nil {
		return types.BaseFee{}, errors.New("fee scheduler parameters are required for FeeScheduler mode")
	}

	return GetFeeSchedulerParams(
		baseFeeParams.FeeSchedulerParam.StartingFeeBps,
		baseFeeParams.FeeSchedulerParam.EndingFeeBps,
		baseFeeParams.BaseFeeMode,
		baseFeeParams.FeeSchedulerParam.NumberOfPeriod,
		baseFeeParams.FeeSchedulerParam.TotalDuration,
	)
}

func GetPercentageSupplyOnMigration(
	initialMarketCap, migrationMarketCap *big.Float,
	lockedVesting dbc.LockedVestingParams,
	totalLeftover, totalTokenSupply *big.Int,
) float64 {
	// formula: x = sqrt(initialMC / migrationMC) * (100 - lockedVesting - leftover) / (1 + sqrt(initialMC / migrationMC))

	// sqrtRatio = sqrt(initial_MC / migration_MC)
	marketCapRatio := new(big.Float).Quo(initialMarketCap, migrationMarketCap)
	sqrtRatio := new(big.Float).Sqrt(marketCapRatio)

	// locked vesting percentage
	totalVestingAmount := GetTotalVestingAmount(lockedVesting)
	totalVestingAmount.Mul(totalVestingAmount, constants.HundredInBigInt)

	vestingPercentage := new(big.Float).Quo(
		new(big.Float).SetInt(totalVestingAmount),
		new(big.Float).SetInt(totalTokenSupply),
	)

	// leftover percentage
	hold := new(big.Int)
	hold.Set(totalLeftover)
	hold.Mul(hold, constants.HundredInBigInt)
	leftoverPercentage := new(big.Float).Quo(
		new(big.Float).SetInt(hold),
		new(big.Float).SetInt(totalTokenSupply),
	)

	// (100 * sqrtRatio - (vestingPercentage + leftoverPercentage) * sqrtRatio) / (1 + sqrtRatio)
	numerator := new(big.Float).Mul(sqrtRatio, new(big.Float).Sub(
		constants.HundredInBigFloat,
		new(big.Float).Add(vestingPercentage, leftoverPercentage),
	))
	denominator := new(big.Float).Add(big.NewFloat(1), sqrtRatio)
	resultBigFloat := new(big.Float).Quo(numerator, denominator)

	result, _ := resultBigFloat.Float64()
	return result
}

// GetMigrationQuoteAmount gets the migration quote amount.
func GetMigrationQuoteAmount(
	migrationMarketCap, percentageSupplyOnMigration *big.Float,
) *big.Float {
	// migrationMC * x / 100
	return new(big.Float).Quo(
		new(big.Float).Mul(migrationMarketCap, percentageSupplyOnMigration),
		constants.HundredInBigFloat,
	)
}

// GetRateLimiterParams gets the rate limiter parameters.
func GetRateLimiterParams(
	baseFeeBps uint64, feeIncrementBps uint16, referenceAmount float64, maxLimiterDuration uint64,
	tokenQuoteDecimal types.TokenDecimal, activationType types.ActivationType,
) (types.BaseFee, error) {

	if baseFeeBps <= 0 || feeIncrementBps <= 0 || referenceAmount <= 0 || maxLimiterDuration <= 0 {
		return types.BaseFee{}, errors.New("all rate limiter parameters must be greater than zero")
	}

	if baseFeeBps > constants.MaxFeeBPS {
		return types.BaseFee{}, fmt.Errorf("base fee (%d bps) exceeds maximum allowed value of %d bps", baseFeeBps, constants.MaxFeeBPS)
	}

	if feeIncrementBps > constants.MaxFeeBPS {
		return types.BaseFee{}, fmt.Errorf("base fee (%d bps) exceeds maximum allowed value of %d bps", feeIncrementBps, constants.MaxFeeBPS)
	}

	cliffFeeNumerator, feeIncrementNumerator :=
		BpsToFeeNumerator(baseFeeBps), BpsToFeeNumerator(uint64(feeIncrementBps))

	if feeIncrementNumerator.Cmp(constants.FeeDenominatorBigInt) >= 0 {
		return types.BaseFee{}, errors.New("fee increment numerator must be less than FEE_DENOMINATOR")
	}

	deltaNumerator := new(big.Int).Sub(big.NewInt(constants.MaxFeeNumerator), cliffFeeNumerator)
	maxIndex := new(big.Int).Quo(deltaNumerator, feeIncrementNumerator)
	if maxIndex.Cmp(big.NewInt(1)) < 0 {
		return types.BaseFee{}, errors.New("fee increment is too large for the given base fee")
	}

	if cliffFeeNumerator.Cmp(big.NewInt(constants.MinFeeNumerator)) < 0 ||
		cliffFeeNumerator.Cmp(big.NewInt(constants.MaxFeeNumerator)) > 0 {
		return types.BaseFee{}, errors.New("base fee must be between 0.01% and 99%")
	}

	maxDuration := constants.MaxRateLimiterDurationInSeconds
	if activationType == types.ActivationTypeSlot {
		maxDuration = constants.MaxRateLimiterDurationInSlots
	}

	if maxLimiterDuration > uint64(maxDuration) {
		return types.BaseFee{}, fmt.Errorf("max duration exceeds maximum allowed value of %d", maxDuration)
	}
	referenceAmountInLamports := ConvertToLamports(referenceAmount, tokenQuoteDecimal)

	if !cliffFeeNumerator.IsUint64() || !referenceAmountInLamports.IsInt64() {
		return types.BaseFee{},
			fmt.Errorf("either cliffFeeNumerator(%s) or referenceAmountInLamports(%s) cannot fit into uint64",
				cliffFeeNumerator, referenceAmountInLamports)
	}

	return types.BaseFee{
			CliffFeeNumerator: cliffFeeNumerator.Uint64(),
			FirstFactor:       feeIncrementBps,
			SecondFactor:      maxLimiterDuration,
			ThirdFactor:       referenceAmountInLamports.Uint64(),
			BaseFeeMode:       types.BaseFeeModeRateLimiter,
		},
		nil
}

func GetDynamicFeeParams(
	baseFeeBp, maxPriceChangeBps uint64,
) (dbc.DynamicFeeParameters, error) {
	if maxPriceChangeBps > constants.MaxPriceChangeBpsDefault {
		return dbc.DynamicFeeParameters{},
			fmt.Errorf("maxPriceChangeBps (%d bps) must be less than or equal to %d",
				maxPriceChangeBps, constants.MaxPriceChangeBpsDefault)
	}

	if maxPriceChangeBps == 0 {
		maxPriceChangeBps = constants.MaxPriceChangeBpsDefault
	}
	priceRatio := float64(maxPriceChangeBps)/constants.BasisPointMax + 1

	hold := new(big.Float).SetFloat64(priceRatio)
	hold.Sqrt(hold).Mul(hold, big.NewFloat(math.Pow(2, 64)))

	sqrtPriceRatioQ64 := new(big.Int)
	hold.Int(sqrtPriceRatioQ64)

	deltaBinId := new(big.Int).Mul(
		new(big.Int).Quo(
			new(big.Int).Sub(sqrtPriceRatioQ64, constants.OneQ64),
			constants.BinStepBpsU128Default,
		),
		big.NewInt(2),
	)

	maxVolatilityAccumulator := deltaBinId.Mul(deltaBinId, big.NewInt(constants.BasisPointMax))

	squareVfaBin := new(big.Int).Mul(maxVolatilityAccumulator, big.NewInt(constants.BinStepBpsDefault))
	squareVfaBin.Exp(squareVfaBin, big.NewInt(2), nil)

	baseFeeNumerator := BpsToFeeNumerator(baseFeeBp)
	maxDynamicFeeNumerator := baseFeeNumerator.Mul(baseFeeNumerator, big.NewInt(20))
	maxDynamicFeeNumerator.Quo(maxDynamicFeeNumerator, constants.HundredInBigInt) // default max dynamic fee = 20% of min base fee

	vFee := maxDynamicFeeNumerator.Mul(maxDynamicFeeNumerator, big.NewInt(100_000_000_000))
	vFee.Sub(vFee, big.NewInt(99_999_999_999))

	variableFeeControl := vFee.Quo(vFee, squareVfaBin)

	if maxVolatilityAccumulator.Cmp(big.NewInt(math.MaxUint32)) > 0 ||
		variableFeeControl.Cmp(big.NewInt(math.MaxUint32)) > 0 {
		return dbc.DynamicFeeParameters{}, fmt.Errorf("either variableFeeControl(%s) or maxVolatilityAccumulator(%s) cannot fit into uint32",
			variableFeeControl, maxVolatilityAccumulator)
	}

	return dbc.DynamicFeeParameters{
		BinStep:                  constants.BinStepBpsDefault,
		BinStepU128:              MustBigIntToUint128(constants.BinStepBpsU128Default),
		FilterPeriod:             constants.DynamicFeeFilterPeriodDefault,
		DecayPeriod:              constants.DynamicFeeDecayPeriodDefault,
		ReductionFactor:          constants.DynamicFeeReductionFactorDefault,
		MaxVolatilityAccumulator: uint32(maxVolatilityAccumulator.Uint64()),
		VariableFeeControl:       uint32(variableFeeControl.Uint64()),
	}, nil
}

// // GetFeeSchedulerParams gets the fee scheduler parameters.
func GetFeeSchedulerParams(
	startingBaseFeeBps, endingBaseFeeBps uint64,
	baseFeeMode types.BaseFeeMode,
	numberOfPeriod uint16, totalDuration uint64,
) (types.BaseFee, error) {
	if startingBaseFeeBps == endingBaseFeeBps {
		if numberOfPeriod != 0 || totalDuration != 0 {
			return types.BaseFee{}, errors.New("numberOfPeriod and totalDuration must both be zero")
		}

		cliffFeeNumerator := BpsToFeeNumerator(startingBaseFeeBps)
		if !cliffFeeNumerator.IsUint64() {
			return types.BaseFee{}, fmt.Errorf("cannot fit cliffFeeNumerator(%s) to uint64", cliffFeeNumerator)
		}

		return types.BaseFee{
			CliffFeeNumerator: cliffFeeNumerator.Uint64(),
			BaseFeeMode:       types.BaseFeeModeFeeSchedulerLinear,
		}, nil
	}

	if numberOfPeriod <= 0 {
		return types.BaseFee{}, errors.New("total periods must be greater than zero")
	}

	if startingBaseFeeBps > constants.MaxFeeBPS {
		return types.BaseFee{}, fmt.Errorf("startingBaseFeeBps (%d bps) exceeds maximum allowed value of %d bps", startingBaseFeeBps, constants.MaxFeeBPS)
	}

	if endingBaseFeeBps > startingBaseFeeBps {
		return types.BaseFee{}, errors.New("endingBaseFeeBps bps must be less than or equal to startingBaseFeeBps bps")
	}

	if numberOfPeriod == 0 || totalDuration == 0 {
		return types.BaseFee{}, errors.New("numberOfPeriod and totalDuration must both greater than zero")
	}

	maxBaseFeeNumerator, minBaseFeeNumerator, periodFrequency :=
		BpsToFeeNumerator(startingBaseFeeBps), BpsToFeeNumerator(endingBaseFeeBps),
		new(big.Int).SetUint64(totalDuration/uint64(numberOfPeriod))

	if baseFeeMode == types.BaseFeeModeFeeSchedulerLinear {
		reductionFactor := new(big.Int).Quo(
			new(big.Int).Sub(maxBaseFeeNumerator, minBaseFeeNumerator),
			new(big.Int).SetUint64(uint64(numberOfPeriod)),
		)

		if !reductionFactor.IsUint64() {
			return types.BaseFee{}, fmt.Errorf("cannot fit reductionFactor(%s) into uint64", reductionFactor)
		}

		return types.BaseFee{
			CliffFeeNumerator: maxBaseFeeNumerator.Uint64(),
			FirstFactor:       numberOfPeriod,
			SecondFactor:      periodFrequency.Uint64(),
			ThirdFactor:       reductionFactor.Uint64(),
			BaseFeeMode:       baseFeeMode,
		}, nil
	}

	minBaseFeeNumeratorFloat64, _ := minBaseFeeNumerator.Float64()
	maxBaseFeeNumeratorFloat64, _ := maxBaseFeeNumerator.Float64()
	ratio := minBaseFeeNumeratorFloat64 / maxBaseFeeNumeratorFloat64
	decayBase := math.Pow(ratio, 1/float64(numberOfPeriod))

	reductionFactor := new(big.Float).Mul(
		big.NewFloat(constants.BasisPointMax),
		new(big.Float).Sub(big.NewFloat(1), big.NewFloat(decayBase)),
	)

	reductionFactorU64, _ := reductionFactor.Uint64()

	return types.BaseFee{
		CliffFeeNumerator: maxBaseFeeNumerator.Uint64(),
		FirstFactor:       numberOfPeriod,
		SecondFactor:      periodFrequency.Uint64(),
		ThirdFactor:       reductionFactorU64,
		BaseFeeMode:       baseFeeMode,
	}, nil
}

// GetSqrtPriceFromPrice gets the sqrt price from the price.
//
//	sqrtPriceQ64 = sqrt(price / 10^(tokenADecimal - tokenBDecimal)) * 2^64
func GetSqrtPriceFromPrice(
	price *big.Float,
	tokenADecimal, tokenBDecimal types.TokenDecimal,
) *big.Int {
	adjustedPrice := new(big.Float).Quo(
		price,
		new(big.Float).SetFloat64(math.Pow10(int(tokenADecimal)-int(tokenBDecimal))),
	)

	sqrtQ64 := new(big.Float).Mul(
		adjustedPrice.Sqrt(adjustedPrice),
		new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), 64)),
	)

	result := new(big.Int)
	sqrtQ64.Int(result)
	return result
}

// GetMigratedPoolFeeParams gets migrated pool fee parameters based on migration options.
func GetMigratedPoolFeeParams(
	migrationOption types.MigrationOption,
	migrationFeeOption types.MigrationFeeOption,
	migratedPoolFee dbc.MigratedPoolFee,
) dbc.MigratedPoolFee {

	// For DAMM_V1: always use default parameters
	if migrationOption == types.MigrationOptionMET_DAMM {
		return dbc.MigratedPoolFee{}
	}

	// For DAMM_V2: use custom parameters only if Customizable option is selected
	if migrationOption == types.MigrationOptionMET_DAMM_V2 {
		if migrationFeeOption == types.MigrationFeeOptionCustomizable {
			return migratedPoolFee
		}

		// For fixed fee options (0-5), always use defaults
		return dbc.MigratedPoolFee{}
	}

	return dbc.MigratedPoolFee{}
}

// GetLockedVestingParams calculates the locked vesting parameters.
func GetLockedVestingParams(
	totalLockedVestingAmount, numberOfVestingPeriod, cliffUnlockAmount,
	totalVestingDuration, cliffDurationFromMigrationTime uint64,
	tokenBaseDecimal types.TokenDecimal,
) (dbc.LockedVestingParams, error) {
	if totalLockedVestingAmount == 0 {
		return dbc.LockedVestingParams{}, nil
	}

	holdAmountPerPeriod, holdCliffUnlockAmount := ConvertToLamports(1, tokenBaseDecimal),
		ConvertToLamports(float64(totalLockedVestingAmount-1), tokenBaseDecimal)

	if !holdAmountPerPeriod.IsInt64() || !holdCliffUnlockAmount.IsUint64() {
		return dbc.LockedVestingParams{},
			fmt.Errorf("either holdAmountPerPeriod(%s) or holdCliffUnlockAmount(%s) cannot fit uint64",
				holdAmountPerPeriod, holdCliffUnlockAmount)
	}

	if totalLockedVestingAmount == cliffUnlockAmount {
		return dbc.LockedVestingParams{
			AmountPerPeriod:                holdAmountPerPeriod.Uint64(),
			CliffDurationFromMigrationTime: cliffDurationFromMigrationTime,
			Frequency:                      1,
			NumberOfPeriod:                 1,
			CliffUnlockAmount:              holdCliffUnlockAmount.Uint64(),
		}, nil
	}

	if numberOfVestingPeriod <= 0 {
		return dbc.LockedVestingParams{}, errors.New("total periods must be greater than zero")
	}

	if numberOfVestingPeriod == 0 || totalLockedVestingAmount == 0 {
		return dbc.LockedVestingParams{}, errors.New("numberOfPeriod and totalVestingDuration must both be greater than zero")
	}

	if cliffUnlockAmount > totalLockedVestingAmount {
		return dbc.LockedVestingParams{}, errors.New("cliff unlock amount cannot be greater than total locked vesting amount")
	}

	// amount_per_period = (total_locked_vesting_amount - cliff_unlock_amount) / number_of_period
	// round amountPerPeriod down to ensure we don't exceed total amount
	amountPerPeriod := (totalLockedVestingAmount - cliffUnlockAmount) / numberOfVestingPeriod

	totalPeriodicAmount := amountPerPeriod * numberOfVestingPeriod
	remainder := totalLockedVestingAmount - (cliffUnlockAmount + totalPeriodicAmount)

	// add the remainder to cliffUnlockAmount to maintain total amount
	adjustedCliffUnlockAmount := cliffUnlockAmount + remainder

	holdAmountPerPeriod, holdCliffUnlockAmount = ConvertToLamports(float64(amountPerPeriod), tokenBaseDecimal),
		ConvertToLamports(float64(adjustedCliffUnlockAmount), tokenBaseDecimal)

	if !holdAmountPerPeriod.IsInt64() || !holdCliffUnlockAmount.IsUint64() {
		return dbc.LockedVestingParams{},
			fmt.Errorf("either holdAmountPerPeriod(%s) or holdCliffUnlockAmount(%s) cannot fit uint64",
				holdAmountPerPeriod, holdCliffUnlockAmount)
	}

	return dbc.LockedVestingParams{
		AmountPerPeriod:                holdAmountPerPeriod.Uint64(),
		CliffDurationFromMigrationTime: cliffDurationFromMigrationTime,
		NumberOfPeriod:                 numberOfVestingPeriod,
		Frequency:                      totalVestingDuration / numberOfVestingPeriod,
		CliffUnlockAmount:              holdCliffUnlockAmount.Uint64(),
	}, nil
}

func CalculateFeeSchedulerEndingBaseFeeBps(
	cliffFeeNumerator, numberOfPeriod, reductionFactor float64,
	baseFeeMode types.BaseFeeMode,
) float64 {
	// linear mode
	baseFeeNumerator := cliffFeeNumerator - numberOfPeriod*reductionFactor
	if baseFeeMode != types.BaseFeeModeFeeSchedulerLinear {
		// exponential mode
		decayRate := 1 - reductionFactor/float64(constants.BasisPointMax)
		baseFeeNumerator = cliffFeeNumerator * math.Pow(decayRate, numberOfPeriod)
	}

	return math.Max(0, baseFeeNumerator/float64(constants.FeeDenominator)*float64(constants.BasisPointMax))
}

func GetMigrationBaseToken(
	migrationQuoteAmount, sqrtMigrationPrice *big.Int,
	migrationOption types.MigrationOption,
) (*big.Int, error) {
	if migrationOption == types.MigrationOptionMET_DAMM {
		price, quote, mod := new(big.Int).Mul(sqrtMigrationPrice, sqrtMigrationPrice),
			new(big.Int).Lsh(migrationQuoteAmount, 128), new(big.Int)

		q, _ := quote.QuoRem(quote, price, mod)
		if mod.Sign() != 0 {
			q.Add(q, big.NewInt(1))
		}
		return q, nil
	}

	if migrationOption == types.MigrationOptionMET_DAMM_V2 {
		liquidity, err := GetInitialLiquidityFromDeltaQuote(
			migrationQuoteAmount,
			constants.MinSqrtPrice,
			sqrtMigrationPrice,
		)
		if err != nil {
			return nil, err
		}

		// calculate base threshold
		baseAmount, err := maths.GetDeltaAmountBaseUnsigned(
			sqrtMigrationPrice,
			constants.MaxSqrtPrice,
			liquidity,
			types.RoundingUp,
		)
		if err != nil {
			return nil, err
		}

		return baseAmount, nil
	}

	return nil, errors.New("invalid migration option")
}

// GetInitialLiquidityFromDeltaQuote gets the initial liquidity from delta quote.
//
//	Formula: Δb = L (√P_upper - √P_lower) => L = Δb / (√P_upper - √P_lower)
func GetInitialLiquidityFromDeltaQuote(
	quoteAmount, sqrtMinPrice, sqrtPrice *big.Int,
) (*big.Int, error) {
	priceDelta := new(big.Int).Sub(sqrtPrice, sqrtMinPrice)
	if priceDelta.Sign() < 0 {
		return nil, fmt.Errorf("safeMath requires value not negative: value is %s", priceDelta.String())
	}

	return priceDelta.Quo( // rounds down
		new(big.Int).Lsh(quoteAmount, constants.RESOLUTION*2), // quoteAmountShifted
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
		return nil, fmt.Errorf("GetInitialLiquidityFromDeltaBase:safeMath requires value not negative: value is %s", priceDelta.String())
	}
	return new(big.Int).Quo( // rounds down
		new(big.Int).Mul(
			new(big.Int).Mul(baseAmount, sqrtPrice),
			sqrtMaxPrice,
		),
		priceDelta,
	), nil
}

func GetBaseTokenForSwap(
	sqrtStartPrice, sqrtMigrationPrice *big.Int,
	curve []dbc.LiquidityDistributionParameters,
) (*big.Int, error) {

	totalAmount := big.NewInt(0)
	for i := range len(curve) {
		lowerSqrtPrice := sqrtStartPrice
		if i != 0 {
			lowerSqrtPrice = curve[i-1].SqrtPrice.BigInt()
		}

		if hold := curve[i].SqrtPrice.BigInt(); hold != nil && hold.Cmp(sqrtMigrationPrice) > 0 {
			deltaAmount, err := maths.GetDeltaAmountBaseUnsigned(
				lowerSqrtPrice,
				sqrtMigrationPrice,
				curve[i].Liquidity.BigInt(),
				types.RoundingUp,
			)
			if err != nil {
				return nil, err
			}

			totalAmount.Add(totalAmount, deltaAmount)
			return totalAmount, nil
		}

		deltaAmount, err := maths.GetDeltaAmountBaseUnsigned(
			lowerSqrtPrice,
			curve[i].SqrtPrice.BigInt(),
			curve[i].Liquidity.BigInt(),
			types.RoundingUp,
		)

		if err != nil {
			return nil, err
		}

		totalAmount.Add(totalAmount, deltaAmount)
	}

	return totalAmount, nil
}

func GetTotalSupplyFromCurve(
	migrationQuoteThreshold *big.Int,
	sqrtStartPrice *big.Int,
	curve []dbc.LiquidityDistributionParameters,
	lockedVesting dbc.LockedVestingParams,
	migrationOption types.MigrationOption,
	leftover *big.Int,
	migrationFeePercent float64,
) (*big.Int, error) {

	sqrtMigrationPrice, err := GetMigrationThresholdPrice(
		migrationQuoteThreshold,
		sqrtStartPrice,
		curve,
	)
	if err != nil {
		return nil, err
	}

	swapBaseAmount, err := GetBaseTokenForSwap(
		sqrtStartPrice,
		sqrtMigrationPrice,
		curve,
	)
	if err != nil {
		return nil, err
	}

	swapBaseAmountBuffer, err := GetSwapAmountWithBuffer(
		swapBaseAmount,
		sqrtStartPrice,
		curve,
	)
	if err != nil {
		return nil, err
	}

	migrationQuoteAmount := GetMigrationQuoteAmountFromMigrationQuoteThreshold(
		new(big.Float).SetInt(migrationQuoteThreshold),
		migrationFeePercent,
	)

	migrationQuoteAmountInt, _ := migrationQuoteAmount.Int(nil)

	migrationBaseAmount, err := GetMigrationBaseToken(
		migrationQuoteAmountInt,
		sqrtMigrationPrice,
		migrationOption,
	)
	if err != nil {
		return nil, err
	}

	totalVestingAmount := GetTotalVestingAmount(lockedVesting)

	totalSupply := new(big.Int).Add(swapBaseAmountBuffer, migrationBaseAmount)
	totalSupply.Add(totalSupply, totalVestingAmount)
	totalSupply.Add(totalSupply, leftover)
	return totalSupply, nil
}

func GetSwapAmountWithBuffer(
	swapBaseAmount, sqrtStartPrice *big.Int,
	curve []dbc.LiquidityDistributionParameters,
) (*big.Int, error) {

	swapAmountBuffer := new(big.Int).Add(
		swapBaseAmount,
		new(big.Int).Quo(
			new(big.Int).Mul(swapBaseAmount, big.NewInt(25)),
			big.NewInt(100),
		),
	)

	maxBaseAmountOnCurve, err := GetBaseTokenForSwap(
		sqrtStartPrice,
		constants.MaxSqrtPrice,
		curve,
	)
	if err != nil {
		return nil, err
	}

	if swapAmountBuffer.Cmp(maxBaseAmountOnCurve) < 0 {
		return swapAmountBuffer, nil
	}

	return maxBaseAmountOnCurve, nil
}

// GetMigrationThresholdPrice gets the migration threshold price.
func GetMigrationThresholdPrice(
	migrationThreshold, sqrtStartPrice *big.Int,
	curve []dbc.LiquidityDistributionParameters,
) (*big.Int, error) {
	if len(curve) == 0 {
		return nil, errors.New("curve is empty")
	}

	nextSqrtPrice := new(big.Int).Set(sqrtStartPrice)
	totalAmount, err := maths.GetDeltaAmountQuoteUnsigned(
		nextSqrtPrice,
		curve[0].SqrtPrice.BigInt(),
		curve[0].Liquidity.BigInt(),
		types.RoundingUp,
	)
	if err != nil {
		return nil, err
	}
	if totalAmount.Cmp(migrationThreshold) > 0 {
		return maths.GetNextSqrtPriceFromInput(
			nextSqrtPrice,
			curve[0].Liquidity.BigInt(),
			migrationThreshold,
			false,
		)
	}

	amountLeft := new(big.Int).Sub(migrationThreshold, totalAmount)
	nextSqrtPrice = curve[0].SqrtPrice.BigInt()
	for i := 1; i < len(curve); i++ {
		maxAmount, err := maths.GetDeltaAmountQuoteUnsigned(
			nextSqrtPrice,
			curve[i].SqrtPrice.BigInt(),
			curve[i].Liquidity.BigInt(),
			types.RoundingUp,
		)
		if err != nil {
			return nil, err
		}

		if maxAmount.Cmp(amountLeft) > 0 {
			if nextSqrtPrice, err = maths.GetNextSqrtPriceFromInput(
				nextSqrtPrice,
				curve[i].Liquidity.BigInt(),
				amountLeft,
				false,
			); err != nil {
				return nil, err
			}
			amountLeft = big.NewInt(0)
			break
		}

		amountLeft = new(big.Int).Sub(amountLeft, maxAmount)
		// amountLeft.Sub(amountLeft, maxAmount)
		nextSqrtPrice = curve[i].SqrtPrice.BigInt()
	}

	if amountLeft.Sign() != 0 {
		return nil, fmt.Errorf(
			"not enough liquidity, migrationThreshold: %s  amountLeft: %s",
			migrationThreshold, amountLeft,
		)
	}

	return nextSqrtPrice, nil
}

// GetFirstCurve gets the first curve.
//
//		Swap_amount = L * (1/Pmin - 1/Pmax) = L * (Pmax - Pmin) / (Pmax * Pmin)      (1)
//		Quote_amount = L * (Pmax - Pmin)                                             (2)
//		(Quote_amount * (1-migrationFeePercent/100) / Migration_amount = Pmax ^ 2    (3)
//
//	 From (1) and (2) => Quote_amount / Swap_amount = (Pmax * Pmin)               (4)
//	 From (3) and (4) => Swap_amount * (1-migrationFeePercent/100) / Migration_amount = Pmax / Pmin
//	 => Pmin = Pmax * Migration_amount / (Swap_amount * (1-migrationFeePercent/100))
func GetFirstCurve(
	migrationSqrtPrice, migrationBaseAmount, swapAmount, migrationQuoteThreshold *big.Int,
	migrationFeePercent float64,
) (types.GetFirstCurveResult, error) {

	denominator := new(big.Float).Quo(
		new(big.Float).Mul(
			new(big.Float).SetInt(swapAmount),
			new(big.Float).Sub(constants.HundredInBigFloat, big.NewFloat(migrationFeePercent)),
		),
		constants.HundredInBigFloat,
	)

	sqrtStartPriceFloat := new(big.Float).Quo(
		new(big.Float).Mul(new(big.Float).SetInt(migrationSqrtPrice), new(big.Float).SetInt(migrationBaseAmount)),
		denominator,
	)

	sqrtStartPrice := new(big.Int)
	sqrtStartPriceFloat.Int(sqrtStartPrice)

	liquidity, err := Liquidity(
		swapAmount,
		migrationQuoteThreshold,
		sqrtStartPrice,
		migrationSqrtPrice,
	)
	if err != nil {
		return types.GetFirstCurveResult{}, err
	}

	return types.GetFirstCurveResult{
		SqrtStartPrice: sqrtStartPrice,
		Curve: []dbc.LiquidityDistributionParameters{
			{
				SqrtPrice: MustBigIntToUint128(migrationSqrtPrice),
				Liquidity: MustBigIntToUint128(liquidity),
			},
		},
	}, nil
}

func GetTwoCurve(
	migrationSqrtPrice, midSqrtPrice, initialSqrtPrice,
	swapAmount, migrationQuoteThreshold *big.Int,
) struct {
	IsoK     bool
	TwoCurve types.GetFirstCurveResult
} {
	p0, p1, p2 := new(big.Float).SetInt(initialSqrtPrice),
		new(big.Float).SetInt(midSqrtPrice), new(big.Float).SetInt(migrationSqrtPrice)

	a1 := new(big.Float).Sub(
		new(big.Float).Quo(big.NewFloat(1), p0),
		new(big.Float).Quo(big.NewFloat(1), p1),
	)

	b1 := new(big.Float).Sub(
		new(big.Float).Quo(big.NewFloat(1), p1),
		new(big.Float).Quo(big.NewFloat(1), p2),
	)

	c1 := new(big.Float).SetInt(swapAmount)

	a2 := new(big.Float).Sub(p1, p0)
	b2 := new(big.Float).Sub(p2, p1)
	c2 := new(big.Float).SetInt(migrationQuoteThreshold)
	c2.Mul(c2, new(big.Float).SetInt(new(big.Int).Lsh(big.NewInt(1), 128)))
	// c2.Mul(c2, big.NewFloat(math.Pow(2, 128)))

	// solve equation to find l0 and l1
	l0 := new(big.Float).Quo(
		new(big.Float).Sub(
			new(big.Float).Mul(c1, b2),
			new(big.Float).Mul(c2, b1),
		),
		new(big.Float).Sub(
			new(big.Float).Mul(a1, b2),
			new(big.Float).Mul(a2, b1),
		),
	)
	l1 := new(big.Float).Quo(
		new(big.Float).Sub(
			new(big.Float).Mul(c1, a2),
			new(big.Float).Mul(c2, a1),
		),
		new(big.Float).Sub(
			new(big.Float).Mul(b1, a2),
			new(big.Float).Mul(b2, a1),
		),
	)

	if l0.Sign() < 0 || l1.Sign() < 0 {
		return struct {
			IsoK     bool
			TwoCurve types.GetFirstCurveResult
		}{
			TwoCurve: types.GetFirstCurveResult{
				SqrtStartPrice: big.NewInt(0),
				Curve:          []dbc.LiquidityDistributionParameters{},
			},
		}
	}

	l0BigInt, l1BigInt := new(big.Int), new(big.Int)
	l0.Int(l0BigInt)
	l1.Int(l1BigInt)

	return struct {
		IsoK     bool
		TwoCurve types.GetFirstCurveResult
	}{
		IsoK: true,
		TwoCurve: types.GetFirstCurveResult{
			SqrtStartPrice: initialSqrtPrice,
			Curve: []dbc.LiquidityDistributionParameters{
				{
					SqrtPrice: MustBigIntToUint128(midSqrtPrice),
					Liquidity: MustBigIntToUint128(l0BigInt),
				},
				{
					SqrtPrice: MustBigIntToUint128(migrationSqrtPrice),
					Liquidity: MustBigIntToUint128(l1BigInt),
				},
			},
		},
	}
}

func Liquidity(
	baseAmount, quoteAmount, minSqrtPrice, maxSqrtPrice *big.Int,
) (*big.Int, error) {

	liquidityFromBase, err := GetInitialLiquidityFromDeltaBase(
		baseAmount,
		maxSqrtPrice,
		minSqrtPrice,
	)
	if err != nil {
		return nil, err
	}
	liquidityFromQuote, err := GetInitialLiquidityFromDeltaQuote(
		quoteAmount,
		minSqrtPrice,
		maxSqrtPrice,
	)
	if err != nil {
		return nil, err
	}

	if liquidityFromBase.Cmp(liquidityFromQuote) < 0 {
		return liquidityFromBase, nil
	}

	return liquidityFromQuote, nil
}

func CalculateRateLimiterFee(
	params types.BaseFee, inputAmount *big.Int,
) *big.Int {
	// for input_amount <= reference_amount
	// --> fee = input_amount * cliff_fee_numerator

	// for input_amount > reference_amount

	// let x0 = reference_amount
	// let c = cliff_fee_numerator
	// let i = fee_increment
	// let a = (input_amount - x0) / x0 (integer division)
	// let b = (input_amount - x0) % x0 (remainder)

	// max_index =
	//     (MAX_FEE_NUMERATOR - cliff_fee_numerator) / fee_increment_numerator
	// where: fee_increment_numerator =
	//     (fee_increment_bps * FEE_DENOMINATOR) / 10_000

	// if a < max_index:
	// --> fee = x0 * (c + c*a + i*a*(a+1)/2) + b * (c + i*(a+1))

	// if a ≥ max_index:
	// --> fee = x0 * (c + c*max_index + i*max_index*(max_index+1)/2) + (d*x0 + b) * MAX_FEE
	// where:
	// d = a - max_index
	// MAX_FEE is the maximum allowed fee (9900 bps)

	// for input_amount <= reference_amount
	if inputAmount.Cmp(new(big.Int).SetUint64(params.ThirdFactor)) <= 0 {
		return new(big.Int).Quo(
			new(big.Int).Mul(inputAmount, new(big.Int).SetUint64(params.CliffFeeNumerator)),
			constants.FeeDenominatorBigInt,
		)
	}

	// for input_amount > reference_amount
	x0, c, feeIncrementNumerator := new(big.Int).SetUint64(params.ThirdFactor),
		new(big.Int).SetUint64(params.CliffFeeNumerator), BpsToFeeNumerator(uint64(params.FirstFactor))

	// calculate a and b
	diff := new(big.Int).Sub(inputAmount, x0)
	a, b := new(big.Int).Quo(diff, x0), new(big.Int).Mod(diff, x0)

	// calculate max_index
	maxFeeNumerator := big.NewInt(constants.MaxFeeNumerator)
	deltaNumerator := new(big.Int).Sub(maxFeeNumerator, c)
	maxIndex := new(big.Int).Quo(deltaNumerator, feeIncrementNumerator)

	if a.Cmp(maxIndex) < 0 {
		// if a < max_index
		numerator1 := new(big.Int).Add(
			new(big.Int).Add(c, new(big.Int).Mul(c, a)),
			new(big.Int).Quo(
				new(big.Int).Mul(
					new(big.Int).Mul(feeIncrementNumerator, a),
					new(big.Int).Add(a, big.NewInt(1)),
				),
				big.NewInt(2),
			),
		)
		numerator2 := new(big.Int).Add(
			c,
			new(big.Int).Mul(
				feeIncrementNumerator,
				new(big.Int).Add(a, big.NewInt(1)),
			),
		)
		firstFee := new(big.Int).Mul(x0, numerator1)
		secondFee := new(big.Int).Mul(b, numerator2)
		fee := new(big.Int).Add(firstFee, secondFee)
		return new(big.Int).Quo(fee, constants.FeeDenominatorBigInt)
	}

	// if a >= max_index
	numerator1, numerator2 := new(big.Int).Add(
		new(big.Int).Add(c, new(big.Int).Mul(c, maxIndex)),
		new(big.Int).Quo(
			new(big.Int).Mul(
				new(big.Int).Mul(feeIncrementNumerator, maxIndex),
				new(big.Int).Add(maxIndex, big.NewInt(1)),
			),
			big.NewInt(2),
		),
	), new(big.Int).Set(maxFeeNumerator)

	firstFee, d := new(big.Int).Mul(x0, numerator1), new(big.Int).Sub(a, maxIndex)
	leftAmount := new(big.Int).Add(new(big.Int).Mul(d, x0), b)
	secondFee := new(big.Int).Mul(leftAmount, numerator2)
	fee := new(big.Int).Add(firstFee, secondFee)
	return new(big.Int).Quo(fee, constants.FeeDenominatorBigInt)
}

// GetCurrentPoint gets the current point based on activation type.
func GetCurrentPoint(
	conn *rpc.Client,
	activationType types.ActivationType,
) (*big.Int, error) {
	currentSlot, err := conn.GetSlot(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		return nil, fmt.Errorf("GetCurrentPoint:%w", err)
	}

	if activationType == types.ActivationTypeSlot {
		return new(big.Int).SetUint64(currentSlot), nil
	}

	currentTime, err := conn.GetBlockTime(context.Background(), currentSlot)
	if err != nil {
		return nil, fmt.Errorf("GetCurrentPoint:%w", err)
	}

	if currentTime == nil {
		return nil, fmt.Errorf("GetCurrentPoint:timestamp is not available for this block(%d)", currentSlot)
	}

	return big.NewInt(int64(*currentTime)), nil
}

// PrepareSwapAmountParam prepares the swap amount param.
func PrepareSwapAmountParam(
	amount float64,
	mintAddress solana.PublicKey,
	conn *rpc.Client,
) (*big.Int, error) {
	mintTokenDecimals, err := GetTokenDecimals(conn, mintAddress)
	if err != nil {
		return nil, err
	}

	return ConvertToLamports(amount, types.TokenDecimal(mintTokenDecimals)), nil
}
