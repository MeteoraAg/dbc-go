package services

import (
	"context"
	"dbcGoSDK/anchor"
	"dbcGoSDK/constants"
	"dbcGoSDK/generated/dbc"
	dynamic_vault "dbcGoSDK/generated/dynamicVault"
	"dbcGoSDK/helpers"
	"dbcGoSDK/types"
	"fmt"
	"sync"

	"github.com/gagliardetto/solana-go"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/rpc"
)

type MigrationService struct {
	state *StateService
}

func NewMigrationService(
	conn *rpc.Client,
	commitment rpc.CommitmentType,
) *MigrationService {
	return &MigrationService{
		state: NewStateService(conn, commitment),
	}
}

func (m *MigrationService) CreateLocker(
	ctx context.Context,
	param types.CreateLockerParam,
) ([]solana.Instruction, error) {

	virtualPoolState, err := m.state.GetPool(ctx, param.VirtualPool)
	if err != nil {
		return nil, fmt.Errorf("pool(%s) not found: err: %w", param.VirtualPool.String(), err)
	}

	poolConfigState, err := m.state.GetPoolConfig(ctx, virtualPoolState.Config)
	if err != nil {
		return nil, fmt.Errorf("pool config(%s) not found: err: %w", param.VirtualPool.String(), err)
	}
	base := helpers.DeriveBaseKeyForLocker(param.VirtualPool)
	escrow := helpers.DeriveBaseKeyForLocker(base)

	tokenProgram := solana.Token2022ProgramID
	if poolConfigState.TokenType == 0 {
		tokenProgram = solana.TokenProgramID
	}

	escrowToken, err := helpers.FindAssociatedTokenAddress(
		escrow,
		virtualPoolState.BaseMint,
		tokenProgram,
	)
	if err != nil {
		return nil, err
	}

	createOwnerEscrowVaultTokenXIx := helpers.CreateAssociatedTokenAccountIdempotentInstruction(
		param.Payer,
		escrowToken,
		escrow,
		virtualPoolState.BaseMint,
		tokenProgram,
	)
	currentIx, err := dbc.NewCreateLockerInstruction(
		param.VirtualPool,
		virtualPoolState.Config,
		m.state.GetPoolAuthority(),
		virtualPoolState.BaseVault,
		virtualPoolState.BaseMint,
		base,
		virtualPoolState.Creator,
		escrow,
		escrowToken,
		param.Payer,
		tokenProgram,
		constants.LockerProgramId,
		helpers.DeriveLockerEventAuthority(),
		solana.SystemProgramID,
	).ValidateAndBuild()
	if err != nil {
		return nil, err
	}
	ixns := make([]solana.Instruction, 0, 2)
	return append(ixns, createOwnerEscrowVaultTokenXIx, currentIx), nil
}

func (m *MigrationService) WithdrawLeftover(
	ctx context.Context,
	param types.WithdrawLeftoverParam,
) ([]solana.Instruction, error) {

	poolState, err := m.state.GetPool(ctx, param.VirtualPool)
	if err != nil {
		return nil, fmt.Errorf("pool(%s) not found: err: %w", param.VirtualPool.String(), err)
	}

	poolConfigState, err := m.state.GetPoolConfig(ctx, poolState.Config)
	if err != nil {
		return nil, fmt.Errorf("pool config(%s) not found: err: %w", param.VirtualPool.String(), err)
	}

	tokenBaseProgram := helpers.GetTokenProgram(poolConfigState.TokenType)

	tokenBaseAccount, ix, err := helpers.GetOrCreateATAInstruction(
		ctx,
		m.state.conn,
		poolState.BaseMint,
		poolConfigState.LeftoverReceiver,
		param.Payer,
		true,
		tokenBaseProgram,
	)
	if err != nil {
		return nil, err
	}

	withdrawLeftovePtr := dbc.NewWithdrawLeftoverInstruction(
		m.state.GetPoolAuthority(),
		poolState.Config,
		param.VirtualPool,
		tokenBaseAccount,
		poolState.BaseVault,
		poolState.BaseMint,
		poolConfigState.LeftoverReceiver,
		tokenBaseProgram,
		solana.PublicKey{},
		constants.DBCProgramId,
	)
	eventAuthPDA, _, err := withdrawLeftovePtr.FindEventAuthorityAddress()
	if err != nil {
		return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
	}

	currentIx, err := withdrawLeftovePtr.
		SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
	if err != nil {
		return nil, err
	}

	ixns := make([]solana.Instruction, 0, 2)
	return append(ixns, ix, currentIx), nil
}

