package types

import (
	"dbcGoSDK/generated/dbc"
	"math/big"

	"github.com/gagliardetto/solana-go"
)

type PrepareTokenAccountParams struct {
	Owner         solana.PublicKey
	Payer         solana.PublicKey
	TokenAMint    solana.PublicKey
	TokenBMint    solana.PublicKey
	TokenAProgram solana.PublicKey
	TokenBProgram solana.PublicKey
}

type FeeBreakdown struct {
	PartnerBaseFee  uint64 `json:"partnerBaseFee"`
	PartnerQuoteFee uint64 `json:"partnerQuoteFee"`
	CreatorBaseFee  uint64 `json:"creatorBaseFee"`
	CreatorQuoteFee uint64 `json:"creatorQuoteFee"`
}

type FeeTotal struct {
	TotalTradingBaseFee  uint64 `json:"totalTradingBaseFee"`
	TotalTradingQuoteFee uint64 `json:"totalTradingQuoteFee"`
}

type PoolFeeMetrics struct {
	Current FeeBreakdown `json:"current"`
	Total   FeeTotal     `json:"total"`
}

type PoolFeeByConfigOrCreator struct {
	PoolAddress          solana.PublicKey
	PartnerBaseFee       uint64
	PartnerQuoteFee      uint64
	CreatorBaseFee       uint64
	CreatorQuoteFee      uint64
	TotalTradingBaseFee  uint64
	TotalTradingQuoteFee uint64
}

type CreateVirtualPoolMetadataParam struct {
	VirtualPool solana.PublicKey
	Name        string
	Website     string
	Logo        string
	Creator     solana.PublicKey
	Payer       solana.PublicKey
}

type ClaimCreatorTradingFeeWithQuoteMintNotSolParam struct {
	Creator           solana.PublicKey
	Payer             solana.PublicKey
	FeeReceiver       solana.PublicKey
	Pool              solana.PublicKey
	PoolState         *dbc.VirtualPoolAccount
	PoolConfigState   *dbc.PoolConfigAccount
	TokenBaseProgram  solana.PublicKey
	TokenQuoteProgram solana.PublicKey
}

type ClaimCreatorTradingFeeWithQuoteMintSolParam struct {
	ClaimCreatorTradingFeeWithQuoteMintNotSolParam
	TempWSolAcc solana.PublicKey
}

type ClaimCreatorTradingFeeParam struct {
	Creator        solana.PublicKey
	Payer          solana.PublicKey
	Pool           solana.PublicKey
	MaxBaseAmount  uint64
	MaxQuoteAmount uint64
	Receiver       solana.PublicKey
	TempWSolAcc    solana.PublicKey
}

type ClaimCreatorTradingFee2Param struct {
	Creator        solana.PublicKey
	Payer          solana.PublicKey
	Pool           solana.PublicKey
	MaxBaseAmount  uint64
	MaxQuoteAmount uint64
	Receiver       solana.PublicKey
}

type CreatorWithdrawSurplusParam struct {
	Creator, VirtualPool solana.PublicKey
}

type TransferPoolCreatorParam struct {
	VirtualPool, Creator, NewCreator solana.PublicKey
}

type WithdrawMigrationFeeParam struct {
	VirtualPool solana.PublicKey
	Sender      solana.PublicKey // sender is creator or partner
	FeePayer    *solana.PublicKey
}

type InitializePoolBaseParam struct {
	Name         string
	Symbol       string
	URI          string
	Pool         solana.PublicKey
	Config       solana.PublicKey
	Payer        solana.PublicKey
	PoolCreator  solana.PublicKey
	BaseMint     solana.PublicKey
	BaseVault    solana.PublicKey
	QuoteVault   solana.PublicKey
	QuoteMint    solana.PublicKey
	MintMetadata solana.PublicKey
}

type VirtualPoolState struct {
	BaseMint solana.PublicKey
	PoolType TokenType
}

