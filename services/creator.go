package services

import (
	"context"
	"dbcGoSDK/constants"
	"dbcGoSDK/generated/dbc"
	"dbcGoSDK/helpers"
	"dbcGoSDK/types"
	"fmt"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
)

type CreatorService struct {
	state *StateService
}

func NewCreatorService(
	conn *rpc.Client,
	commitment rpc.CommitmentType,
) *CreatorService {
	return &CreatorService{
		state: NewStateService(conn, commitment),
	}
}

// CreatePoolMetadata create virtual pool metadata.
func (c *CreatorService) CreatePoolMetadata(
	param types.CreateVirtualPoolMetadataParam,
) (*dbc.Instruction, error) {
	createVirtualPoolMetadataPtr := dbc.NewCreateVirtualPoolMetadataInstruction(
		dbc.CreateVirtualPoolMetadataParameters{
			Padding: [96]uint8{},
			Name:    param.Name,
			Website: param.Website,
			Logo:    param.Logo,
		},
		param.VirtualPool,
		helpers.DeriveDbcPoolMetadata(param.VirtualPool),
		param.Creator,
		param.Payer,
		system.ProgramID,
		solana.PublicKey{},
		constants.DBCProgramId,
	)
	eventAuthPDA, _, err := createVirtualPoolMetadataPtr.FindEventAuthorityAddress()
	if err != nil {
		return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
	}

	return createVirtualPoolMetadataPtr.
		SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
}