///////////////////////
// DAMM V1 FUNCTIONS //
///////////////////////

func (m *MigrationService) CreateDammV1MigrationMetadata(
	param types.CreateDammV1MigrationMetadataParam,
) (*dbc.Instruction, error) {

	migrationMeteoraDammCreateMetadataPtr := dbc.NewMigrationMeteoraDammCreateMetadataInstruction(
		param.VirtualPool,
		param.Config,
		helpers.DeriveDammV1MigrationMetadataAddress(param.VirtualPool),
		param.Payer,
		solana.SystemProgramID,
		solana.PublicKey{},
		constants.DBCProgramId,
	)
	eventAuthPDA, _, err := migrationMeteoraDammCreateMetadataPtr.FindEventAuthorityAddress()
	if err != nil {
		return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
	}

	return migrationMeteoraDammCreateMetadataPtr.
		SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
}

func (m *MigrationService) MigrateToDammV1(
	ctx context.Context,
	param types.MigrateToDammV1Param,
) ([]solana.Instruction, error) {

	poolState, err := m.state.GetPool(ctx, param.VirtualPool)
	if err != nil {
		return nil, fmt.Errorf("pool(%s) not found: err: %w", param.VirtualPool.String(), err)
	}

	poolConfigState, err := m.state.GetPoolConfig(ctx, poolState.Config)
	if err != nil {
		return nil, fmt.Errorf("pool config(%s) not found: err: %w", param.VirtualPool.String(), err)
	}

	vaultPDAsA, err := helpers.DeriveVaultPdas(poolState.BaseMint, solana.PublicKey{})
	if err != nil {
		return nil, err
	}

	vaultPDAsB, err := helpers.DeriveVaultPdas(poolConfigState.QuoteMint, solana.PublicKey{})
	if err != nil {
		return nil, err
	}
	var (
		wg                           sync.WaitGroup
		aVaultAccount, bVaultAccount *dynamic_vault.VaultAccount
	)

	wg.Add(2)
	go func(out **dynamic_vault.VaultAccount) {
		defer wg.Done()
		result, _ := anchor.NewPgAccounts(
			m.state.conn,
			func() *dynamic_vault.VaultAccount { return &dynamic_vault.VaultAccount{} },
		).Fetch(ctx, vaultPDAsA.VaultPDA, nil)
		*out = result
	}(&aVaultAccount)

	go func(out **dynamic_vault.VaultAccount) {
		defer wg.Done()
		result, _ := anchor.NewPgAccounts(
			m.state.conn,
			func() *dynamic_vault.VaultAccount { return &dynamic_vault.VaultAccount{} },
		).Fetch(ctx, vaultPDAsB.VaultPDA, nil)
		*out = result
	}(&bVaultAccount)

	dammPool := helpers.DeriveDammV1PoolAddress(
		param.DammConfig,
		poolState.BaseMint,
		poolConfigState.QuoteMint,
	)

	lpMint := helpers.DeriveDammV1LpMintAddress(dammPool)
	mintMetadata := helpers.DeriveMintMetadata(lpMint)
	protocolTokenAFee := helpers.DeriveDammV1ProtocolFeeAddress(poolState.BaseMint, dammPool)
	protocolTokenBFee := helpers.DeriveDammV1ProtocolFeeAddress(poolConfigState.QuoteMint, dammPool)

	wg.Wait()

	preInstructions := make([]solana.Instruction, 0, 2)
	aVaultLpMint, bVaultLpMint := aVaultAccount.LpMint, bVaultAccount.LpMint
	if aVaultAccount == nil {
		createVaultAIx, err := helpers.CreateInitializePermissionlessDynamicVaultIx(
			poolState.BaseMint,
			param.Payer,
		)
		if err != nil {
			return nil, err
		}

		preInstructions = append(preInstructions, createVaultAIx.Ix)
	}

	if bVaultAccount == nil {
		createVaultAIx, err := helpers.CreateInitializePermissionlessDynamicVaultIx(
			poolConfigState.QuoteMint,
			param.Payer,
		)
		if err != nil {
			return nil, err
		}

		preInstructions = append(preInstructions, createVaultAIx.Ix)
	}

	aVaultLp := helpers.DeriveDammV1VaultLPAddress(vaultPDAsA.VaultPDA, dammPool)
	bVaultLp := helpers.DeriveDammV1VaultLPAddress(vaultPDAsB.VaultPDA, dammPool)

	virtualPoolLp, err := helpers.GetAssociatedTokenAddressSync(
		lpMint,
		m.state.GetPoolAuthority(),
		true,
		solana.TokenProgramID,
		solana.SPLAssociatedTokenAccountProgramID,
	)
	if err != nil {
		return nil, err
	}

	currentIx, err := dbc.NewMigrateMeteoraDammInstruction(
		param.VirtualPool,
		helpers.DeriveDammV1MigrationMetadataAddress(param.VirtualPool),
		poolState.Config,
		m.state.GetPoolAuthority(),
		dammPool,
		param.DammConfig,
		lpMint,
		poolState.BaseMint,
		poolConfigState.QuoteMint,
		vaultPDAsA.VaultPDA,
		vaultPDAsB.VaultPDA,
		vaultPDAsA.TokenVaultPDA,
		vaultPDAsB.TokenVaultPDA,
		aVaultLpMint,
		bVaultLpMint,
		aVaultLp,
		bVaultLp,
		poolState.BaseVault,
		poolState.QuoteVault,
		virtualPoolLp,
		protocolTokenAFee,
		protocolTokenBFee,
		param.Payer,
		solana.SysVarRentPubkey,
		mintMetadata,
		constants.MetaplexProgramId,
		constants.DammV1ProgramId,
		constants.VaultProgramId,
		solana.TokenProgramID,
		solana.SPLAssociatedTokenAccountProgramID,
		solana.SystemProgramID,
	).ValidateAndBuild()
	if err != nil {
		return nil, err
	}

	ixns := make([]solana.Instruction, 0, 4)
	ixns = append(ixns, preInstructions...)
	return append(ixns,
			currentIx,
			computebudget.NewSetComputeUnitPriceInstruction(
				500_000,
			).Build(),
		),
		nil
}

