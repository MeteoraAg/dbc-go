package services

import (
	"context"
	"dbcGoSDK/anchor"
	"dbcGoSDK/generated/dbc"
	"dbcGoSDK/helpers"
	"dbcGoSDK/types"
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type StateService struct {
	*DBCProgram
}

func NewStateService(
	conn *rpc.Client,
	commitment rpc.CommitmentType,
) *StateService {
	return &StateService{
		DBCProgram: NewDBCProgram(
			conn,
			commitment,
		),
	}
}

// GetPoolConfig get pool config data (partner config).
func (s *StateService) GetPoolConfig(
	ctx context.Context,
	configAddress solana.PublicKey,
) (*dbc.PoolConfigAccount, error) {
	return anchor.NewPgAccounts(
		s.conn,
		func() *dbc.PoolConfigAccount { return &dbc.PoolConfigAccount{} },
	).Fetch(ctx, configAddress, &rpc.GetAccountInfoOpts{Commitment: s.commitment,})
}

// GetPoolConfigs all config keys.
func (s *StateService) GetPoolConfigs(
	ctx context.Context,
) ([]anchor.ProgramAccount[*dbc.PoolConfigAccount], error) {
	return anchor.NewPgAccounts(
		s.conn,
		func() *dbc.PoolConfigAccount { return &dbc.PoolConfigAccount{} },
	).All(
		ctx,
		s.GetProgramID(),
		dbc.PoolConfigAccountDiscriminator,
		rpc.GetProgramAccountsOpts{},
		nil,
	)
}

// GetPoolConfigsByOwner all config keys of an owner wallet address.
func (s *StateService) GetPoolConfigsByOwner(
	ctx context.Context,
	owner solana.PublicKey,
) ([]anchor.ProgramAccount[*dbc.PoolConfigAccount], error) {
	return anchor.NewPgAccounts(
		s.conn,
		func() *dbc.PoolConfigAccount { return &dbc.PoolConfigAccount{} },
	).All(
		ctx,
		s.GetProgramID(),
		dbc.PoolConfigAccountDiscriminator,
		helpers.CreateProgramAccountFilter(owner, 72),
		nil,
	)
}

// GetPool gvirtual pool data.
func (s *StateService) GetPool(
	ctx context.Context,
	poolAddress solana.PublicKey,
) (*dbc.VirtualPoolAccount, error) {
	return anchor.NewPgAccounts(
		s.conn,
		func() *dbc.VirtualPoolAccount { return &dbc.VirtualPoolAccount{} },
	).Fetch(ctx, poolAddress, &rpc.GetAccountInfoOpts{Commitment: s.commitment})
}

// GetPools get all dynamic bonding curve pools.
func (s *StateService) GetPools(
	ctx context.Context,
) ([]anchor.ProgramAccount[*dbc.VirtualPoolAccount], error) {
	return anchor.NewPgAccounts(
		s.conn,
		func() *dbc.VirtualPoolAccount { return &dbc.VirtualPoolAccount{} },
	).All(
		ctx,
		s.GetProgramID(),
		dbc.VirtualPoolAccountDiscriminator,
		rpc.GetProgramAccountsOpts{},
		nil,
	)
}

// GetPoolsByConfig all dynamic bonding curve pools by config key address.
func (s *StateService) GetPoolsByConfig(
	ctx context.Context,
	configAddress solana.PublicKey,
) ([]anchor.ProgramAccount[*dbc.VirtualPoolAccount], error) {
	return anchor.NewPgAccounts(
		s.conn,
		func() *dbc.VirtualPoolAccount { return &dbc.VirtualPoolAccount{} },
	).All(
		ctx,
		s.GetProgramID(),
		dbc.VirtualPoolAccountDiscriminator,
		helpers.CreateProgramAccountFilter(configAddress, 72),
		nil,
	)
}

// GetPoolsByCreator all dynamic bonding curve pools by creator address.
func (s *StateService) GetPoolsByCreator(
	ctx context.Context,
	creatorAddress solana.PublicKey,
) ([]anchor.ProgramAccount[*dbc.VirtualPoolAccount], error) {
	return anchor.NewPgAccounts(
		s.conn,
		func() *dbc.VirtualPoolAccount { return &dbc.VirtualPoolAccount{} },
	).All(
		ctx,
		s.GetProgramID(),
		dbc.VirtualPoolAccountDiscriminator,
		helpers.CreateProgramAccountFilter(creatorAddress, 104),
		nil,
	)
}

// GetPoolByBaseMint by a base mint.
func (s *StateService) GetPoolByBaseMint(
	ctx context.Context,
	baseMint solana.PublicKey,
) (anchor.ProgramAccount[*dbc.PoolConfigAccount], error) {
	pools, err := anchor.NewPgAccounts(
		s.conn,
		func() *dbc.PoolConfigAccount { return &dbc.PoolConfigAccount{} },
	).All(
		ctx,
		s.GetProgramID(),
		dbc.VirtualPoolAccountDiscriminator,
		helpers.CreateProgramAccountFilter(baseMint, 136),
		nil,
	)
	if err != nil {
		return anchor.ProgramAccount[*dbc.PoolConfigAccount]{}, err
	}
	if len(pools) == 0 {
		return anchor.ProgramAccount[*dbc.PoolConfigAccount]{}, errors.New("len of pool as zero")
	}

	return pools[0], nil
}

// GetPoolMigrationQuoteThreshold get pool migration quote threshold.
func (s *StateService) GetPoolMigrationQuoteThreshold(
	ctx context.Context,
	poolAddress solana.PublicKey,
) (uint64, error) {
	pool, err := s.GetPool(ctx, poolAddress)
	if err != nil {
		return 0, fmt.Errorf("pool not found: error: %w", err)
	}

	config, err := s.GetPoolConfig(ctx, pool.Config)
	if err != nil {
		return 0, err
	}

	return config.MigrationQuoteThreshold, nil
}

// GetPoolCurveProgress get the progress of the curve by comparing current quote reserve to migration threshold.
func (s *StateService) GetPoolCurveProgress(
	ctx context.Context,
	poolAddress solana.PublicKey,
) (float64, error) {
	pool, err := s.GetPool(ctx, poolAddress)
	if err != nil {
		return 0, fmt.Errorf("pool not found: error: %w", err)
	}

	config, err := s.GetPoolConfig(ctx, pool.Config)
	if err != nil {
		return 0, err
	}

	quoteReserve, migrationThreshold := new(big.Float).SetUint64(pool.QuoteReserve),
		new(big.Float).SetUint64(config.MigrationQuoteThreshold)

	if migrationThreshold.Sign() == 0 {
		return 0, errors.New("migration threshold is zero")
	}
	progress := new(big.Float).Quo(quoteReserve, migrationThreshold)

	f64, _ := progress.Float64()
	// u64, _ := progress.Uint64()
	return math.Min(math.Max(f64, 0), 1), nil
}

// GetPoolMetadata get pool metadata.
func (s *StateService) GetPoolMetadata(
	ctx context.Context,
	poolAddress solana.PublicKey,
) ([]*dbc.VirtualPoolMetadataAccount, error) {
	pgAAccs, err := anchor.NewPgAccounts(
		s.conn,
		func() *dbc.VirtualPoolMetadataAccount { return &dbc.VirtualPoolMetadataAccount{} },
	).All(
		ctx,
		s.GetProgramID(),
		dbc.VirtualPoolMetadataAccountDiscriminator,
		helpers.CreateProgramAccountFilter(poolAddress, 8),
		nil,
	)
	if err != nil {
		return nil, err
	}

	accs := make([]*dbc.VirtualPoolMetadataAccount, 0, len(pgAAccs))
	for _, v := range pgAAccs {
		accs = append(accs, v.Account)
	}

	return accs, nil
}

// GetPartnerMetadata get partner metadata.
func (s *StateService) GetPartnerMetadata(
	ctx context.Context,
	walletAddress solana.PublicKey,
) ([]*dbc.PartnerMetadataAccount, error) {
	pgAAccs, err := anchor.NewPgAccounts(
		s.conn,
		func() *dbc.PartnerMetadataAccount { return &dbc.PartnerMetadataAccount{} },
	).All(
		ctx,
		s.GetProgramID(),
		dbc.PartnerMetadataAccountDiscriminator,
		helpers.CreateProgramAccountFilter(walletAddress, 8),
		nil,
	)
	if err != nil {
		return nil, err
	}

	accs := make([]*dbc.PartnerMetadataAccount, 0, len(pgAAccs))
	for _, v := range pgAAccs {
		accs = append(accs, v.Account)
	}

	return accs, nil
}

// GetDammV1LockEscrow get DAMM V1 lock escrow details.
func (s *StateService) GetDammV1LockEscrow(
	ctx context.Context,
	lockEscrowAddress solana.PublicKey,
) (*dbc.LockEscrowAccount, error) {
	return anchor.NewPgAccounts(
		s.conn,
		func() *dbc.LockEscrowAccount { return &dbc.LockEscrowAccount{} },
	).Fetch(ctx, lockEscrowAddress, &rpc.GetAccountInfoOpts{})
}

// GetPoolFeeMetrics get fee metrics for a specific pool.
func (s *StateService) GetPoolFeeMetrics(
	ctx context.Context,
	poolAddress solana.PublicKey,
) (types.PoolFeeMetrics, error) {

	pool, err := s.GetPool(ctx, poolAddress)
	if err != nil {
		return types.PoolFeeMetrics{}, fmt.Errorf("pool not found: error: %w", err)
	}

	return types.PoolFeeMetrics{
		Current: types.FeeBreakdown{
			PartnerBaseFee:  pool.PartnerBaseFee,
			PartnerQuoteFee: pool.PartnerQuoteFee,
			CreatorBaseFee:  pool.CreatorBaseFee,
			CreatorQuoteFee: pool.CreatorQuoteFee,
		},
		Total: types.FeeTotal{
			TotalTradingBaseFee:  pool.Metrics.TotalTradingBaseFee,
			TotalTradingQuoteFee: pool.Metrics.TotalTradingQuoteFee,
		},
	}, nil
}

// GetPoolsFeesByConfig get all fees for pools linked to a specific config key.
func (s *StateService) GetPoolsFeesByConfig(
	ctx context.Context,
	configAddress solana.PublicKey,
) ([]types.PoolFeeByConfigOrCreator, error) {
	filteredPools, err := s.GetPoolsByConfig(ctx, configAddress)
	if err != nil {
		return nil, err
	}

	res := make([]types.PoolFeeByConfigOrCreator, 0, len(filteredPools))
	for _, v := range filteredPools {
		res = append(res, types.PoolFeeByConfigOrCreator{
			PoolAddress:          v.PublicKey,
			PartnerBaseFee:       v.Account.PartnerBaseFee,
			PartnerQuoteFee:      v.Account.PartnerQuoteFee,
			CreatorBaseFee:       v.Account.CreatorBaseFee,
			CreatorQuoteFee:      v.Account.CreatorQuoteFee,
			TotalTradingBaseFee:  v.Account.Metrics.TotalTradingBaseFee,
			TotalTradingQuoteFee: v.Account.Metrics.TotalTradingQuoteFee,
		})
	}

	return res, nil
}

// GetPoolsFeesByCreator get all fees for pools linked to a specific creator.
func (s *StateService) GetPoolsFeesByCreator(
	ctx context.Context,
	creatorAddress solana.PublicKey,
) ([]types.PoolFeeByConfigOrCreator, error) {
	filteredPools, err := s.GetPoolsByCreator(ctx, creatorAddress)
	if err != nil {
		return nil, err
	}

	res := make([]types.PoolFeeByConfigOrCreator, 0, len(filteredPools))
	for _, v := range filteredPools {
		res = append(res, types.PoolFeeByConfigOrCreator{
			PoolAddress:          v.PublicKey,
			PartnerBaseFee:       v.Account.PartnerBaseFee,
			PartnerQuoteFee:      v.Account.PartnerQuoteFee,
			CreatorBaseFee:       v.Account.CreatorBaseFee,
			CreatorQuoteFee:      v.Account.CreatorQuoteFee,
			TotalTradingBaseFee:  v.Account.Metrics.TotalTradingBaseFee,
			TotalTradingQuoteFee: v.Account.Metrics.TotalTradingQuoteFee,
		})
	}

	return res, nil
}

// GetDammV1MigrationMetadata gets DAMM V1 migration metadata.
func (s *StateService) GetDammV1MigrationMetadata(
	ctx context.Context,
	poolAdress solana.PublicKey,
) (*dbc.MeteoraDammMigrationMetadataAccount, error) {
	migrationMetadataAddress := helpers.DeriveDammV1MigrationMetadataAddress(poolAdress)

	return anchor.NewPgAccounts(
		s.conn, func() *dbc.MeteoraDammMigrationMetadataAccount { return &dbc.MeteoraDammMigrationMetadataAccount{} },
	).Fetch(ctx, migrationMetadataAddress, &rpc.GetAccountInfoOpts{})

}