type PoolConfigState struct {
	QuoteMint      solana.PublicKey
	QuoteTokenFlag TokenType
}

type PrepareSwapParams struct {
	InputMint          solana.PublicKey
	OutputMint         solana.PublicKey
	InputTokenProgram  solana.PublicKey
	OutputTokenProgram solana.PublicKey
}

type CreatePoolParam struct {
	PreCreatePoolParam
	Payer  solana.PublicKey
	Config solana.PublicKey
}

type FirstBuyParam struct {
	Buyer                solana.PublicKey
	Receiver             solana.PublicKey
	BuyAmount            uint64
	MinimumAmountOut     uint64
	ReferralTokenAccount solana.PublicKey
}

type PreCreatePoolParam struct {
	Name        string
	Symbol      string
	URI         string
	PoolCreator solana.PublicKey
	BaseMint    solana.PublicKey
}

type CreateConfigAndPoolParam struct {
	PreCreatePoolParam
	CreateConfigParam
	TokenType TokenType
}
type CreateConfigParam struct {
	dbc.ConfigParameters
	Config           solana.PublicKey
	FeeClaimer       solana.PublicKey
	LeftoverReceiver solana.PublicKey
	QuoteMint        solana.PublicKey
	Payer            solana.PublicKey
}

type CreateConfigAndPoolWithFirstBuyParam struct {
	FirstBuyParam
	CreateConfigAndPoolParam
	PreCreatePoolParam
	BaseFeeMode
}

type CreatePoolWithFirstBuyParam struct {
	CreatePoolParam
	FirstBuyParam
}

type CreatePoolWithPartnerAndCreatorFirstBuyParam struct {
	CreatePoolParam      CreatePoolParam
	PartnerFirstBuyParam PartnerFirstBuyParam
	CreatorFirstBuyParam CreatorFirstBuyParam
}

type PartnerFirstBuyParam struct {
	Partner              solana.PublicKey
	Receiver             solana.PublicKey
	BuyAmount            uint64
	MinimumAmountOut     uint64
	ReferralTokenAccount solana.PublicKey
}
type CreatorFirstBuyParam struct {
	Creator              solana.PublicKey
	Receiver             solana.PublicKey
	BuyAmount            uint64
	MinimumAmountOut     uint64
	ReferralTokenAccount solana.PublicKey
}

type SwapParam struct {
	Owner                solana.PublicKey
	Pool                 solana.PublicKey
	AmountIn             uint64
	MinimumAmountOut     uint64
	SwapBaseForQuote     bool
	ReferralTokenAccount solana.PublicKey
	Payer                solana.PublicKey
}

type Swap2Param struct {
	Owner                solana.PublicKey
	Pool                 solana.PublicKey
	SwapBaseForQuote     bool
	ReferralTokenAccount solana.PublicKey
	Payer                solana.PublicKey
	SwapMode             SwapMode

	// ExactIn & PartialFill
	AmountIn         *big.Int
	MinimumAmountOut *big.Int

	// ExactOut
	AmountOut       *big.Int
	MaximumAmountIn *big.Int
}

type SwapQuote2Param struct {
	VirtualPool      *dbc.VirtualPoolAccount
	Config           *dbc.PoolConfigAccount
	SwapBaseForQuote bool
	HasReferral      bool
	CurrentPoint     *big.Int
	SlippageBps      uint64
	SwapMode         SwapMode

	// ExactIn & PartialFill
	AmountIn         *big.Int
	MinimumAmountOut *big.Int

	// ExactOut
	AmountOut       *big.Int
	MaximumAmountIn *big.Int
}

type SwapQuoteResult struct {
	dbc.SwapResult
	MinimumAmountOut uint64
}

type QuoteFee struct {
	Trading  *big.Int
	Protocol *big.Int
	Referral *big.Int // Can be nil if optional
}

type QuotePrice struct {
	BeforeSwap *big.Int
	AfterSwap  *big.Int
}