// LockDammV1LpToken locks DAMM V1 LP token for creator or partner.
func (m *MigrationService) LockDammV1LpToken(
	ctx context.Context,
	param types.DammLpTokenParam,
) ([]solana.Instruction, error) {
	poolState, err := m.state.GetPool(ctx, param.VirtualPool)
	if err != nil {
		return nil, fmt.Errorf("pool(%s) not found: err: %w", param.VirtualPool.String(), err)
	}

	poolConfigState, err := m.state.GetPoolConfig(ctx, poolState.Config)
	if err != nil {
		return nil, fmt.Errorf("pool config(%s) not found: err: %w", param.VirtualPool.String(), err)
	}

	vaultPDAsA, err := helpers.DeriveVaultPdas(poolState.BaseMint, solana.PublicKey{})
	if err != nil {
		return nil, err
	}

	vaultPDAsB, err := helpers.DeriveVaultPdas(poolConfigState.QuoteMint, solana.PublicKey{})
	if err != nil {
		return nil, err
	}
	var (
		wg                           sync.WaitGroup
		aVaultAccount, bVaultAccount *dynamic_vault.VaultAccount
	)

	wg.Add(2)
	go func(out **dynamic_vault.VaultAccount) {
		defer wg.Done()
		result, _ := anchor.NewPgAccounts(
			m.state.conn,
			func() *dynamic_vault.VaultAccount { return &dynamic_vault.VaultAccount{} },
		).Fetch(ctx, vaultPDAsA.VaultPDA, nil)
		*out = result
	}(&aVaultAccount)

	go func(out **dynamic_vault.VaultAccount) {
		defer wg.Done()
		result, _ := anchor.NewPgAccounts(
			m.state.conn,
			func() *dynamic_vault.VaultAccount { return &dynamic_vault.VaultAccount{} },
		).Fetch(ctx, vaultPDAsB.VaultPDA, nil)
		*out = result
	}(&bVaultAccount)

	preInstructions := make([]solana.Instruction, 0, 2)
	aVaultLpMint, bVaultLpMint := aVaultAccount.LpMint, bVaultAccount.LpMint
	if aVaultAccount == nil {
		createVaultAIx, err := helpers.CreateInitializePermissionlessDynamicVaultIx(
			poolState.BaseMint,
			param.Payer,
		)
		if err != nil {
			return nil, err
		}

		preInstructions = append(preInstructions, createVaultAIx.Ix)
	}

	if bVaultAccount == nil {
		createVaultAIx, err := helpers.CreateInitializePermissionlessDynamicVaultIx(
			poolConfigState.QuoteMint,
			param.Payer,
		)
		if err != nil {
			return nil, err
		}

		preInstructions = append(preInstructions, createVaultAIx.Ix)
	}

	dammPool := helpers.DeriveDammV1PoolAddress(
		param.DammConfig,
		poolState.BaseMint,
		poolConfigState.QuoteMint,
	)
	aVaultLp := helpers.DeriveDammV1VaultLPAddress(vaultPDAsA.VaultPDA, dammPool)
	bVaultLp := helpers.DeriveDammV1VaultLPAddress(vaultPDAsB.VaultPDA, dammPool)

	lpMint := helpers.DeriveDammV1LpMintAddress(dammPool)

	var lockEscrowKey solana.PublicKey
	if param.IsPartner {
		lockEscrowKey = helpers.DeriveDammV1LockEscrowAddress(dammPool, poolConfigState.FeeClaimer)
		if lockEscrowData, _ := m.state.conn.GetAccountInfo(ctx, lockEscrowKey); lockEscrowData == nil {
			ix, err := helpers.CreateLockEscrowIx(
				param.Payer,
				dammPool,
				lpMint,
				poolConfigState.FeeClaimer,
				lockEscrowKey,
			)
			if err != nil {
				return nil, err
			}
			preInstructions = append(preInstructions, ix)
		}
	} else {
		lockEscrowKey = helpers.DeriveDammV1LockEscrowAddress(
			dammPool, poolState.Creator,
		)
		if lockEscrowData, _ := m.state.conn.GetAccountInfo(ctx, lockEscrowKey); lockEscrowData == nil {
			ix, err := helpers.CreateLockEscrowIx(
				param.Payer,
				dammPool,
				lpMint,
				poolState.Creator,
				lockEscrowKey,
			)
			if err != nil {
				return nil, err
			}
			preInstructions = append(preInstructions, ix)
		}
	}

	escrowVault, err := helpers.GetAssociatedTokenAddressSync(
		lpMint,
		lockEscrowKey,
		true,
		solana.TokenProgramID,
		solana.SPLAssociatedTokenAccountProgramID,
	)
	if err != nil {
		return nil, err
	}

	createEscrowVaultIx := helpers.CreateAssociatedTokenAccountIdempotentInstruction(
		param.Payer,
		escrowVault,
		lockEscrowKey,
		lpMint,
		solana.TokenProgramID,
	)

	preInstructions = append(preInstructions, createEscrowVaultIx)

	sourceTokens, err := helpers.GetAssociatedTokenAddressSync(
		lpMint,
		m.state.GetPoolAuthority(),
		true,
		solana.TokenProgramID,
		solana.SPLAssociatedTokenAccountProgramID,
	)

	if err != nil {
		return nil, err
	}
	owner := poolState.Creator
	if param.IsPartner {
		owner = poolConfigState.FeeClaimer
	}

	currentIx, err := dbc.NewMigrateMeteoraDammLockLpTokenInstruction(
		param.VirtualPool,
		helpers.DeriveDammV1MigrationMetadataAddress(param.VirtualPool),
		m.state.GetPoolAuthority(),
		dammPool,
		lpMint,
		lockEscrowKey,
		owner,
		sourceTokens,
		escrowVault,
		constants.DammV1ProgramId,
		vaultPDAsA.VaultPDA,
		vaultPDAsB.VaultPDA,
		aVaultLp,
		bVaultLp,
		aVaultLpMint,
		bVaultLpMint,
		solana.TokenProgramID,
	).ValidateAndBuild()
	if err != nil {
		return nil, err
	}

	ixns := make([]solana.Instruction, 0, 5)
	ixns = append(ixns, preInstructions...)
	return append(ixns, currentIx), nil
}

