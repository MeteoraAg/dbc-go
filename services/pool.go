package services

import (
	"context"
	"dbcGoSDK/constants"
	"dbcGoSDK/generated/dbc"
	"dbcGoSDK/helpers"
	"dbcGoSDK/maths"
	mathsPoolfees "dbcGoSDK/maths/poolFees"
	"dbcGoSDK/types"
	"errors"
	"fmt"
	"math/big"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"golang.org/x/sync/errgroup"
)

type PoolService struct {
	state *StateService
}

func NewPoolService(
	conn *rpc.Client,
	commitment rpc.CommitmentType,
) *PoolService {
	return &PoolService{
		state: NewStateService(conn, commitment),
	}
}

// initializeSplPool initialize a pool with SPL token.
func (p *PoolService) initializeSplPool(
	param types.InitializePoolBaseParam,
) (*dbc.Instruction, error) {
	initializeVirtualPoolWithSplTokenPtr := dbc.NewInitializeVirtualPoolWithSplTokenInstruction(
		dbc.InitializePoolParameters{
			Name:   param.Name,
			Symbol: param.Symbol,
			Uri:    param.URI,
		},
		param.Config,
		p.state.GetPoolAuthority(),
		param.PoolCreator,
		param.BaseMint,
		param.QuoteMint,
		param.Pool,
		param.BaseVault,
		param.QuoteVault,
		param.MintMetadata,
		constants.MetaplexProgramId,
		param.Payer,
		solana.TokenProgramID,
		solana.TokenProgramID,
		solana.SystemProgramID,
		solana.PublicKey{},
		constants.DBCProgramId,
	)
	eventAuthPDA, _, err := initializeVirtualPoolWithSplTokenPtr.FindEventAuthorityAddress()
	if err != nil {
		return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
	}

	return initializeVirtualPoolWithSplTokenPtr.
		SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
}

// initializeToken2022Pool initialize a pool with token22.
func (p *PoolService) initializeToken2022Pool(
	param types.InitializePoolBaseParam,
) (*dbc.Instruction, error) {
	initializeVirtualPoolWithToken2022Ptr := dbc.NewInitializeVirtualPoolWithToken2022Instruction(
		dbc.InitializePoolParameters{
			Name:   param.Name,
			Symbol: param.Symbol,
			Uri:    param.URI,
		},
		param.Config,
		p.state.GetPoolAuthority(),
		param.PoolCreator,
		param.BaseMint,
		param.QuoteMint,
		param.Pool,
		param.BaseVault,
		param.QuoteVault,
		param.Payer,
		solana.TokenProgramID,
		solana.Token2022ProgramID,
		solana.SystemProgramID,
		solana.PublicKey{},
		constants.DBCProgramId,
	)

	eventAuthPDA, _, err := initializeVirtualPoolWithToken2022Ptr.FindEventAuthorityAddress()
	if err != nil {
		return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
	}

	return initializeVirtualPoolWithToken2022Ptr.
		SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
}

// prepareSwapParams prepares swap parameters.
func (p *PoolService) prepareSwapParams(
	swapBaseForQuote bool,
	virtualPooolState types.VirtualPoolState,
	poolConfigState types.PoolConfigState,
) types.PrepareSwapParams {
	if swapBaseForQuote {
		return types.PrepareSwapParams{
			InputMint:          virtualPooolState.BaseMint,
			OutputMint:         poolConfigState.QuoteMint,
			InputTokenProgram:  helpers.GetTokenProgram(uint8(virtualPooolState.PoolType)),
			OutputTokenProgram: helpers.GetTokenProgram(uint8(poolConfigState.QuoteTokenFlag)),
		}
	}
	return types.PrepareSwapParams{
		InputMint:          poolConfigState.QuoteMint,
		OutputMint:         virtualPooolState.BaseMint,
		InputTokenProgram:  helpers.GetTokenProgram(uint8(poolConfigState.QuoteTokenFlag)),
		OutputTokenProgram: helpers.GetTokenProgram(uint8(virtualPooolState.PoolType)),
	}
}

// createConfigIx creates config transaction.
func (p *PoolService) createConfigIx(
	configParam dbc.ConfigParameters,
	config, feeClaimer, leftoverReceiver, quoteMint, payer solana.PublicKey,
) (*dbc.Instruction, error) {

	// TODO: validation func

	createConfigPtr := dbc.NewCreateConfigInstruction(
		configParam,
		config,
		feeClaimer,
		leftoverReceiver,
		quoteMint,
		payer,
		solana.SystemProgramID,
		solana.PublicKey{},
		constants.DBCProgramId,
	)
	eventAuthPDA, _, err := createConfigPtr.FindEventAuthorityAddress()
	if err != nil {
		return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
	}

	return createConfigPtr.
		SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
}

