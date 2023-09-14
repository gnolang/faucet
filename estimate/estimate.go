package estimate

import "github.com/gnolang/gno/tm2/pkg/std"

// Estimator defines the transaction gas estimator
type Estimator interface {
	// EstimateGasFee estimates the optimal gas fee for the specified transaction
	EstimateGasFee(tx std.Tx) std.Coins

	// EstimateGasWanted estimates the optimal gas wanted for the specified transaction
	EstimateGasWanted(tx std.Tx) int64
}
