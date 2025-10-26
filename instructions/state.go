package instructions

import (
	"bytes"
	"context"
	"fmt"

	"github.com/gagliardetto/solana-go"
	solRpc "github.com/gagliardetto/solana-go/rpc"

	"github.com/MeteoraAg/dbc-go/common"
	"github.com/MeteoraAg/dbc-go/helpers"
)

func GetPoolConfig(ctx context.Context, configAddress solana.PublicKey, rpcClient *solRpc.Client) (*common.PoolConfig, error) {
	account, err := rpcClient.GetAccountInfo(ctx, configAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool config account: %w", err)
	}

	if account == nil || account.Value == nil {
		return nil, fmt.Errorf("pool config account not found")
	}

	data := account.Value.Data.GetBinary()

	if len(data) < 8 {
		return nil, fmt.Errorf("data too short")
	}

	expectedDiscriminator := []byte{26, 108, 14, 123, 116, 230, 129, 43}
	if !bytes.Equal(data[:8], expectedDiscriminator) {
		return nil, fmt.Errorf("invalid discriminator, not a pool config account")
	}

	return helpers.DeserializePoolConfig(data)
}

func GetPoolFeeMetrics(ctx context.Context, poolAddress solana.PublicKey, rpcClient *solRpc.Client) (*common.PoolFeeMetrics, error) {
	pool, err := GetPool(ctx, poolAddress, rpcClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool: %w", err)
	}

	if pool == nil {
		return nil, fmt.Errorf("pool not found: %s", poolAddress.String())
	}

	metrics := &common.PoolFeeMetrics{}
	metrics.Current.PartnerBaseFee = pool.PartnerBaseFee
	metrics.Current.PartnerQuoteFee = pool.PartnerQuoteFee
	metrics.Current.CreatorBaseFee = pool.CreatorBaseFee
	metrics.Current.CreatorQuoteFee = pool.CreatorQuoteFee
	metrics.Total.TotalTradingBaseFee = pool.Metrics.TotalTradingBaseFee
	metrics.Total.TotalTradingQuoteFee = pool.Metrics.TotalTradingQuoteFee

	return metrics, nil
}

func GetPool(ctx context.Context, poolAddress solana.PublicKey, rpcClient *solRpc.Client) (*common.Pool, error) {
	account, err := rpcClient.GetAccountInfo(ctx, poolAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get pool account: %w", err)
	}

	if account == nil || account.Value == nil {
		return nil, fmt.Errorf("pool account not found")
	}

	data := account.Value.Data.GetBinary()

	if len(data) < 8 {
		return nil, fmt.Errorf("data too short")
	}

	expectedDiscriminator := []byte{213, 224, 5, 209, 98, 69, 119, 92}
	if !bytes.Equal(data[:8], expectedDiscriminator) {
		return nil, fmt.Errorf("invalid discriminator, not a pool account")
	}

	return helpers.DeserializePool(data)
}
