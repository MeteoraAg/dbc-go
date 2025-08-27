package helpers

import (
	"context"
	"errors"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
)

var (
	// NativeMint2022 address of the special mint for wrapped native SOL in spl-token-2022.
	//   NativeMint2022 = solana.MustPublicKeyFromBase58("9pan9bMn5HatX4EJdBwg9VgCa7Uz5HL8N1m5D3NdXejP")
	NativeMint2022 = solana.MustPublicKeyFromBase58("9pan9bMn5HatX4EJdBwg9VgCa7Uz5HL8N1m5D3NdXejP")
)

// GetAssociatedTokenAddressSync returns the address of the associated token account for a given mint and owner.
func GetAssociatedTokenAddressSync(
	mint, owner solana.PublicKey,
	allowOwnerOffCurve bool,
	programId, associatedTokenProgramId solana.PublicKey,
) (solana.PublicKey, error) {
	if !allowOwnerOffCurve && !solana.IsOnCurve(owner.Bytes()) {
		return solana.PublicKey{}, errors.New("token owner is off-curve")
	}

	if programId.IsZero() {
		programId = solana.TokenProgramID
	}

	if associatedTokenProgramId.IsZero() {
		associatedTokenProgramId = solana.SPLAssociatedTokenAccountProgramID
	}

	addr, _, err := solana.FindProgramAddress(
		[][]byte{
			owner.Bytes(),
			programId.Bytes(),
			mint.Bytes(),
		},
		associatedTokenProgramId,
	)

	if err != nil {
		return solana.PublicKey{}, err
	}

	return addr, nil
}

// GetAccount retrieve information about a token account.
func GetAccount(
	ctx context.Context,
	conn *rpc.Client,
	address solana.PublicKey,
	commitment rpc.CommitmentType,
	programId solana.PublicKey,
) (token.Account, error) {

	acc, err := conn.GetAccountInfoWithOpts(
		ctx,
		address,
		&rpc.GetAccountInfoOpts{
			Commitment: commitment,
		})
	if err != nil {
		return token.Account{}, err
	}

	if acc == nil || len(acc.GetBinary()) == 0 {
		return token.Account{}, errors.New("empty data account from GetAccountInfo")
	}

	var t token.Account
	if err := ag_binary.NewBorshDecoder(acc.GetBinary()).Decode(&t); err != nil {
		return token.Account{}, err
	}

	return t, nil
}

func CreateAssociatedTokenAccountIdempotentInstruction(
	payer,
	associatedToken,
	owner,
	mint,
	programId solana.PublicKey,
) *solana.GenericInstruction {

	if programId.IsZero() {
		programId = solana.TokenProgramID
	}
	return buildAssociatedTokenAccountInstruction(
		payer,
		associatedToken,
		owner,
		mint,
		[]byte{1},
		programId,
		solana.SPLAssociatedTokenAccountProgramID,
	)
}

func buildAssociatedTokenAccountInstruction(
	payer solana.PublicKey,
	associatedToken solana.PublicKey,
	owner solana.PublicKey,
	mint solana.PublicKey,
	instructionData []byte,
	programId solana.PublicKey,
	associatedTokenProgramId solana.PublicKey,
) *solana.GenericInstruction {
	keys := solana.AccountMetaSlice{
		{PublicKey: payer, IsSigner: true, IsWritable: true},
		{PublicKey: associatedToken, IsSigner: false, IsWritable: true},
		{PublicKey: owner, IsSigner: false, IsWritable: false},
		{PublicKey: mint, IsSigner: false, IsWritable: false},
		{PublicKey: system.ProgramID, IsSigner: false, IsWritable: false},
		{PublicKey: programId, IsSigner: false, IsWritable: false},
	}

	return solana.NewInstruction(
		associatedTokenProgramId,
		keys,
		instructionData,
	)
}
