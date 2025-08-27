package types

type TokenType uint8

const (
	TokenTypeSPL TokenType = iota
	TokenTypeToken2022
)

type SwapMode uint8

const (
	SwapModeExactIn = iota
	SwapModePartialFill
	SwapModeExactOut
)

type BaseFeeMode uint8

const (
	BaseFeeModeFeeSchedulerLinear BaseFeeMode = iota
	BaseFeeModeFeeSchedulerExponential
	BaseFeeModeRateLimiter
)

type ActivationType uint8

const (
	ActivationTypeSlot ActivationType = iota
	ActivationTypeTimestamp
)

type TradeDirection uint8

const (
	TradeDirectionBaseToQuote TradeDirection = iota
	TradeDirectionQuoteToBase
)

type CollectFeeMode uint8

const (
	CollectFeeModeQuoteToken CollectFeeMode = iota
	CollectFeeModeOutputToken
)

type Rounding uint

const (
	RoundingUp Rounding = iota
	RoundingDown
)

type MigrationOption uint8

const (
	MigrationOptionMET_DAMM MigrationOption = iota
	MigrationOptionMET_DAMM_V2
)

type TokenDecimal uint8

const (
	TokenDecimalSIX   TokenDecimal = 6
	TokenDecimalSEVEN TokenDecimal = 7
	TokenDecimalEIGHT TokenDecimal = 8
	TokenDecimalNINE  TokenDecimal = 9
)

type MigrationFeeOption uint8

const (
	MigrationFeeOptionFixedBps25 MigrationFeeOption = iota
	MigrationFeeOptionFixedBps30
	MigrationFeeOptionFixedBps100
	MigrationFeeOptionFixedBps200
	MigrationFeeOptionFixedBps400
	MigrationFeeOptionFixedBps600
	MigrationFeeOptionCustomizable // only for DAMM v2

)

// TokenUpdateAuthorityOption represents the update authority permission options for a token.
type TokenUpdateAuthorityOption uint8

const (
	// TokenUpdateAuthorityOptionCreatorUpdateAuthority means the creator can update the update_authority.
	TokenUpdateAuthorityOptionCreatorUpdateAuthority TokenUpdateAuthorityOption = iota

	// TokenUpdateAuthorityOptionImmutable means no one can update the update_authority.
	TokenUpdateAuthorityOptionImmutable

	// TokenUpdateAuthorityOptionPartnerUpdateAuthority means the partner can update the update_authority.
	TokenUpdateAuthorityOptionPartnerUpdateAuthority

	// TokenUpdateAuthorityOptionCreatorUpdateAndMintAuthority means the creator can update both update_authority and mint_authority.
	TokenUpdateAuthorityOptionCreatorUpdateAndMintAuthority

	// TokenUpdateAuthorityOptionPartnerUpdateAndMintAuthority means the partner can update both update_authority and mint_authority.
	TokenUpdateAuthorityOptionPartnerUpdateAndMintAuthority
)