// createPoolIx creates pool transaction.
func (p *PoolService) createPoolIx(
	createPoolParam types.CreatePoolParam,
	tokenType types.TokenType, quoteMint solana.PublicKey,
) (*dbc.Instruction, error) {

	// TODO: validation func

	pool := helpers.DeriveDbcPoolAddress(quoteMint, createPoolParam.BaseMint, createPoolParam.Config)
	baseVault := helpers.DeriveDbcTokenVaultAddress(pool, createPoolParam.BaseMint)
	quoteVault := helpers.DeriveDbcTokenVaultAddress(pool, quoteMint)

	if tokenType == types.TokenTypeSPL {
		return p.initializeSplPool(types.InitializePoolBaseParam{
			Name:         createPoolParam.Name,
			Symbol:       createPoolParam.Symbol,
			URI:          createPoolParam.URI,
			Config:       createPoolParam.Config,
			Payer:        createPoolParam.Payer,
			PoolCreator:  createPoolParam.PoolCreator,
			BaseMint:     createPoolParam.BaseMint,
			QuoteMint:    quoteMint,
			BaseVault:    baseVault,
			QuoteVault:   quoteVault,
			Pool:         pool,
			MintMetadata: helpers.DeriveMintMetadata(createPoolParam.BaseMint),
		})
	}

	return p.initializeToken2022Pool(types.InitializePoolBaseParam{
		Name:        createPoolParam.Name,
		Symbol:      createPoolParam.Symbol,
		URI:         createPoolParam.URI,
		Config:      createPoolParam.Config,
		Payer:       createPoolParam.Payer,
		PoolCreator: createPoolParam.PoolCreator,
		BaseMint:    createPoolParam.BaseMint,
		QuoteMint:   quoteMint,
		BaseVault:   baseVault,
		QuoteVault:  quoteVault,
		Pool:        pool,
	})
}

// cswapBuyIx
// reates first buy transaction.
func (p *PoolService) swapBuyIx(
	ctx context.Context,
	firstBuyParam types.FirstBuyParam,
	baseMint, config solana.PublicKey,
	baseFee dbc.BaseFeeParameters,
	swapBaseForQuote bool,
	currentPoint *big.Int,
	tokenType types.TokenType,
	quoteMint solana.PublicKey,
) ([]solana.Instruction, error) {

	// TODO: validation func

	quoteTokenFlag, err := helpers.GetTokenType(p.state.conn, quoteMint)
	if err != nil {
		return nil, err
	}

	prepareSwapParams := p.prepareSwapParams(
		false,
		types.VirtualPoolState{
			BaseMint: baseMint,
			PoolType: tokenType,
		},
		types.PoolConfigState{
			QuoteMint:      quoteMint,
			QuoteTokenFlag: quoteTokenFlag,
		},
	)

	type res struct {
		AtaPubkey solana.PublicKey
		Ix        *solana.GenericInstruction
		Err       error
	}
	var (
		a, b res
	)
	{
		g, ctx := errgroup.WithContext(ctx)
		g.Go(func() error {
			ata, ix, err := helpers.GetOrCreateATAInstruction(
				ctx,
				p.state.conn,
				prepareSwapParams.InputMint,
				firstBuyParam.Buyer,
				firstBuyParam.Buyer,
				true,
				prepareSwapParams.InputTokenProgram,
			)

			if err != nil {
				return err
			}

			a.AtaPubkey = ata
			a.Ix = ix
			return nil
		})

		g.Go(func() error {
			owner := firstBuyParam.Buyer
			if !firstBuyParam.Receiver.IsZero() {
				owner = firstBuyParam.Receiver
			}
			ata, ix, err := helpers.GetOrCreateATAInstruction(
				ctx,
				p.state.conn,
				prepareSwapParams.OutputMint,
				owner,
				firstBuyParam.Buyer,
				true,
				prepareSwapParams.OutputTokenProgram,
			)
			if err != nil {
				return err
			}

			b.AtaPubkey = ata
			b.Ix = ix
			return nil
		})

		if err := g.Wait(); err != nil {
			return nil, err
		}
	}

	preInstructions := make([]solana.Instruction, 0, 4)
	preInstructions = append(preInstructions, a.Ix, b.Ix)

	// add SOL wrapping instructions if needed
	if prepareSwapParams.InputMint.Equals(solana.WrappedSol) {
		preInstructions = append(preInstructions,
			helpers.WrapSOLInstruction(
				firstBuyParam.Buyer,
				a.AtaPubkey,
				firstBuyParam.BuyAmount,
			)...,
		)
	}

	// add postInstructions for SOL unwrapping if needed
	postInstructions := make([]solana.Instruction, 0, 4)
	if prepareSwapParams.InputMint.Equals(solana.WrappedSol) ||
		prepareSwapParams.OutputMint.Equals(solana.WrappedSol) {
		ix, err := helpers.UnwrapSOLInstruction(firstBuyParam.Buyer, firstBuyParam.Buyer, false)
		if err != nil {
			return nil, err
		}
		postInstructions = append(postInstructions, ix)
	}

	// check if rate limiter is applied
	// this swapBuyTx is only QuoteToBase direction
	// this swapBuyTx does not check poolState, so there is no check for activation point
	tradeDirection := types.TradeDirectionQuoteToBase
	if swapBaseForQuote {
		tradeDirection = types.TradeDirectionBaseToQuote
	}

	rateLimiterApplied := mathsPoolfees.IsRateLimiterApplied(
		currentPoint,
		big.NewInt(0),
		tradeDirection,
		new(big.Int).SetUint64(baseFee.SecondFactor),
		new(big.Int).SetUint64(baseFee.ThirdFactor),
		new(big.Int).SetUint64(uint64(baseFee.FirstFactor)),
	)

	// add remaining accounts if rate limiter is applied
	var remainingAccounts solana.AccountMetaSlice
	if rateLimiterApplied {
		remainingAccounts = []*solana.AccountMeta{
			{
				PublicKey: solana.SysVarInstructionsPubkey,
			},
		}
	}

	pool := helpers.DeriveDbcPoolAddress(quoteMint, baseMint, config)
	swapPtr := dbc.NewSwapInstruction(
		dbc.SwapParameters{
			AmountIn:         firstBuyParam.BuyAmount,
			MinimumAmountOut: firstBuyParam.MinimumAmountOut,
		},
		p.state.GetPoolAuthority(),
		config,
		pool,
		a.AtaPubkey,
		b.AtaPubkey,
		helpers.DeriveDbcTokenVaultAddress(pool, baseMint),
		helpers.DeriveDbcTokenVaultAddress(pool, quoteMint),
		baseMint,
		quoteMint,
		firstBuyParam.Buyer,
		prepareSwapParams.OutputTokenProgram,
		prepareSwapParams.InputTokenProgram,
		firstBuyParam.ReferralTokenAccount,
		solana.PublicKey{},
		constants.DBCProgramId,
	)

	eventAuthPDA, _, err := swapPtr.FindEventAuthorityAddress()
	if err != nil {
		return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
	}

	swapPtr.AccountMetaSlice = append(swapPtr.AccountMetaSlice, remainingAccounts...)

	currentIx, err := swapPtr.
		SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	finalIxns := make([]solana.Instruction, 0, len(preInstructions)+1+len(postInstructions))
	finalIxns = append(finalIxns, preInstructions...)
	finalIxns = append(finalIxns, currentIx)
	finalIxns = append(finalIxns, postInstructions...)
	return finalIxns, nil
}

