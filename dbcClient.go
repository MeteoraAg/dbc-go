package dbcgosdk

import (
	"dbcGoSDK/services"

	"github.com/gagliardetto/solana-go/rpc"
)

type DynamicBondingCurveClient struct {
	Conn       *rpc.Client
	Commitment rpc.CommitmentType
	State      *services.StateService
	Pool       *services.PoolService
	Partner    *services.PartnerService
	Creator    *services.CreatorService
	Migration  *services.MigrationService
}

func NewDynamicBondingCurveClient(
	conn *rpc.Client,
	commitment rpc.CommitmentType,
) *DynamicBondingCurveClient {
	return &DynamicBondingCurveClient{
		Conn:       conn,
		Commitment: commitment,
		State:      services.NewStateService(conn, commitment),
		Pool:       services.NewPoolService(conn, commitment),
		Partner:    services.NewPartnerService(conn, commitment),
		Creator:    services.NewCreatorService(conn, commitment),
		Migration:  services.NewMigrationService(conn, commitment),
	}
}
