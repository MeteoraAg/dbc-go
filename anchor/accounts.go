package anchor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/btcsuite/btcutil/base58"
	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"golang.org/x/sync/errgroup"
)

type PgAccountI interface {
	MarshalWithEncoder(encoder *ag_binary.Encoder) (err error)
	UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error)
}

type PgAccounts[T PgAccountI] struct {
	conn    *rpc.Client
	account func() T
}

func NewPgAccounts[T PgAccountI](conn *rpc.Client, account func() T) *PgAccounts[T] {
	return &PgAccounts[T]{
		conn:    conn,
		account: account,
	}
}

func (ac *PgAccounts[T]) fetchNullable(ctx context.Context, address solana.PublicKey, opts *rpc.GetAccountInfoOpts) (T, error) {
	account, _, err := ac.fetchNullableAndContext(ctx, address, opts)

	return account, err

}

func (ac *PgAccounts[T]) fetchNullableAndContext(ctx context.Context, address solana.PublicKey, opts *rpc.GetAccountInfoOpts) (T, *rpc.RPCContext, error) {
	accoutnInfo, rpcCtx, err := ac.conn.GetAccountInfoWithRpcContext(ctx, address, opts)
	var zeroValue T
	if err != nil {
		return zeroValue, nil, err
	}
	if accoutnInfo == nil || accoutnInfo.Data == nil || len(accoutnInfo.Data.GetBinary()) == 0 {
		return zeroValue, nil, fmt.Errorf("fetchNullableAndContext: account does not exist: %s", address.String())
	}
	concrate := ac.account()
	if err := concrate.UnmarshalWithDecoder(ag_binary.NewBorshDecoder(accoutnInfo.Data.GetBinary())); err != nil {
		return zeroValue, nil, fmt.Errorf("fetchNullableAndContext: err decoding account data: %s", err.Error())
	}

	return concrate, rpcCtx, nil

}

func (ac *PgAccounts[T]) Fetch(ctx context.Context, address solana.PublicKey, opts *rpc.GetAccountInfoOpts) (T, error) {
	return ac.fetchNullable(ctx, address, opts)
}
func (ac *PgAccounts[T]) FetchWithRpcCtx(ctx context.Context, address solana.PublicKey, opts *rpc.GetAccountInfoOpts) (T, *rpc.RPCContext, error) {
	return ac.fetchNullableAndContext(ctx, address, opts)
}

func (ac *PgAccounts[T]) FetchMultiple(ctx context.Context, addresses []solana.PublicKey, opts *rpc.GetMultipleAccountsOpts) ([]T, error) {
	var (
		batchSize = 99
		res       = make([]T, len(addresses))
		g, newCtx = errgroup.WithContext(ctx)
	)

	for i := 0; i < len(addresses); i += batchSize {
		batchStart := i // for Go version below 1.22
		end := min(batchStart+batchSize, len(addresses))
		batchKeys := addresses[batchStart:end]

		g.SetLimit(runtime.NumCPU())

		g.Go(func() error {
			out, err := ac.conn.GetMultipleAccountsWithOpts(newCtx, batchKeys, opts)
			if err != nil {
				return err
			}

			if out == nil || len(out.Value) == 0 {
				return errors.New("empty result from GetMultipleAccounts")
			}

			if len(out.Value) > len(batchKeys) {
				return errors.New("GetMultipleAccounts returned more result than expected")
			}

			for j, value := range out.Value {

				concrete := ac.account()
				if value != nil && value.Data != nil && len(value.Data.GetBinary()) != 0 {
					if err := concrete.UnmarshalWithDecoder(ag_binary.NewBorshDecoder(value.Data.GetBinary())); err != nil {
						continue
					}
				}

				res[batchStart+j] = concrete
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return res, nil
}

func (ac *PgAccounts[T]) All(
	ctx context.Context,
	programID solana.PublicKey,
	accountDiscriminator [8]byte,
	opts rpc.GetProgramAccountsOpts,
	bufferParam []byte,
) ([]ProgramAccount[T], error) {

	apppendData := bytes.NewBuffer(accountDiscriminator[:])

	if len(bufferParam) != 0 {
		apppendData.Write(bufferParam)
	}

	allFilters := make([]rpc.RPCFilter, 0, 1+len(opts.Filters))
	allFilters = append(allFilters, rpc.RPCFilter{
		Memcmp: &rpc.RPCFilterMemcmp{
			Offset: 0,
			Bytes:  solana.Base58(base58.Encode(apppendData.Bytes())),
		},
	})
	allFilters = append(allFilters, opts.Filters...)
	opts.Filters = allFilters

	// opts.Filters = append(
	// 	[]rpc.RPCFilter{{
	// 		Memcmp: &rpc.RPCFilterMemcmp{
	// 			Offset: 0,
	// 			Bytes:  solana.Base58(base58.Encode(apppendData.Bytes())),
	// 		},
	// 	}},
	// 	opts.Filters...,
	// )

	out, err := ac.conn.GetProgramAccountsWithOpts(ctx, programID, &opts)
	if err != nil {
		return nil, err
	}
	res := make([]ProgramAccount[T], 0, len(out))

	for _, v := range out {
		if v == nil || v.Account == nil || v.Account.Data == nil || len(v.Account.Data.GetBinary()) == 0 {
			continue
		}

		concrete := ac.account()
		if err := concrete.UnmarshalWithDecoder(ag_binary.NewBorshDecoder(v.Account.Data.GetBinary())); err != nil {
			continue
		}
		res = append(res, ProgramAccount[T]{
			PublicKey: v.Pubkey,
			Account:   concrete,
		})
	}

	return res, nil
}

type ProgramAccount[T any] struct {
	PublicKey solana.PublicKey
	Account   T
}
