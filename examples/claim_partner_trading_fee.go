package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	"github.com/gagliardetto/solana-go/rpc"

	"github.com/MeteoraAg/dbc-go/common"
	"github.com/MeteoraAg/dbc-go/helpers"
	"github.com/MeteoraAg/dbc-go/instructions"
)

func ClaimPartnerTradingFee() {
	ctx := context.Background()
	client := rpc.New("https://api.mainnet-beta.solana.com")

	// 1) load payer and fee claimer PKs
	payer := solana.MustPrivateKeyFromBase58("YOUR_PAYER_PRIVATE_KEY")
	feeClaimer := solana.MustPrivateKeyFromBase58("YOUR_FEE_CLAIMER_PRIVATE_KEY")

	// 2) config and pool addresses
	config := solana.MustPublicKeyFromBase58("YOUR_CONFIG_KEY")
	pool := solana.MustPublicKeyFromBase58("YOUR_POOL_ADDRESS")

	baseMint := solana.MustPublicKeyFromBase58("YOUR_BASE_MINT")
	quoteMint := solana.MustPublicKeyFromBase58(common.NativeMint) // SOL (switch to USDC if needed)

	// 3) derive PDAs
	baseVault := helpers.DeriveTokenVaultPDA(pool, baseMint)
	quoteVault := helpers.DeriveTokenVaultPDA(pool, quoteMint)

	// 4) get token accounts
	tokenAAccount, _, _ := solana.FindAssociatedTokenAddress(
		feeClaimer.PublicKey(),
		baseMint,
	)
	tokenBAccount, _, _ := solana.FindAssociatedTokenAddress(
		feeClaimer.PublicKey(),
		quoteMint,
	)

	// 5) create ATAs if they don't exist
	var createTokenAAtaIx, createTokenBAtaIx solana.Instruction

	// Check if token A ATA exists
	accountInfo, err := client.GetAccountInfo(ctx, tokenAAccount)
	if err != nil || accountInfo == nil || accountInfo.Value == nil {
		createTokenAAtaIx = associatedtokenaccount.NewCreateInstruction(
			payer.PublicKey(),
			feeClaimer.PublicKey(),
			baseMint,
		).Build()
	}

	// check if token B ATA exists
	accountInfo, err = client.GetAccountInfo(ctx, tokenBAccount)
	if err != nil || accountInfo == nil || accountInfo.Value == nil {
		createTokenBAtaIx = associatedtokenaccount.NewCreateInstruction(
			payer.PublicKey(),
			feeClaimer.PublicKey(),
			quoteMint,
		).Build()
	}

	// 6) build claim partner trading fee instruction
	ixClaim := instructions.ClaimPartnerTradingFee(
		config,
		pool,
		tokenAAccount,
		tokenBAccount,
		baseVault,
		quoteVault,
		baseMint,
		quoteMint,
		feeClaimer.PublicKey(),
		10000, // maxAmountA
		10000, // maxAmountB
	)

	// 7) assemble transaction
	bh, err := client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		log.Fatalf("GetLatestBlockhash: %v", err)
	}

	// create instructions slice
	instructions := []solana.Instruction{ixClaim}

	// add ATA creation instructions only if needed
	if createTokenAAtaIx != nil {
		instructions = append([]solana.Instruction{createTokenAAtaIx}, instructions...)
	}
	if createTokenBAtaIx != nil {
		instructions = append([]solana.Instruction{createTokenBAtaIx}, instructions...)
	}

	tx, err := solana.NewTransaction(
		instructions,
		bh.Value.Blockhash,
		solana.TransactionPayer(payer.PublicKey()),
	)
	if err != nil {
		log.Fatalf("NewTransaction: %v", err)
	}

	// 8) sign with payer and fee claimer
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		switch {
		case key.Equals(payer.PublicKey()):
			return &payer
		case key.Equals(feeClaimer.PublicKey()):
			return &feeClaimer
		default:
			return nil
		}
	})
	if err != nil {
		log.Fatalf("Sign: %v", err)
	}

	// 9) send & confirm
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
// 	ClaimPartnerTradingFee()
// }