// claimWithQuoteMintSol claims trading fee with quote mint SOL.
func (c *CreatorService) claimWithQuoteMintSol(
	param types.ClaimCreatorTradingFeeWithQuoteMintSolParam,
) (struct {
	Accounts struct {
		PoolAuthority     solana.PublicKey
		Pool              solana.PublicKey
		TokenAAccount     solana.PublicKey
		TokenBAccount     solana.PublicKey
		BaseVault         solana.PublicKey
		QuoteVault        solana.PublicKey
		BaseMint          solana.PublicKey
		QuoteMint         solana.PublicKey
		Creator           solana.PublicKey
		TokenBaseProgram  solana.PublicKey
		TokenQuoteProgram solana.PublicKey
	}
	PreInstructions []solana.Instruction
	PostInstruction *token.Instruction
}, error) {

	tokenBaseAccount, err := helpers.FindAssociatedTokenAddress(
		param.FeeReceiver,
		param.PoolState.BaseMint,
		param.TokenBaseProgram,
	)
	if err != nil {
		return struct {
			Accounts struct {
				PoolAuthority     solana.PublicKey
				Pool              solana.PublicKey
				TokenAAccount     solana.PublicKey
				TokenBAccount     solana.PublicKey
				BaseVault         solana.PublicKey
				QuoteVault        solana.PublicKey
				BaseMint          solana.PublicKey
				QuoteMint         solana.PublicKey
				Creator           solana.PublicKey
				TokenBaseProgram  solana.PublicKey
				TokenQuoteProgram solana.PublicKey
			}
			PreInstructions []solana.Instruction
			PostInstruction *token.Instruction
		}{}, err
	}

	tokenQuoteAccount, err := helpers.FindAssociatedTokenAddress(
		param.TempWSolAcc,
		param.PoolConfigState.QuoteMint,
		param.TokenQuoteProgram,
	)
	if err != nil {
		return struct {
			Accounts struct {
				PoolAuthority     solana.PublicKey
				Pool              solana.PublicKey
				TokenAAccount     solana.PublicKey
				TokenBAccount     solana.PublicKey
				BaseVault         solana.PublicKey
				QuoteVault        solana.PublicKey
				BaseMint          solana.PublicKey
				QuoteMint         solana.PublicKey
				Creator           solana.PublicKey
				TokenBaseProgram  solana.PublicKey
				TokenQuoteProgram solana.PublicKey
			}
			PreInstructions []solana.Instruction
			PostInstruction *token.Instruction
		}{}, err
	}

	preInstructions := make([]solana.Instruction, 0, 2)
	preInstructions = append(preInstructions,
		helpers.CreateAssociatedTokenAccountIdempotentInstruction(
			param.Payer,
			tokenBaseAccount,
			param.FeeReceiver,
			param.PoolState.BaseMint,
			param.TokenBaseProgram,
		),
		helpers.CreateAssociatedTokenAccountIdempotentInstruction(
			param.Payer,
			tokenQuoteAccount,
			param.TempWSolAcc,
			param.PoolConfigState.QuoteMint,
			param.TokenQuoteProgram,
		),
	)
	postInstruction, err := helpers.UnwrapSOLInstruction(
		param.TempWSolAcc,
		param.FeeReceiver,
		false,
	)
	if err != nil {
		return struct {
			Accounts struct {
				PoolAuthority     solana.PublicKey
				Pool              solana.PublicKey
				TokenAAccount     solana.PublicKey
				TokenBAccount     solana.PublicKey
				BaseVault         solana.PublicKey
				QuoteVault        solana.PublicKey
				BaseMint          solana.PublicKey
				QuoteMint         solana.PublicKey
				Creator           solana.PublicKey
				TokenBaseProgram  solana.PublicKey
				TokenQuoteProgram solana.PublicKey
			}
			PreInstructions []solana.Instruction
			PostInstruction *token.Instruction
		}{}, err
	}

	return struct {
		Accounts struct {
			PoolAuthority     solana.PublicKey
			Pool              solana.PublicKey
			TokenAAccount     solana.PublicKey
			TokenBAccount     solana.PublicKey
			BaseVault         solana.PublicKey
			QuoteVault        solana.PublicKey
			BaseMint          solana.PublicKey
			QuoteMint         solana.PublicKey
			Creator           solana.PublicKey
			TokenBaseProgram  solana.PublicKey
			TokenQuoteProgram solana.PublicKey
		}
		PreInstructions []solana.Instruction
		PostInstruction *token.Instruction
	}{
		Accounts: struct {
			PoolAuthority     solana.PublicKey
			Pool              solana.PublicKey
			TokenAAccount     solana.PublicKey
			TokenBAccount     solana.PublicKey
			BaseVault         solana.PublicKey
			QuoteVault        solana.PublicKey
			BaseMint          solana.PublicKey
			QuoteMint         solana.PublicKey
			Creator           solana.PublicKey
			TokenBaseProgram  solana.PublicKey
			TokenQuoteProgram solana.PublicKey
		}{
			PoolAuthority:     c.state.GetPoolAuthority(),
			Pool:              param.Pool,
			TokenAAccount:     tokenBaseAccount,
			TokenBAccount:     tokenQuoteAccount,
			BaseVault:         param.PoolState.BaseVault,
			QuoteVault:        param.PoolState.QuoteVault,
			BaseMint:          param.PoolState.BaseMint,
			QuoteMint:         param.PoolConfigState.QuoteMint,
			Creator:           param.Creator,
			TokenBaseProgram:  param.TokenBaseProgram,
			TokenQuoteProgram: param.TokenQuoteProgram,
		},
		PreInstructions: preInstructions,
		PostInstruction: postInstruction,
	}, nil
}

