package testUtils

import (
	"errors"
	"strings"
	"testing"

	"github.com/gagliardetto/solana-go/rpc/jsonrpc"
)

func PrettyPrintTxnErrorLog(t *testing.T, err error) {
	if err == nil {
		return
	}

	// ── 1.  Try to locate an *jsonrpc.RPCError anywhere in the chain ────────────
	var rpcErr *jsonrpc.RPCError
	if errors.As(err, &rpcErr) {
		t.Logf("► RPC-error code %d\n", rpcErr.Code)
		printLogLines(t, rpcErr.Message)
		return
	}

	// ── 2.  Fallback: use the raw error string we already have ──────────────────
	t.Log("► raw error:")
	printLogLines(t, err.Error())
}

// helper: split by '\n' literals and print each trimmed line
func printLogLines(t *testing.T, msg string) {
	for _, ln := range strings.Split(msg, `\n`) { // literal backslash-n
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		t.Log(ln)
	}
}
