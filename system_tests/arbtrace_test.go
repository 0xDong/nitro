package arbtest

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/offchainlabs/nitro/arbnode"
	"github.com/offchainlabs/nitro/util/testhelpers"
)

type callTxArgs struct {
	From       *common.Address `json:"from"`
	To         *common.Address `json:"to"`
	Gas        *hexutil.Uint64 `json:"gas"`
	GasPrice   *hexutil.Big    `json:"gasPrice"`
	Value      *hexutil.Big    `json:"value"`
	Data       *hexutil.Bytes  `json:"data"`
	Aggregator *common.Address `json:"aggregator"`
}
type traceAction struct {
	CallType string          `json:"callType,omitempty"`
	From     common.Address  `json:"from"`
	Gas      hexutil.Uint64  `json:"gas"`
	Input    *hexutil.Bytes  `json:"input,omitempty"`
	Init     hexutil.Bytes   `json:"init,omitempty"`
	To       *common.Address `json:"to,omitempty"`
	Value    *hexutil.Big    `json:"value"`
}

type traceCallResult struct {
	Address *common.Address `json:"address,omitempty"`
	Code    *hexutil.Bytes  `json:"code,omitempty"`
	GasUsed hexutil.Uint64  `json:"gasUsed"`
	Output  *hexutil.Bytes  `json:"output,omitempty"`
}

type traceFrame struct {
	Action              traceAction      `json:"action"`
	BlockHash           *hexutil.Bytes   `json:"blockHash,omitempty"`
	BlockNumber         *uint64          `json:"blockNumber,omitempty"`
	Result              *traceCallResult `json:"result,omitempty"`
	Error               *string          `json:"error,omitempty"`
	Subtraces           int              `json:"subtraces"`
	TraceAddress        []int            `json:"traceAddress"`
	TransactionHash     *hexutil.Bytes   `json:"transactionHash,omitempty"`
	TransactionPosition *uint64          `json:"transactionPosition,omitempty"`
	Type                string           `json:"type"`
}

type traceResult struct {
	Output             hexutil.Bytes     `json:"output"`
	StateDiff          *int              `json:"stateDiff"`
	Trace              []traceFrame      `json:"trace"`
	VmTrace            *int              `json:"vmTrace"`
	DestroyedContracts *[]common.Address `json:"destroyedContracts"`
}

type ArbTraceAPIStub struct {
	t *testing.T
}

func (s *ArbTraceAPIStub) Call(ctx context.Context, callArgs callTxArgs, traceTypes []string, blockNum rpc.BlockNumberOrHash) (*traceResult, error) {
	return &traceResult{}, nil
}

func TestArbTraceForwarding(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ipcPath := filepath.Join(t.TempDir(), "redirect.ipc")
	var apis []rpc.API
	apis = append(apis, rpc.API{
		Namespace: "arbtrace",
		Version:   "1.0",
		Service:   &ArbTraceAPIStub{t: t},
		Public:    false,
	})
	listener, srv, err := rpc.StartIPCEndpoint(ipcPath, apis)
	testhelpers.RequireImpl(t, err)
	defer srv.Stop()
	defer listener.Close()

	nodeConfig := arbnode.ConfigDefaultL1Test()
	nodeConfig.RPC.ClassicRedirect = ipcPath
	nodeConfig.RPC.ClassicRedirectTimeout = time.Second
	_, _, _, l2stack, _, _, _, l1stack := createTestNodeOnL1WithConfig(t, ctx, true, nodeConfig, nil, nil)
	defer requireClose(t, l1stack)
	defer requireClose(t, l2stack)

	l2rpc, _ := l2stack.Attach()
	txArgs := callTxArgs{}
	traceTypes := []string{}
	blockNum := rpc.BlockNumberOrHash{}
	var result json.RawMessage
	err = l2rpc.CallContext(ctx, &result, "arbtrace_call", txArgs, traceTypes, blockNum)
	testhelpers.RequireImpl(t, err)
	t.Log(fmt.Sprint(result))
}