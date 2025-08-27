package dbcgosdk_test

import (
	"context"
	dbcgosdk "dbcGoSDK"
	"math/big"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"github.com/stretchr/testify/assert"

	"dbcGoSDK/generated/dbc"
	"dbcGoSDK/helpers"
	testUtils "dbcGoSDK/internal/test/utils"
	"dbcGoSDK/types"
)

const (
	surfPoolRPCClient = "http://127.0.0.1:8899"
	surfPoolWSlient   = "ws://127.0.0.1:8900"
)

func TestSwap(t *testing.T) {
	conn := rpc.New(surfPoolRPCClient)
	wsClient, err := ws.Connect(context.Background(), surfPoolWSlient)
	if err != nil {
		t.Fatalf("err creating ws client: %s", err.Error())
	}

	t.Cleanup(func() {
		conn.Close()
		wsClient.Close()
	})

	actors, err := testUtils.SetupTestContext(
		t,
		conn,
		wsClient,
	)
	if err != nil {
		t.Fatalf("err from SetupTestContext: %s", err.Error())
	}

	dbcClient := dbcgosdk.NewDynamicBondingCurveClient(
		conn,
		rpc.CommitmentConfirmed,
	)
	ctx := context.TODO()

	l0, _ := new(big.Int).SetString("622226417996106429201027821619672729", 10)
	s1, _ := new(big.Int).SetString("79226673521066979257578248091", 10)

	createConfigAndPoolIxnx, err := dbcClient.Pool.CreateConfigAndPool(
		types.CreateConfigAndPoolParam{
			TokenType: types.TokenTypeSPL,
			CreateConfigParam: types.CreateConfigParam{
				Payer:            actors.Admin.PublicKey(),
				Config:           actors.Config.PublicKey(),
				FeeClaimer:       actors.Admin.PublicKey(),
				LeftoverReceiver: actors.Admin.PublicKey(),
				QuoteMint:        solana.WrappedSol,
				ConfigParameters: dbc.ConfigParameters{
					PoolFees: dbc.PoolFeeParameters{
						BaseFee: dbc.BaseFeeParameters{
							CliffFeeNumerator: 2_500_000,
							FirstFactor:       0,
							SecondFactor:      0,
							ThirdFactor:       0,
							BaseFeeMode:       0,
						},
						DynamicFee: &dbc.DynamicFeeParameters{
							BinStep:                  1,
							BinStepU128:              helpers.MustBigIntToUint128(big.NewInt(1844674407370955)),
							FilterPeriod:             10,
							DecayPeriod:              120,
							ReductionFactor:          1_000,
							VariableFeeControl:       100_000,
							MaxVolatilityAccumulator: 100_000,
						},
					},
					TokenDecimal:              9,
					MigrationQuoteThreshold:   1_000_000_000_000,
					PartnerLpPercentage:       25,
					CreatorLpPercentage:       25,
					PartnerLockedLpPercentage: 25,
					CreatorLockedLpPercentage: 25,
					SqrtStartPrice:            helpers.MustBigIntToUint128(big.NewInt(58333726687135158)),
					TokenSupply: &dbc.TokenSupplyParams{
						PreMigrationTokenSupply:  10_000_000_000_000_000_000,
						PostMigrationTokenSupply: 10_000_000_000_000_000_000,
					},
					TokenUpdateAuthority: 1,
					MigrationFee: dbc.MigrationFee{
						FeePercentage:        25,
						CreatorFeePercentage: 50,
					},
					MigratedPoolFee: dbc.MigratedPoolFee{
						CollectFeeMode: 0,
						DynamicFee:     0,
						PoolFeeBps:     0,
					},
					Curve: []dbc.LiquidityDistributionParameters{
						{
							SqrtPrice: helpers.MustBigIntToUint128(big.NewInt(233334906748540631)),
							Liquidity: helpers.MustBigIntToUint128(l0),
						},
						{
							SqrtPrice: helpers.MustBigIntToUint128(s1),
							Liquidity: helpers.MustBigIntToUint128(big.NewInt(1)),
						},
					},
				},
			},
			PreCreatePoolParam: types.PreCreatePoolParam{
				Name:        "The Excocist",
				Symbol:      "ext",
				PoolCreator: actors.Admin.PublicKey(),
				BaseMint:    actors.BaseMint.PublicKey(),
			},
		},
	)
	if err != nil {
		t.Fatalf("CreateConfigAndPool errored: %s", err.Error())
	}

	_, err = testUtils.ExecuteTransaction(
		conn,
		wsClient,
		createConfigAndPoolIxnx,
		actors.Admin,
		actors.Config, actors.BaseMint,
	)
	if err != nil {
		testUtils.PrettyPrintTxnErrorLog(t, err)
		t.FailNow()
	}

	t.Log("createConfigAndPool successful ✅")

	actors.Pool = helpers.DeriveDbcPoolAddress(
		solana.WrappedSol,
		actors.BaseMint.PublicKey(),
		actors.Config.PublicKey(),
	)

	poolState, err := dbcClient.State.GetPool(ctx, actors.Pool)
	if err != nil {
		t.Fatalf("GetPool errored: %s", err.Error())
	}

	poolConfigState, err := dbcClient.State.GetPoolConfig(
		ctx,
		poolState.Config,
	)
	if err != nil {
		t.Fatalf("GetPoolConfig errored: %s", err.Error())
	}

	// create referralTokenAccount ata
	ixns := make([]solana.Instruction, 0, 5)
	referralTokenAccount, err := helpers.GetAssociatedTokenAddressSync(
		solana.WrappedSol,
		actors.User.PublicKey(),
		false,
		solana.TokenProgramID,
		solana.PublicKey{},
	)
	if err != nil {
		t.Fatalf("err from helpers.GetAssociatedTokenAddressSync: %s", err.Error())
	}

	createAtaIx := helpers.CreateAssociatedTokenAccountIdempotentInstruction(
		actors.User.PublicKey(),
		referralTokenAccount,
		actors.User.PublicKey(),
		solana.WrappedSol,
		solana.TokenProgramID,
	)

	ixns = append(ixns, createAtaIx)

	t.Run("swapV1", func(t *testing.T) {
		swapIxns, err := dbcClient.Pool.Swap(
			ctx,
			types.SwapParam{
				Owner:                actors.User.PublicKey(),
				Pool:                 actors.Pool,
				AmountIn:             1_000_000_000,
				MinimumAmountOut:     0,
				SwapBaseForQuote:     false,
				ReferralTokenAccount: referralTokenAccount,
				Payer:                actors.User.PublicKey(),
			},
		)
		if err != nil {
			t.Fatalf("Swap errored: %s", err.Error())
		}

		_, err = testUtils.ExecuteTransaction(
			conn,
			wsClient,
			append(ixns, swapIxns...),
			actors.User,
		)
		if err != nil {
			testUtils.PrettyPrintTxnErrorLog(t, err)
			t.FailNow()
		}
	})

	t.Run("swap2ExactIn", func(t *testing.T) {
		currentPoint, err := helpers.GetCurrentPoint(
			conn,
			types.ActivationType(poolConfigState.ActivationType),
		)
		if err != nil {
			t.Fatalf("GetCurrentPoint errored: %s", err.Error())
		}

		swapQuote, err := dbcClient.Pool.SwapQuote2(types.SwapQuote2Param{
			VirtualPool:      poolState,
			Config:           poolConfigState,
			SwapBaseForQuote: true,
			AmountIn:         big.NewInt(1_000_000_000),
			SlippageBps:      50,
			CurrentPoint:     currentPoint,
			SwapMode:         types.SwapModeExactIn,
		})
		if err != nil {
			t.Fatalf("SwapQuote2 errored: %s", err.Error())
		}

		t.Logf("swapQuote ----> %+v\n\n", swapQuote)
		assert.Equal(t, uint64(9974), swapQuote.OutputAmount)

		swap2Ixns, err := dbcClient.Pool.Swap2(
			ctx,
			types.Swap2Param{
				SwapMode:             types.SwapModeExactIn,
				AmountIn:             big.NewInt(1_000_000_000),
				MinimumAmountOut:     new(big.Int).SetUint64(swapQuote.MinimumAmountOut),
				Owner:                actors.User.PublicKey(),
				Pool:                 actors.Pool,
				ReferralTokenAccount: referralTokenAccount,
				Payer:                actors.User.PublicKey(),
			},
		)

		if err != nil {
			t.Fatalf("Swap errored: %s", err.Error())
		}

		_, err = testUtils.ExecuteTransaction(
			conn,
			wsClient,
			append(ixns, swap2Ixns...),
			actors.User,
		)
		if err != nil {
			testUtils.PrettyPrintTxnErrorLog(t, err)
			t.FailNow()
		}
	})

	t.Run("swap2PartialFill", func(t *testing.T) {
		currentPoint, err := helpers.GetCurrentPoint(
			conn,
			types.ActivationType(poolConfigState.ActivationType),
		)
		if err != nil {
			t.Fatalf("GetCurrentPoint errored: %s", err.Error())
		}

		swapQuote, err := dbcClient.Pool.SwapQuote2(types.SwapQuote2Param{
			VirtualPool:  poolState,
			Config:       poolConfigState,
			AmountIn:     big.NewInt(1_000_000_000),
			SlippageBps:  50,
			CurrentPoint: currentPoint,
			SwapMode:     types.SwapModePartialFill,
		})
		if err != nil {
			t.Fatalf("SwapQuote2 errored: %s", err.Error())
		}

		t.Logf("%+v\n\n", swapQuote)
		assert.Equal(t, uint64(99749067190242), swapQuote.OutputAmount)

		swap2Ixns, err := dbcClient.Pool.Swap2(
			ctx,
			types.Swap2Param{
				AmountIn:             big.NewInt(1_000_000_000),
				MinimumAmountOut:     new(big.Int).SetUint64(swapQuote.MinimumAmountOut),
				SwapMode:             types.SwapModePartialFill,
				Owner:                actors.User.PublicKey(),
				Pool:                 actors.Pool,
				ReferralTokenAccount: referralTokenAccount,
				Payer:                actors.User.PublicKey(),
			},
		)

		if err != nil {
			t.Fatalf("Swap errored: %s", err.Error())
		}

		_, err = testUtils.ExecuteTransaction(
			conn,
			wsClient,
			append(ixns, swap2Ixns...),
			actors.User,
		)
		if err != nil {
			testUtils.PrettyPrintTxnErrorLog(t, err)
			t.FailNow()
		}
	})

	t.Run("swapVExactOut", func(t *testing.T) {
		currentPoint, err := helpers.GetCurrentPoint(
			conn,
			types.ActivationType(poolConfigState.ActivationType),
		)
		if err != nil {
			t.Fatalf("GetCurrentPoint errored: %s", err.Error())
		}

		swapQuote, err := dbcClient.Pool.SwapQuote2(types.SwapQuote2Param{
			VirtualPool:  poolState,
			Config:       poolConfigState,
			AmountOut:    big.NewInt(1_000_000_000),
			SlippageBps:  50,
			CurrentPoint: currentPoint,
			SwapMode:     types.SwapModeExactOut,
		})
		if err != nil {
			t.Fatalf("SwapQuote2 errored: %s", err.Error())
		}

		t.Logf("%+v\n\n", swapQuote)
		assert.Equal(t, uint64(1_000_000_000), swapQuote.OutputAmount)

		swap2Ixns, err := dbcClient.Pool.Swap2(
			ctx,
			types.Swap2Param{
				AmountOut:            big.NewInt(1_000_000_000),
				MaximumAmountIn:      new(big.Int).SetUint64(swapQuote.MaximumAmountIn),
				SwapMode:             types.SwapModeExactOut,
				Owner:                actors.User.PublicKey(),
				Pool:                 actors.Pool,
				ReferralTokenAccount: referralTokenAccount,
				Payer:                actors.User.PublicKey(),
			},
		)

		if err != nil {
			t.Fatalf("Swap errored: %s", err.Error())
		}

		_, err = testUtils.ExecuteTransaction(
			conn,
			wsClient,
			append(ixns, swap2Ixns...),
			actors.User,
		)
		if err != nil {
			testUtils.PrettyPrintTxnErrorLog(t, err)
			t.FailNow()
		}
	})
}