type FeeMode struct {
	FeesOnInput     bool
	FeesOnBaseToken bool
	HasReferral     bool
}

type FeeOnAmountResult struct {
	Amount      *big.Int
	ProtocolFee *big.Int
	TradingFee  *big.Int
	ReferralFee *big.Int
}

type SwapAmount struct {
	OutputAmount  *big.Int
	NextSqrtPrice *big.Int
	AmountLeft    *big.Int
}

type SwapQuoteParam struct {
	VirtualPool      *dbc.VirtualPoolAccount
	Config           *dbc.PoolConfigAccount
	SwapBaseForQuote bool
	AmountIn         *big.Int
	SlippageBps      uint64 // optional
	HasReferral      bool
	CurrentPoint     *big.Int
}

type SwapQuoteExactInParam struct {
	VirtualPool  *dbc.VirtualPoolAccount
	Config       *dbc.PoolConfigAccount
	CurrentPoint *big.Int
}

type SwapQuoteExactOutParam struct {
	VirtualPool      *dbc.VirtualPoolAccount
	Config           *dbc.PoolConfigAccount
	SwapBaseForQuote bool
	OutAmount        *big.Int
	SlippageBps      uint64 // optional, use pointer to distinguish zero vs unset
	HasReferral      bool
	CurrentPoint     *big.Int
}

type CreatePartnerMetadataParam struct {
	Name       string
	Website    string
	Logo       string
	FeeClaimer solana.PublicKey
	Payer      solana.PublicKey
}

type ClaimPartnerTradingFeeWithQuoteMintNotSolParam struct {
	FeeClaimer        solana.PublicKey
	Payer             solana.PublicKey
	FeeReceiver       solana.PublicKey
	Config            solana.PublicKey
	Pool              solana.PublicKey
	PoolState         *dbc.VirtualPoolAccount
	PoolConfigState   *dbc.PoolConfigAccount
	TokenBaseProgram  solana.PublicKey
	TokenQuoteProgram solana.PublicKey
}

type ClaimPartnerTradingFeeWithQuoteMintSolParam struct {
	ClaimPartnerTradingFeeWithQuoteMintNotSolParam
	TempWSolAcc solana.PublicKey
}

type Accounts struct {
	PoolAuthority     solana.PublicKey
	Config            solana.PublicKey
	Pool              solana.PublicKey
	TokenAAccount     solana.PublicKey
	TokenBAccount     solana.PublicKey
	BaseVault         solana.PublicKey
	QuoteVault        solana.PublicKey
	BaseMint          solana.PublicKey
	QuoteMint         solana.PublicKey
	FeeClaimer        solana.PublicKey
	TokenBaseProgram  solana.PublicKey
	TokenQuoteProgram solana.PublicKey
}

type ClaimTradingFeeParam struct {
	FeeClaimer     solana.PublicKey
	Payer          solana.PublicKey
	Pool           solana.PublicKey
	MaxBaseAmount  *big.Int
	MaxQuoteAmount *big.Int
	Receiver       solana.PublicKey
	TempWSolAcc    solana.PublicKey
}

type PartnerWithdrawSurplusParam struct {
	FeeClaimer  solana.PublicKey
	VirtualPool solana.PublicKey
}
type CreateLockerParam struct {
	Payer       solana.PublicKey
	VirtualPool solana.PublicKey
}

type WithdrawLeftoverParam struct {
	Payer       solana.PublicKey
	VirtualPool solana.PublicKey
}

type CreateDammV1MigrationMetadataParam struct {
	Payer       solana.PublicKey
	VirtualPool solana.PublicKey
	Config      solana.PublicKey
}

type MigrateToDammV1Param struct {
	Payer       solana.PublicKey
	VirtualPool solana.PublicKey
	DammConfig  solana.PublicKey
}

type DammLpTokenParam struct {
	Payer       solana.PublicKey
	VirtualPool solana.PublicKey
	DammConfig  solana.PublicKey
	IsPartner   bool
}