// CreatePool creates a new pool.
func (p *PoolService) CreatePool(
	ctx context.Context,
	param types.CreatePoolParam,
) (*dbc.Instruction, error) {

	poolConfigState, err := p.state.GetPoolConfig(ctx, param.Config)
	if err != nil {
		return nil, err
	}

	pool := helpers.DeriveDbcPoolAddress(poolConfigState.QuoteMint, param.BaseMint, param.Config)
	baseVault := helpers.DeriveDbcTokenVaultAddress(pool, param.BaseMint)
	quoteVault := helpers.DeriveDbcTokenVaultAddress(pool, poolConfigState.QuoteMint)

	if poolConfigState.TokenType == uint8(types.TokenTypeSPL) {
		return p.initializeSplPool(
			types.InitializePoolBaseParam{
				Name:         param.Name,
				Symbol:       param.Symbol,
				URI:          param.URI,
				Pool:         pool,
				Config:       param.Config,
				Payer:        param.Payer,
				PoolCreator:  param.PoolCreator,
				BaseMint:     param.BaseMint,
				BaseVault:    baseVault,
				QuoteVault:   quoteVault,
				QuoteMint:    poolConfigState.QuoteMint,
				MintMetadata: helpers.DeriveMintMetadata(param.BaseMint),
			},
		)
	}

	return p.initializeToken2022Pool(
		types.InitializePoolBaseParam{
			Name:        param.Name,
			Symbol:      param.Symbol,
			URI:         param.URI,
			Pool:        pool,
			Config:      param.Config,
			Payer:       param.Payer,
			PoolCreator: param.PoolCreator,
			BaseMint:    param.BaseMint,
			BaseVault:   baseVault,
			QuoteVault:  quoteVault,
			QuoteMint:   poolConfigState.QuoteMint,
		},
	)
}

// CreateConfigAndPool creates a new config and pool.
func (p *PoolService) CreateConfigAndPool(
	param types.CreateConfigAndPoolParam,

) ([]solana.Instruction, error) {

	createConfigIx, err := p.createConfigIx(
		param.CreateConfigParam.ConfigParameters,
		param.Config,
		param.FeeClaimer,
		param.LeftoverReceiver,
		param.QuoteMint,
		param.Payer,
	)
	if err != nil {
		return nil, err
	}

	createPoolIx, err := p.createPoolIx(
		types.CreatePoolParam{
			PreCreatePoolParam: param.PreCreatePoolParam,
			Config:             param.Config,
			Payer:              param.Payer,
		},
		param.TokenType,
		param.QuoteMint,
	)
	if err != nil {
		return nil, err
	}

	return []solana.Instruction{createConfigIx, createPoolIx}, nil
}