// ClaimDammV1LpToken claims DAMM V1 LP token for creator or partner.
func (m *MigrationService) ClaimDammV1LpToken(
	ctx context.Context,
	param types.DammLpTokenParam,
) ([]solana.Instruction, error) {

	virtualPoolState, err := m.state.GetPool(ctx, param.VirtualPool)
	if err != nil {
		return nil, fmt.Errorf("pool(%s) not found: err: %w", param.VirtualPool.String(), err)
	}

	poolConfigState, err := m.state.GetPoolConfig(ctx, virtualPoolState.Config)
	if err != nil {
		return nil, fmt.Errorf("pool config(%s) not found: err: %w", param.VirtualPool.String(), err)
	}

	dammPool := helpers.DeriveDammV1PoolAddress(
		param.DammConfig,
		virtualPoolState.BaseMint,
		poolConfigState.QuoteMint,
	)

	lpMint := helpers.DeriveDammV1LpMintAddress(dammPool)

	var destinationToken solana.PublicKey
	if param.IsPartner {
		destinationToken, err = helpers.FindAssociatedTokenAddress(
			poolConfigState.FeeClaimer,
			lpMint,
			solana.TokenProgramID,
		)
		if err != nil {
			return nil, err
		}

	} else {
		destinationToken, err = helpers.FindAssociatedTokenAddress(
			poolConfigState.FeeClaimer,
			lpMint,
			solana.TokenProgramID,
		)
		if err != nil {
			return nil, err
		}
	}
	owner := virtualPoolState.Creator
	if param.IsPartner {
		owner = poolConfigState.FeeClaimer
	}

	createDestinationTokenIx := helpers.CreateAssociatedTokenAccountIdempotentInstruction(
		param.Payer,
		destinationToken,
		owner,
		lpMint,
		solana.TokenProgramID,
	)

	sourceToken, err := helpers.GetAssociatedTokenAddressSync(
		lpMint,
		m.state.GetPoolAuthority(),
		true,
		solana.PublicKey{},
		solana.PublicKey{},
	)
	if err != nil {
		return nil, err
	}

	currentIx, err := dbc.NewMigrateMeteoraDammClaimLpTokenInstruction(
		param.VirtualPool,
		helpers.DeriveDammV1MigrationMetadataAddress(param.VirtualPool),
		m.state.GetPoolAuthority(),
		lpMint,
		sourceToken,
		destinationToken,
		owner,
		param.Payer,
		solana.TokenProgramID,
	).ValidateAndBuild()
	if err != nil {
		return nil, err
	}

	ixns := make([]solana.Instruction, 0, 2)
	return append(ixns, createDestinationTokenIx, currentIx), nil
}