// claimWithQuoteMintNotSol claims trading fee with quote mint not SOL.
func (c *CreatorService) claimWithQuoteMintNotSol(
	ctx context.Context,
	param types.ClaimCreatorTradingFeeWithQuoteMintNotSolParam,
) (struct {
	Accounts struct {
		PoolAuthority     solana.PublicKey
		Pool              solana.PublicKey
		TokenAAccount     solana.PublicKey
		TokenBAccount     solana.PublicKey
		BaseVault         solana.PublicKey
		QuoteVault        solana.PublicKey
		BaseMint          solana.PublicKey
		QuoteMint         solana.PublicKey
		Creator           solana.PublicKey
		TokenBaseProgram  solana.PublicKey
		TokenQuoteProgram solana.PublicKey
	}
	PreInstructions []solana.Instruction
}, error) {

	out, err := c.state.prepareTokenAccounts(
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
			Accounts struct {
				PoolAuthority     solana.PublicKey
				Pool              solana.PublicKey
				TokenAAccount     solana.PublicKey
				TokenBAccount     solana.PublicKey
				BaseVault         solana.PublicKey
				QuoteVault        solana.PublicKey
				BaseMint          solana.PublicKey
				QuoteMint         solana.PublicKey
				Creator           solana.PublicKey
				TokenBaseProgram  solana.PublicKey
				TokenQuoteProgram solana.PublicKey
			}
			PreInstructions []solana.Instruction
		}{}, err
	}

	return struct {
		Accounts struct {
			PoolAuthority     solana.PublicKey
			Pool              solana.PublicKey
			TokenAAccount     solana.PublicKey
			TokenBAccount     solana.PublicKey
			BaseVault         solana.PublicKey
			QuoteVault        solana.PublicKey
			BaseMint          solana.PublicKey
			QuoteMint         solana.PublicKey
			Creator           solana.PublicKey
			TokenBaseProgram  solana.PublicKey
			TokenQuoteProgram solana.PublicKey
		}
		PreInstructions []solana.Instruction
	}{
		Accounts: struct {
			PoolAuthority     solana.PublicKey
			Pool              solana.PublicKey
			TokenAAccount     solana.PublicKey
			TokenBAccount     solana.PublicKey
			BaseVault         solana.PublicKey
			QuoteVault        solana.PublicKey
			BaseMint          solana.PublicKey
			QuoteMint         solana.PublicKey
			Creator           solana.PublicKey
			TokenBaseProgram  solana.PublicKey
			TokenQuoteProgram solana.PublicKey
		}{
			PoolAuthority:     c.state.GetPoolAuthority(),
			Pool:              param.Pool,
			TokenAAccount:     out.TokenAAta,
			TokenBAccount:     out.TokenBAta,
			BaseVault:         param.PoolState.BaseVault,
			QuoteVault:        param.PoolState.QuoteVault,
			BaseMint:          param.PoolState.BaseMint,
			QuoteMint:         param.PoolConfigState.QuoteMint,
			Creator:           param.Creator,
			TokenBaseProgram:  param.TokenBaseProgram,
			TokenQuoteProgram: param.TokenQuoteProgram,
		},
		PreInstructions: out.CreateATAIxns,
	}, nil

}