// CreateConfigAndPoolWithFirstBuy creates a new config and pool and buy tokens.
func (p *PoolService) CreateConfigAndPoolWithFirstBuy(
	ctx context.Context,
	param types.CreateConfigAndPoolWithFirstBuyParam,
) (struct {
	CreateConfigIx, CreatePoolIx *dbc.Instruction
	SwapBuyIxns                  []solana.Instruction
}, error) {

	createConfigIx, err := p.createConfigIx(
		param.CreateConfigParam.ConfigParameters,
		param.Config,
		param.FeeClaimer,
		param.LeftoverReceiver,
		param.QuoteMint,
		param.Payer,
	)
	if err != nil {
		return struct {
			CreateConfigIx *dbc.Instruction
			CreatePoolIx   *dbc.Instruction
			SwapBuyIxns    []solana.Instruction
		}{}, err
	}

	createPoolIx, err := p.createPoolIx(
		types.CreatePoolParam{
			PreCreatePoolParam: param.PreCreatePoolParam,
			Config:             param.Config,
		},
		param.TokenType,
		param.QuoteMint,
	)
	if err != nil {
		return struct {
			CreateConfigIx *dbc.Instruction
			CreatePoolIx   *dbc.Instruction
			SwapBuyIxns    []solana.Instruction
		}{}, err
	}

	currentPoint, err := helpers.GetCurrentPoint(
		p.state.conn,
		types.ActivationType(param.ActivationType),
	)
	if err != nil {
		return struct {
			CreateConfigIx *dbc.Instruction
			CreatePoolIx   *dbc.Instruction
			SwapBuyIxns    []solana.Instruction
		}{}, err
	}

	// create first buy transaction
	var swapBuyIxns []solana.Instruction
	if param.FirstBuyParam.BuyAmount > 0 {
		if swapBuyIxns, err = p.swapBuyIx(
			ctx,
			param.FirstBuyParam,
			param.PreCreatePoolParam.BaseMint,
			param.Config,
			param.PoolFees.BaseFee,
			false,
			currentPoint,
			param.TokenType,
			param.QuoteMint,
		); err != nil {
			return struct {
				CreateConfigIx *dbc.Instruction
				CreatePoolIx   *dbc.Instruction
				SwapBuyIxns    []solana.Instruction
			}{}, err
		}
	}

	return struct {
		CreateConfigIx *dbc.Instruction
		CreatePoolIx   *dbc.Instruction
		SwapBuyIxns    []solana.Instruction
	}{
		CreateConfigIx: createConfigIx,
		CreatePoolIx:   createPoolIx,
		SwapBuyIxns:    swapBuyIxns,
	}, nil
}

// CreatePoolWithFirstBuy creates a new pool and buy tokens.
func (p *PoolService) CreatePoolWithFirstBuy(
	ctx context.Context,
	param types.CreatePoolWithFirstBuyParam,
) (struct {
	CreatePoolIx *dbc.Instruction
	SwapBuyIxns  []solana.Instruction
}, error) {
	poolConfigState, err := p.state.GetPoolConfig(ctx, param.Config)
	if err != nil {
		return struct {
			CreatePoolIx *dbc.Instruction
			SwapBuyIxns  []solana.Instruction
		}{}, err
	}
	createPoolIx, err := p.createPoolIx(
		types.CreatePoolParam{
			PreCreatePoolParam: param.PreCreatePoolParam,
			Config:             param.Config,
		},
		types.TokenType(poolConfigState.TokenType),
		poolConfigState.QuoteMint,
	)
	if err != nil {
		return struct {
			CreatePoolIx *dbc.Instruction
			SwapBuyIxns  []solana.Instruction
		}{}, err
	}

	currentPoint, err := helpers.GetCurrentPoint(
		p.state.conn,
		types.ActivationType(poolConfigState.ActivationType),
	)
	if err != nil {
		return struct {
			CreatePoolIx *dbc.Instruction
			SwapBuyIxns  []solana.Instruction
		}{}, err
	}

	var swapBuyIxns []solana.Instruction
	if param.FirstBuyParam.BuyAmount > 0 {
		if swapBuyIxns, err = p.swapBuyIx(
			ctx,
			param.FirstBuyParam,
			param.PreCreatePoolParam.BaseMint,
			param.Config,
			dbc.BaseFeeParameters{
				CliffFeeNumerator: poolConfigState.PoolFees.BaseFee.CliffFeeNumerator,
				FirstFactor:       poolConfigState.PoolFees.BaseFee.FirstFactor,
				SecondFactor:      poolConfigState.PoolFees.BaseFee.SecondFactor,
				ThirdFactor:       poolConfigState.PoolFees.BaseFee.ThirdFactor,
				BaseFeeMode:       poolConfigState.PoolFees.BaseFee.BaseFeeMode,
			},
			false,
			currentPoint,
			types.TokenType(poolConfigState.TokenType),
			poolConfigState.QuoteMint,
		); err != nil {
			return struct {
				CreatePoolIx *dbc.Instruction
				SwapBuyIxns  []solana.Instruction
			}{}, err
		}
	}

	return struct {
		CreatePoolIx *dbc.Instruction
		SwapBuyIxns  []solana.Instruction
	}{
		CreatePoolIx: createPoolIx,
		SwapBuyIxns:  swapBuyIxns,
	}, nil
}

