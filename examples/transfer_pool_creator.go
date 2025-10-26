package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"github.com/MeteoraAg/dbc-go/helpers"
	"github.com/MeteoraAg/dbc-go/instructions"
)

func TransferPoolCreator() {
	ctx := context.Background()
	client := rpc.New("https://api.mainnet-beta.solana.com")

	// 1) load payer and creator PKs
	payer := solana.MustPrivateKeyFromBase58("YOUR_PAYER_PRIVATE_KEY")
	creator := solana.MustPrivateKeyFromBase58("YOUR_CREATOR_PRIVATE_KEY")
	newCreator := solana.MustPublicKeyFromBase58("NEW_CREATOR_PUBLIC_KEY")

	// 2) virtual pool address
	virtualPool := solana.MustPublicKeyFromBase58("YOUR_VIRTUAL_POOL_ADDRESS")

	// 3) get pool state to get config
	poolState, err := client.GetAccountInfo(ctx, virtualPool)
	if err != nil {
		log.Fatalf("GetAccountInfo: %v", err)
	}
	if poolState == nil || poolState.Value == nil {
		log.Fatalf("Pool not found")
	}

	// 4) derive PDAs
	migrationMetadata := helpers.DeriveDammV1MigrationMetadataPda(virtualPool)

	// 5) build transfer pool creator instruction
	ixTransfer := instructions.TransferPoolCreator(
		virtualPool,
		poolState.Value.Owner, // config
		creator.PublicKey(),
		newCreator,
		migrationMetadata,
	)

	// 6) assemble transaction
	bh, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		log.Fatalf("GetLatestBlockhash: %v", err)
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{ixTransfer},
		bh.Value.Blockhash,
		solana.TransactionPayer(payer.PublicKey()),
	)
	if err != nil {
		log.Fatalf("NewTransaction: %v", err)
	}

	// 7) sign with payer and creator
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		switch {
		case key.Equals(payer.PublicKey()):
			return &payer
		case key.Equals(creator.PublicKey()):
			return &creator
		default:
			return nil
		}
	})
	if err != nil {
		log.Fatalf("Sign: %v", err)
	}

	// 8) send & confirm
	sig, err := client.SendTransaction(ctx, tx)
	if err != nil {
		log.Fatalf("SendTransaction: %v", err)
	}
	fmt.Printf("Transaction sent: %s\n", sig)

	// wait for confirmation by polling
	for i := 0; i < 30; i++ { // try for 30 secs
		time.Sleep(time.Second)
		resp, err := client.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
			Commitment: rpc.CommitmentFinalized,
		})
		if err != nil {
			continue
		}
		if resp != nil {
			if resp.Meta != nil && resp.Meta.Err != nil {
				log.Fatalf("Transaction failed: %v", resp.Meta.Err)
			}
			fmt.Printf("Transaction confirmed: %s\n", `https://solscan.io/tx/`+sig.String())
			return
		}
	}
	log.Fatalf("Transaction confirmation timeout")
}

// func main() {
// 	TransferPoolCreator()
// }