type CreateDammV2MigrationMetadataParam struct {
	CreateDammV1MigrationMetadataParam
}

type MigrateToDammV2Param struct {
	MigrateToDammV1Param
}

type MigrateToDammV2Response struct {
	FirstPositionNftKeypair, SecondPositionNftKeypair solana.PrivateKey
	Ixns                                              []solana.Instruction
}

type BuildCurveBaseParam struct {
	TotalTokenSupply            uint64
	MigrationOption             MigrationOption
	TokenBaseDecimal            TokenDecimal
	TokenQuoteDecimal           TokenDecimal
	LockedVestingParam          LockedVestingParams
	BaseFeeParams               BaseFeeParams
	DynamicFeeEnabled           bool
	ActivationType              ActivationType
	CollectFeeMode              CollectFeeMode
	MigrationFeeOption          MigrationFeeOption
	TokenType                   TokenType
	PartnerLpPercentage         uint8
	CreatorLpPercentage         uint8
	PartnerLockedLpPercentage   uint8
	CreatorLockedLpPercentage   uint8
	CreatorTradingFeePercentage uint8
	Leftover                    uint64
	TokenUpdateAuthority        uint8
	MigrationFee                MigrationFee
	MigratedPoolFee             dbc.MigratedPoolFee
}

type MigrationFee struct {
	FeePercentage        float64
	CreatorFeePercentage float64
}

type BuildCurveParam struct {
	BuildCurveBaseParam
	PercentageSupplyOnMigration float64
	MigrationQuoteThreshold     float64
}

type BuildCurveWithMarketCapParam struct {
	BuildCurveBaseParam
	InitialMarketCap   float64
	MigrationMarketCap float64
}

type BuildCurveWithTwoSegmentsParam struct {
	BuildCurveBaseParam
	InitialMarketCap            uint64
	MigrationMarketCap          uint64
	PercentageSupplyOnMigration uint8
}

type BuildCurveWithLiquidityWeightsParam struct {
	BuildCurveBaseParam
	InitialMarketCap   uint64
	MigrationMarketCap uint64
	LiquidityWeights   []float64
}

type LockedVestingParams struct {
	TotalLockedVestingAmount       uint64
	NumberOfVestingPeriod          uint64
	CliffUnlockAmount              uint64
	TotalVestingDuration           uint64
	CliffDurationFromMigrationTime uint64
}

type BaseFeeParams struct {
	BaseFeeMode       BaseFeeMode
	FeeSchedulerParam *FeeSchedulerParams
	RateLimiterParam  *RateLimiterParams
}

type FeeSchedulerParams struct {
	StartingFeeBps uint64
	EndingFeeBps   uint64
	NumberOfPeriod uint16
	TotalDuration  uint64
}
type RateLimiterParams struct {
	BaseFeeBps         uint64
	FeeIncrementBps    uint16
	ReferenceAmount    float64
	MaxLimiterDuration uint64
}

type BaseFee struct {
	CliffFeeNumerator uint64
	FirstFactor       uint16 // feeScheduler: numberOfPeriod, rateLimiter: feeIncrementBps
	SecondFactor      uint64 // feeScheduler: periodFrequency, rateLimiter: maxLimiterDuration
	ThirdFactor       uint64 // feeScheduler: reductionFactor, rateLimiter: referenceAmount
	BaseFeeMode       BaseFeeMode
}

type LockedVestingParamsBigInt struct {
	AmountPerPeriod                *big.Int
	CliffDurationFromMigrationTime *big.Int
	Frequency                      *big.Int
	NumberOfPeriod                 *big.Int
	CliffUnlockAmount              *big.Int
}

type GetFirstCurveResult struct {
	SqrtStartPrice *big.Int
	Curve          []dbc.LiquidityDistributionParameters
}

type SwapQuote2Result struct {
	dbc.SwapResult2
	MinimumAmountOut uint64
	MaximumAmountIn  uint64
}
