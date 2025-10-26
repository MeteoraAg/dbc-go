package helpers

import (
	"bytes"
	"log"

	"github.com/gagliardetto/solana-go"

	"github.com/MeteoraAg/dbc-go/common"
)

// Derives the dbc pool address
func DeriveDbcPoolPDA(quoteMint, baseMint, config solana.PublicKey) solana.PublicKey {
	// pda order: the larger public key bytes goes first
	var mintA, mintB solana.PublicKey
	if bytes.Compare(quoteMint.Bytes(), baseMint.Bytes()) > 0 {
		mintA = quoteMint
		mintB = baseMint
	} else {
		mintA = baseMint
		mintB = quoteMint
	}
	seeds := [][]byte{
		[]byte("pool"),
		config.Bytes(),
		mintA.Bytes(),
		mintB.Bytes(),
	}
	pda, _, err := solana.FindProgramAddress(seeds, solana.MustPublicKeyFromBase58(common.DbcProgramID))
	if err != nil {
		log.Fatalf("find pool PDA: %v", err)
	}
	return pda
}

// Derives the DAMM V1 pool address
func DeriveDammV1PoolPDA(config, tokenAMint, tokenBMint solana.PublicKey) solana.PublicKey {
	// Get the first and second keys based on byte comparison
	var firstKey, secondKey solana.PublicKey
	if bytes.Compare(tokenAMint.Bytes(), tokenBMint.Bytes()) > 0 {
		firstKey = tokenAMint
		secondKey = tokenBMint
	} else {
		firstKey = tokenBMint
		secondKey = tokenAMint
	}

	seeds := [][]byte{
		firstKey.Bytes(),
		secondKey.Bytes(),
		config.Bytes(),
	}
	pda, _, err := solana.FindProgramAddress(seeds, solana.MustPublicKeyFromBase58(common.DammV1ProgramID))
	if err != nil {
		log.Fatalf("find DAMM V1 pool PDA: %v", err)
	}
	return pda
}

// Derives the DAMM V2 pool address
func DeriveDammV2PoolPDA(config, tokenAMint, tokenBMint solana.PublicKey) solana.PublicKey {
	// Get the first and second keys based on byte comparison
	var firstKey, secondKey solana.PublicKey
	if bytes.Compare(tokenAMint.Bytes(), tokenBMint.Bytes()) > 0 {
		firstKey = tokenAMint
		secondKey = tokenBMint
	} else {
		firstKey = tokenBMint
		secondKey = tokenAMint
	}

	seeds := [][]byte{
		[]byte("pool"),
		config.Bytes(),
		firstKey.Bytes(),
		secondKey.Bytes(),
	}
	pda, _, err := solana.FindProgramAddress(seeds, solana.MustPublicKeyFromBase58(common.DammV2ProgramID))
	if err != nil {
		log.Fatalf("find DAMM V2 pool PDA: %v", err)
	}
	return pda
}

// Derives the dbc token vault address
func DeriveTokenVaultPDA(pool, mint solana.PublicKey) solana.PublicKey {
	seed := [][]byte{
		[]byte("token_vault"),
		mint.Bytes(),
		pool.Bytes(),
	}
	pda, _, err := solana.FindProgramAddress(seed, solana.MustPublicKeyFromBase58(common.DbcProgramID))
	if err != nil {
		log.Fatalf("find vault PDA: %v", err)
	}
	return pda
}

// Derives the event authority PDA
func DeriveEventAuthorityPDA() solana.PublicKey {
	seeds := [][]byte{[]byte("__event_authority")}
	address, _, err := solana.FindProgramAddress(seeds, solana.MustPublicKeyFromBase58(common.DbcProgramID))
	if err != nil {
		panic(err)
	}
	return address
}

// Derives the pool authority PDA
func DerivePoolAuthorityPDA() solana.PublicKey {
	seeds := [][]byte{[]byte("pool_authority")}
	address, _, err := solana.FindProgramAddress(seeds, solana.MustPublicKeyFromBase58(common.DbcProgramID))
	if err != nil {
		panic(err)
	}
	return address
}

// Derives the mint metadata address
func DeriveMintMetadataPDA(mint solana.PublicKey) solana.PublicKey {
	seeds := [][]byte{
		[]byte("metadata"),
		solana.MustPublicKeyFromBase58(common.MetadataProgram).Bytes(),
		mint.Bytes(),
	}
	pda, _, err := solana.FindProgramAddress(seeds, solana.MustPublicKeyFromBase58(common.MetadataProgram))
	if err != nil {
		log.Fatalf("find mint metadata PDA: %v", err)
	}
	return pda
}

// Derives the DAMM V1 migration metadata PDA
func DeriveDammV1MigrationMetadataPda(pool solana.PublicKey) solana.PublicKey {
	seeds := [][]byte{
		[]byte("meteora"),
		pool.Bytes(),
	}
	pda, _, err := solana.FindProgramAddress(seeds, solana.MustPublicKeyFromBase58(common.DbcProgramID))
	if err != nil {
		log.Fatalf("find DAMM V1 migration metadata PDA: %v", err)
	}
	return pda
}
