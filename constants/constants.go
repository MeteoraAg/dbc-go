package constants

import (
	"dbcGoSDK/generated/dammv1"
	"dbcGoSDK/generated/dammv2"
	"dbcGoSDK/generated/dbc"
	"math"
	"math/big"

	"github.com/gagliardetto/solana-go"
)

const (
	OFFSET         = 64
	RESOLUTION     = 64
	FeeDenominator = 1_000_000_000

	// MaxFeeBPS
	//  99%
	MaxFeeBPS = 9900

	// MinFeeBPS
	//  0.0001%
	MinFeeBPS = 1

	// MinFeeNumerator
	//  0.0001%
	MinFeeNumerator = 100_000

	// MaxFeeNumerator
	//  99%
	MaxFeeNumerator = 990_000_000

	// BasisPointMax
	BasisPointMax = 10_000

	// MaxCurvePoint
	MaxCurvePoint = 16

	// PartnerSurplusShare
	//  80%
	PartnerSurplusShare = 80

	// SwapBufferPercentage
	//  25%
	SwapBufferPercentage = 25

	// MaxMigrationFeePercentage
	//  50%
	MaxMigrationFeePercentage = 50

	// MaxCreatorMigrationFeePercentage
	//  100%
	MaxCreatorMigrationFeePercentage = 100

	// MaxRateLimiterDurationInSeconds
	//  12 hours
	MaxRateLimiterDurationInSeconds = 43200

	// MaxRateLimiterDurationInSlots
	//  12 hours
	MaxRateLimiterDurationInSlots = 108000

	// SlotDuration
	SlotDuration = 400

	// TimestampDuration
	TimestampDuration = 1000

	// DynamicFee constants
	//

	DynamicFeeFilterPeriodDefault = 10  // 10 seconds
	DynamicFeeDecayPeriodDefault  = 120 // 120 seconds

	MaxDynamicPercentage = 20 // 20% of base fee
	MaxSwallowPercentage = 20 // 20%

	MinMigratedPoolFeeBps = 10   // 0.1%
	MaxMigratedPoolFeeBps = 1000 // 10%

	DynamicFeeReductionFactorDefault = 5000 // 50%

	BinStepBpsDefault = 1

	// MaxPriceChangeBpsDefault
	//  15%
	MaxPriceChangeBpsDefault = 1500
)

