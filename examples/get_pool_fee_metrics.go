package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"github.com/MeteoraAg/dbc-go/instructions"
)

func GetPoolFeeMetrics() {
	rpcClient := rpc.New("https://api.mainnet-beta.solana.com")

	poolAddressStr := "YOUR_POOL_ADDRESS"

	fmt.Println("Getting pool fee metrics...")
	poolAddress := solana.MustPublicKeyFromBase58(poolAddressStr)

	ctx := context.Background()

	metrics, err := instructions.GetPoolFeeMetrics(ctx, poolAddress, rpcClient)
	if err != nil {
		log.Fatalf("Failed to get pool fee metrics: %v", err)
	}

	// Marshal the metrics to JSON
	jsonData, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal metrics to JSON: %v", err)
	}

	fmt.Printf("Pool Fee Metrics JSON: %s\n", string(jsonData))
}

// func main() {
// 	GetPoolFeeMetrics()
// }
