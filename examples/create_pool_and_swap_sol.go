package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	system "github.com/gagliardetto/solana-go/programs/system"
	token "github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"

	"github.com/MeteoraAg/dbc-go/common"
	"github.com/MeteoraAg/dbc-go/helpers"
	"github.com/MeteoraAg/dbc-go/instructions"
)

func CreatePoolAndSwapSol() {
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

	// 4) quote mint = wrapped SOL
	quoteMint := solana.MustPublicKeyFromBase58(common.NativeMint)

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

	// wrap and swap quote mint (0.01 SOL)
	amountIn := uint64(1e7)
	rentExemptAmount := uint64(2039280) // minimum rent-exempt balance for WSOL account
	totalAmount := amountIn + rentExemptAmount

	// create WSOL associated token account (ATA)
	createWSOLIx := associatedtokenaccount.NewCreateInstruction(
		payer.PublicKey(),
		poolCreator.PublicKey(),
		quoteMint,
	).Build()

	// wrap SOL by transferring lamports into the WSOL ATA
	wrapSOLIx := system.NewTransferInstruction(
		totalAmount,
		poolCreator.PublicKey(),
		userInputTokenAccount,
	).Build()

	// sync the WSOL account to update its balance
	syncNativeIx := token.NewSyncNativeInstruction(
		userInputTokenAccount,
	).Build()

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

	// close the WSOL account after swap to recover rent
	closeWSOLIx := token.NewCloseAccountInstruction(
		userInputTokenAccount,
		poolCreator.PublicKey(),
		poolCreator.PublicKey(),
		[]solana.PublicKey{},
	).Build()

	// 6) assemble transaction
	bh, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		log.Fatalf("GetLatestBlockhash: %v", err)
	}
	tx, err := solana.NewTransaction(
		[]solana.Instruction{
			ixInit,          // your pool init
			createWSOLIx,    // create WSOL ATA
			wrapSOLIx,       // wrap SOL
			syncNativeIx,    // sync WSOL balance
			createBaseAtaIx, // create output base mint ATA
			ixSwap,          // your swap
			closeWSOLIx,     // close WSOL ATA after swap
		},
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
// 	CreatePoolAndSwapSol()
// }
