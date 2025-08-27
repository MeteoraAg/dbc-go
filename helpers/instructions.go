package helpers

import (
	"dbcGoSDK/constants"
	"dbcGoSDK/generated/dammv1"
	dynamic_vault "dbcGoSDK/generated/dynamicVault"

	"github.com/gagliardetto/solana-go"
)

// CreateInitializePermissionlessDynamicVaultIx creates a permissionless dynamic vault instruction.
func CreateInitializePermissionlessDynamicVaultIx(
	mint, payer solana.PublicKey,
) (struct {
	VaultKey, TokenVaultKey, LPMintKey solana.PublicKey
	Ix                                 *dynamic_vault.Instruction
}, error) {
	vaultKey := DeriveVaultAddress(mint, constants.BaseAddress)
	tokenVaultKey := DeriveTokenVaultKey(vaultKey)
	lpMintKey := DeriveVaultLpMintAddress(vaultKey)

	ix, err := dynamic_vault.NewInitializeInstruction(
		vaultKey,
		payer,
		tokenVaultKey,
		mint,
		lpMintKey,
		solana.SysVarRentPubkey,
		solana.TokenProgramID,
		solana.SystemProgramID,
	).ValidateAndBuild()
	if err != nil {
		return struct {
			VaultKey      solana.PublicKey
			TokenVaultKey solana.PublicKey
			LPMintKey     solana.PublicKey
			Ix            *dynamic_vault.Instruction
		}{}, err
	}
	return struct {
		VaultKey      solana.PublicKey
		TokenVaultKey solana.PublicKey
		LPMintKey     solana.PublicKey
		Ix            *dynamic_vault.Instruction
	}{
		VaultKey:      vaultKey,
		TokenVaultKey: tokenVaultKey,
		LPMintKey:     lpMintKey,
		Ix:            ix,
	}, nil
}

func CreateLockEscrowIx(
	payer, pool, lpMint, escrowOwner, lockEscrowKey solana.PublicKey,
) (*dammv1.Instruction, error) {
	return dammv1.NewCreateLockEscrowInstruction(
		pool,
		lockEscrowKey,
		escrowOwner,
		lpMint,
		payer,
		solana.SystemProgramID,
	).ValidateAndBuild()
}
