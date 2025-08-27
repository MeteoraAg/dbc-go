package helpers

import (
	"bytes"
	"dbcGoSDK/constants"

	"github.com/gagliardetto/solana-go"
)

func DeriveDbcPoolAuthority() solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedPoolAuthority),
		},
		constants.DBCProgramId,
	)
	return pda
}

func DeriveLockerEventAuthority() solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedEventAuthority),
			constants.LockerProgramId.Bytes(),
		},
		constants.DBCProgramId,
	)
	return pda
}

// DeriveBaseKeyForLocker derives base key for the locker.
func DeriveBaseKeyForLocker(virtualPool solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedBaseLocker),
			virtualPool.Bytes(),
		},
		constants.DBCProgramId,
	)
	return pda
}

func DeriveEscrow(base solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedEscrow),
			base.Bytes(),
		},
		constants.DBCProgramId,
	)
	return pda
}

func DeriveDbcPoolMetadata(pool solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedVirtualPoolMetadata),
			pool.Bytes(),
		},
		constants.DBCProgramId,
	)
	return pda
}

func DeriveDammV2TokenVaultAddress(pool, mint solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedTokenVault),
			mint.Bytes(),
			pool.Bytes(),
		},
		constants.DammV2ProgramId,
	)
	return pda
}
func DeriveDammV1LpMintAddress(pool solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedLpMint),
			pool.Bytes(),
		},
		constants.DammV1ProgramId,
	)
	return pda
}

// DerivePositionAddress derives DAMM V2 position address.
func DerivePositionAddress(positionNft solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedPosition),
			positionNft.Bytes(),
		},
		constants.DammV2ProgramId,
	)
	return pda
}

// DerivePositionNftAccount derives DAMM V2 position NFT account.
func DerivePositionNftAccount(positionNft solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedPositionNFTAccount),
			positionNft.Bytes(),
		},
		constants.DammV2ProgramId,
	)
	return pda
}

func DeriveDammV2PoolAuthority() solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedPoolAuthority),
		},
		constants.DammV2ProgramId,
	)
	return pda
}

func DeriveDammV2EventAuthority() solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedEventAuthority),
		},
		constants.DammV2ProgramId,
	)
	return pda
}
func DeriveDammV2MigrationMetadataAddress(virtualPool solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedDammV2MigrationMetadata),
			virtualPool.Bytes(),
		},
		constants.DBCProgramId,
	)
	return pda
}

func DeriveDammV1LockEscrowAddress(dammPool, creator solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedLockEscrow),
			dammPool.Bytes(),
			creator.Bytes(),
		},
		constants.DammV1ProgramId,
	)
	return pda
}

func DeriveDbcPartnerMetadata(feeClaimer solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedPartnerMetadata),
			feeClaimer.Bytes(),
		},
		constants.DBCProgramId,
	)
	return pda
}

func DeriveDammV1MigrationMetadataAddress(virtualPool solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedDammV1MigrationMetadata),
			virtualPool.Bytes(),
		},
		constants.DBCProgramId,
	)
	return pda
}

func DeriveDammV1ProtocolFeeAddress(mint, pool solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedFee),
			mint.Bytes(),
			pool.Bytes(),
		},
		constants.DammV1ProgramId,
	)
	return pda
}

func DeriveDammV1PoolAddress(
	config, tokenAMint, tokenBMintt solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedDammV1MigrationMetadata),
			GetFirstkey(tokenAMint, tokenBMintt),
			GetSecondkey(tokenAMint, tokenBMintt),
			config.Bytes(),
		},
		constants.DammV1ProgramId,
	)
	return pda
}

func DeriveDammV1VaultLPAddress(
	vault, pool solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			vault.Bytes(),
			pool.Bytes(),
		},
		constants.DammV1ProgramId,
	)
	return pda
}

// DeriveDbcTokenVaultAddress derives DBC token vault address.
func DeriveDbcTokenVaultAddress(pool, mint solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedTokenVault),
			mint.Bytes(),
			pool.Bytes(),
		},
		constants.DBCProgramId,
	)
	return pda
}

// DeriveMintMetadata derives mint metadata address.
func DeriveMintMetadata(mint solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedMetadata),
			constants.MetaplexProgramId.Bytes(),
			mint.Bytes(),
		},
		constants.MetaplexProgramId,
	)
	return pda
}

func DeriveTokenVaultKey(vaultKey solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedTokenVault),
			vaultKey.Bytes(),
		},
		constants.VaultProgramId,
	)
	return pda
}
func DeriveVaultAddress(mint, payer solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedVault),
			mint.Bytes(),
			payer.Bytes(),
		},
		constants.VaultProgramId,
	)
	return pda
}

func DeriveVaultLpMintAddress(pool solana.PublicKey) solana.PublicKey {
	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedLpMint),
			pool.Bytes(),
		},
		constants.VaultProgramId,
	)
	return pda
}

func DeriveDbcPoolAddress(quoteMint, baseMint, config solana.PublicKey) solana.PublicKey {
	isQuoteMintBiggerThanBaseMint := bytes.Compare(quoteMint.Bytes(), baseMint.Bytes()) > 0

	if isQuoteMintBiggerThanBaseMint {
		pda, _, _ := solana.FindProgramAddress(
			[][]byte{
				[]byte(constants.SeedPool),
				config.Bytes(),
				quoteMint.Bytes(),
				baseMint.Bytes(),
			},
			constants.DBCProgramId,
		)
		return pda
	}

	pda, _, _ := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedPool),
			config.Bytes(),
			baseMint.Bytes(),
			quoteMint.Bytes(),
		},
		constants.DBCProgramId,
	)
	return pda
}

func DeriveVaultPdas(
	tokenMint, seedBaseKey solana.PublicKey,
) (struct{ VaultPDA, TokenVaultPDA, LPMintPDA solana.PublicKey }, error) {

	bbb := constants.BaseAddress
	if !seedBaseKey.IsZero() {
		bbb = seedBaseKey
	}
	vault, _, err := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedVault),
			tokenMint.Bytes(),
			bbb.Bytes(),
		},
		constants.VaultProgramId,
	)
	if err != nil {
		return struct {
			VaultPDA      solana.PublicKey
			TokenVaultPDA solana.PublicKey
			LPMintPDA     solana.PublicKey
		}{}, err
	}

	tokenVault, _, err := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedTokenVault),
			vault.Bytes(),
		},
		constants.VaultProgramId,
	)
	if err != nil {
		return struct {
			VaultPDA      solana.PublicKey
			TokenVaultPDA solana.PublicKey
			LPMintPDA     solana.PublicKey
		}{}, err
	}

	lpMint, _, err := solana.FindProgramAddress(
		[][]byte{
			[]byte(constants.SeedLpMint),
			vault.Bytes(),
		},
		constants.VaultProgramId,
	)
	if err != nil {
		return struct {
			VaultPDA      solana.PublicKey
			TokenVaultPDA solana.PublicKey
			LPMintPDA     solana.PublicKey
		}{}, err
	}

	return struct {
		VaultPDA      solana.PublicKey
		TokenVaultPDA solana.PublicKey
		LPMintPDA     solana.PublicKey
	}{
		VaultPDA:      vault,
		TokenVaultPDA: tokenVault,
		LPMintPDA:     lpMint,
	}, nil
}