var (
	// DynamicFeeScalingFactor
	//  DynamicFeeScalingFactor = new(big.Int).SetUint64(100_000_000_000)
	DynamicFeeScalingFactor = new(big.Int).SetUint64(100_000_000_000)

	// DynamicFeeRoundingOffset
	//  DynamicFeeRoundingOffset = new(big.Int).SetUint64(99_999_999_999)
	DynamicFeeRoundingOffset = new(big.Int).SetUint64(99_999_999_999)

	// HundredInBigFloat
	//  HundredInBigFloat = big.NewFloat(100)
	HundredInBigFloat = big.NewFloat(100)

	// U64MaxBigInt
	//  U64MaxBigInt = new(big.Int).SetUint64(math.MaxUint64)
	U64MaxBigInt = new(big.Int).SetUint64(math.MaxUint64)

	// U128MaxBigInt
	//  U128MaxBigInt = new(big.Int).SetString("340282366920938463463374607431768211455", 10)
	U128MaxBigInt, _ = new(big.Int).SetString("340282366920938463463374607431768211455", 10)

	// HundredInBigInt
	//  HundredInBigInt = big.NewInt(100)
	HundredInBigInt = big.NewInt(100)

	// FeeDenominatorBigInt
	//  FeeDenominatorBigInt = big.NewInt(FeeDenominator)
	FeeDenominatorBigInt = big.NewInt(FeeDenominator)

	// Offset resolution //

	// MinSqrtPrice
	//  MinSqrtPrice = big.NewInt(4295048016)
	MinSqrtPrice = big.NewInt(4295048016)

	// MaxSqrtPrice
	// MaxSqrtPrice, _ = new(big.Int).SetString("79226673521066979257578248091", 10)
	MaxSqrtPrice, _ = new(big.Int).SetString("79226673521066979257578248091", 10)

	// OneQ64 = 1 << RESOLUTION
	//  OneQ64 = new(big.Int).Lsh(big.NewInt(1), RESOLUTION)
	OneQ64 = new(big.Int).Lsh(big.NewInt(1), RESOLUTION)

	// DBC program ID.
	//  DBCProgramId = solana.MustPublicKeyFromBase58("dbcij3LWUppWqq96dh6gJWwBifmcGfLSB5D4DuSMaqN")
	DBCProgramId = dbc.ProgramID

	// Metaplex program ID.
	// MetaplexProgramId = solana.MustPublicKeyFromBase58("metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s")
	MetaplexProgramId = solana.MustPublicKeyFromBase58("metaqbxxUerdq28cj1RbAWkYQm3ybzjb6a8bt518x1s")

	// DAMM v1 program ID.
	//  DammV1ProgramId = solana.MustPublicKeyFromBase58("Eo7WjKq67rjJQSZxS6z3YkapzY3eMj6Xy8X5EQVn5UaB")
	DammV1ProgramId = dammv1.ProgramID

	// DAMM v2 program ID.
	//  CpAMMProgramId = solana.MustPublicKeyFromBase58("cpamdpZCGKUy5JxQXB4dcpGPiikHawvSWAd6mEn1sGG")
	DammV2ProgramId = dammv2.ProgramID

	// Vault program ID.
	//  VaultProgramId = solana.MustPublicKeyFromBase58("24Uqj9JCLxUeoC3hGfh5W3s9FM9uCHDS2SG3LYwBpyTi")
	VaultProgramId = solana.MustPublicKeyFromBase58("24Uqj9JCLxUeoC3hGfh5W3s9FM9uCHDS2SG3LYwBpyTi")

	// Locker program ID.
	//  LockerProgramId = solana.MustPublicKeyFromBase58("LocpQgucEQHbqNABEYvBvwoxCPsSbG91A1QaQhQQqjn")
	LockerProgramId = solana.MustPublicKeyFromBase58("LocpQgucEQHbqNABEYvBvwoxCPsSbG91A1QaQhQQqjn")

	// Base address.
	//  BaseAddress = solana.MustPublicKeyFromBase58("HWzXGcGHy4tcpYfaRDCyLNzXqBTv3E6BttpCH2vJxArv")
	BaseAddress = solana.MustPublicKeyFromBase58("HWzXGcGHy4tcpYfaRDCyLNzXqBTv3E6BttpCH2vJxArv")

	// BinStepBpsU128Default
	//  BinStepBpsU128Default, _ = new(big.Int).SetString("1844674407370955", 10)
	BinStepBpsU128Default, _ = new(big.Int).SetString("1844674407370955", 10)

	// DAMM V1 Migration Fee Addresses
	//
	// DammV1MigrationFeeAddresses = []solana.PublicKey{
	// 	solana.MustPublicKeyFromBase58("8f848CEy8eY6PhJ3VcemtBDzPPSD4Vq7aJczLZ3o8MmX"), FixedBps25
	// 	solana.MustPublicKeyFromBase58("HBxB8Lf14Yj8pqeJ8C4qDb5ryHL7xwpuykz31BLNYr7S"), FixedBps30
	// 	solana.MustPublicKeyFromBase58("7v5vBdUQHTNeqk1HnduiXcgbvCyVEZ612HLmYkQoAkik"), FixedBps100
	// 	solana.MustPublicKeyFromBase58("EkvP7d5yKxovj884d2DwmBQbrHUWRLGK6bympzrkXGja"), FixedBps200
	// 	solana.MustPublicKeyFromBase58("9EZYAJrcqNWNQzP2trzZesP7XKMHA1jEomHzbRsdX8R2"), FixedBps400
	// 	solana.MustPublicKeyFromBase58("8cdKo87jZU2R12KY1BUjjRPwyjgdNjLGqSGQyrDshhud"), FixedBps600
	// }
	DammV1MigrationFeeAddresses = []solana.PublicKey{
		solana.MustPublicKeyFromBase58("8f848CEy8eY6PhJ3VcemtBDzPPSD4Vq7aJczLZ3o8MmX"),
		solana.MustPublicKeyFromBase58("HBxB8Lf14Yj8pqeJ8C4qDb5ryHL7xwpuykz31BLNYr7S"),
		solana.MustPublicKeyFromBase58("7v5vBdUQHTNeqk1HnduiXcgbvCyVEZ612HLmYkQoAkik"),
		solana.MustPublicKeyFromBase58("EkvP7d5yKxovj884d2DwmBQbrHUWRLGK6bympzrkXGja"),
		solana.MustPublicKeyFromBase58("9EZYAJrcqNWNQzP2trzZesP7XKMHA1jEomHzbRsdX8R2"),
		solana.MustPublicKeyFromBase58("8cdKo87jZU2R12KY1BUjjRPwyjgdNjLGqSGQyrDshhud"),
	}

	// DAMM V2 Migration Fee Addresses
	// DammV2MigrationFeeAddresses = []solana.PublicKey{
	// 	solana.MustPublicKeyFromBase58("7F6dnUcRuyM2TwR8myT1dYypFXpPSxqwKNSFNkxyNESd"), FixedBps25
	// 	solana.MustPublicKeyFromBase58("2nHK1kju6XjphBLbNxpM5XRGFj7p9U8vvNzyZiha1z6k"), FixedBps30
	// 	solana.MustPublicKeyFromBase58("Hv8Lmzmnju6m7kcokVKvwqz7QPmdX9XfKjJsXz8RXcjp"), FixedBps100
	// 	solana.MustPublicKeyFromBase58("2c4cYd4reUYVRAB9kUUkrq55VPyy2FNQ3FDL4o12JXmq"), FixedBps200
	// 	solana.MustPublicKeyFromBase58("AkmQWebAwFvWk55wBoCr5D62C6VVDTzi84NJuD9H7cFD"), FixedBps400
	// 	solana.MustPublicKeyFromBase58("DbCRBj8McvPYHJG1ukj8RE15h2dCNUdTAESG49XpQ44u"), FixedBps600
	// solana.MustPublicKeyFromBase58("A8gMrEPJkacWkcb3DGwtJwTe16HktSEfvwtuDh2MCtck"), Customizable
	// }
	DammV2MigrationFeeAddresses = []solana.PublicKey{
		solana.MustPublicKeyFromBase58("7F6dnUcRuyM2TwR8myT1dYypFXpPSxqwKNSFNkxyNESd"),
		solana.MustPublicKeyFromBase58("2nHK1kju6XjphBLbNxpM5XRGFj7p9U8vvNzyZiha1z6k"),
		solana.MustPublicKeyFromBase58("Hv8Lmzmnju6m7kcokVKvwqz7QPmdX9XfKjJsXz8RXcjp"),
		solana.MustPublicKeyFromBase58("2c4cYd4reUYVRAB9kUUkrq55VPyy2FNQ3FDL4o12JXmq"),
		solana.MustPublicKeyFromBase58("AkmQWebAwFvWk55wBoCr5D62C6VVDTzi84NJuD9H7cFD"),
		solana.MustPublicKeyFromBase58("DbCRBj8McvPYHJG1ukj8RE15h2dCNUdTAESG49XpQ44u"),
		solana.MustPublicKeyFromBase58("A8gMrEPJkacWkcb3DGwtJwTe16HktSEfvwtuDh2MCtck"),
	}
)

const (
	SeedPoolAuthority           = "pool_authority"
	SeedEventAuthority          = "__event_authority"
	SeedPool                    = "pool"
	SeedTokenVault              = "token_vault"
	SeedMetadata                = "metadata"
	SeedPartnerMetadata         = "partner_metadata"
	SeedClaimFeeOperator        = "cf_operator"
	SeedDammV1MigrationMetadata = "meteora"
	SeedDammV2MigrationMetadata = "damm_v2"
	SeedLpMint                  = "lp_mint"
	SeedFee                     = "fee"
	SeedPosition                = "position"
	SeedPositionNFTAccount      = "position_nft_account"
	SeedLockEscrow              = "lock_escrow"
	SeedVirtualPoolMetadata     = "virtual_pool_metadata"
	SeedEscrow                  = "escrow"
	SeedBaseLocker              = "base_locker"
	SeedVault                   = "vault"
)
