package helpers

import (
	"dbcGoSDK/constants"
	"dbcGoSDK/types"
	"fmt"
	"math/big"

	ag_binary "github.com/gagliardetto/binary"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

func CreateProgramAccountFilter(owner solana.PublicKey, offset uint64) rpc.GetProgramAccountsOpts {
	return rpc.GetProgramAccountsOpts{
		Encoding: solana.EncodingBase58,
		Filters: []rpc.RPCFilter{
			{
				Memcmp: &rpc.RPCFilterMemcmp{
					Bytes:  owner.Bytes(),
					Offset: offset,
				},
			},
		},
	}
}

func BpsToFeeNumerator(bps uint64) *big.Int {
	return new(big.Int).Quo(
		new(big.Int).Mul(new(big.Int).SetUint64(bps), constants.FeeDenominatorBigInt),
		big.NewInt(constants.BasisPointMax))
}

func ConvertToLamports(amount float64, tokenDecimal types.TokenDecimal) *big.Int {
	floatVal := new(big.Float).Mul(
		big.NewFloat(amount),
		new(big.Float).SetPrec(256).SetInt(
			new(big.Int).Exp(big.NewInt(10), new(big.Int).SetUint64(uint64(tokenDecimal)), nil),
		),
	)
	result := new(big.Int)
	floatVal.Int(result) // truncate/floor
	return result
}

func BigIntToUint128(b *big.Int) (ag_binary.Uint128, error) {
	if b.Sign() < 0 {
		return ag_binary.Uint128{}, fmt.Errorf("value must be unsigned")
	}

	if b.BitLen() > 128 {
		return ag_binary.Uint128{}, fmt.Errorf("value %s exceeds 128 bits", b.String())
	}

	var buf [16]byte
	b.FillBytes(buf[:]) // zero-pads on the left

	ag_binary.ReverseBytes(buf[:])

	// ag_binary.NewBinDecoder(buf[:]).ReadUint128()
	var u ag_binary.Uint128
	if err := u.UnmarshalWithDecoder(ag_binary.NewBinDecoder(buf[:])); err != nil {
		return ag_binary.Uint128{}, err
	}
	return u, nil
}

// Must helper
func MustBigIntToUint128(b *big.Int) ag_binary.Uint128 {
	v, err := BigIntToUint128(b)
	if err != nil {
		panic(fmt.Errorf("cannot fit big.Int into Uint128: %s", err.Error()))
	}
	return v
}
