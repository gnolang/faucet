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

func (m *mockPubKey) VerifyBytes(msg, sig []byte) bool {
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
	statusDelegate                func() (*coreTypes.ResultStatus, error)
)

type mockClient struct {
	getAccountFn            getAccountDelegate
	sendTransactionSyncFn   sendTransactionSyncDelegate
	sendTransactionCommitFn sendTransactionCommitDelegate
	statusFn                statusDelegate
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

func (m *mockClient) Status() (*coreTypes.ResultStatus, error) {
	if m.statusFn != nil {
		return m.statusFn()
	}

	return nil, nil
}

type (
	getAddressDelegate       func() crypto.Address
	setAddressDelegate       func(crypto.Address) error
	getPubKeyDelegate        func() crypto.PubKey
	setPubKeyDelegate        func(crypto.PubKey) error
	getAccountNumberDelegate func() uint64
	setAccountNumberDelegate func(uint64) error
	getSequenceDelegate      func() uint64
	setSequenceDelegate      func(uint64) error
	getCoinsDelegate         func() std.Coins
	setCoinsDelegate         func(std.Coins) error
)

type mockAccount struct {
	getAddressFn       getAddressDelegate
	setAddressFn       setAddressDelegate
	getPubKeyFn        getPubKeyDelegate
	setPubKeyFn        setPubKeyDelegate
	getAccountNumberFn getAccountNumberDelegate
	setAccountNumberFn setAccountNumberDelegate
	getSequenceFn      getSequenceDelegate
	setSequenceFn      setSequenceDelegate
	getCoinsFn         getCoinsDelegate
	setCoinsFn         setCoinsDelegate
	stringFn           stringDelegate
}

func (m *mockAccount) GetAddress() crypto.Address {
	if m.getAddressFn != nil {
		return m.getAddressFn()
	}

	return crypto.Address{}
}

func (m *mockAccount) SetAddress(address crypto.Address) error {
	if m.setAddressFn != nil {
		return m.setAddressFn(address)
	}

	return nil
}

func (m *mockAccount) GetPubKey() crypto.PubKey {
	if m.getPubKeyFn != nil {
		return m.getPubKeyFn()
	}

	return nil
}

func (m *mockAccount) SetPubKey(key crypto.PubKey) error {
	if m.setPubKeyFn != nil {
		return m.setPubKeyFn(key)
	}

	return nil
}

func (m *mockAccount) GetAccountNumber() uint64 {
	if m.getAccountNumberFn != nil {
		return m.getAccountNumberFn()
	}

	return 0
}

func (m *mockAccount) SetAccountNumber(number uint64) error {
	if m.setAccountNumberFn != nil {
		return m.setAccountNumberFn(number)
	}

	return nil
}

func (m *mockAccount) GetSequence() uint64 {
	if m.getSequenceFn != nil {
		return m.getSequenceFn()
	}

	return 0
}

func (m *mockAccount) SetSequence(sequence uint64) error {
	if m.setSequenceFn != nil {
		return m.setSequenceFn(sequence)
	}

	return nil
}

func (m *mockAccount) GetCoins() std.Coins {
	if m.getCoinsFn != nil {
		return m.getCoinsFn()
	}

	return std.Coins{}
}

func (m *mockAccount) SetCoins(coins std.Coins) error {
	if m.setCoinsFn != nil {
		return m.setCoinsFn(coins)
	}

	return nil
}

func (m *mockAccount) String() string {
	if m.stringFn != nil {
		return m.stringFn()
	}

	return ""
}

type (
	getAddressesDelegate func() []crypto.Address
	getKeyDelegate       func(crypto.Address) crypto.PrivKey
)

type mockKeyring struct {
	getAddressesFn getAddressesDelegate
	getKeyFn       getKeyDelegate
}

func (m *mockKeyring) GetAddresses() []crypto.Address {
	if m.getAddressesFn != nil {
		return m.getAddressesFn()
	}

	return nil
}

func (m *mockKeyring) GetKey(address crypto.Address) crypto.PrivKey {
	if m.getKeyFn != nil {
		return m.getKeyFn(address)
	}

	return nil
}
