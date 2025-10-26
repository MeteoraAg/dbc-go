package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/rpc"

	"github.com/MeteoraAg/dbc-go/helpers"
	"github.com/MeteoraAg/dbc-go/instructions"
)

func CreatePoolAndSwapUsdc() {
	ctx := context.Background()
	client := rpc.New("https://api.mainnet-beta.solana.com")

	// 1) load payer and pool creator PKs
	payer := solana.MustPrivateKeyFromBase58("YOUR_PAYER_PRIVATE_KEY")
	poolCreator := solana.MustPrivateKeyFromBase58("YOUR_POOL_CREATOR_PRIVATE_KEY")

	// 2) config key (generate on launch.meteora.ag)
	config := solana.MustPublicKeyFromBase58("YOUR_CONFIG_PUBLIC_KEY")

	// 3) generate baseMint (can be vanity)
	baseMintWallet := solana.NewWallet()
	baseMint := baseMintWallet.PublicKey()
	fmt.Println("Base mint:", baseMint)

	// 4) quote mint = USDC
	quoteMint := solana.MustPublicKeyFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")

	userInputTokenAccount, _, _ := solana.FindAssociatedTokenAddress(
		poolCreator.PublicKey(),
		quoteMint,
	)
	userOutputTokenAccount, _, _ := solana.FindAssociatedTokenAddress(
		poolCreator.PublicKey(),
		baseMint,
	)

	// 5) derive PDAs
	pool := helpers.DeriveDbcPoolPDA(quoteMint, baseMint, config)
	baseVault := helpers.DeriveTokenVaultPDA(pool, baseMint)
	quoteVault := helpers.DeriveTokenVaultPDA(pool, quoteMint)
	mintMetadata := helpers.DeriveMintMetadataPDA(baseMint)

	// build initialize virtual pool instruction
	ixInit := instructions.InitializeVirtualPoolWithSplToken(
		config,
		poolCreator.PublicKey(),
		baseMint,
		quoteMint,
		pool,
		baseVault,
		quoteVault,
		mintMetadata,
		payer.PublicKey(),
		"test",
		"TEST",
		"https://test.fun",
	)

	// amount in USDC (e.g., 1 USDC = 1_000_000)
	amountIn := uint64(1_000_000) // 1 USDC

	// initialize with a nil instruction
	var createUSDCAtaIx solana.Instruction

	// check if USDC ATA exists and create if needed
	accountInfo, err := client.GetAccountInfo(ctx, userInputTokenAccount)
	if err != nil || accountInfo == nil || accountInfo.Value == nil {
		createUSDCAtaIx = associatedtokenaccount.NewCreateInstruction(
			payer.PublicKey(),
			poolCreator.PublicKey(),
			quoteMint,
		).Build()
	}

	// create base-mint ATA for swap output
	createBaseAtaIx := associatedtokenaccount.NewCreateInstruction(
		payer.PublicKey(),
		poolCreator.PublicKey(),
		baseMint,
	).Build()

	// build swap instruction
	ixSwap := instructions.Swap(
		config,
		pool,
		userInputTokenAccount,
		userOutputTokenAccount,
		baseVault,
		quoteVault,
		baseMint,
		quoteMint,
		poolCreator.PublicKey(),
		userInputTokenAccount, // using input token account as referral
		amountIn,
		1, // minOut
	)

	// 6) assemble transaction
	bh, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		log.Fatalf("GetLatestBlockhash: %v", err)
	}

	// create instructions slice
	instructions := []solana.Instruction{
		ixInit,          // your pool init
		createBaseAtaIx, // create output base mint ATA
		ixSwap,          // your swap
	}

	// add USDC ATA creation instruction only if needed
	if createUSDCAtaIx != nil {
		instructions = append([]solana.Instruction{createUSDCAtaIx}, instructions...)
	}

	tx, err := solana.NewTransaction(
		instructions,
		bh.Value.Blockhash,
		solana.TransactionPayer(payer.PublicKey()),
	)
	if err != nil {
		log.Fatalf("NewTransaction: %v", err)
	}
	// 7) sign with payer, poolCreator, baseMint
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		switch {
		case key.Equals(payer.PublicKey()):
			return &payer
		case key.Equals(poolCreator.PublicKey()):
			return &poolCreator
		case key.Equals(baseMint):
			return &baseMintWallet.PrivateKey
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
// 	CreatePoolAndSwapUsdc()
// }