///////////////////////
// DAMM V2 FUNCTIONS //
///////////////////////

// CreateDammV2MigrationMetadata creates metadata for the migration of Meteora DAMM V2.
func (m *MigrationService) CreateDammV2MigrationMetadata(
	param types.CreateDammV2MigrationMetadataParam,
) (*dbc.Instruction, error) {
	migrationDammV2CreateMetadataPtr := dbc.NewMigrationDammV2CreateMetadataInstruction(
		param.VirtualPool,
		param.Config,
		helpers.DeriveDammV2MigrationMetadataAddress(param.VirtualPool),
		param.Payer,
		solana.SystemProgramID,
		solana.PublicKey{},
		constants.DBCProgramId,
	)
	eventAuthPDA, _, err := migrationDammV2CreateMetadataPtr.FindEventAuthorityAddress()
	if err != nil {
		return nil, fmt.Errorf("err deriving eventAuthPDA: %w", err)
	}

	return migrationDammV2CreateMetadataPtr.
		SetEventAuthorityAccount(eventAuthPDA).ValidateAndBuild()
}

// MigrateToDammV2 migrates to DAMM V2.
func (m *MigrationService) MigrateToDammV2(
	ctx context.Context,
	param types.MigrateToDammV2Param,
) (types.MigrateToDammV2Response, error) {

	dammPoolAuthority := helpers.DeriveDammV2PoolAuthority()
	dammEventAuthority := helpers.DeriveDammV2EventAuthority()

	virtualPoolState, err := m.state.GetPool(ctx, param.VirtualPool)
	if err != nil {
		return types.MigrateToDammV2Response{}, fmt.Errorf("pool(%s) not found: err: %w", param.VirtualPool.String(), err)
	}

	poolConfigState, err := m.state.GetPoolConfig(ctx, virtualPoolState.Config)
	if err != nil {
		return types.MigrateToDammV2Response{}, fmt.Errorf("pool config(%s) not found: err: %w", param.VirtualPool.String(), err)
	}

	dammPool := helpers.DeriveDammV1PoolAddress(
		param.DammConfig,
		virtualPoolState.BaseMint,
		poolConfigState.QuoteMint,
	)

	firstPositionNftKP := solana.NewWallet()
	firstPosition := helpers.DerivePositionAddress(firstPositionNftKP.PublicKey())
	firstPositionNftAccount := helpers.DerivePositionNftAccount(firstPosition)

	secondPositionNftKP := solana.NewWallet()
	secondPosition := helpers.DerivePositionAddress(secondPositionNftKP.PublicKey())
	secondPositionNftAccount := helpers.DerivePositionNftAccount(secondPosition)

	tokenAVault := helpers.DeriveDammV2TokenVaultAddress(
		dammPool,
		virtualPoolState.BaseMint,
	)

	tokenBVault := helpers.DeriveDammV2TokenVaultAddress(
		dammPool,
		poolConfigState.QuoteMint,
	)

	tokenBaseProgram, tokenQuoteProgram := solana.Token2022ProgramID, solana.Token2022ProgramID

	if poolConfigState.TokenType == 0 {
		tokenBaseProgram = solana.TokenProgramID
	}

	if poolConfigState.QuoteTokenFlag == 0 {
		tokenQuoteProgram = solana.TokenProgramID
	}

	migrationDammV2Ptr := dbc.NewMigrationDammV2Instruction(
		param.VirtualPool,
		helpers.DeriveDammV2MigrationMetadataAddress(param.VirtualPool),
		virtualPoolState.Config,
		m.state.GetPoolAuthority(),
		dammPool,
		firstPositionNftKP.PublicKey(),
		firstPositionNftAccount,
		firstPosition,
		secondPositionNftKP.PublicKey(),
		secondPositionNftAccount,
		secondPosition,
		dammPoolAuthority,
		constants.DammV2ProgramId,
		virtualPoolState.BaseMint,
		poolConfigState.QuoteMint,
		tokenAVault,
		tokenBVault,
		virtualPoolState.BaseVault,
		virtualPoolState.QuoteVault,
		param.Payer,
		tokenBaseProgram,
		tokenQuoteProgram,
		solana.Token2022ProgramID,
		dammEventAuthority,
		solana.SystemProgramID,
	)

	migrationDammV2Ptr.AccountMetaSlice = append(
		migrationDammV2Ptr.AccountMetaSlice, &solana.AccountMeta{
			PublicKey: param.DammConfig,
		})

	currentIx, err := migrationDammV2Ptr.ValidateAndBuild()
	if err != nil {
		return types.MigrateToDammV2Response{}, err
	}

	return types.MigrateToDammV2Response{
		FirstPositionNftKeypair:  firstPositionNftKP.PrivateKey,
		SecondPositionNftKeypair: secondPositionNftKP.PrivateKey,
		Ixns: append([]solana.Instruction{}, currentIx,
			computebudget.NewSetComputeUnitLimitInstruction(500_000).Build()),
	}, nil
}
