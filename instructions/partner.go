package instructions

import (
	"encoding/binary"

	"github.com/gagliardetto/solana-go"

	"github.com/MeteoraAg/dbc-go/common"
	"github.com/MeteoraAg/dbc-go/helpers"
)

func ClaimPartnerTradingFee(
	config solana.PublicKey,
	pool solana.PublicKey,
	tokenAAccount solana.PublicKey,
	tokenBAccount solana.PublicKey,
	baseVault solana.PublicKey,
	quoteVault solana.PublicKey,
	baseMint solana.PublicKey,
	quoteMint solana.PublicKey,
	feeClaimer solana.PublicKey,
	maxAmountA uint64,
	maxAmountB uint64,
) solana.Instruction {
	disc := []byte{8, 236, 89, 49, 152, 125, 177, 81}
	buf := make([]byte, 8+8+8)
	copy(buf, disc)
	binary.LittleEndian.PutUint64(buf[8:], maxAmountA)
	binary.LittleEndian.PutUint64(buf[16:], maxAmountB)

	tokenBaseProgram := solana.MustPublicKeyFromBase58(common.TokenProgram)
	tokenQuoteProgram := solana.MustPublicKeyFromBase58(common.TokenProgram)
	poolAuthority := helpers.DerivePoolAuthorityPDA()
	eventAuthority := helpers.DeriveEventAuthorityPDA()

	acctMeta := solana.AccountMetaSlice{
		// 1. pool_authority
		{PublicKey: poolAuthority, IsSigner: false, IsWritable: false},
		// 2. config
		{PublicKey: config, IsSigner: false, IsWritable: false},
		// 3. pool (writable)
		{PublicKey: pool, IsSigner: false, IsWritable: true},
		// 4. token_a_account (writable)
		{PublicKey: tokenAAccount, IsSigner: false, IsWritable: true},
		// 5. token_b_account (writable)
		{PublicKey: tokenBAccount, IsSigner: false, IsWritable: true},
		// 6. base_vault (writable)
		{PublicKey: baseVault, IsSigner: false, IsWritable: true},
		// 7. quote_vault (writable)
		{PublicKey: quoteVault, IsSigner: false, IsWritable: true},
		// 8. base_mint
		{PublicKey: baseMint, IsSigner: false, IsWritable: false},
		// 9. quote_mint
		{PublicKey: quoteMint, IsSigner: false, IsWritable: false},
		// 10. fee_claimer (signer)
		{PublicKey: feeClaimer, IsSigner: true, IsWritable: false},
		// 11. token_base_program
		{PublicKey: tokenBaseProgram, IsSigner: false, IsWritable: false},
		// 12. token_quote_program
		{PublicKey: tokenQuoteProgram, IsSigner: false, IsWritable: false},
		// 13. event_authority
		{PublicKey: eventAuthority, IsSigner: false, IsWritable: false},
		// 14. program
		{PublicKey: solana.MustPublicKeyFromBase58(common.DbcProgramID), IsSigner: false, IsWritable: false},
	}

	return solana.NewInstruction(
		solana.MustPublicKeyFromBase58(common.DbcProgramID),
		acctMeta,
		buf,
	)
}