// ClaimCreatorTradingFee claims creator trading fee.
func (c *CreatorService) ClaimCreatorTradingFee(
	ctx context.Context,
	param types.ClaimCreatorTradingFeeParam,
) ([]solana.Instruction, error) {

	poolState, err := c.state.GetPool(ctx, param.Pool)
	if err != nil {
		return nil, fmt.Errorf("pool(%s) not found: error: %w", param.Pool.String(), err)
	}

	poolConfigState, err := c.state.GetPoolConfig(ctx, poolState.Config)
	if err != nil {
		return nil, fmt.Errorf("pool config(%s) not found: error: %w", param.Pool.String(), err)
	}

	tokenBaseProgram := helpers.GetTokenProgram(poolConfigState.TokenType)
	tokenQuoteProgram := helpers.GetTokenProgram(poolConfigState.QuoteTokenFlag)
	isSOLQuoteMint := poolConfigState.QuoteMint.Equals(solana.WrappedSol)

	// if receiver is present and not equal to creator, use tempWSolAcc, otherwise use creator
	tempWSol, feeReceiver := param.Creator, param.Creator
	if !param.Receiver.IsZero() && !param.Receiver.Equals(param.Creator) {
		tempWSol = param.TempWSolAcc
	}

	// if receiver is provided, use receiver, otherwise use creator
	if !param.Receiver.IsZero() {
		feeReceiver = param.Receiver
	}
	if isSOLQuoteMint {
		out, err := c.claimWithQuoteMintSol(
			types.ClaimCreatorTradingFeeWithQuoteMintSolParam{
				ClaimCreatorTradingFeeWithQuoteMintNotSolParam: types.ClaimCreatorTradingFeeWithQuoteMintNotSolParam{
					Creator:           param.Creator,
					Payer:             param.Payer,
					FeeReceiver:       feeReceiver,
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

		claimCreatorTradingFeePtr := dbc.NewClaimCreatorTradingFeeInstruction(
			param.MaxBaseAmount,
			param.MaxQuoteAmount,
			out.Accounts.PoolAuthority,
			out.Accounts.Pool,
			out.Accounts.TokenAAccount,
			out.Accounts.TokenBAccount,
			out.Accounts.BaseVault,
			out.Accounts.QuoteVault,
			out.Accounts.BaseMint,
			out.Accounts.QuoteMint,
			out.Accounts.Creator,
			out.Accounts.TokenBaseProgram,
			out.Accounts.TokenQuoteProgram,
			solana.PublicKey{},
			constants.DBCProgramId,
		)
		eventAuthPDA, _, err := claimCreatorTradingFeePtr.FindEventAuthorityAddress()
		if err != nil {
			return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
		}

		currentIx, err := claimCreatorTradingFeePtr.
			SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
		if err != nil {
			return nil, err
		}

		finalIxns := make([]solana.Instruction, 0, len(out.PreInstructions)+1+1)
		finalIxns = append(finalIxns, out.PreInstructions...)
		finalIxns = append(finalIxns, currentIx, out.PostInstruction)
		return finalIxns, nil
	}

	out, err := c.claimWithQuoteMintNotSol(
		ctx,
		types.ClaimCreatorTradingFeeWithQuoteMintNotSolParam{
			Creator:           param.Creator,
			Payer:             param.Payer,
			FeeReceiver:       feeReceiver,
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
	claimCreatorTradingFeePtr := dbc.NewClaimCreatorTradingFeeInstruction(
		param.MaxBaseAmount,
		param.MaxQuoteAmount,
		out.Accounts.PoolAuthority,
		out.Accounts.Pool,
		out.Accounts.TokenAAccount,
		out.Accounts.TokenBAccount,
		out.Accounts.BaseVault,
		out.Accounts.QuoteVault,
		out.Accounts.BaseMint,
		out.Accounts.QuoteMint,
		out.Accounts.Creator,
		out.Accounts.TokenBaseProgram,
		out.Accounts.TokenQuoteProgram,
		solana.PublicKey{},
		constants.DBCProgramId,
	)
	eventAuthPDA, _, err := claimCreatorTradingFeePtr.FindEventAuthorityAddress()
	if err != nil {
		return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
	}

	currentIx, err := claimCreatorTradingFeePtr.
		SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
	if err != nil {
		return nil, err
	}

	finalIxns := make([]solana.Instruction, 0, len(out.PreInstructions)+1)
	finalIxns = append(finalIxns, out.PreInstructions...)
	finalIxns = append(finalIxns, currentIx)
	return finalIxns, nil
}

// ClaimCreatorTradingFee2 claims creator trading fee.
func (c *CreatorService) ClaimCreatorTradingFee2(
	ctx context.Context,
	param types.ClaimCreatorTradingFee2Param,
) ([]solana.Instruction, error) {

	poolState, err := c.state.GetPool(ctx, param.Pool)
	if err != nil {
		return nil, fmt.Errorf("pool(%s) not found: error: %w", param.Pool.String(), err)
	}

	poolConfigState, err := c.state.GetPoolConfig(ctx, poolState.Config)
	if err != nil {
		return nil, fmt.Errorf("pool config(%s) not found: error: %w", param.Pool.String(), err)
	}

	tokenBaseProgram := helpers.GetTokenProgram(poolConfigState.TokenType)
	tokenQuoteProgram := helpers.GetTokenProgram(poolConfigState.QuoteTokenFlag)
	isSOLQuoteMint := poolConfigState.QuoteMint.Equals(solana.WrappedSol)

	tokenBaseAccount, err := helpers.FindAssociatedTokenAddress(
		param.Receiver,
		poolState.BaseMint,
		tokenBaseProgram,
	)
	if err != nil {
		return nil, err
	}

	tokenQuoteAccount, err := helpers.FindAssociatedTokenAddress(
		param.Creator,
		poolConfigState.QuoteMint,
		tokenQuoteProgram,
	)
	if err != nil {
		return nil, err
	}

	preInstructions := make([]solana.Instruction, 0, 2)
	preInstructions = append(preInstructions,
		helpers.CreateAssociatedTokenAccountIdempotentInstruction(
			param.Payer,
			tokenBaseAccount,
			param.Receiver,
			poolState.BaseMint,
			tokenBaseProgram,
		),
		helpers.CreateAssociatedTokenAccountIdempotentInstruction(
			param.Payer,
			tokenQuoteAccount,
			param.Creator,
			poolConfigState.QuoteMint,
			tokenQuoteProgram,
		),
	)

	postInstruction, err := helpers.UnwrapSOLInstruction(
		param.Creator,
		param.Receiver,
		false,
	)
	if err != nil {
		return nil, err
	}

	if isSOLQuoteMint {
		claimCreatorTradingFeePtr := dbc.NewClaimCreatorTradingFeeInstruction(
			param.MaxBaseAmount,
			param.MaxQuoteAmount,
			c.state.GetPoolAuthority(),
			param.Pool,
			tokenBaseAccount,
			tokenQuoteAccount,
			poolState.BaseVault,
			poolState.QuoteVault,
			poolState.BaseMint,
			poolConfigState.QuoteMint,
			param.Creator,
			tokenBaseProgram,
			tokenQuoteProgram,
			solana.PublicKey{},
			constants.DBCProgramId,
		)
		eventAuthPDA, _, err := claimCreatorTradingFeePtr.FindEventAuthorityAddress()
		if err != nil {
			return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
		}

		currentIx, err := claimCreatorTradingFeePtr.
			SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
		if err != nil {
			return nil, err
		}

		finalIxns := make([]solana.Instruction, 0, len(preInstructions)+1+1)
		finalIxns = append(finalIxns, preInstructions...)
		finalIxns = append(finalIxns, currentIx, postInstruction)
		return finalIxns, nil
	}

	out, err := c.claimWithQuoteMintNotSol(
		ctx,
		types.ClaimCreatorTradingFeeWithQuoteMintNotSolParam{
			Creator:           param.Creator,
			Payer:             param.Payer,
			FeeReceiver:       param.Receiver,
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
	claimCreatorTradingFeePtr := dbc.NewClaimCreatorTradingFeeInstruction(
		param.MaxBaseAmount,
		param.MaxQuoteAmount,
		out.Accounts.PoolAuthority,
		out.Accounts.Pool,
		out.Accounts.TokenAAccount,
		out.Accounts.TokenBAccount,
		out.Accounts.BaseVault,
		out.Accounts.QuoteMint,
		out.Accounts.BaseMint,
		out.Accounts.QuoteMint,
		out.Accounts.Creator,
		out.Accounts.TokenBaseProgram,
		out.Accounts.TokenQuoteProgram,
		solana.PublicKey{},
		constants.DBCProgramId,
	)
	eventAuthPDA, _, err := claimCreatorTradingFeePtr.FindEventAuthorityAddress()
	if err != nil {
		return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
	}

	currentIx, err := claimCreatorTradingFeePtr.
		SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
	if err != nil {
		return nil, err
	}

	finalIxns := make([]solana.Instruction, 0, len(out.PreInstructions)+1)
	finalIxns = append(finalIxns, out.PreInstructions...)
	finalIxns = append(finalIxns, currentIx)
	return finalIxns, nil
}

// CreatorWithdrawSurplus withdraw creator surplus.
func (c *CreatorService) CreatorWithdrawSurplus(
	ctx context.Context,
	param types.CreatorWithdrawSurplusParam,
) ([]solana.Instruction, error) {
	poolState, err := c.state.GetPool(ctx, param.VirtualPool)
	if err != nil {
		return nil, fmt.Errorf("pool(%s) not found: error: %w", param.VirtualPool.String(), err)
	}

	poolConfigState, err := c.state.GetPoolConfig(ctx, poolState.Config)
	if err != nil {
		return nil, fmt.Errorf("pool config(%s) not found: error: %w", param.VirtualPool.String(), err)
	}

	tokenQuoteAccount, err := helpers.FindAssociatedTokenAddress(
		param.Creator,
		poolConfigState.QuoteMint,
		solana.TokenProgramID,
	)
	if err != nil {
		return nil, err
	}

	createQuoteTokenAccountIx := helpers.CreateAssociatedTokenAccountIdempotentInstruction(
		param.Creator,
		tokenQuoteAccount,
		param.Creator,
		poolConfigState.QuoteMint,
		solana.TokenProgramID,
	)

	var postInstruction *token.Instruction
	isSOLQuoteMint := poolConfigState.QuoteMint.Equals(solana.WrappedSol)
	if isSOLQuoteMint {
		unwrapIx, err := helpers.UnwrapSOLInstruction(
			param.Creator, param.Creator, false,
		)
		if err != nil {
			return nil, err
		}
		postInstruction = unwrapIx
	}

	creatorWithdrawSurplusPtr := dbc.NewCreatorWithdrawSurplusInstruction(
		c.state.GetPoolAuthority(),
		poolState.Config,
		param.VirtualPool,
		tokenQuoteAccount,
		poolState.QuoteVault,
		poolConfigState.QuoteMint,
		param.Creator,
		solana.TokenProgramID,
		solana.PublicKey{},
		constants.DBCProgramId,
	)

	eventAuthPDA, _, err := creatorWithdrawSurplusPtr.FindEventAuthorityAddress()
	if err != nil {
		return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
	}

	currentIx, err := creatorWithdrawSurplusPtr.
		SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
	if err != nil {
		return nil, err
	}

	finalIxns := make([]solana.Instruction, 0, 1+1+1)
	finalIxns = append(finalIxns, createQuoteTokenAccountIx)
	finalIxns = append(finalIxns, currentIx, postInstruction)
	return finalIxns, nil
}

// TransferPoolCreator transfers pool creator.
func (c *CreatorService) TransferPoolCreator(
	ctx context.Context,
	param types.TransferPoolCreatorParam,
) (*dbc.Instruction, error) {
	virtualPoolState, err := c.state.GetPool(ctx, param.VirtualPool)
	if err != nil {
		return nil, err
	}
	transferPoolCreatorPtr := dbc.NewTransferPoolCreatorInstruction(
		param.VirtualPool,
		virtualPoolState.Config,
		param.Creator,
		param.NewCreator,
		solana.PublicKey{},
		constants.DBCProgramId,
	)
	eventAuthPDA, _, err := transferPoolCreatorPtr.FindEventAuthorityAddress()
	if err != nil {
		return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
	}
	transferPoolCreatorPtr.SetEventAuthorityAccount(eventAuthPDA)

	transferPoolCreatorPtr.AccountMetaSlice = append(
		transferPoolCreatorPtr.AccountMetaSlice,
		&solana.AccountMeta{
			PublicKey: helpers.DeriveDammV1MigrationMetadataAddress(param.VirtualPool),
		},
	)

	return transferPoolCreatorPtr.ValidateAndBuild()
}

// CreatorWithdrawMigrationFee allows creator to withdraw migration fee.
func (c *CreatorService) CreatorWithdrawMigrationFee(
	ctx context.Context,
	param types.WithdrawMigrationFeeParam,
) ([]solana.Instruction, error) {
	virtualPoolState, err := c.state.GetPool(ctx, param.VirtualPool)
	if err != nil {
		return nil, err
	}

	configState, err := c.state.GetPoolConfig(ctx, virtualPoolState.Config)
	if err != nil {
		return nil, err
	}

	payer := param.Sender
	if !param.FeePayer.IsZero() {
		payer = *param.FeePayer
	}

	tokenQuoteAccount, ix, err := helpers.GetOrCreateATAInstruction(
		ctx,
		c.state.conn,
		configState.QuoteMint,
		param.Sender,
		payer,
		true,
		helpers.GetTokenProgram(configState.QuoteTokenFlag),
	)
	if err != nil {
		return nil, err
	}

	var postInstruction *token.Instruction
	if configState.QuoteMint.Equals(solana.WrappedSol) {
		unwarpSOLIx, err := helpers.UnwrapSOLInstruction(
			param.Sender,
			param.Sender,
			false,
		)
		if err != nil {
			return nil, err
		}
		postInstruction = unwarpSOLIx
	}

	withdrawMigrationFeePtr := dbc.NewWithdrawMigrationFeeInstruction(
		1, // 0 as partner and 1 as creator
		c.state.GetPoolAuthority(),
		virtualPoolState.Config,
		param.VirtualPool,
		tokenQuoteAccount,
		virtualPoolState.QuoteVault,
		configState.QuoteMint,
		param.Sender,
		helpers.GetTokenProgram(configState.QuoteTokenFlag),
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

	return append(make([]solana.Instruction, 0, 3), ix, currentIx, postInstruction), nil
}