// CreatePoolWithPartnerAndCreatorFirstBuy creates a new pool and buy tokens with partner and creator.
func (p *PoolService) CreatePoolWithPartnerAndCreatorFirstBuy(
	ctx context.Context,
	param types.CreatePoolWithPartnerAndCreatorFirstBuyParam,
) (struct {
	CreatorPoolIx                      *dbc.Instruction
	PartnerSwapBuyIx, CreatorSwapBuyIx []solana.Instruction
}, error) {

	poolConfigState, err := p.state.GetPoolConfig(ctx, param.CreatePoolParam.Config)
	if err != nil {
		return struct {
			CreatorPoolIx    *dbc.Instruction
			PartnerSwapBuyIx []solana.Instruction
			CreatorSwapBuyIx []solana.Instruction
		}{}, err
	}

	createPoolIx, err := p.createPoolIx(
		types.CreatePoolParam{
			PreCreatePoolParam: param.CreatePoolParam.PreCreatePoolParam,
			Config:             param.CreatePoolParam.Config,
		},
		types.TokenType(poolConfigState.TokenType),
		poolConfigState.QuoteMint,
	)
	if err != nil {
		return struct {
			CreatorPoolIx    *dbc.Instruction
			PartnerSwapBuyIx []solana.Instruction
			CreatorSwapBuyIx []solana.Instruction
		}{}, err
	}

	currentPoint, err := helpers.GetCurrentPoint(
		p.state.conn,
		types.ActivationType(poolConfigState.ActivationType),
	)
	if err != nil {
		return struct {
			CreatorPoolIx    *dbc.Instruction
			PartnerSwapBuyIx []solana.Instruction
			CreatorSwapBuyIx []solana.Instruction
		}{}, err
	}

	var partnerSwapBuyIx []solana.Instruction
	if param := param; param.PartnerFirstBuyParam.BuyAmount > 0 {
		// create partner first buy transaction
		if partnerSwapBuyIx, err = p.swapBuyIx(
			ctx,
			types.FirstBuyParam{
				Buyer:                param.PartnerFirstBuyParam.Partner,
				Receiver:             param.PartnerFirstBuyParam.Receiver,
				BuyAmount:            param.PartnerFirstBuyParam.BuyAmount,
				MinimumAmountOut:     param.PartnerFirstBuyParam.MinimumAmountOut,
				ReferralTokenAccount: param.PartnerFirstBuyParam.ReferralTokenAccount,
			},
			param.CreatePoolParam.BaseMint,
			param.CreatePoolParam.Config,
			dbc.BaseFeeParameters{
				CliffFeeNumerator: poolConfigState.PoolFees.BaseFee.CliffFeeNumerator,
				FirstFactor:       poolConfigState.PoolFees.BaseFee.FirstFactor,
				SecondFactor:      poolConfigState.PoolFees.BaseFee.SecondFactor,
				ThirdFactor:       poolConfigState.PoolFees.BaseFee.ThirdFactor,
				BaseFeeMode:       poolConfigState.PoolFees.BaseFee.BaseFeeMode,
			},
			false,
			currentPoint,
			types.TokenType(poolConfigState.TokenType),
			poolConfigState.QuoteMint,
		); err != nil {
			return struct {
				CreatorPoolIx    *dbc.Instruction
				PartnerSwapBuyIx []solana.Instruction
				CreatorSwapBuyIx []solana.Instruction
			}{}, err
		}
	}

	var creatorSwapBuyIx []solana.Instruction
	// TODO: check createConfigAndPoolWithFirstBuyParam.FirstBuyParam is not empty
	if param := param; param.CreatorFirstBuyParam.BuyAmount > 0 {
		// create partner first buy transaction
		if creatorSwapBuyIx, err = p.swapBuyIx(
			ctx,
			types.FirstBuyParam{
				Buyer:                param.CreatorFirstBuyParam.Creator,
				Receiver:             param.CreatorFirstBuyParam.Receiver,
				BuyAmount:            param.CreatorFirstBuyParam.BuyAmount,
				MinimumAmountOut:     param.CreatorFirstBuyParam.MinimumAmountOut,
				ReferralTokenAccount: param.CreatorFirstBuyParam.ReferralTokenAccount,
			},
			param.CreatePoolParam.BaseMint,
			param.CreatePoolParam.Config,
			dbc.BaseFeeParameters{
				CliffFeeNumerator: poolConfigState.PoolFees.BaseFee.CliffFeeNumerator,
				FirstFactor:       poolConfigState.PoolFees.BaseFee.FirstFactor,
				SecondFactor:      poolConfigState.PoolFees.BaseFee.SecondFactor,
				ThirdFactor:       poolConfigState.PoolFees.BaseFee.ThirdFactor,
				BaseFeeMode:       poolConfigState.PoolFees.BaseFee.BaseFeeMode,
			},
			false,
			currentPoint,
			types.TokenType(poolConfigState.TokenType),
			poolConfigState.QuoteMint,
		); err != nil {
			return struct {
				CreatorPoolIx    *dbc.Instruction
				PartnerSwapBuyIx []solana.Instruction
				CreatorSwapBuyIx []solana.Instruction
			}{}, err
		}
	}

	return struct {
		CreatorPoolIx    *dbc.Instruction
		PartnerSwapBuyIx []solana.Instruction
		CreatorSwapBuyIx []solana.Instruction
	}{
		CreatorPoolIx:    createPoolIx,
		PartnerSwapBuyIx: partnerSwapBuyIx,
		CreatorSwapBuyIx: creatorSwapBuyIx,
	}, nil

}

