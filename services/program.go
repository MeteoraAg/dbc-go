package services

import (
	"context"
	"dbcGoSDK/constants"
	"dbcGoSDK/helpers"
	"fmt"
	"slices"
	"sync"

	"dbcGoSDK/types"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type DBCProgram struct {
	conn          *rpc.Client
	poolAuthority solana.PublicKey
	commitment    rpc.CommitmentType
}

func NewDBCProgram(
	conn *rpc.Client,
	commitment rpc.CommitmentType,
) *DBCProgram {
	return &DBCProgram{
		conn:          conn,
		poolAuthority: helpers.DeriveDbcPoolAuthority(),
		commitment:    commitment,
	}
}

func (d *DBCProgram) prepareTokenAccounts(
	ctx context.Context,
	param types.PrepareTokenAccountParams,

) (struct {
	TokenAAta, TokenBAta solana.PublicKey
	CreateATAIxns        []solana.Instruction
}, error) {
	type res struct {
		AtaPubkey solana.PublicKey
		Ix        *solana.GenericInstruction
		Err       error
	}
	var (
		wg   sync.WaitGroup
		a, b res
	)

	wg.Add(2)

	go func(p *res, wg *sync.WaitGroup) {
		defer wg.Done()
		ata, ix, err := helpers.GetOrCreateATAInstruction(
			ctx,
			d.conn,
			param.TokenAMint,
			param.Owner,
			param.Payer,
			true,
			param.TokenAProgram,
		)
		if err != nil {
			p.Err = err
			return
		}
		p.AtaPubkey = ata
		p.Ix = ix
	}(&a, &wg)

	go func(p *res, wg *sync.WaitGroup) {
		defer wg.Done()
		ata, ix, err := helpers.GetOrCreateATAInstruction(
			ctx,
			d.conn,
			param.TokenBMint,
			param.Owner,
			param.Payer,
			true,
			param.TokenBProgram,
		)
		if err != nil {
			p.Err = err
			return
		}
		p.AtaPubkey = ata
		p.Ix = ix
	}(&b, &wg)

	wg.Wait()

	handleNilErr := func(err error) string {
		if err == nil {
			return ""
		}
		return err.Error()
	}

	if a.Err != nil || b.Err != nil {
		return struct {
				TokenAAta     solana.PublicKey
				TokenBAta     solana.PublicKey
				CreateATAIxns []solana.Instruction
			}{},
			fmt.Errorf("err from token A— %s: err from tokenB— %s", handleNilErr(a.Err), handleNilErr(b.Err))
	}

	prepareATAIxns := func() []solana.Instruction {
		ixns := make([]solana.Instruction, 0, 2)
		if a.Ix != nil {
			ixns = append(ixns, a.Ix)
		}
		if b.Ix != nil {
			ixns = append(ixns, b.Ix)
		}

		return slices.Clip(ixns)
	}

	return struct {
		TokenAAta     solana.PublicKey
		TokenBAta     solana.PublicKey
		CreateATAIxns []solana.Instruction
	}{
		TokenAAta:     a.AtaPubkey,
		TokenBAta:     b.AtaPubkey,
		CreateATAIxns: prepareATAIxns(),
	}, nil
}

func (d DBCProgram) GetProgramID() solana.PublicKey {
	return constants.DBCProgramId
}

func (d DBCProgram) GetPoolAuthority() solana.PublicKey {
	return d.poolAuthority
}
