package estimate

import "github.com/gnolang/gno/tm2/pkg/std"

// Estimator defines the transaction gas estimator
type Estimator interface {
	// EstimateGasFee estimates the current network gas fee for transactions
	EstimateGasFee() std.Coin

	// EstimateGasWanted estimates the optimal gas wanted for the specified transaction
	EstimateGasWanted(tx std.Tx) int64
}
