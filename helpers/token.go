package helpers

import (
	"context"
	"dbcGoSDK/types"
	"errors"
	"fmt"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
)

func GetOrCreateATAInstruction(
	ctx context.Context,
	conn *rpc.Client,
	tokenMint, owner, payer solana.PublicKey,
	allowOwnerOffCurve bool,
	tokenProgram solana.PublicKey,
) (solana.PublicKey, *solana.GenericInstruction, error) {

	toAccount, err := GetAssociatedTokenAddressSync(
		tokenMint,
		owner,
		allowOwnerOffCurve,
		tokenProgram,
		solana.PublicKey{},
	)
	if err != nil {
		return solana.PublicKey{}, nil, err
	}

	if _, err = GetAccount(
		ctx,
		conn,
		toAccount,
		rpc.CommitmentConfirmed,
		solana.PublicKey{},
	); err == nil {
		return toAccount, nil, nil
	}

	ix := CreateAssociatedTokenAccountIdempotentInstruction(
		payer,
		toAccount,
		owner,
		tokenMint,
		tokenProgram,
	)
	return toAccount, ix, nil
}

func GetTokenProgram(flag uint8) solana.PublicKey {
	if flag == 0 {
		return solana.TokenProgramID
	}
	return solana.Token2022ProgramID
}

func WrapSOLInstruction(from, to solana.PublicKey, amount uint64) []solana.Instruction {
	return []solana.Instruction{
		system.NewTransferInstruction(
			amount,
			from,
			to,
		).Build(),
		solana.NewInstruction(
			solana.TokenProgramID,
			solana.AccountMetaSlice{
				{PublicKey: to, IsWritable: true},
			},
			[]byte{17},
		),
	}
}

func UnwrapSOLInstruction(
	owner, receiver solana.PublicKey,
	allowOwnerOffCurve bool,
) (*token.Instruction, error) {

	wSolATAAccount, err := GetAssociatedTokenAddressSync(
		solana.WrappedSol,
		owner,
		allowOwnerOffCurve,
		solana.PublicKey{},
		solana.PublicKey{},
	)
	if err != nil {
		return nil, err
	}

	return token.NewCloseAccountInstruction(
		wSolATAAccount,
		receiver,
		owner,
		nil,
	).Build(), nil

}

func GetAllPositionNftAccountByOwner(
	ctx context.Context,
	conn *rpc.Client, user solana.PublicKey,
) ([]struct{ PositionNft, PositionNftAccount solana.PublicKey }, error) {

	tokenAccounts, err := conn.GetTokenAccountsByOwner(
		ctx,
		user,
		&rpc.GetTokenAccountsConfig{
			ProgramId: &solana.Token2022ProgramID,
		},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get token accounts: %w", err)
	}

	if tokenAccounts == nil || len(tokenAccounts.Value) == 0 {
		return nil, errors.New("empty result from getTokenAccountsByOwner")
	}

	res := make([]struct{ PositionNft, PositionNftAccount solana.PublicKey }, 0, len(tokenAccounts.Value))

	for _, v := range tokenAccounts.Value {
		var tokenAcc token.Account

		if len(v.Account.Data.GetBinary()) == 0 {
			continue
		}

		if err := ag_binary.NewBorshDecoder(v.Account.Data.GetBinary()).Decode(&tokenAcc); err != nil {
			continue
		}

		if tokenAcc.Amount != 1 {
			continue
		}

		res = append(res, struct {
			PositionNft        solana.PublicKey
			PositionNftAccount solana.PublicKey
		}{
			PositionNft:        tokenAcc.Mint,
			PositionNftAccount: v.Pubkey,
		})
	}

	return res, nil
}

func FindAssociatedTokenAddress(
	walletAddress solana.PublicKey,
	tokenMintAddress solana.PublicKey,
	tokenProgramID solana.PublicKey,
) (solana.PublicKey, error) {
	seeds := [][]byte{
		walletAddress.Bytes(),
		tokenProgramID.Bytes(),
		tokenMintAddress.Bytes(),
	}

	addr, _, err := solana.FindProgramAddress(seeds, solana.SPLAssociatedTokenAccountProgramID)
	return addr, err
}

func GetTokenType(conn *rpc.Client, tokenMint solana.PublicKey) (types.TokenType, error) {
	var accInfo token.Account
	if err := conn.GetAccountDataBorshInto(
		context.Background(), tokenMint, &accInfo); err != nil {
		return types.TokenTypeSPL, err
	}

	if accInfo.Owner.Equals(solana.Token2022ProgramID) {
		return types.TokenTypeToken2022, nil
	}

	return types.TokenTypeSPL, nil
}

func GetTokenDecimals(
	conn *rpc.Client, mintAddress solana.PublicKey,
) (uint8, error) {
	var mint token.Mint
	if err := conn.GetAccountDataBorshInto(
		context.Background(),
		mintAddress,
		&mint,
	); err != nil {
		return 0, err
	}

	return mint.Decimals, nil
}
