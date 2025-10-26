package instructions

import (
	"encoding/binary"

	"github.com/gagliardetto/solana-go"

	"github.com/MeteoraAg/dbc-go/common"
	"github.com/MeteoraAg/dbc-go/helpers"
)

func InitializeVirtualPoolWithSplToken(
	config solana.PublicKey,
	poolCreator solana.PublicKey,
	baseMint solana.PublicKey,
	quoteMint solana.PublicKey,
	pool solana.PublicKey,
	baseVault solana.PublicKey,
	quoteVault solana.PublicKey,
	mintMetadata solana.PublicKey,
	payer solana.PublicKey,
	name string,
	symbol string,
	uri string,
) solana.Instruction {
	disc := []byte{140, 85, 215, 176, 102, 54, 104, 79}

	packString := func(s string) []byte {
		b := make([]byte, 4+len(s))
		binary.LittleEndian.PutUint32(b[:4], uint32(len(s)))
		copy(b[4:], []byte(s))
		return b
	}
	data := append(append(append(disc, packString(name)...), packString(symbol)...), packString(uri)...)

	tokenQuoteProgram := solana.MustPublicKeyFromBase58(common.TokenProgram)
	tokenProgram := solana.MustPublicKeyFromBase58(common.TokenProgram)
	poolAuthority := helpers.DerivePoolAuthorityPDA()
	eventAuthority := helpers.DeriveEventAuthorityPDA()

	acctMeta := solana.AccountMetaSlice{
		// 1. config
		{PublicKey: config, IsSigner: false, IsWritable: false},
		// 2. pool_authority
		{PublicKey: poolAuthority, IsSigner: false, IsWritable: false},
		// 3. creator (signer)
		{PublicKey: poolCreator, IsSigner: true, IsWritable: false},
		// 4. base_mint (signer, writable)
		{PublicKey: baseMint, IsSigner: true, IsWritable: true},
		// 5. quote_mint
		{PublicKey: quoteMint, IsSigner: false, IsWritable: false},
		// 6. pool (writable)
		{PublicKey: pool, IsSigner: false, IsWritable: true},
		// 7. base_vault (writable)
		{PublicKey: baseVault, IsSigner: false, IsWritable: true},
		// 8. quote_vault (writable)
		{PublicKey: quoteVault, IsSigner: false, IsWritable: true},
		// 9. mint_metadata (writable)
		{PublicKey: mintMetadata, IsSigner: false, IsWritable: true},
		// 10. metadata_program
		{PublicKey: solana.MustPublicKeyFromBase58(common.MetadataProgram), IsSigner: false, IsWritable: false},
		// 11. payer (signer, writable)
		{PublicKey: payer, IsSigner: true, IsWritable: true},
		// 12. token_quote_program
		{PublicKey: tokenQuoteProgram, IsSigner: false, IsWritable: false},
		// 13. token_program
		{PublicKey: tokenProgram, IsSigner: false, IsWritable: false},
		// 14. system_program
		{PublicKey: solana.SystemProgramID, IsSigner: false, IsWritable: false},
		// 15. event_authority (PDA, same as pool_authority)
		{PublicKey: eventAuthority, IsSigner: false, IsWritable: false},
		// 16. program (ProgramID)
		{PublicKey: solana.MustPublicKeyFromBase58(common.DbcProgramID), IsSigner: false, IsWritable: false},
	}

	return solana.NewInstruction(
		solana.MustPublicKeyFromBase58(common.DbcProgramID),
		acctMeta,
		data,
	)
}

func Swap(
	config solana.PublicKey,
	pool solana.PublicKey,
	userInputTokenAccount solana.PublicKey,
	userOutputTokenAccount solana.PublicKey,
	baseVault solana.PublicKey,
	quoteVault solana.PublicKey,
	baseMint solana.PublicKey,
	quoteMint solana.PublicKey,
	payer solana.PublicKey,
	referralTokenAccount solana.PublicKey,
	amountIn uint64,
	minOut uint64,
) solana.Instruction {
	swapDisc := []byte{248, 198, 158, 145, 225, 117, 135, 200}
	buf := make([]byte, 8+8+8)
	copy(buf, swapDisc)
	binary.LittleEndian.PutUint64(buf[8:], amountIn)
	binary.LittleEndian.PutUint64(buf[16:], minOut)

	poolAuthority := helpers.DerivePoolAuthorityPDA()
	tokenBaseProgram := solana.MustPublicKeyFromBase58(common.TokenProgram)
	tokenQuoteProgram := solana.MustPublicKeyFromBase58(common.TokenProgram)
	eventAuthority := helpers.DeriveEventAuthorityPDA()

	acctMetaSwap := solana.AccountMetaSlice{
		// 1. pool_authority
		{PublicKey: poolAuthority, IsSigner: false, IsWritable: false},
		// 2. config
		{PublicKey: config, IsSigner: false, IsWritable: false},
		// 3. pool
		{PublicKey: pool, IsSigner: false, IsWritable: true},
		// 4. input_token_account (user's token account for input token)
		{PublicKey: userInputTokenAccount, IsSigner: false, IsWritable: true},
		// 5. output_token_account (user's token account for output token)
		{PublicKey: userOutputTokenAccount, IsSigner: false, IsWritable: true},
		// 6. base_vault
		{PublicKey: baseVault, IsSigner: false, IsWritable: true},
		// 7. quote_vault
		{PublicKey: quoteVault, IsSigner: false, IsWritable: true},
		// 8. base_mint
		{PublicKey: baseMint, IsSigner: false, IsWritable: false},
		// 9. quote_mint
		{PublicKey: quoteMint, IsSigner: false, IsWritable: false},
		// 10. payer
		{PublicKey: payer, IsSigner: true, IsWritable: true},
		// 11. token_base_program
		{PublicKey: tokenBaseProgram, IsSigner: false, IsWritable: false},
		// 12. token_quote_program
		{PublicKey: tokenQuoteProgram, IsSigner: false, IsWritable: false},
		// 13. referral_token_account (optional; use a valid token account)
		{PublicKey: referralTokenAccount, IsSigner: false, IsWritable: true},
		// 14. event_authority
		{PublicKey: eventAuthority, IsSigner: false, IsWritable: false},
		// 15. program
		{PublicKey: solana.MustPublicKeyFromBase58(common.DbcProgramID), IsSigner: false, IsWritable: false},
	}

	return solana.NewInstruction(
		solana.MustPublicKeyFromBase58(common.DbcProgramID),
		acctMetaSwap,
		buf,
	)
}
