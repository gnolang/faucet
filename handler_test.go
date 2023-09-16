package faucet

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gnolang/faucet/config"
	"github.com/gnolang/faucet/estimate/static"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	coreTypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func decodeResponse[T Response | Responses](t *testing.T, responseBody []byte) *T {
	t.Helper()

	var response *T

	require.NoError(t, json.NewDecoder(bytes.NewReader(responseBody)).Decode(&response))

	return response
}

// getFreePort fetches a currently free port on the OS
func getFreePort(t *testing.T) int {
	t.Helper()

	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	return l.Addr().(*net.TCPAddr).Port
}

func waitForServer(t *testing.T, URL string, timeout time.Duration) {
	t.Helper()

	ch := make(chan bool)
	go func() {
		for {
			if _, err := http.Get(URL); err == nil {
				ch <- true

			}

			time.Sleep(10 * time.Millisecond)
		}
	}()

	select {
	case <-ch:
		return
	case <-time.After(timeout):
		t.Fatalf("server wait timeout exceeded")
	}
}

func TestFaucet_Serve_ValidRequests(t *testing.T) {
	t.Parallel()

	var (
		gasFee     = std.MustParseCoin(config.DefaultGasFee)
		sendAmount = std.MustParseCoins(config.DefaultSendAmount)
	)

	var (
		validAddress = crypto.MustAddressFromString("g155n659f89cfak0zgy575yqma64sm4tv6exqk99")

		singleValidRequest = Request{
			To: validAddress.String(),
		}

		bulkValidRequests = Requests{
			singleValidRequest,
			singleValidRequest,
		}
	)

	encodedSingleValidRequest, err := json.Marshal(
		singleValidRequest,
	)
	require.NoError(t, err)

	encodedBulkValidRequests, err := json.Marshal(
		bulkValidRequests,
	)
	require.NoError(t, err)

	getFaucetURL := func(address string) string {
		return fmt.Sprintf("http://%s", address)
	}

	testTable := []struct {
		name              string
		expectedNumTxs    int
		requestValidateFn func(response []byte)
		request           []byte
	}{
		{
			"single request",
			1,
			func(resp []byte) {
				response := decodeResponse[Response](t, resp)

				assert.Empty(t, response.Error)
				assert.Equal(t, faucetSuccess, response.Result)
			},
			encodedSingleValidRequest,
		},
		{
			"bulk request",
			len(bulkValidRequests),
			func(resp []byte) {
				responses := decodeResponse[Responses](t, resp)
				require.Len(t, *responses, len(bulkValidRequests))

				for _, response := range *responses {
					assert.Empty(t, response.Error)
					assert.Equal(t, faucetSuccess, response.Result)
				}
			},
			encodedBulkValidRequests,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				requiredFunds = sendAmount.Add(std.NewCoins(gasFee))
				fundAccount   = crypto.Address{1}

				broadcastResponse = &coreTypes.ResultBroadcastTxCommit{
					CheckTx: abci.ResponseCheckTx{
						ResponseBase: abci.ResponseBase{
							Error: nil, // no error
						},
					},
					DeliverTx: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Error: nil, // no error
						},
					},
				}

				signature   = []byte("signature")
				capturedTxs []*std.Tx

				mockPubKey = &mockPubKey{
					addressFn: func() crypto.Address {
						return fundAccount
					},
				}
				mockPrivKey = &mockPrivKey{
					signFn: func(_ []byte) ([]byte, error) {
						return signature, nil
					},
					pubKeyFn: func() crypto.PubKey {
						return mockPubKey
					},
				}
				mockKeyring = &mockKeyring{
					getAddressesFn: func() []crypto.Address {
						return []crypto.Address{
							fundAccount,
						}
					},
					getKeyFn: func(address crypto.Address) crypto.PrivKey {
						if address == fundAccount {
							return mockPrivKey
						}

						return nil
					},
				}
				mockAccount = &mockAccount{
					getAddressFn: func() crypto.Address {
						return fundAccount
					},
					getCoinsFn: func() std.Coins {
						return requiredFunds // enough funds
					},
				}
				mockClient = &mockClient{
					getAccountFn: func(address crypto.Address) (std.Account, error) {
						if address == fundAccount {
							return mockAccount, nil
						}

						return nil, errors.New("account not found")
					},
					sendTransactionCommitFn: func(tx *std.Tx) (*coreTypes.ResultBroadcastTxCommit, error) {
						capturedTxs = append(capturedTxs, tx)

						return broadcastResponse, nil
					},
				}
			)

			// Create a new faucet with default params
			cfg := config.DefaultConfig()
			cfg.ListenAddress = fmt.Sprintf("127.0.0.1:%d", getFreePort(t))

			f, err := NewFaucet(
				static.New(gasFee, 100000),
				mockClient,
				WithConfig(cfg),
			)

			require.NoError(t, err)
			require.NotNil(t, f)

			// Update the keyring
			f.keyring = mockKeyring

			// Start the faucet
			ctx, cancelFn := context.WithCancel(context.Background())
			defer cancelFn()

			g, gCtx := errgroup.WithContext(ctx)

			g.Go(func() error {
				return f.Serve(gCtx)
			})

			url := getFaucetURL(f.config.ListenAddress)

			// Wait for the faucet to be started
			waitForServer(t, url, time.Second*5)

			// Execute the request
			respRaw, err := http.Post(
				url,
				jsonMimeType,
				bytes.NewBuffer(testCase.request),
			)
			require.NoError(t, err)

			respBytes, err := io.ReadAll(respRaw.Body)
			require.NoError(t, err)

			testCase.requestValidateFn(respBytes)

			// Stop the faucet and wait for it to finish
			cancelFn()
			assert.NoError(t, g.Wait())

			// Validate the broadcast txs
			require.Len(t, capturedTxs, testCase.expectedNumTxs)

			for _, capturedTx := range capturedTxs {
				assert.Len(t, capturedTx.Signatures, 1)
				require.Len(t, capturedTx.Msgs, 1)

				msg := capturedTx.Msgs[0]

				// Make sure the message is a send message
				msgSend, ok := msg.(bank.MsgSend)
				require.True(t, ok)

				assert.Equal(t, fundAccount, msgSend.FromAddress)
				assert.Equal(t, sendAmount, msgSend.Amount)
			}
		})
	}
}

