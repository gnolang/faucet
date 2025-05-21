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
	"github.com/gnolang/faucet/spec"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	coreTypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

// decodeResponse decodes the JSON response
func decodeResponse[T spec.BaseJSONResponse | spec.BaseJSONResponses](t *testing.T, responseBody []byte) *T {
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

// waitForServer waits for the web server to start up
func waitForServer(t *testing.T, url string) {
	t.Helper()

	ch := make(chan bool)

	go func() {
		for {
			if _, err := http.Get(url); err == nil {
				ch <- true
			}

			time.Sleep(10 * time.Millisecond)
		}
	}()

	select {
	case <-ch:
		return
	case <-time.After(5 * time.Second):
		t.Fatalf("server wait timeout exceeded")
	}
}

func TestFaucet_Serve_ValidRequests(t *testing.T) {
	t.Parallel()

	var (
		gasFee     = std.MustParseCoin("1ugnot")
		sendAmount = std.MustParseCoins(config.DefaultMaxSendAmount)
	)

	var (
		validAddress = crypto.MustAddressFromString("g155n659f89cfak0zgy575yqma64sm4tv6exqk99")

		singleValidRequest = spec.NewJSONRequest(0, DefaultDripMethod, []any{validAddress.String()})

		bulkValidRequests = spec.BaseJSONRequests{
			spec.NewJSONRequest(0, DefaultDripMethod, []any{validAddress.String()}),
			spec.NewJSONRequest(1, DefaultDripMethod, []any{validAddress.String()}),
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
		requestValidateFn func(response []byte)
		name              string
		request           []byte
		expectedNumTxs    int
	}{
		{
			func(resp []byte) {
				response := decodeResponse[spec.BaseJSONResponse](t, resp)

				assert.Empty(t, response.Error)
				assert.Equal(t, faucetSuccess, response.Result)
			},
			"single request",
			encodedSingleValidRequest,
			1,
		},
		{
			func(resp []byte) {
				responses := decodeResponse[spec.BaseJSONResponses](t, resp)
				require.Len(t, *responses, len(bulkValidRequests))

				for _, response := range *responses {
					assert.Empty(t, response.Error)
					assert.Equal(t, faucetSuccess, response.Result)
				}
			},
			"bulk request",
			encodedBulkValidRequests,
			len(bulkValidRequests),
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			//nolint:dupl // It's fine to leave this setup the same
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
			waitForServer(t, url)

			// Execute the request
			respRaw, err := http.Post(
				url,
				JSONMimeType,
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

func TestFaucet_Serve_Middleware(t *testing.T) {
	t.Parallel()

	var (
		gasFee     = std.MustParseCoin("1ugnot")
		sendAmount = std.MustParseCoins(config.DefaultMaxSendAmount)
		id         = uint(10)
	)

	var (
		validAddress       = crypto.MustAddressFromString("g155n659f89cfak0zgy575yqma64sm4tv6exqk99")
		singleValidRequest = spec.NewJSONRequest(id, DefaultDripMethod, []any{validAddress.String()})

		requiredFunds = sendAmount.Add(std.NewCoins(gasFee))
		fundAccount   = crypto.Address{1}
	)

	encodedReq, err := json.Marshal(singleValidRequest)
	require.NoError(t, err)

	getURL := func(addr string) string {
		return fmt.Sprintf("http://%s", addr)
	}

	var (
		mockPubKey = &mockPubKey{
			addressFn: func() crypto.Address {
				return fundAccount
			},
		}
		mockPrivKey = &mockPrivKey{
			signFn: func(_ []byte) ([]byte, error) {
				return []byte("signature"), nil
			},
			pubKeyFn: func() crypto.PubKey {
				return mockPubKey
			},
		}
		mockKeyring = &mockKeyring{
			getAddressesFn: func() []crypto.Address {
				return []crypto.Address{fundAccount}
			},
			getKeyFn: func(address crypto.Address) crypto.PrivKey {
				if address == fundAccount {
					return mockPrivKey
				}

				return nil
			},
		}
		mockAccount = &mockAccount{
			getAddressFn: func() crypto.Address { return fundAccount },
			getCoinsFn:   func() std.Coins { return requiredFunds },
		}
		mockClient = &mockClient{
			getAccountFn: func(address crypto.Address) (std.Account, error) {
				if address == fundAccount {
					return mockAccount, nil
				}

				return nil, errors.New("account not found")
			},
			sendTransactionCommitFn: func(_ *std.Tx) (*coreTypes.ResultBroadcastTxCommit, error) {
				return &coreTypes.ResultBroadcastTxCommit{
					CheckTx: abci.ResponseCheckTx{
						ResponseBase: abci.ResponseBase{Error: nil},
					},
					DeliverTx: abci.ResponseDeliverTx{
						ResponseBase: abci.ResponseBase{Error: nil},
					},
				}, nil
			},
		}
	)

	t.Run("all pass", func(t *testing.T) {
		t.Parallel()

		var (
			executed = 0

			idMW = func(next HandlerFunc) HandlerFunc {
				return func(ctx context.Context, req *spec.BaseJSONRequest) *spec.BaseJSONResponse {
					require.Equal(t, id, req.ID)

					executed++

					return next(ctx, req)
				}
			}

			addrMW = func(next HandlerFunc) HandlerFunc {
				return func(ctx context.Context, req *spec.BaseJSONRequest) *spec.BaseJSONResponse {
					require.Equal(t, validAddress.String(), req.Params[0].(string))

					executed++

					return next(ctx, req)
				}
			}
		)

		cfg := config.DefaultConfig()
		cfg.ListenAddress = fmt.Sprintf("127.0.0.1:%d", getFreePort(t))

		f, err := NewFaucet(
			static.New(gasFee, 100000),
			mockClient,
			WithConfig(cfg),
			WithMiddlewares([]Middleware{idMW, addrMW}),
		)
		require.NoError(t, err)

		f.keyring = mockKeyring

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		g, gCtx := errgroup.WithContext(ctx)
		g.Go(func() error {
			return f.Serve(gCtx)
		})

		waitForServer(t, getURL(f.config.ListenAddress))

		resp, err := http.Post(getURL(f.config.ListenAddress), JSONMimeType, bytes.NewBuffer(encodedReq))
		require.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		r := decodeResponse[spec.BaseJSONResponse](t, body)

		assert.Empty(t, r.Error)
		assert.Equal(t, faucetSuccess, r.Result)
		assert.Equal(t, 2, executed)

		cancel()
		assert.NoError(t, g.Wait())
	})

	t.Run("middleware fails", func(t *testing.T) {
		t.Parallel()

		var (
			executed = 0
			mwErr    = errors.New("middleware error")

			failMW = func(_ HandlerFunc) HandlerFunc {
				return func(_ context.Context, req *spec.BaseJSONRequest) *spec.BaseJSONResponse {
					return spec.NewJSONResponse(
						req.ID,
						nil,
						spec.NewJSONError(mwErr.Error(), spec.ServerErrorCode),
					)
				}
			}

			idMW = func(next HandlerFunc) HandlerFunc {
				return func(ctx context.Context, req *spec.BaseJSONRequest) *spec.BaseJSONResponse {
					executed++

					return next(ctx, req)
				}
			}
		)

		cfg := config.DefaultConfig()
		cfg.ListenAddress = fmt.Sprintf("127.0.0.1:%d", getFreePort(t))

		f, err := NewFaucet(
			static.New(gasFee, 100000),
			mockClient,
			WithConfig(cfg),
			WithMiddlewares([]Middleware{failMW, idMW}), // first mw should fail
		)
		require.NoError(t, err)

		f.keyring = mockKeyring

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		g, gCtx := errgroup.WithContext(ctx)
		g.Go(func() error { return f.Serve(gCtx) })

		waitForServer(t, getURL(f.config.ListenAddress))

		resp, err := http.Post(getURL(f.config.ListenAddress), JSONMimeType, bytes.NewBuffer(encodedReq))
		require.NoError(t, err)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		r := decodeResponse[spec.BaseJSONResponse](t, body)

		require.NotEmpty(t, r.Error)

		assert.Equal(t, mwErr.Error(), r.Error.Message)
		assert.Equal(t, 0, executed)

		cancel()
		assert.NoError(t, g.Wait())
	})
}

func TestFaucet_Serve_InvalidRequests(t *testing.T) {
	t.Parallel()

	var (
		gasFee     = std.MustParseCoin("1ugnot")
		sendAmount = std.MustParseCoins(config.DefaultMaxSendAmount)
	)

	var (
		invalidAddress = "invalid-address"

		singleInvalidRequest = spec.NewJSONRequest(0, DefaultDripMethod, []any{invalidAddress})

		bulkInvalidRequests = spec.BaseJSONRequests{
			singleInvalidRequest,
			spec.NewJSONRequest(1, DefaultDripMethod, []any{"", sendAmount.String()}), // empty address
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
		requestValidateFn func(response []byte)
		name              string
		request           []byte
		expectedNumTxs    int
	}{
		{
			func(resp []byte) {
				response := decodeResponse[spec.BaseJSONResponse](t, resp)

				require.NotNil(t, response.Error)

				assert.Contains(t, response.Error.Message, errInvalidBeneficiary.Error())
				assert.Nil(t, response.Result)
			},
			"single request",
			encodedSingleInvalidRequest,
			1,
		},
		{
			func(resp []byte) {
				responses := decodeResponse[spec.BaseJSONResponses](t, resp)
				require.Len(t, *responses, len(bulkInvalidRequests))

				for _, response := range *responses {
					require.NotNil(t, response.Error)

					assert.Contains(t, response.Error.Message, errInvalidBeneficiary.Error())
					assert.Nil(t, response.Result)
				}
			},
			"bulk request",
			encodedBulkInvalidRequests,
			len(bulkInvalidRequests),
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			//nolint:dupl // It's fine to leave this setup the same
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
			waitForServer(t, url)

			// Execute the request
			respRaw, err := http.Post(
				url,
				JSONMimeType,
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
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create a new faucet with default params
			cfg := config.DefaultConfig()
			cfg.ListenAddress = fmt.Sprintf("127.0.0.1:%d", getFreePort(t))

			f, err := NewFaucet(
				static.New(std.MustParseCoin("1ugnot"), 100000),
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
			waitForServer(t, url)

			// Execute the request
			respRaw, err := http.Post(
				url,
				JSONMimeType,
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
		gasFee     = std.MustParseCoin("1ugnot")
		sendAmount = std.MustParseCoins(config.DefaultMaxSendAmount)
	)

	var (
		validAddress = crypto.MustAddressFromString("g155n659f89cfak0zgy575yqma64sm4tv6exqk99")

		singleValidRequest = spec.NewJSONRequest(0, DefaultDripMethod, []any{validAddress.String()})
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
	waitForServer(t, url)

	// Execute the request
	respRaw, err := http.Post(
		url,
		JSONMimeType,
		bytes.NewBuffer(encodedSingleValidRequest),
	)
	require.NoError(t, err)

	respBytes, err := io.ReadAll(respRaw.Body)
	require.NoError(t, err)

	response := decodeResponse[spec.BaseJSONResponse](t, respBytes)

	require.NotNil(t, response.Error)
	assert.Contains(t, response.Error.Message, errNoFundedAccount.Error())
	assert.Nil(t, response.Result)

	// Stop the faucet and wait for it to finish
	cancelFn()
	assert.NoError(t, g.Wait())

	// Validate the broadcast tx
	assert.Nil(t, capturedTxs)
}

func TestFaucet_Serve_InvalidSendAmount(t *testing.T) {
	t.Parallel()

	// Extract the default send amount
	maxSendAmount := std.MustParseCoins(config.DefaultMaxSendAmount)

	testTable := []struct {
		name       string
		sendAmount std.Coins
	}{
		{
			"invalid send amount",
			std.NewCoins(std.NewCoin("atom", 10)),
		},
		{
			"excessive send amount",
			maxSendAmount.Add(std.MustParseCoins("100ugnot")),
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				validAddress = crypto.MustAddressFromString("g155n659f89cfak0zgy575yqma64sm4tv6exqk99")
				gasFee       = std.MustParseCoin("1ugnot")

				singleInvalidRequest = spec.NewJSONRequest(
					0,
					DefaultDripMethod,
					[]any{validAddress.String(), testCase.sendAmount.String()},
				)
			)

			encodedSingleInvalidRequest, err := json.Marshal(
				singleInvalidRequest,
			)
			require.NoError(t, err)

			getFaucetURL := func(address string) string {
				return fmt.Sprintf("http://%s", address)
			}

			// Create a new faucet with default params
			cfg := config.DefaultConfig()
			cfg.ListenAddress = fmt.Sprintf("127.0.0.1:%d", getFreePort(t))
			cfg.MaxSendAmount = maxSendAmount.String()

			f, err := NewFaucet(
				static.New(gasFee, 100000),
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
			waitForServer(t, url)

			// Execute the request
			respRaw, err := http.Post(
				url,
				JSONMimeType,
				bytes.NewBuffer(encodedSingleInvalidRequest),
			)
			require.NoError(t, err)

			respBytes, err := io.ReadAll(respRaw.Body)
			require.NoError(t, err)

			response := decodeResponse[spec.BaseJSONResponse](t, respBytes)

			require.NotNil(t, response.Error)

			assert.Contains(t, response.Error.Message, errInvalidSendAmount.Error())
			assert.Nil(t, response.Result)

			// Stop the faucet and wait for it to finish
			cancelFn()
			assert.NoError(t, g.Wait())
		})
	}
}