// Swap swaps between base and quote.
func (p *PoolService) Swap(
	ctx context.Context,
	param types.SwapParam,
) ([]solana.Instruction, error) {

	poolState, err := p.state.GetPool(ctx, param.Pool)
	if err != nil {
		return nil, fmt.Errorf("swap:pool (%s) not found: error: %w", param.Pool.String(), err)
	}

	poolConfigState, err := p.state.GetPoolConfig(ctx, poolState.Config)
	if err != nil {
		return nil, fmt.Errorf("swap:pool config (%s) not found: error: %w", param.Pool.String(), err)
	}

	// TODO: validation checks

	currentPoint, err := helpers.GetCurrentPoint(
		p.state.conn,
		types.ActivationType(poolConfigState.ActivationType),
	)
	if err != nil {
		return nil, err
	}

	// check if rate limiter is applied if:
	// 1. rate limiter mode
	// 2. swap direction is QuoteToBase
	// 3. current point is greater than activation point
	// 4. current point is less than activation point + maxLimiterDuration
	tradeDirection := types.TradeDirectionQuoteToBase
	if param.SwapBaseForQuote {
		tradeDirection = types.TradeDirectionBaseToQuote
	}
	rateLimiterApplied := mathsPoolfees.IsRateLimiterApplied(
		currentPoint,
		new(big.Int).SetUint64(poolState.ActivationPoint),
		tradeDirection,
		new(big.Int).SetUint64(poolConfigState.PoolFees.BaseFee.SecondFactor),
		new(big.Int).SetUint64(poolConfigState.PoolFees.BaseFee.ThirdFactor),
		new(big.Int).SetUint64(uint64(poolConfigState.PoolFees.BaseFee.FirstFactor)),
	)

	prepareSwapParams := p.prepareSwapParams(
		param.SwapBaseForQuote,
		types.VirtualPoolState{
			BaseMint: poolState.BaseMint,
			PoolType: types.TokenType(poolState.PoolType),
		},
		types.PoolConfigState{
			QuoteMint:      poolConfigState.QuoteMint,
			QuoteTokenFlag: types.TokenType(poolConfigState.QuoteTokenFlag),
		},
	)

	// add preInstructions for ATA creation and SOL wrapping
	payer := param.Owner
	if !param.Payer.IsZero() {
		payer = param.Payer
	}
	prepareTokenAccounts, err := p.state.prepareTokenAccounts(
		ctx,
		types.PrepareTokenAccountParams{
			Owner:         param.Owner,
			Payer:         payer,
			TokenAMint:    prepareSwapParams.InputMint,
			TokenBMint:    prepareSwapParams.OutputMint,
			TokenAProgram: prepareSwapParams.InputTokenProgram,
			TokenBProgram: prepareSwapParams.OutputTokenProgram,
		},
	)
	if err != nil {
		return nil, err
	}
	preInstructions := make([]solana.Instruction, 0, len(prepareTokenAccounts.CreateATAIxns)+2)
	preInstructions = append(preInstructions, prepareTokenAccounts.CreateATAIxns...)

	// add SOL wrapping instructions if needed
	if prepareSwapParams.InputMint.Equals(solana.WrappedSol) {
		preInstructions = append(preInstructions,
			helpers.WrapSOLInstruction(
				param.Owner,
				prepareTokenAccounts.TokenAAta,
				param.AmountIn,
			)...,
		)
	}

	postInstructions := make([]solana.Instruction, 0, 1)
	if prepareSwapParams.InputMint.Equals(solana.WrappedSol) ||
		prepareSwapParams.OutputMint.Equals(solana.WrappedSol) {
		ix, err := helpers.UnwrapSOLInstruction(
			param.Owner,
			param.Owner,
			false,
		)
		if err != nil {
			return nil, err
		}
		postInstructions = append(postInstructions, ix)
	}

	var remainingAccounts solana.AccountMetaSlice
	if rateLimiterApplied {
		remainingAccounts = []*solana.AccountMeta{
			{
				PublicKey: solana.SysVarInstructionsPubkey,
			},
		}
	}
	tokenBaseProgram, tokenQuoteProgram :=
		prepareSwapParams.OutputTokenProgram, prepareSwapParams.InputTokenProgram

	if param.SwapBaseForQuote {
		tokenBaseProgram, tokenQuoteProgram =
			prepareSwapParams.InputTokenProgram, prepareSwapParams.OutputTokenProgram
	}

	swapPtr := dbc.NewSwapInstruction(
		dbc.SwapParameters{
			AmountIn:         param.AmountIn,
			MinimumAmountOut: param.MinimumAmountOut,
		},
		p.state.GetPoolAuthority(),
		poolState.Config,
		param.Pool,
		prepareTokenAccounts.TokenAAta,
		prepareTokenAccounts.TokenBAta,
		poolState.BaseVault,
		poolState.QuoteVault,
		poolState.BaseMint,
		poolConfigState.QuoteMint,
		param.Payer,
		tokenBaseProgram,
		tokenQuoteProgram,
		param.ReferralTokenAccount,
		solana.PublicKey{},
		constants.DBCProgramId,
	)

	eventAuthPDA, _, err := swapPtr.FindEventAuthorityAddress()
	if err != nil {
		return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
	}

	swapPtr.AccountMetaSlice = append(swapPtr.AccountMetaSlice, remainingAccounts...)

	currentIx, err := swapPtr.
		SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	finalIxns := make([]solana.Instruction, 0, len(preInstructions)+1+len(postInstructions))
	finalIxns = append(finalIxns, preInstructions...)
	finalIxns = append(finalIxns, currentIx)
	finalIxns = append(finalIxns, postInstructions...)
	return finalIxns, nil
}

