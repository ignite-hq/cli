package cosmosclient_test

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/p2p"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ignite/cli/v29/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v29/ignite/pkg/cosmosclient"
	"github.com/ignite/cli/v29/ignite/pkg/cosmosclient/mocks"
	"github.com/ignite/cli/v29/ignite/pkg/cosmosclient/testutil"
	"github.com/ignite/cli/v29/ignite/pkg/cosmosfaucet"
	"github.com/ignite/cli/v29/ignite/pkg/errors"
)

const (
	defaultFaucetDenom     = "token"
	defaultFaucetMinAmount = 100
)

type suite struct {
	rpcClient        *mocks.RPCClient
	accountRetriever *mocks.AccountRetriever
	bankQueryClient  *mocks.BankQueryClient
	gasometer        *mocks.Gasometer
	faucetClient     *mocks.FaucetClient
	signer           *mocks.Signer
}

func newClient(t *testing.T, setup func(suite), opts ...cosmosclient.Option) cosmosclient.Client {
	t.Helper()

	s := suite{
		rpcClient:        mocks.NewRPCClient(t),
		accountRetriever: mocks.NewAccountRetriever(t),
		bankQueryClient:  mocks.NewBankQueryClient(t),
		gasometer:        mocks.NewGasometer(t),
		faucetClient:     mocks.NewFaucetClient(t),
		signer:           mocks.NewSigner(t),
	}
	// Because rpcClient is passed as argument inside clientContext of mocked
	// methods, we must EXPECT a call to String (because testify/mock is calling
	// String() on mocked methods' args)
	s.rpcClient.EXPECT().String().Return("plop").Maybe()
	// cosmosclient.New always makes a call to Status
	s.rpcClient.EXPECT().Status(mock.Anything).
		Return(&ctypes.ResultStatus{
			NodeInfo: p2p.DefaultNodeInfo{Network: "mychain"},
		}, nil).Once()
	if setup != nil {
		setup(s)
	}
	opts = append(opts, []cosmosclient.Option{
		cosmosclient.WithKeyringBackend(cosmosaccount.KeyringMemory),
		cosmosclient.WithRPCClient(s.rpcClient),
		cosmosclient.WithAccountRetriever(s.accountRetriever),
		cosmosclient.WithBankQueryClient(s.bankQueryClient),
		cosmosclient.WithGasometer(s.gasometer),
		cosmosclient.WithFaucetClient(s.faucetClient),
		cosmosclient.WithSigner(s.signer),
	}...)
	c, err := cosmosclient.New(context.Background(), opts...)
	require.NoError(t, err)
	return c
}

func TestNew(t *testing.T) {
	c := newClient(t, nil)

	ctx := c.Context()
	require.Equal(t, "mychain", ctx.ChainID)
	require.NotNil(t, ctx.InterfaceRegistry)
	require.NotNil(t, ctx.Codec)
	require.NotNil(t, ctx.TxConfig)
	require.NotNil(t, ctx.LegacyAmino)
	require.Equal(t, bufio.NewReader(os.Stdin), ctx.Input)
	require.Equal(t, io.Discard, ctx.Output)
	require.NotNil(t, ctx.AccountRetriever)
	require.Equal(t, flags.BroadcastSync, ctx.BroadcastMode)
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	require.Equal(t, home+"/.mychain", ctx.HomeDir)
	require.NotNil(t, ctx.Client)
	require.True(t, ctx.SkipConfirm)
	require.Equal(t, c.AccountRegistry.Keyring, ctx.Keyring)
	require.False(t, ctx.GenerateOnly)
	txf := c.TxFactory
	require.Equal(t, "mychain", txf.ChainID())
	require.Equal(t, c.AccountRegistry.Keyring, txf.Keybase())
	require.EqualValues(t, 300000, txf.Gas())
	require.Equal(t, 1.0, txf.GasAdjustment())
	require.Equal(t, signing.SignMode_SIGN_MODE_UNSPECIFIED, txf.SignMode())
	require.NotNil(t, txf.AccountRetriever())
}

