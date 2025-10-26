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

func GetPoolConfig() {
	rpcClient := rpc.New("https://api.mainnet-beta.solana.com")

	configAddressStr := "YOUR_CONFIG_KEY"

	fmt.Println("Getting pool config...")
	configAddress := solana.MustPublicKeyFromBase58(configAddressStr)

	ctx := context.Background()

	poolConfig, err := instructions.GetPoolConfig(ctx, configAddress, rpcClient)
	if err != nil {
		log.Fatalf("Failed to get pool config: %v", err)
	}

	// Marshal the pool config to JSON
	jsonData, err := json.MarshalIndent(poolConfig, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal pool config to JSON: %v", err)
	}

	fmt.Printf("Pool config JSON: %s\n", string(jsonData))

	fmt.Printf("SqrtStartPrice: %s\n", poolConfig.SqrtStartPrice.String())

	fmt.Println("Curve points:")
	for i, point := range poolConfig.Curve {
		fmt.Printf("Curve[%d] Liquidity: %s\n", i, point.Liquidity.String())
		fmt.Printf("Curve[%d] SqrtPrice: %s\n", i, point.SqrtPrice.String())
	}

}

// func main() {
// 	GetPoolConfig()
// }