// Swap2 swaps between base and quote (included SwapMode: ExactIn, PartialFill, ExactOut).
func (p *PoolService) Swap2(
	ctx context.Context,
	param types.Swap2Param,
) ([]solana.Instruction, error) {

	// NOTE: remove this, validation checks func should cover
	switch param.SwapMode {
	case types.SwapModeExactIn, types.SwapModePartialFill:
		if param.AmountIn == nil || param.MinimumAmountOut == nil {
			return nil,
				fmt.Errorf("AmountIn(%s) or MinimumAmountOut(%s) cannot be nil for swapMode(%d)",
					param.AmountIn, param.MinimumAmountOut, param.SwapMode,
				)
		}

		if !param.AmountIn.IsUint64() {
			return nil, fmt.Errorf(
				"cannot fit AmountIn(%s) into uint64", param.AmountIn,
			)
		}
	case types.SwapModeExactOut:
		if param.AmountOut == nil || param.MaximumAmountIn == nil {
			return nil,
				fmt.Errorf("AmountOut(%s) or MaximumAmountIn(%s) cannot be nil for swapMode(%d)",
					param.AmountOut, param.MaximumAmountIn, param.SwapMode,
				)
		}
		if !param.AmountOut.IsUint64() {
			return nil, fmt.Errorf(
				"cannot fit AmountOut(%s) into uint64", param.AmountOut,
			)
		}
	}

	// TODO: validation checks

	poolState, err := p.state.GetPool(ctx, param.Pool)
	if err != nil {
		return nil, fmt.Errorf("swap2:pool (%s) not found: error: %w", param.Pool.String(), err)
	}

	poolConfigState, err := p.state.GetPoolConfig(ctx, poolState.Config)
	if err != nil {
		return nil, fmt.Errorf("swap2:pool config (%s) not found: error: %w", param.Pool.String(), err)
	}

	currentPoint, err := helpers.GetCurrentPoint(
		p.state.conn,
		types.ActivationType(poolConfigState.ActivationType),
	)
	if err != nil {
		return nil, err
	}

	// check if rate limiter is applied if:
	// 1. rate limiter mode
	// 2. swap direction is QuoteToBase
	// 3. current point is greater than activation point
	// 4. current point is less than activation point + maxLimiterDuration
	tradeDirection := types.TradeDirectionQuoteToBase
	if param.SwapBaseForQuote {
		tradeDirection = types.TradeDirectionBaseToQuote
	}
	rateLimiterApplied := mathsPoolfees.IsRateLimiterApplied(
		currentPoint,
		new(big.Int).SetUint64(poolState.ActivationPoint),
		tradeDirection,
		new(big.Int).SetUint64(poolConfigState.PoolFees.BaseFee.SecondFactor),
		new(big.Int).SetUint64(poolConfigState.PoolFees.BaseFee.ThirdFactor),
		new(big.Int).SetUint64(uint64(poolConfigState.PoolFees.BaseFee.FirstFactor)),
	)

	prepareSwapParams := p.prepareSwapParams(
		param.SwapBaseForQuote,
		types.VirtualPoolState{
			BaseMint: poolState.BaseMint,
			PoolType: types.TokenType(poolState.PoolType),
		},
		types.PoolConfigState{
			QuoteMint:      poolConfigState.QuoteMint,
			QuoteTokenFlag: types.TokenType(poolConfigState.QuoteTokenFlag),
		},
	)

	// add preInstructions for ATA creation and SOL wrapping
	payer := param.Owner
	if !param.Payer.IsZero() {
		payer = param.Payer
	}
	prepareTokenAccounts, err := p.state.prepareTokenAccounts(
		ctx,
		types.PrepareTokenAccountParams{
			Owner:         param.Owner,
			Payer:         payer,
			TokenAMint:    prepareSwapParams.InputMint,
			TokenBMint:    prepareSwapParams.OutputMint,
			TokenAProgram: prepareSwapParams.InputTokenProgram,
			TokenBProgram: prepareSwapParams.OutputTokenProgram,
		},
	)
	if err != nil {
		return nil, err
	}
	preInstructions := make([]solana.Instruction, 0, len(prepareTokenAccounts.CreateATAIxns)+2)
	preInstructions = append(preInstructions, prepareTokenAccounts.CreateATAIxns...)

	amount0, amount1 := param.AmountIn, param.MinimumAmountOut
	if param.SwapMode == types.SwapModeExactOut {
		amount0, amount1 = param.AmountOut, param.MaximumAmountIn
	}

	// add SOL wrapping instructions if needed
	if prepareSwapParams.InputMint.Equals(solana.WrappedSol) {
		amount := amount0
		if param.SwapMode == types.SwapModeExactOut {
			amount = amount1
		}
		preInstructions = append(preInstructions,
			helpers.WrapSOLInstruction(
				param.Owner,
				prepareTokenAccounts.TokenAAta,
				amount.Uint64(),
			)...,
		)
	}

	// add postInstructions for SOL unwrapping
	postInstructions := make([]solana.Instruction, 0, 1)
	if prepareSwapParams.InputMint.Equals(solana.WrappedSol) ||
		prepareSwapParams.OutputMint.Equals(solana.WrappedSol) {
		ix, err := helpers.UnwrapSOLInstruction(
			param.Owner,
			param.Owner,
			false,
		)
		if err != nil {
			return nil, err
		}
		postInstructions = append(postInstructions, ix)
	}

	// add remaining accounts if rate limiter is applied
	var remainingAccounts solana.AccountMetaSlice
	if rateLimiterApplied {
		remainingAccounts = []*solana.AccountMeta{
			{
				PublicKey: solana.SysVarInstructionsPubkey,
			},
		}
	}
	tokenBaseProgram, tokenQuoteProgram :=
		prepareSwapParams.OutputTokenProgram, prepareSwapParams.InputTokenProgram

	if param.SwapBaseForQuote {
		tokenBaseProgram, tokenQuoteProgram =
			prepareSwapParams.InputTokenProgram, prepareSwapParams.OutputTokenProgram
	}

	swapPtr := dbc.NewSwap2Instruction(
		dbc.SwapParameters2{
			Amount0:  amount0.Uint64(),
			Amount1:  amount1.Uint64(),
			SwapMode: uint8(param.SwapMode),
		},
		p.state.GetPoolAuthority(),
		poolState.Config,
		param.Pool,
		prepareTokenAccounts.TokenAAta,
		prepareTokenAccounts.TokenBAta,
		poolState.BaseVault,
		poolState.QuoteVault,
		poolState.BaseMint,
		poolConfigState.QuoteMint,
		param.Owner,
		tokenBaseProgram,
		tokenQuoteProgram,
		param.ReferralTokenAccount,
		solana.PublicKey{},
		constants.DBCProgramId,
	)

	eventAuthPDA, _, err := swapPtr.FindEventAuthorityAddress()
	if err != nil {
		return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
	}

	swapPtr.AccountMetaSlice = append(swapPtr.AccountMetaSlice, remainingAccounts...)

	currentIx, err := swapPtr.
		SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	finalIxns := make([]solana.Instruction, 0, len(preInstructions)+1+len(postInstructions))
	finalIxns = append(finalIxns, preInstructions...)
	finalIxns = append(finalIxns, currentIx)
	finalIxns = append(finalIxns, postInstructions...)
	return finalIxns, nil
}

