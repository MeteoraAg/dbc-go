package anchor

import (
	"dbcGoSDK/generated/dbc"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type PgMethodI interface {
	PgAccountI
	Build() *dbc.Instruction
	ValidateAndBuild() (*dbc.Instruction, error)
}

type PgMethods[T PgMethodI] struct {
	programID            solana.PublicKey
	accountDiscriminator [8]byte
	conn                 *rpc.Client
	account              func() T
}