func TestClientWaitForBlockHeight(t *testing.T) {
	targetBlockHeight := int64(42)
	tests := []struct {
		name          string
		timeout       time.Duration
		expectedError string
		setup         func(suite)
	}{
		{
			name: "ok: no wait",
			setup: func(s suite) {
				s.rpcClient.EXPECT().Status(mock.Anything).Return(&ctypes.ResultStatus{
					SyncInfo: ctypes.SyncInfo{LatestBlockHeight: targetBlockHeight},
				}, nil)
			},
		},
		{
			name:    "ok: wait 1 time",
			timeout: time.Second * 2, // must exceed the wait loop duration
			setup: func(s suite) {
				s.rpcClient.EXPECT().Status(mock.Anything).Return(&ctypes.ResultStatus{
					SyncInfo: ctypes.SyncInfo{LatestBlockHeight: targetBlockHeight - 1},
				}, nil).Once()
				s.rpcClient.EXPECT().Status(mock.Anything).Return(&ctypes.ResultStatus{
					SyncInfo: ctypes.SyncInfo{LatestBlockHeight: targetBlockHeight},
				}, nil).Once()
			},
		},
		{
			name:          "fail: wait expired",
			timeout:       time.Millisecond,
			expectedError: "timeout exceeded waiting for block: context deadline exceeded",
			setup: func(s suite) {
				s.rpcClient.EXPECT().Status(mock.Anything).Return(&ctypes.ResultStatus{
					SyncInfo: ctypes.SyncInfo{LatestBlockHeight: targetBlockHeight - 1},
				}, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newClient(t, tt.setup)
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			err := c.WaitForBlockHeight(ctx, targetBlockHeight)

			if tt.expectedError != "" {
				require.EqualError(t, err, tt.expectedError)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestClientWaitForTx(t *testing.T) {
	var (
		ctx          = context.Background()
		hash         = "abcd"
		hashBytes, _ = hex.DecodeString(hash)
		result       = &ctypes.ResultTx{
			Hash: hashBytes,
		}
	)
	tests := []struct {
		name           string
		hash           string
		expectedError  string
		expectedResult *ctypes.ResultTx
		setup          func(suite)
	}{
		{
			name:          "fail: hash not in hex format",
			hash:          "zzz",
			expectedError: "unable to decode tx hash 'zzz': encoding/hex: invalid byte: U+007A 'z'",
		},
		{
			name:           "ok: tx found immediately",
			hash:           hash,
			expectedResult: result,
			setup: func(s suite) {
				s.rpcClient.EXPECT().Tx(ctx, hashBytes, false).Return(result, nil)
			},
		},
		{
			name:          "fail: tx returns an unexpected error",
			hash:          hash,
			expectedError: "fetching tx 'abcd': error while requesting node 'http://localhost:26657': oups",
			setup: func(s suite) {
				s.rpcClient.EXPECT().Tx(ctx, hashBytes, false).Return(nil, errors.New("oups"))
			},
		},
		{
			name:           "ok: tx found after 1 block",
			hash:           hash,
			expectedResult: result,
			setup: func(s suite) {
				// tx is not found
				s.rpcClient.EXPECT().Tx(ctx, hashBytes, false).Return(nil, errors.New("tx abcd not found")).Once()
				// wait for next block
				s.rpcClient.EXPECT().Status(ctx).Return(&ctypes.ResultStatus{
					SyncInfo: ctypes.SyncInfo{LatestBlockHeight: 1},
				}, nil).Once()
				s.rpcClient.EXPECT().Status(ctx).Return(&ctypes.ResultStatus{
					SyncInfo: ctypes.SyncInfo{LatestBlockHeight: 2},
				}, nil).Once()
				// next block reached, check tx again, this time it's found.
				s.rpcClient.EXPECT().Tx(ctx, hashBytes, false).Return(result, nil).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newClient(t, tt.setup)

			res, err := c.WaitForTx(ctx, tt.hash)

			if tt.expectedError != "" {
				require.EqualError(t, err, tt.expectedError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expectedResult, res)
		})
	}
}

func TestClientAccount(t *testing.T) {
	var (
		accountName = "bob"
		passphrase  = "passphrase"
	)
	r, err := cosmosaccount.NewInMemory()
	require.NoError(t, err)
	expectedAccount, _, err := r.Create(accountName)
	require.NoError(t, err)
	expectedAddr, err := expectedAccount.Address("cosmos")
	require.NoError(t, err)
	// Export created account to we can import it in the Client below.
	key, err := r.Export(accountName, passphrase)
	require.NoError(t, err)

	tests := []struct {
		name          string
		addressOrName string
		expectedError string
	}{
		{
			name:          "ok: find by name",
			addressOrName: expectedAccount.Name,
		},
		{
			name:          "ok: find by address",
			addressOrName: expectedAddr,
		},
		{
			name:          "fail: name not found",
			addressOrName: "unknown",
			expectedError: "decoding bech32 failed: invalid bech32 string length 7",
		},
		{
			name:          "fail: address not found",
			addressOrName: "cosmos1cs4hpwrpna6ucsgsa78jfp403l7gdynukrxkrv",
			expectedError: `account "cosmos1cs4hpwrpna6ucsgsa78jfp403l7gdynukrxkrv" does not exist`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newClient(t, nil)
			_, err := c.AccountRegistry.Import(accountName, key, passphrase)
			require.NoError(t, err)

			account, err := c.Account(tt.addressOrName)

			if tt.expectedError != "" {
				require.EqualError(t, err, tt.expectedError)
				return
			}
			require.NoError(t, err)
			require.Equal(t, expectedAccount, account)
		})
	}
}

func TestClientAddress(t *testing.T) {
	var (
		accountName = "bob"
		passphrase  = "passphrase"
	)
	r, err := cosmosaccount.NewInMemory()
	require.NoError(t, err)
	expectedAccount, _, err := r.Create(accountName)
	require.NoError(t, err)
	// Export created account to we can import it in the Client below.
	key, err := r.Export(accountName, passphrase)
	require.NoError(t, err)

	tests := []struct {
		name           string
		accountName    string
		opts           []cosmosclient.Option
		expectedError  string
		expectedPrefix string
	}{
		{
			name:           "ok: name exists",
			accountName:    expectedAccount.Name,
			expectedPrefix: "cosmos",
		},
		{
			name: "ok: name exists with different prefix",
			opts: []cosmosclient.Option{
				cosmosclient.WithAddressPrefix("test"),
			},
			accountName:    expectedAccount.Name,
			expectedPrefix: "test",
		},
		{
			name:          "fail: name not found",
			accountName:   "unknown",
			expectedError: `account "unknown" does not exist`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newClient(t, nil, tt.opts...)
			_, err := c.AccountRegistry.Import(accountName, key, passphrase)
			require.NoError(t, err)

			address, err := c.Address(tt.accountName)

			if tt.expectedError != "" {
				require.EqualError(t, err, tt.expectedError)
				return
			}
			require.NoError(t, err)
			expectedAddr, err := expectedAccount.Address(tt.expectedPrefix)
			require.NoError(t, err)
			require.Equal(t, expectedAddr, address)
		})
	}
}

func TestClientStatus(t *testing.T) {
	var (
		ctx            = context.Background()
		expectedStatus = &ctypes.ResultStatus{
			NodeInfo: p2p.DefaultNodeInfo{Network: "mychain"},
		}
	)
	tests := []struct {
		name          string
		expectedError string
		setup         func(suite)
	}{
		{
			name: "ok",
			setup: func(s suite) {
				s.rpcClient.EXPECT().Status(ctx).Return(expectedStatus, nil).Once()
			},
		},
		{
			name:          "fail",
			expectedError: "error while requesting node 'http://localhost:26657': oups",
			setup: func(s suite) {
				s.rpcClient.EXPECT().Status(ctx).Return(expectedStatus, errors.New("oups")).Once()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newClient(t, tt.setup)

			status, err := c.Status(ctx)

			if tt.expectedError != "" {
				require.EqualError(t, err, tt.expectedError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, expectedStatus, status)
		})
	}
}

func TestClientCreateTx(t *testing.T) {
	var (
		ctx         = context.Background()
		accountName = "bob"
		passphrase  = "passphrase"
	)
	r, err := cosmosaccount.NewInMemory()
	require.NoError(t, err)
	a, _, err := r.Create(accountName)
	require.NoError(t, err)
	// Export created account to we can import it in the Client below.
	key, err := r.Export(accountName, passphrase)
	require.NoError(t, err)
	sdkaddr, err := a.Record.GetAddress()
	require.NoError(t, err)

	tests := []struct {
		name           string
		opts           []cosmosclient.Option
		msg            sdktypes.Msg
		expectedJSONTx string
		expectedError  string
		setup          func(s suite)
	}{
		{
			name:          "fail: account doesn't exist",
			expectedError: "nope",
			setup: func(s suite) {
				s.accountRetriever.EXPECT().
					EnsureExists(mock.Anything, sdkaddr).Return(errors.New("nope"))
			},
		},
		{
			name: "ok: with default values",
			msg: &banktypes.MsgSend{
				FromAddress: "from",
				ToAddress:   "to",
				Amount: sdktypes.NewCoins(
					sdktypes.NewCoin("token", math.NewIntFromUint64(1)),
				),
			},
			expectedJSONTx: `{"body":{"messages":[{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"from","to_address":"to","amount":[{"denom":"token","amount":"1"}]}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"300000","payer":"","granter":""},"tip":null},"signatures":[]}`,
			setup: func(s suite) {
				s.expectPrepareFactory(sdkaddr)
			},
		},
		{
			name: "ok: with faucet enabled, account balance is high enough",
			opts: []cosmosclient.Option{
				cosmosclient.WithUseFaucet("localhost:1234", "", 0),
			},
			msg: &banktypes.MsgSend{
				FromAddress: "from",
				ToAddress:   "to",
				Amount: sdktypes.NewCoins(
					sdktypes.NewCoin("token", math.NewIntFromUint64(1)),
				),
			},
			expectedJSONTx: `{"body":{"messages":[{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"from","to_address":"to","amount":[{"denom":"token","amount":"1"}]}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"300000","payer":"","granter":""},"tip":null},"signatures":[]}`,
			setup: func(s suite) {
				s.expectMakeSureAccountHasToken(sdkaddr.String(), defaultFaucetMinAmount)

				s.expectPrepareFactory(sdkaddr)
			},
		},
		{
			name: "ok: with faucet enabled, account balance is too low",
			opts: []cosmosclient.Option{
				cosmosclient.WithUseFaucet("localhost:1234", "", 0),
			},
			msg: &banktypes.MsgSend{
				FromAddress: "from",
				ToAddress:   "to",
				Amount: sdktypes.NewCoins(
					sdktypes.NewCoin("token", math.NewIntFromUint64(1)),
				),
			},
			expectedJSONTx: `{"body":{"messages":[{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"from","to_address":"to","amount":[{"denom":"token","amount":"1"}]}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"300000","payer":"","granter":""},"tip":null},"signatures":[]}`,
			setup: func(s suite) {
				s.expectMakeSureAccountHasToken(sdkaddr.String(), defaultFaucetMinAmount-1)
				s.expectPrepareFactory(sdkaddr)
			},
		},
		{
			name: "ok: with fees",
			opts: []cosmosclient.Option{
				cosmosclient.WithFees("10token"),
			},
			msg: &banktypes.MsgSend{
				FromAddress: "from",
				ToAddress:   "to",
				Amount: sdktypes.NewCoins(
					sdktypes.NewCoin("token", math.NewIntFromUint64(1)),
				),
			},
			expectedJSONTx: `{"body":{"messages":[{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"from","to_address":"to","amount":[{"denom":"token","amount":"1"}]}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[{"denom":"token","amount":"10"}],"gas_limit":"300000","payer":"","granter":""},"tip":null},"signatures":[]}`,
			setup: func(s suite) {
				s.expectPrepareFactory(sdkaddr)
			},
		},
		{
			name: "ok: with gas price",
			opts: []cosmosclient.Option{
				// Should set fees to 3*defaultGasLimit
				cosmosclient.WithGasPrices("3token"),
			},
			msg: &banktypes.MsgSend{
				FromAddress: "from",
				ToAddress:   "to",
				Amount: sdktypes.NewCoins(
					sdktypes.NewCoin("token", math.NewIntFromUint64(1)),
				),
			},
			expectedJSONTx: `{"body":{"messages":[{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"from","to_address":"to","amount":[{"denom":"token","amount":"1"}]}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[{"denom":"token","amount":"900000"}],"gas_limit":"300000","payer":"","granter":""},"tip":null},"signatures":[]}`,
			setup: func(s suite) {
				s.expectPrepareFactory(sdkaddr)
			},
		},
		{
			name: "fail: with fees, gas prices and gas adjustment",
			opts: []cosmosclient.Option{
				cosmosclient.WithFees("10token"),
				cosmosclient.WithGasPrices("3token"),
				cosmosclient.WithGasAdjustment(2.1),
			},
			msg: &banktypes.MsgSend{
				FromAddress: "from",
				ToAddress:   "to",
				Amount: sdktypes.NewCoins(
					sdktypes.NewCoin("token", math.NewIntFromUint64(1)),
				),
			},
			expectedError: "cannot provide both fees and gas prices",
			setup: func(s suite) {
				s.expectPrepareFactory(sdkaddr)
			},
		},
		{
			name: "ok: without empty gas limit",
			opts: []cosmosclient.Option{
				cosmosclient.WithGas(""),
			},
			msg: &banktypes.MsgSend{
				FromAddress: "from",
				ToAddress:   "to",
				Amount: sdktypes.NewCoins(
					sdktypes.NewCoin("token", math.NewIntFromUint64(1)),
				),
			},
			expectedJSONTx: `{"body":{"messages":[{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"from","to_address":"to","amount":[{"denom":"token","amount":"1"}]}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"20042","payer":"","granter":""},"tip":null},"signatures":[]}`,
			setup: func(s suite) {
				s.expectPrepareFactory(sdkaddr)
				s.gasometer.EXPECT().
					CalculateGas(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, 42, nil)
			},
		},
		{
			name: "ok: without auto gas limit",
			opts: []cosmosclient.Option{
				cosmosclient.WithGas("auto"),
			},
			msg: &banktypes.MsgSend{
				FromAddress: "from",
				ToAddress:   "to",
				Amount: sdktypes.NewCoins(
					sdktypes.NewCoin("token", math.NewIntFromUint64(1)),
				),
			},
			expectedJSONTx: `{"body":{"messages":[{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"from","to_address":"to","amount":[{"denom":"token","amount":"1"}]}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"20042","payer":"","granter":""},"tip":null},"signatures":[]}`,
			setup: func(s suite) {
				s.expectPrepareFactory(sdkaddr)
				s.gasometer.EXPECT().
					CalculateGas(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, 42, nil)
			},
		},
		{
			name: "ok: with gas adjustment",
			opts: []cosmosclient.Option{
				cosmosclient.WithGasAdjustment(2.4),
			},
			msg: &banktypes.MsgSend{
				FromAddress: "from",
				ToAddress:   "to",
				Amount: sdktypes.NewCoins(
					sdktypes.NewCoin("token", math.NewIntFromUint64(1)),
				),
			},
			expectedJSONTx: `{"body":{"messages":[{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"from","to_address":"to","amount":[{"denom":"token","amount":"1"}]}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"300000","payer":"","granter":""},"tip":null},"signatures":[]}`,
			setup: func(s suite) {
				s.expectPrepareFactory(sdkaddr)
			},
		},
		{
			name: "ok: without gas price and zero gas adjustment",
			opts: []cosmosclient.Option{
				cosmosclient.WithGas("auto"),
				cosmosclient.WithGasAdjustment(0),
			},
			msg: &banktypes.MsgSend{
				FromAddress: "from",
				ToAddress:   "to",
				Amount: sdktypes.NewCoins(
					sdktypes.NewCoin("token", math.NewIntFromUint64(1)),
				),
			},
			expectedJSONTx: `{"body":{"messages":[{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"from","to_address":"to","amount":[{"denom":"token","amount":"1"}]}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"20042","payer":"","granter":""},"tip":null},"signatures":[]}
`,
			setup: func(s suite) {
				s.expectPrepareFactory(sdkaddr)
				s.gasometer.EXPECT().
					CalculateGas(mock.Anything, mock.Anything, mock.Anything).
					Return(nil, 42, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newClient(t, tt.setup, tt.opts...)
			account, err := c.AccountRegistry.Import(accountName, key, passphrase)
			require.NoError(t, err)

			txs, err := c.CreateTx(ctx, account, tt.msg)

			if tt.expectedError != "" {
				require.EqualError(t, err, tt.expectedError)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, txs)
			bz, err := txs.EncodeJSON()
			require.NoError(t, err)
			require.JSONEq(t, tt.expectedJSONTx, string(bz))
		})
	}
}

func TestGetBlockTXs(t *testing.T) {
	m := testutil.NewTendermintClientMock(t)
	ctx := context.Background()

	// Mock the Block RPC endpoint
	block := createTestBlock(1)

	m.On("Block", ctx, &block.Height).Return(&ctypes.ResultBlock{Block: &block}, nil)

	// Mock the TxSearch RPC endpoint
	searchQry := fmt.Sprintf("tx.height=%d", block.Height)
	page := 1
	perPage := 30
	rtx := ctypes.ResultTx{}
	resSearch := ctypes.ResultTxSearch{
		Txs:        []*ctypes.ResultTx{&rtx},
		TotalCount: 1,
	}

	m.On("TxSearch", ctx, searchQry, false, &page, &perPage, "asc").Return(&resSearch, nil)

	// Create a cosmos client that uses the RPC mock
	client := cosmosclient.Client{RPC: m}

	txs, err := client.GetBlockTXs(ctx, block.Height)

	// Assert
	require.NoError(t, err)
	require.Equal(t, txs, []cosmosclient.TX{
		{
			BlockTime: block.Time,
			Raw:       &rtx,
		},
	})

	m.AssertNumberOfCalls(t, "Block", 1)
	m.AssertNumberOfCalls(t, "TxSearch", 1)
}

func TestGetBlockTXsWithBlockError(t *testing.T) {
	m := testutil.NewTendermintClientMock(t)

	wantErr := errors.New("expected error")

	// Mock the Block RPC endpoint to return an error
	m.OnBlock().Return(nil, wantErr)

	// Create a cosmos client that uses the RPC mock
	client := cosmosclient.Client{RPC: m}

	txs, err := client.GetBlockTXs(context.Background(), 1)

	// Assert
	require.ErrorIs(t, err, wantErr)
	require.Nil(t, txs)

	m.AssertNumberOfCalls(t, "Block", 1)
	m.AssertNumberOfCalls(t, "TxSearch", 0)
}

func TestGetBlockTXsPagination(t *testing.T) {
	m := testutil.NewTendermintClientMock(t)

	// Mock the Block RPC endpoint
	block := createTestBlock(1)

	m.OnBlock().Return(&ctypes.ResultBlock{Block: &block}, nil)

	// Mock the TxSearch RPC endpoint and fake the number of
	// transactions, so it is called twice to fetch two pages
	ctx := context.Background()
	searchQry := fmt.Sprintf("tx.height=%d", block.Height)
	perPage := 30
	fakeCount := perPage + 1
	first := 1
	second := 2
	firstPage := ctypes.ResultTxSearch{
		Txs:        []*ctypes.ResultTx{{}},
		TotalCount: fakeCount,
	}
	secondPage := ctypes.ResultTxSearch{
		Txs:        []*ctypes.ResultTx{{}},
		TotalCount: fakeCount,
	}

	m.On("TxSearch", ctx, searchQry, false, &first, &perPage, "asc").Return(&firstPage, nil)
	m.On("TxSearch", ctx, searchQry, false, &second, &perPage, "asc").Return(&secondPage, nil)

	// Create a cosmos client that uses the RPC mock
	client := cosmosclient.Client{RPC: m}

	txs, err := client.GetBlockTXs(ctx, block.Height)

	// Assert
	require.NoError(t, err)
	require.Equal(t, txs, []cosmosclient.TX{
		{
			BlockTime: block.Time,
			Raw:       firstPage.Txs[0],
		},
		{
			BlockTime: block.Time,
			Raw:       secondPage.Txs[0],
		},
	})

	m.AssertNumberOfCalls(t, "Block", 1)
	m.AssertNumberOfCalls(t, "TxSearch", 2)
}

func TestGetBlockTXsWithSearchError(t *testing.T) {
	m := testutil.NewTendermintClientMock(t)

	wantErr := errors.New("expected error")

	// Mock the Block RPC endpoint
	block := createTestBlock(1)

	m.OnBlock().Return(&ctypes.ResultBlock{Block: &block}, nil)

	// Mock the TxSearch RPC endpoint to return an error
	m.OnTxSearch().Return(nil, wantErr)

	// Create a cosmos client that uses the RPC mock
	client := cosmosclient.Client{RPC: m}

	txs, err := client.GetBlockTXs(context.Background(), block.Height)

	// Assert
	require.ErrorIs(t, err, wantErr)
	require.Nil(t, txs)

	m.AssertNumberOfCalls(t, "Block", 1)
	m.AssertNumberOfCalls(t, "TxSearch", 1)
}

func TestCollectTXs(t *testing.T) {
	m := testutil.NewTendermintClientMock(t)
	ctx := context.Background()

	// Mock the Status RPC endpoint to report that only two blocks exists
	status := ctypes.ResultStatus{
		SyncInfo: ctypes.SyncInfo{
			LatestBlockHeight: 2,
		},
	}

	m.On("Status", ctx).Return(&status, nil)

	// Mock the Block RPC endpoint to return two blocks
	b1 := createTestBlock(1)
	b2 := createTestBlock(2)

	m.On("Block", ctx, &b1.Height).Return(&ctypes.ResultBlock{Block: &b1}, nil)
	m.On("Block", ctx, &b2.Height).Return(&ctypes.ResultBlock{Block: &b2}, nil)

	// Mock the TxSearch RPC endpoint to return each of the two block.
	// Transactions are empty because only the pointer address is required to assert.
	page := 1
	perPage := 30
	q1 := "tx.height=1"
	r1 := ctypes.ResultTxSearch{
		Txs:        []*ctypes.ResultTx{{}},
		TotalCount: 1,
	}
	q2 := "tx.height=2"
	r2 := ctypes.ResultTxSearch{
		Txs:        []*ctypes.ResultTx{{}, {}},
		TotalCount: 2,
	}

	m.On("TxSearch", ctx, q1, false, &page, &perPage, "asc").Return(&r1, nil)
	m.On("TxSearch", ctx, q2, false, &page, &perPage, "asc").Return(&r2, nil)

	// Prepare expected values
	wantTXs := []cosmosclient.TX{
		{
			BlockTime: b1.Time,
			Raw:       r1.Txs[0],
		},
		{
			BlockTime: b2.Time,
			Raw:       r2.Txs[0],
		},
		{
			BlockTime: b2.Time,
			Raw:       r2.Txs[1],
		},
	}

	// Create a cosmos client that uses the RPC mock
	client := cosmosclient.Client{RPC: m}

	// Create a channel to receive the transactions from the two blocks.
	// The channel must be closed after the call to collect.
	tc := make(chan []cosmosclient.TX)

	// Collect all transactions
	var (
		txs  []cosmosclient.TX
		open bool
	)

	finished := make(chan struct{})
	go func() {
		defer close(finished)

		for t := range tc {
			txs = append(txs, t...)
		}
	}()

	err := client.CollectTXs(ctx, 1, tc)

	select {
	case <-time.After(time.Second):
		t.Fatal("expected CollectTXs to finish sooner")
	case <-finished:
	}

	select {
	case _, open = <-tc:
	default:
	}

	// Assert
	require.NoError(t, err)
	require.Equal(t, wantTXs, txs)
	require.False(t, open, "expected transaction channel to be closed")
}

func TestCollectTXsWithStatusError(t *testing.T) {
	m := testutil.NewTendermintClientMock(t)

	wantErr := errors.New("expected error")

	// Mock the Status RPC endpoint to return an error
	m.OnStatus().Return(nil, wantErr)

	// Create a cosmos client that uses the RPC mock
	client := cosmosclient.Client{RPC: m}

	// Create a channel to receive the transactions from the two blocks.
	// The channel must be closed after the call to collect.
	tc := make(chan []cosmosclient.TX)

	open := false
	ctx := context.Background()
	err := client.CollectTXs(ctx, 1, tc)

	select {
	case _, open = <-tc:
	default:
	}

	// Assert
	require.ErrorIs(t, err, wantErr)
	require.False(t, open, "expected transaction channel to be closed")
}

func TestCollectTXsWithBlockError(t *testing.T) {
	m := testutil.NewTendermintClientMock(t)

	wantErr := errors.New("expected error")

	// Mock the Status RPC endpoint
	status := ctypes.ResultStatus{
		SyncInfo: ctypes.SyncInfo{
			LatestBlockHeight: 1,
		},
	}

	m.OnStatus().Return(&status, nil)

	// Mock the Block RPC endpoint to return an error
	m.OnBlock().Return(nil, wantErr)

	// Create a cosmos client that uses the RPC mock
	client := cosmosclient.Client{RPC: m}

	// Create a channel to receive the transactions from the two blocks.
	// The channel must be closed after the call to collect.
	tc := make(chan []cosmosclient.TX)

	open := false
	ctx := context.Background()
	err := client.CollectTXs(ctx, 1, tc)

	select {
	case _, open = <-tc:
	default:
	}

	// Assert
	require.ErrorIs(t, err, wantErr)
	require.False(t, open, "expected transaction channel to be closed")
}

func TestCollectTXsWithContextDone(t *testing.T) {
	m := testutil.NewTendermintClientMock(t)

	// Mock the Status RPC endpoint
	status := ctypes.ResultStatus{
		SyncInfo: ctypes.SyncInfo{
			LatestBlockHeight: 1,
		},
	}

	m.OnStatus().Return(&status, nil)

	// Mock the Block RPC endpoint
	block := createTestBlock(1)

	m.OnBlock().Return(&ctypes.ResultBlock{Block: &block}, nil)

	// Mock the TxSearch RPC endpoint
	rs := ctypes.ResultTxSearch{
		Txs:        []*ctypes.ResultTx{{}},
		TotalCount: 1,
	}

	m.OnTxSearch().Return(&rs, nil)

	// Create a cosmos client that uses the RPC mock
	client := cosmosclient.Client{RPC: m}

	// Create a channel to receive the transactions from the two blocks.
	// The channel must be closed after the call to collect.
	tc := make(chan []cosmosclient.TX)

	// Create a context and cancel it so the collect call finishes because the context is done
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	open := false
	err := client.CollectTXs(ctx, 1, tc)

	select {
	case _, open = <-tc:
	default:
	}

	// Assert
	require.ErrorIs(t, err, ctx.Err())
	require.False(t, open, "expected transaction channel to be closed")
}

func (s suite) expectMakeSureAccountHasToken(address string, balance int64) {
	currentBalance := sdktypes.NewInt64Coin(defaultFaucetDenom, balance)
	s.bankQueryClient.EXPECT().Balance(
		context.Background(),
		&banktypes.QueryBalanceRequest{
			Address: address,
			Denom:   defaultFaucetDenom,
		},
	).Return(
		&banktypes.QueryBalanceResponse{
			Balance: &currentBalance,
		},
		nil,
	).Once()
	if balance >= defaultFaucetMinAmount {
		// balance is high enough, faucet won't be called
		return
	}

	s.faucetClient.EXPECT().Transfer(context.Background(),
		cosmosfaucet.TransferRequest{AccountAddress: address},
	).Return(
		cosmosfaucet.TransferResponse{}, nil,
	)

	newBalance := sdktypes.NewInt64Coin(defaultFaucetDenom, defaultFaucetMinAmount)
	s.bankQueryClient.EXPECT().Balance(
		mock.Anything,
		&banktypes.QueryBalanceRequest{
			Address: address,
			Denom:   defaultFaucetDenom,
		},
	).Return(
		&banktypes.QueryBalanceResponse{
			Balance: &newBalance,
		},
		nil,
	).Once()
}

func (s suite) expectPrepareFactory(sdkaddr sdktypes.Address) {
	s.accountRetriever.EXPECT().
		EnsureExists(mock.Anything, sdkaddr).
		Return(nil)
	s.accountRetriever.EXPECT().
		GetAccountNumberSequence(mock.Anything, sdkaddr).
		Return(1, 2, nil)
}

func createTestBlock(height int64) tmtypes.Block {
	return tmtypes.Block{
		Header: tmtypes.Header{
			Height: height,
			Time:   time.Now(),
		},
	}
}
