package services

import (
	"context"
	"dbcGoSDK/constants"
	"dbcGoSDK/generated/dbc"
	"dbcGoSDK/helpers"
	"dbcGoSDK/types"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
)

type PartnerService struct {
	state *StateService
}

func NewPartnerService(
	conn *rpc.Client,
	commitment rpc.CommitmentType,
) *PartnerService {
	return &PartnerService{
		state: NewStateService(conn, commitment),
	}
}

// CreateConfigParam create a new config.
func (p *PartnerService) CreateConfigParam(
	param types.CreateConfigParam,
) (*dbc.Instruction, error) {

	// TODO: validityChecks
	createConfigPtr := dbc.NewCreateConfigInstruction(
		param.ConfigParameters,
		param.Config,
		param.FeeClaimer,
		param.LeftoverReceiver,
		param.QuoteMint,
		param.Payer,
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

// CreatePartnerMetadata creates partner metadata.
func (p *PartnerService) CreatePartnerMetadata(
	param types.CreatePartnerMetadataParam,
) (*dbc.Instruction, error) {

	CreatePartnerMetadataPtr := dbc.NewCreatePartnerMetadataInstruction(
		dbc.CreatePartnerMetadataParameters{
			Name:    param.Name,
			Website: param.Website,
			Logo:    param.Logo,
		},
		helpers.DeriveDbcPartnerMetadata(param.FeeClaimer),
		param.Payer,
		param.FeeClaimer,
		solana.SystemProgramID,
		solana.PublicKey{},
		constants.DBCProgramId,
	)

	eventAuthPDA, _, err := CreatePartnerMetadataPtr.FindEventAuthorityAddress()
	if err != nil {
		return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
	}

	return CreatePartnerMetadataPtr.
		SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
}

// claimWithQuoteMintSol method to claim trading fee with quote mint SOL.
func (p *PartnerService) claimWithQuoteMintSol(
	param types.ClaimPartnerTradingFeeWithQuoteMintSolParam,
) (struct {
	types.Accounts
	PreInstructions  *solana.GenericInstruction
	PostInstructions *token.Instruction
}, error) {

	tokenBaseAccount, err := helpers.FindAssociatedTokenAddress(
		param.FeeReceiver,
		param.PoolState.BaseMint,
		param.TokenBaseProgram,
	)
	if err != nil {
		return struct {
			types.Accounts
			PreInstructions  *solana.GenericInstruction
			PostInstructions *token.Instruction
		}{}, err
	}

	tokenQuoteAccount, err := helpers.FindAssociatedTokenAddress(
		param.TempWSolAcc,
		param.PoolConfigState.QuoteMint,
		param.TokenQuoteProgram,
	)
	if err != nil {
		return struct {
			types.Accounts
			PreInstructions  *solana.GenericInstruction
			PostInstructions *token.Instruction
		}{}, err
	}

	preInstructions := helpers.CreateAssociatedTokenAccountIdempotentInstruction(
		param.Payer,
		tokenQuoteAccount,
		param.TempWSolAcc,
		param.PoolConfigState.QuoteMint,
		param.TokenQuoteProgram,
	)

	postInstructions, err := helpers.UnwrapSOLInstruction(
		param.TempWSolAcc, param.FeeReceiver, false,
	)
	if err != nil {
		return struct {
			types.Accounts
			PreInstructions  *solana.GenericInstruction
			PostInstructions *token.Instruction
		}{}, err
	}

	return struct {
		types.Accounts
		PreInstructions  *solana.GenericInstruction
		PostInstructions *token.Instruction
	}{
		Accounts: types.Accounts{
			PoolAuthority:     p.state.GetPoolAuthority(),
			Config:            param.Config,
			Pool:              param.Pool,
			TokenAAccount:     tokenBaseAccount,
			TokenBAccount:     tokenQuoteAccount,
			BaseVault:         param.PoolState.BaseVault,
			QuoteVault:        param.PoolState.QuoteVault,
			BaseMint:          param.PoolState.BaseMint,
			QuoteMint:         param.PoolConfigState.QuoteMint,
			FeeClaimer:        param.FeeClaimer,
			TokenBaseProgram:  param.TokenBaseProgram,
			TokenQuoteProgram: param.TokenQuoteProgram,
		},
		PreInstructions:  preInstructions,
		PostInstructions: postInstructions,
	}, nil
}

// claimWithQuoteMintNotSol method to claim trading fee with quote mint not SOL.
func (p *PartnerService) claimWithQuoteMintNotSol(
	ctx context.Context,
	param types.ClaimPartnerTradingFeeWithQuoteMintNotSolParam,
) (struct {
	types.Accounts
	PreInstructions []solana.Instruction
}, error) {

	out, err := p.state.prepareTokenAccounts(
		ctx,
		types.PrepareTokenAccountParams{
			Owner:         param.FeeReceiver,
			Payer:         param.Payer,
			TokenAMint:    param.PoolState.BaseMint,
			TokenBMint:    param.PoolConfigState.QuoteMint,
			TokenAProgram: param.TokenBaseProgram,
			TokenBProgram: param.TokenQuoteProgram,
		},
	)
	if err != nil {
		return struct {
			types.Accounts
			PreInstructions []solana.Instruction
		}{}, err
	}

	return struct {
		types.Accounts
		PreInstructions []solana.Instruction
	}{
		Accounts: types.Accounts{
			PoolAuthority:     p.state.GetPoolAuthority(),
			Config:            param.Config,
			Pool:              param.Pool,
			TokenAAccount:     out.TokenAAta,
			TokenBAccount:     out.TokenBAta,
			BaseVault:         param.PoolState.BaseVault,
			QuoteVault:        param.PoolState.QuoteVault,
			BaseMint:          param.PoolState.BaseMint,
			QuoteMint:         param.PoolConfigState.QuoteMint,
			FeeClaimer:        param.FeeClaimer,
			TokenBaseProgram:  param.TokenBaseProgram,
			TokenQuoteProgram: param.TokenQuoteProgram,
		},
		PreInstructions: out.CreateATAIxns,
	}, nil
}

// ClaimPartnerTradingFee claims partner trading fee.
func (p *PartnerService) ClaimPartnerTradingFee(
	ctx context.Context,
	param types.ClaimTradingFeeParam,
) ([]solana.Instruction, error) {

	poolState, err := p.state.GetPool(ctx, param.Pool)
	if err != nil {
		return nil, fmt.Errorf("pool not found: %s", param.Pool.String())
	}

	poolConfigState, err := p.state.GetPoolConfig(ctx, poolState.Config)
	if err != nil {
		return nil, fmt.Errorf("pool config not found: %s", param.Pool.String())
	}

	tokenBaseProgram := helpers.GetTokenProgram(poolConfigState.TokenType)
	tokenQuoteProgram := helpers.GetTokenProgram(poolConfigState.QuoteTokenFlag)

	isSOLQuoteMint := poolConfigState.QuoteMint.Equals(solana.WrappedSol)

	if isSOLQuoteMint {
		// if receiver is present and not equal to feeClaimer, use tempWSolAcc, otherwise use feeClaimer
		tempWSol := param.FeeClaimer
		if !param.Receiver.IsZero() && !param.Receiver.Equals(param.FeeClaimer) {
			tempWSol = param.TempWSolAcc
		}

		feeReceiver := param.FeeClaimer
		if !param.Receiver.IsZero() {
			feeReceiver = param.Receiver
		}

		out, err := p.claimWithQuoteMintSol(
			types.ClaimPartnerTradingFeeWithQuoteMintSolParam{
				ClaimPartnerTradingFeeWithQuoteMintNotSolParam: types.ClaimPartnerTradingFeeWithQuoteMintNotSolParam{
					FeeClaimer:        param.FeeClaimer,
					Payer:             param.Payer,
					FeeReceiver:       feeReceiver,
					Config:            poolState.Config,
					Pool:              param.Pool,
					PoolState:         poolState,
					PoolConfigState:   poolConfigState,
					TokenBaseProgram:  tokenBaseProgram,
					TokenQuoteProgram: tokenQuoteProgram,
				},
				TempWSolAcc: tempWSol,
			},
		)
		if err != nil {
			return nil, err
		}

		createTradingFeePtr := dbc.NewClaimTradingFeeInstruction(
			param.MaxBaseAmount.Uint64(),
			param.MaxQuoteAmount.Uint64(),
			out.PoolAuthority,
			out.Config,
			out.Pool,
			out.TokenAAccount,
			out.TokenBAccount,
			out.BaseVault,
			out.QuoteVault,
			out.BaseMint,
			out.QuoteMint,
			out.FeeClaimer,
			out.TokenBaseProgram,
			out.TokenQuoteProgram,
			solana.PublicKey{},
			constants.DBCProgramId,
		)

		eventAuthPDA, _, err := createTradingFeePtr.FindEventAuthorityAddress()
		if err != nil {
			return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
		}

		currentIx, err := createTradingFeePtr.
			SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
		if err != nil {
			return nil, err
		}

		ixns := make([]solana.Instruction, 0, 1+1+1)
		return append(ixns, out.PreInstructions, currentIx, out.PostInstructions), nil
	}

	feeReceiver := param.FeeClaimer
	if !param.Receiver.IsZero() {
		feeReceiver = param.Receiver
	}

	out, err := p.claimWithQuoteMintNotSol(
		ctx,
		types.ClaimPartnerTradingFeeWithQuoteMintNotSolParam{
			FeeClaimer:        param.FeeClaimer,
			Payer:             param.Payer,
			FeeReceiver:       feeReceiver,
			Config:            poolState.Config,
			Pool:              param.Pool,
			PoolState:         poolState,
			PoolConfigState:   poolConfigState,
			TokenBaseProgram:  tokenBaseProgram,
			TokenQuoteProgram: tokenQuoteProgram,
		},
	)
	if err != nil {
		return nil, err
	}

	createTradingFeePtr := dbc.NewClaimTradingFeeInstruction(
		param.MaxBaseAmount.Uint64(),
		param.MaxQuoteAmount.Uint64(),
		out.PoolAuthority,
		out.Config,
		out.Pool,
		out.TokenAAccount,
		out.TokenBAccount,
		out.BaseVault,
		out.QuoteVault,
		out.BaseMint,
		out.QuoteMint,
		out.FeeClaimer,
		out.TokenBaseProgram,
		out.TokenQuoteProgram,
		solana.PublicKey{},
		constants.DBCProgramId,
	)

	eventAuthPDA, _, err := createTradingFeePtr.FindEventAuthorityAddress()
	if err != nil {
		return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
	}

	currentIx, err := createTradingFeePtr.
		SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
	if err != nil {
		return nil, err
	}

	ixns := make([]solana.Instruction, 0, len(out.PreInstructions)+1+1)
	ixns = append(ixns, out.PreInstructions...)
	return append(ixns, currentIx), nil

}

// ClaimPartnerTradingFee2 claims partner trading fee.
func (p *PartnerService) ClaimPartnerTradingFee2(
	ctx context.Context,
	param types.ClaimTradingFeeParam,
) ([]solana.Instruction, error) {

	poolState, err := p.state.GetPool(ctx, param.Pool)
	if err != nil {
		return nil, fmt.Errorf("pool not found: %s", param.Pool.String())
	}

	poolConfigState, err := p.state.GetPoolConfig(ctx, poolState.Config)
	if err != nil {
		return nil, fmt.Errorf("pool config not found: %s", param.Pool.String())
	}

	tokenBaseProgram := helpers.GetTokenProgram(poolConfigState.TokenType)
	tokenQuoteProgram := helpers.GetTokenProgram(poolConfigState.QuoteTokenFlag)

	isSOLQuoteMint := poolConfigState.QuoteMint.Equals(solana.WrappedSol)

	if isSOLQuoteMint {
		preInstructions := make([]solana.Instruction, 0, 2)

		tokenBaseAccount, err := helpers.FindAssociatedTokenAddress(
			param.Receiver,
			poolState.BaseMint,
			tokenBaseProgram,
		)
		if err != nil {
			return nil, err
		}

		tokenQuoteAccount, err := helpers.FindAssociatedTokenAddress(
			param.TempWSolAcc,
			poolConfigState.QuoteMint,
			tokenQuoteProgram,
		)
		if err != nil {
			return nil, err
		}

		preInstructions = append(preInstructions,
			helpers.CreateAssociatedTokenAccountIdempotentInstruction(
				param.Payer,
				tokenBaseAccount,
				param.Receiver,
				poolState.BaseMint,
				tokenBaseAccount,
			),
			helpers.CreateAssociatedTokenAccountIdempotentInstruction(
				param.Payer,
				tokenQuoteAccount,
				param.FeeClaimer,
				poolConfigState.QuoteMint,
				tokenQuoteProgram,
			),
		)

		unwrapSolIx, err := helpers.UnwrapSOLInstruction(
			param.FeeClaimer,
			param.Receiver,
			false,
		)
		if err != nil {
			return nil, err
		}

		createTradingFeePtr := dbc.NewClaimTradingFeeInstruction(
			param.MaxBaseAmount.Uint64(),
			param.MaxQuoteAmount.Uint64(),
			p.state.GetPoolAuthority(),
			poolState.Config,
			param.Pool,
			tokenBaseAccount,
			tokenQuoteAccount,
			poolState.BaseVault,
			poolState.QuoteVault,
			poolState.BaseMint,
			poolConfigState.QuoteMint,
			param.FeeClaimer,
			tokenBaseProgram,
			tokenQuoteProgram,
			solana.PublicKey{},
			constants.DBCProgramId,
		)
		eventAuthPDA, _, err := createTradingFeePtr.FindEventAuthorityAddress()
		if err != nil {
			return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
		}

		currentIx, err := createTradingFeePtr.
			SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
		if err != nil {
			return nil, err
		}

		ixns := make([]solana.Instruction, 0, len(preInstructions)+1+1)
		ixns = append(ixns, preInstructions...)
		return append(ixns, currentIx, unwrapSolIx), nil
	}

	out, err := p.claimWithQuoteMintNotSol(
		ctx,
		types.ClaimPartnerTradingFeeWithQuoteMintNotSolParam{
			FeeClaimer:        param.FeeClaimer,
			Payer:             param.Payer,
			FeeReceiver:       param.Receiver,
			Config:            poolState.Config,
			Pool:              param.Pool,
			PoolState:         poolState,
			PoolConfigState:   poolConfigState,
			TokenBaseProgram:  tokenBaseProgram,
			TokenQuoteProgram: tokenQuoteProgram,
		},
	)
	if err != nil {
		return nil, err
	}

	createTradingFeePtr := dbc.NewClaimTradingFeeInstruction(
		param.MaxBaseAmount.Uint64(),
		param.MaxQuoteAmount.Uint64(),
		out.PoolAuthority,
		out.Config,
		out.Pool,
		out.TokenAAccount,
		out.TokenBAccount,
		out.BaseVault,
		out.QuoteVault,
		out.BaseMint,
		out.QuoteMint,
		out.FeeClaimer,
		out.TokenBaseProgram,
		out.TokenQuoteProgram,
		solana.PublicKey{},
		constants.DBCProgramId,
	)

	eventAuthPDA, _, err := createTradingFeePtr.FindEventAuthorityAddress()
	if err != nil {
		return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
	}

	currentIx, err := createTradingFeePtr.
		SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
	if err != nil {
		return nil, err
	}

	ixns := make([]solana.Instruction, 0, len(out.PreInstructions)+1+1)
	ixns = append(ixns, out.PreInstructions...)
	return append(ixns, currentIx), nil
}

// PartnerWithdrawSurplus partner withdraw surplus.
func (p *PartnerService) PartnerWithdrawSurplus(
	ctx context.Context,
	param types.PartnerWithdrawSurplusParam,
) ([]solana.Instruction, error) {

	poolState, err := p.state.GetPool(ctx, param.VirtualPool)
	if err != nil {
		return nil, fmt.Errorf("pool not found: %s", param.VirtualPool.String())
	}

	poolConfigState, err := p.state.GetPoolConfig(ctx, poolState.Config)
	if err != nil {
		return nil, fmt.Errorf("pool config not found: %s", param.VirtualPool.String())
	}

	tokenQuoteProgram := helpers.GetTokenProgram(poolConfigState.QuoteTokenFlag)

	tokenQuoteAccount, ix, err := helpers.GetOrCreateATAInstruction(
		ctx,
		p.state.conn,
		poolConfigState.QuoteMint,
		param.FeeClaimer,
		param.FeeClaimer,
		true,
		tokenQuoteProgram,
	)
	if err != nil {
		return nil, err
	}
	var unwrapSolIx *token.Instruction
	if !poolConfigState.QuoteMint.Equals(solana.WrappedSol) {
		if unwrapSolIx, err = helpers.UnwrapSOLInstruction(
			param.FeeClaimer,
			param.FeeClaimer,
			false,
		); err != nil {
			return nil, err
		}
	}

	partnerWithdrawSurplusPtr := dbc.NewPartnerWithdrawSurplusInstruction(
		p.state.GetPoolAuthority(),
		poolState.Config,
		param.VirtualPool,
		tokenQuoteAccount,
		poolState.QuoteVault,
		poolConfigState.QuoteMint,
		param.FeeClaimer,
		tokenQuoteProgram,
		solana.PublicKey{},
		constants.DBCProgramId,
	)
	eventAuthPDA, _, err := partnerWithdrawSurplusPtr.FindEventAuthorityAddress()
	if err != nil {
		return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
	}

	currentIx, err := partnerWithdrawSurplusPtr.
		SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
	if err != nil {
		return nil, err
	}

	ixns := make([]solana.Instruction, 0, 1+1+1)
	ixns = append(ixns, ix, currentIx, unwrapSolIx)
	return append(ixns, currentIx), nil
}

// PartnerWithdrawMigrationFee partner  withdraw migration fee.
func (p *PartnerService) PartnerWithdrawMigrationFee(
	ctx context.Context,
	param types.WithdrawMigrationFeeParam,
) ([]solana.Instruction, error) {

	virtualPoolState, err := p.state.GetPool(ctx, param.VirtualPool)
	if err != nil {
		return nil, fmt.Errorf("pool not found: %s", param.VirtualPool.String())
	}

	configState, err := p.state.GetPoolConfig(ctx, virtualPoolState.Config)
	if err != nil {
		return nil, fmt.Errorf("pool config not found: %s", param.VirtualPool.String())
	}

	tokenQuoteProgram := helpers.GetTokenProgram(configState.QuoteTokenFlag)

	feePayer := param.Sender
	if !param.FeePayer.IsZero() {
		feePayer = *param.FeePayer
	}
	tokenQuoteAccount, ix, err := helpers.GetOrCreateATAInstruction(
		ctx,
		p.state.conn,
		configState.QuoteMint,
		param.Sender,
		feePayer,
		true,
		tokenQuoteProgram,
	)
	if err != nil {
		return nil, err
	}
	var unwrapSolIx *token.Instruction
	if !configState.QuoteMint.Equals(solana.WrappedSol) {
		if unwrapSolIx, err = helpers.UnwrapSOLInstruction(
			param.Sender,
			param.Sender,
			false,
		); err != nil {
			return nil, err
		}
	}

	withdrawMigrationFeePtr := dbc.NewWithdrawMigrationFeeInstruction(
		0, // 0 as partner and 1 as creator
		p.state.GetPoolAuthority(),
		virtualPoolState.Config,
		param.VirtualPool,
		tokenQuoteAccount,
		virtualPoolState.QuoteVault,
		configState.QuoteMint,
		param.Sender,
		tokenQuoteProgram,
		solana.PublicKey{},
		constants.DBCProgramId,
	)
	eventAuthPDA, _, err := withdrawMigrationFeePtr.FindEventAuthorityAddress()
	if err != nil {
		return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
	}

	currentIx, err := withdrawMigrationFeePtr.
		SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
	if err != nil {
		return nil, err
	}

	ixns := make([]solana.Instruction, 0, 1+1+1)
	ixns = append(ixns, ix, currentIx, unwrapSolIx)
	return append(ixns, currentIx), nil
}