// SwapQuote calculates the amount out for a swap (quote) for swap1.
func (p *PoolService) SwapQuote(
	param types.SwapQuoteParam,
) (types.SwapQuoteResult, error) {
	return maths.SwapQuote(
		param.VirtualPool,
		param.Config,
		param.SwapBaseForQuote,
		param.AmountIn,
		param.SlippageBps,
		param.HasReferral,
		param.CurrentPoint,
	)
}

// SwapQuote2 calculates the amount out for a swap (quote) based on swap mode (for swap2).
func (p *PoolService) SwapQuote2(
	param types.SwapQuote2Param,
) (types.SwapQuote2Result, error) {

	switch param.SwapMode {
	case types.SwapModeExactIn:
		if param.AmountIn == nil {
			return types.SwapQuote2Result{}, errors.New("SwapQuote2:amountIn cannot be nil for SwapModeExactIn")
		}
		return maths.SwapQuoteExactIn(
			param.VirtualPool,
			param.Config,
			param.SwapBaseForQuote,
			param.AmountIn,
			param.SlippageBps,
			param.HasReferral,
			param.CurrentPoint,
		)

	case types.SwapModePartialFill:
		if param.AmountIn == nil {
			return types.SwapQuote2Result{}, errors.New("SwapQuote2:amountIn cannot be nil for SwapModePartialFill")
		}
		return maths.SwapQuotePartialFill(
			param.VirtualPool,
			param.Config,
			param.SwapBaseForQuote,
			param.AmountIn,
			param.SlippageBps,
			param.HasReferral,
			param.CurrentPoint,
		)

	case types.SwapModeExactOut:
		if param.AmountOut == nil {
			return types.SwapQuote2Result{}, errors.New("SwapQuote2:actualReferralFeemountOut cannot be nil for SwapModeExactOut")
		}
		return maths.SwapQuoteExactOut(
			param.VirtualPool,
			param.Config,
			param.SwapBaseForQuote,
			param.AmountOut,
			param.SlippageBps,
			param.HasReferral,
			param.CurrentPoint,
		)
	}

	return types.SwapQuote2Result{}, fmt.Errorf("unsupported swapMode(%d)", param.SwapMode)
}
