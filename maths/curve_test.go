package maths_test

import (
	"dbcGoSDK/maths"
	"dbcGoSDK/types"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

// func TestCurve(t *testing.T) {
// 	t.Run("base amount calculation:getDeltaAmountBaseUnsigned", func(t *testing.T) {
// 		// Lower test liquidity value to prevent overflow
// 		liquidity, _ := new(big.Int).SetString("1293129312931923921293912", 10)
// 		lower, upper := maths.Q64(1.0), maths.Q64(1.0001)

// 		got, err := maths.GetDeltaAmountBaseUnsigned(
// 			lower,
// 			upper,
// 			liquidity,
// 			types.RoundingDown,
// 		)
// 		if err != nil {
// 			t.Fatalf("GetDeltaAmountBaseUnsigned errored: %s", err.Error())
// 		}
// 		assert.Equal(t, "7", got.String())
// 	})

// 	t.Run("quote amount calculation:getDeltaAmountQuoteUnsigned", func(t *testing.T) {
// 		liquidity := new(big.Int).Exp(big.NewInt(10), big.NewInt(25), nil)
// 		lower, upper := maths.Q64(1.0), maths.Q64(1.0001)

// 		got, err := maths.GetDeltaAmountBaseUnsigned(
// 			lower,
// 			upper,
// 			liquidity,
// 			types.RoundingDown,
// 		)
// 		if err != nil {
// 			t.Fatalf("GetDeltaAmountBaseUnsigned errored: %s", err.Error())
// 		}
// 		assert.True(t, got.Cmp(big.NewInt(0)) > 0)
// 	})

// 	t.Run("price update from base input:getNextSqrtPriceFromInput", func(t *testing.T) {

// 		newPrice, err := maths.GetNextSqrtPriceFromInput(
// 			maths.Q64(1.0),
// 			big.NewInt(100_000),
// 			big.NewInt(50_000),
// 			false,
// 		)
// 		if err != nil {
// 			t.Fatalf("GetNextSqrtPriceFromInput errored: %s", err.Error())
// 		}
// 		expectedPrice := new(big.Int).Quo(
// 			new(big.Int).Mul(
// 				maths.Q64(1.0), big.NewInt(2),
// 			),
// 			big.NewInt(3),
// 		)

// 		var diff *big.Int
// 		if newPrice.Cmp(expectedPrice) > 0 {
// 			diff = new(big.Int).Sub(newPrice, expectedPrice)
// 		} else {
// 			diff = new(big.Int).Sub(expectedPrice, newPrice)
// 		}

// 		assert.Equal(t, "170141183460469231737836218407120622934", diff.String())
// 	})

// 	t.Run("edge case: identical prices:getDeltaAmountQuoteUnsigned", func(t *testing.T) {
// 		// With identical prices, the delta is zero
// 		got, err := maths.GetDeltaAmountQuoteUnsigned(
// 			maths.Q64(1),
// 			maths.Q64(1),
// 			big.NewInt(1_000),
// 			types.RoundingDown,
// 		)
// 		if err != nil {
// 			t.Fatalf("GetDeltaAmountBaseUnsigned errored: %s", err.Error())
// 		}

// 		assert.True(t, got.Sign() == 0)
// 	})

// 	t.Run("edge case: identical prices:getDeltaAmountQuoteUnsigned", func(t *testing.T) {
// 		// Test for zero price case which should return an error
// 		_, err := maths.GetDeltaAmountBaseUnsigned(
// 			big.NewInt(0),
// 			maths.Q64(1),
// 			big.NewInt(1_000),
// 			types.RoundingDown,
// 		)

// 		assert.NotNil(t, err)
// 		assert.Contains(t, err.Error(), "sqrt price cannot be zero")
// 	})
// }

func TestGetDeltaAmountXXXUnsigned(t *testing.T) {
	p0, p1 := big.NewInt(10312044770285001), big.NewInt(41248173712355948)
	p2, _ := new(big.Int).SetString("79226673521066979257578248091", 10)

	l0, _ := new(big.Int).SetString("10999513467186856574015959876923", 10)
	l1, _ := new(big.Int).SetString("3436021254348803974616125", 10)

	t.Run("getDeltaAmountBaseUnsigned", func(t *testing.T) {
		base1, err := maths.GetDeltaAmountBaseUnsigned(
			p0,
			p1,
			l0,
			types.RoundingUp,
		)
		if err != nil {
			t.Fatalf("GetDeltaAmountBaseUnsigned errored: %s", err.Error())
		}

		base2, err := maths.GetDeltaAmountBaseUnsigned(
			p1,
			p2,
			l1,
			types.RoundingUp,
		)
		if err != nil {
			t.Fatalf("GetDeltaAmountBaseUnsigned errored: %s", err.Error())
		}

		assert.True(t, big.NewInt(799999979174704).Cmp(base1.Add(base1, base2)) == 0)
	})

	t.Run("getDeltaAmountQuoteUnsigned", func(t *testing.T) {
		quote1, err := maths.GetDeltaAmountQuoteUnsigned(
			p0,
			p1,
			l0,
			types.RoundingUp,
		)
		if err != nil {
			t.Fatalf("GetDeltaAmountQuoteUnsigned errored: %s", err.Error())
		}

		quote2, err := maths.GetDeltaAmountQuoteUnsigned(
			p1,
			p2,
			l1,
			types.RoundingUp,
		)
		if err != nil {
			t.Fatalf("GetDeltaAmountQuoteUnsigned errored: %s", err.Error())
		}

		assert.True(t, big.NewInt(799997005061429).Cmp(quote1.Add(quote1, quote2)) == 0)
	})
}