func TestFaucet_Serve_InvalidRequests(t *testing.T) {
	t.Parallel()

	var (
		gasFee     = std.MustParseCoin(config.DefaultGasFee)
		sendAmount = std.MustParseCoins(config.DefaultSendAmount)
	)

	var (
		invalidAddress = "invalid-address"

		singleInvalidRequest = Request{
			To: invalidAddress,
		}

		bulkInvalidRequests = Requests{
			singleInvalidRequest,
			Request{
				To: "", // empty address
			},
		}
	)

	encodedSingleInvalidRequest, err := json.Marshal(
		singleInvalidRequest,
	)
	require.NoError(t, err)

	encodedBulkInvalidRequests, err := json.Marshal(
		bulkInvalidRequests,
	)
	require.NoError(t, err)

	getFaucetURL := func(address string) string {
		return fmt.Sprintf("http://%s", address)
	}

	testTable := []struct {
		name              string
		expectedNumTxs    int
		requestValidateFn func(response []byte)
		request           []byte
	}{
		{
			"single request",
			1,
			func(resp []byte) {
				response := decodeResponse[Response](t, resp)

				assert.Contains(t, response.Error, errInvalidBeneficiary.Error())
				assert.Equal(t, response.Result, unableToHandleRequest)
			},
			encodedSingleInvalidRequest,
		},
		{
			"bulk request",
			len(bulkInvalidRequests),
			func(resp []byte) {
				responses := decodeResponse[Responses](t, resp)
				require.Len(t, *responses, len(bulkInvalidRequests))

				for _, response := range *responses {
					assert.Contains(t, response.Error, errInvalidBeneficiary.Error())
					assert.Equal(t, response.Result, unableToHandleRequest)
				}
			},
			encodedBulkInvalidRequests,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				requiredFunds = sendAmount.Add(std.NewCoins(gasFee))
				fundAccount   = crypto.Address{1}

				broadcastResponse = &coreTypes.ResultBroadcastTxCommit{
					CheckTx: abci.ResponseCheckTx{
						ResponseBase: abci.ResponseBase{
							Error: nil, // no error
						},
					},
					DeliverTx: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{
							Error: nil, // no error
						},
					},
				}

				signature   = []byte("signature")
				capturedTxs []*std.Tx

				mockPubKey = &mockPubKey{
					addressFn: func() crypto.Address {
						return fundAccount
					},
				}
				mockPrivKey = &mockPrivKey{
					signFn: func(_ []byte) ([]byte, error) {
						return signature, nil
					},
					pubKeyFn: func() crypto.PubKey {
						return mockPubKey
					},
				}
				mockKeyring = &mockKeyring{
					getAddressesFn: func() []crypto.Address {
						return []crypto.Address{
							fundAccount,
						}
					},
					getKeyFn: func(address crypto.Address) crypto.PrivKey {
						if address == fundAccount {
							return mockPrivKey
						}

						return nil
					},
				}
				mockAccount = &mockAccount{
					getAddressFn: func() crypto.Address {
						return fundAccount
					},
					getCoinsFn: func() std.Coins {
						return requiredFunds // enough funds
					},
				}
				mockClient = &mockClient{
					getAccountFn: func(address crypto.Address) (std.Account, error) {
						if address == fundAccount {
							return mockAccount, nil
						}

						return nil, errors.New("account not found")
					},
					sendTransactionCommitFn: func(tx *std.Tx) (*coreTypes.ResultBroadcastTxCommit, error) {
						capturedTxs = append(capturedTxs, tx)

						return broadcastResponse, nil
					},
				}
			)

			// Create a new faucet with default params
			cfg := config.DefaultConfig()
			cfg.ListenAddress = fmt.Sprintf("127.0.0.1:%d", getFreePort(t))

			f, err := NewFaucet(
				static.New(gasFee, 100000),
				mockClient,
				WithConfig(cfg),
			)

			require.NoError(t, err)
			require.NotNil(t, f)

			// Update the keyring
			f.keyring = mockKeyring

			// Start the faucet
			ctx, cancelFn := context.WithCancel(context.Background())
			defer cancelFn()

			g, gCtx := errgroup.WithContext(ctx)

			g.Go(func() error {
				return f.Serve(gCtx)
			})

			url := getFaucetURL(f.config.ListenAddress)

			// Wait for the faucet to be started
			waitForServer(t, url, time.Second*5)

			// Execute the request
			respRaw, err := http.Post(
				url,
				jsonMimeType,
				bytes.NewBuffer(testCase.request),
			)
			require.NoError(t, err)

			respBytes, err := io.ReadAll(respRaw.Body)
			require.NoError(t, err)

			testCase.requestValidateFn(respBytes)

			// Stop the faucet and wait for it to finish
			cancelFn()
			assert.NoError(t, g.Wait())

			// Validate the broadcast tx
			assert.Nil(t, capturedTxs)
		})
	}
}

