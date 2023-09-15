package faucet

import (
	coreTypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type (
	bytesDelegate         func() []byte
	signDelegate          func([]byte) ([]byte, error)
	pubKeyDelegate        func() crypto.PubKey
	privKeyEqualsDelegate func(key crypto.PrivKey) bool
)

type mockPrivKey struct {
	bytesFn  bytesDelegate
	signFn   signDelegate
	pubKeyFn pubKeyDelegate
	equalsFn privKeyEqualsDelegate
}

func (m *mockPrivKey) Bytes() []byte {
	if m.bytesFn != nil {
		return m.bytesFn()
	}

	return nil
}

func (m *mockPrivKey) Sign(msg []byte) ([]byte, error) {
	if m.signFn != nil {
		return m.signFn(msg)
	}

	return nil, nil
}

func (m *mockPrivKey) PubKey() crypto.PubKey {
	if m.pubKeyFn != nil {
		return m.pubKeyFn()
	}

	return nil
}

func (m *mockPrivKey) Equals(key crypto.PrivKey) bool {
	if m.equalsFn != nil {
		return m.equalsFn(key)
	}

	return false
}

type (
	addressDelegate      func() crypto.Address
	verifyBytesDelegate  func([]byte, []byte) bool
	pubKeyEqualsDelegate func(crypto.PubKey) bool
	stringDelegate       func() string
)

type mockPubKey struct {
	addressFn addressDelegate
	bytesFn   bytesDelegate
	verifyFn  verifyBytesDelegate
	equalsFn  pubKeyEqualsDelegate
	stringFn  stringDelegate
}

func (m *mockPubKey) Address() crypto.Address {
	if m.addressFn != nil {
		return m.addressFn()
	}

	return crypto.Address{}
}

func (m *mockPubKey) Bytes() []byte {
	if m.bytesFn != nil {
		return m.bytesFn()
	}

	return nil
}

func (m *mockPubKey) VerifyBytes(msg []byte, sig []byte) bool {
	if m.verifyFn != nil {
		return m.verifyFn(msg, sig)
	}

	return false
}

func (m *mockPubKey) Equals(key crypto.PubKey) bool {
	if m.equalsFn != nil {
		return m.equalsFn(key)
	}

	return false
}

func (m *mockPubKey) String() string {
	if m.stringFn != nil {
		return m.stringFn()
	}

	return ""
}

type (
	estimateGasFeeDelegate    func() std.Coin
	estimateGasWantedDelegate func(*std.Tx) int64
)

type mockEstimator struct {
	estimateGasFeeFn    estimateGasFeeDelegate
	estimateGasWantedFn estimateGasWantedDelegate
}

func (m *mockEstimator) EstimateGasFee() std.Coin {
	if m.estimateGasFeeFn != nil {
		return m.estimateGasFeeFn()
	}

	return std.Coin{}
}

func (m *mockEstimator) EstimateGasWanted(tx *std.Tx) int64 {
	if m.estimateGasWantedFn != nil {
		return m.estimateGasWantedFn(tx)
	}

	return 0
}

type (
	getAccountDelegate            func(crypto.Address) (std.Account, error)
	sendTransactionSyncDelegate   func(tx *std.Tx) (*coreTypes.ResultBroadcastTx, error)
	sendTransactionCommitDelegate func(tx *std.Tx) (*coreTypes.ResultBroadcastTxCommit, error)
)

type mockClient struct {
	getAccountFn            getAccountDelegate
	sendTransactionSyncFn   sendTransactionSyncDelegate
	sendTransactionCommitFn sendTransactionCommitDelegate
}

func (m *mockClient) GetAccount(address crypto.Address) (std.Account, error) {
	if m.getAccountFn != nil {
		return m.getAccountFn(address)
	}

	return nil, nil
}

func (m *mockClient) SendTransactionSync(tx *std.Tx) (*coreTypes.ResultBroadcastTx, error) {
	if m.sendTransactionSyncFn != nil {
		return m.sendTransactionSyncFn(tx)
	}

	return nil, nil
}

func (m *mockClient) SendTransactionCommit(tx *std.Tx) (*coreTypes.ResultBroadcastTxCommit, error) {
	if m.sendTransactionCommitFn != nil {
		return m.sendTransactionCommitFn(tx)
	}

	return nil, nil
}
