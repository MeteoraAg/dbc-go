package instructions

import (
	"encoding/binary"

	"github.com/gagliardetto/solana-go"

	"github.com/MeteoraAg/dbc-go/common"
	"github.com/MeteoraAg/dbc-go/helpers"
)

func ClaimCreatorTradingFee(
	pool solana.PublicKey,
	tokenAAccount solana.PublicKey,
	tokenBAccount solana.PublicKey,
	baseVault solana.PublicKey,
	quoteVault solana.PublicKey,
	baseMint solana.PublicKey,
	quoteMint solana.PublicKey,
	creator solana.PublicKey,
	maxBaseAmount uint64,
	maxQuoteAmount uint64,
) solana.Instruction {
	disc := []byte{82, 220, 250, 189, 3, 85, 107, 45}
	buf := make([]byte, 8+8+8)
	copy(buf, disc)
	binary.LittleEndian.PutUint64(buf[8:], maxBaseAmount)
	binary.LittleEndian.PutUint64(buf[16:], maxQuoteAmount)

	tokenBaseProgram := solana.MustPublicKeyFromBase58(common.TokenProgram)
	tokenQuoteProgram := solana.MustPublicKeyFromBase58(common.TokenProgram)
	poolAuthority := helpers.DerivePoolAuthorityPDA()
	eventAuthority := helpers.DeriveEventAuthorityPDA()

	acctMeta := solana.AccountMetaSlice{
		// 1. pool_authority
		{PublicKey: poolAuthority, IsSigner: false, IsWritable: false},
		// 2. pool (writable)
		{PublicKey: pool, IsSigner: false, IsWritable: true},
		// 3. token_a_account (writable)
		{PublicKey: tokenAAccount, IsSigner: false, IsWritable: true},
		// 4. token_b_account (writable)
		{PublicKey: tokenBAccount, IsSigner: false, IsWritable: true},
		// 5. base_vault (writable)
		{PublicKey: baseVault, IsSigner: false, IsWritable: true},
		// 6. quote_vault (writable)
		{PublicKey: quoteVault, IsSigner: false, IsWritable: true},
		// 7. base_mint
		{PublicKey: baseMint, IsSigner: false, IsWritable: false},
		// 8. quote_mint
		{PublicKey: quoteMint, IsSigner: false, IsWritable: false},
		// 9. creator (signer)
		{PublicKey: creator, IsSigner: true, IsWritable: false},
		// 10. token_base_program
		{PublicKey: tokenBaseProgram, IsSigner: false, IsWritable: false},
		// 11. token_quote_program
		{PublicKey: tokenQuoteProgram, IsSigner: false, IsWritable: false},
		// 12. event_authority
		{PublicKey: eventAuthority, IsSigner: false, IsWritable: false},
		// 13. program
		{PublicKey: solana.MustPublicKeyFromBase58(common.DbcProgramID), IsSigner: false, IsWritable: false},
	}

	return solana.NewInstruction(
		solana.MustPublicKeyFromBase58(common.DbcProgramID),
		acctMeta,
		buf,
	)
}

func TransferPoolCreator(
	virtualPool solana.PublicKey,
	config solana.PublicKey,
	creator solana.PublicKey,
	newCreator solana.PublicKey,
	migrationMetadata solana.PublicKey,
) solana.Instruction {
	disc := []byte{20, 7, 169, 33, 58, 147, 166, 33}
	eventAuthority := helpers.DeriveEventAuthorityPDA()

	acctMeta := solana.AccountMetaSlice{
		// 1. virtual_pool (writable)
		{PublicKey: virtualPool, IsSigner: false, IsWritable: true},
		// 2. config
		{PublicKey: config, IsSigner: false, IsWritable: false},
		// 3. creator (signer)
		{PublicKey: creator, IsSigner: true, IsWritable: false},
		// 4. new_creator
		{PublicKey: newCreator, IsSigner: false, IsWritable: false},
		// 5. event_authority
		{PublicKey: eventAuthority, IsSigner: false, IsWritable: false},
		// 6. program
		{PublicKey: solana.MustPublicKeyFromBase58(common.DbcProgramID), IsSigner: false, IsWritable: false},
	}

	acctMeta = append(acctMeta, &solana.AccountMeta{
		PublicKey:  migrationMetadata,
		IsSigner:   false,
		IsWritable: false,
	})

	return solana.NewInstruction(
		solana.MustPublicKeyFromBase58(common.DbcProgramID),
		acctMeta,
		disc,
	)
}