func TestFaucet_Serve_MalformedRequests(t *testing.T) {
	t.Parallel()

	getFaucetURL := func(address string) string {
		return fmt.Sprintf("http://%s", address)
	}

	testTable := []struct {
		name    string
		request []byte
	}{
		{
			"nil request",
			nil,
		},
		{
			"empty request",
			[]byte{},
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create a new faucet with default params
			cfg := config.DefaultConfig()
			cfg.ListenAddress = fmt.Sprintf("127.0.0.1:%d", getFreePort(t))

			f, err := NewFaucet(
				static.New(std.MustParseCoin(config.DefaultGasFee), 100000),
				&mockClient{},
				WithConfig(cfg),
			)

			require.NoError(t, err)
			require.NotNil(t, f)

			// Start the faucet
			ctx, cancelFn := context.WithCancel(context.Background())
			defer cancelFn()

			g, gCtx := errgroup.WithContext(ctx)

			g.Go(func() error {
				return f.Serve(gCtx)
			})

			url := getFaucetURL(f.config.ListenAddress)

			// Wait for the faucet to be started
			waitForServer(t, url, time.Second*5)

			// Execute the request
			respRaw, err := http.Post(
				url,
				jsonMimeType,
				bytes.NewBuffer(testCase.request),
			)
			require.NoError(t, err)

			// Make sure the request errored out
			assert.Equal(t, http.StatusBadRequest, respRaw.StatusCode)

			// Stop the faucet and wait for it to finish
			cancelFn()
			assert.NoError(t, g.Wait())
		})
	}
}

func TestFaucet_Serve_NoFundedAccounts(t *testing.T) {
	t.Parallel()

	var (
		gasFee     = std.MustParseCoin(config.DefaultGasFee)
		sendAmount = std.MustParseCoins(config.DefaultSendAmount)
	)

	var (
		validAddress = crypto.MustAddressFromString("g155n659f89cfak0zgy575yqma64sm4tv6exqk99")

		singleValidRequest = Request{
			To: validAddress.String(),
		}
	)

	encodedSingleValidRequest, err := json.Marshal(
		singleValidRequest,
	)
	require.NoError(t, err)

	getFaucetURL := func(address string) string {
		return fmt.Sprintf("http://%s", address)
	}

	var (
		fundAccount = crypto.Address{1}

		broadcastResponse = &coreTypes.ResultBroadcastTxCommit{
			CheckTx: abci.ResponseCheckTx{
				ResponseBase: abci.ResponseBase{
					Error: nil, // no error
				},
			},
			DeliverTx: abci.ResponseDeliverTx{
				ResponseBase: abci.ResponseBase{
					Error: nil, // no error
				},
			},
		}

		signature   = []byte("signature")
		capturedTxs []*std.Tx

		mockPubKey = &mockPubKey{
			addressFn: func() crypto.Address {
				return fundAccount
			},
		}
		mockPrivKey = &mockPrivKey{
			signFn: func(_ []byte) ([]byte, error) {
				return signature, nil
			},
			pubKeyFn: func() crypto.PubKey {
				return mockPubKey
			},
		}
		mockKeyring = &mockKeyring{
			getAddressesFn: func() []crypto.Address {
				return []crypto.Address{
					fundAccount,
				}
			},
			getKeyFn: func(address crypto.Address) crypto.PrivKey {
				if address == fundAccount {
					return mockPrivKey
				}

				return nil
			},
		}
		mockAccount = &mockAccount{
			getAddressFn: func() crypto.Address {
				return fundAccount
			},
			getCoinsFn: func() std.Coins {
				return sendAmount // NOT enough funds
			},
		}
		mockClient = &mockClient{
			getAccountFn: func(address crypto.Address) (std.Account, error) {
				if address == fundAccount {
					return mockAccount, nil
				}

				return nil, errors.New("account not found")
			},
			sendTransactionCommitFn: func(tx *std.Tx) (*coreTypes.ResultBroadcastTxCommit, error) {
				capturedTxs = append(capturedTxs, tx)

				return broadcastResponse, nil
			},
		}
	)

	// Create a new faucet with default params
	cfg := config.DefaultConfig()
	cfg.ListenAddress = fmt.Sprintf("127.0.0.1:%d", getFreePort(t))

	f, err := NewFaucet(
		static.New(gasFee, 100000),
		mockClient,
		WithConfig(cfg),
	)

	require.NoError(t, err)
	require.NotNil(t, f)

	// Update the keyring
	f.keyring = mockKeyring

	// Start the faucet
	ctx, cancelFn := context.WithCancel(context.Background())
	defer cancelFn()

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return f.Serve(gCtx)
	})

	url := getFaucetURL(f.config.ListenAddress)

	// Wait for the faucet to be started
	waitForServer(t, url, time.Second*5)

	// Execute the request
	respRaw, err := http.Post(
		url,
		jsonMimeType,
		bytes.NewBuffer(encodedSingleValidRequest),
	)
	require.NoError(t, err)

	respBytes, err := io.ReadAll(respRaw.Body)
	require.NoError(t, err)

	response := decodeResponse[Response](t, respBytes)

	assert.Contains(t, response.Error, errNoFundedAccount.Error())
	assert.Equal(t, response.Result, unableToHandleRequest)

	// Stop the faucet and wait for it to finish
	cancelFn()
	assert.NoError(t, g.Wait())

	// Validate the broadcast tx
	assert.Nil(t, capturedTxs)
}
