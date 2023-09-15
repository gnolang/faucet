package static

import (
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Estimator is a static gas estimator (returns static values)
type Estimator struct {
	gasFee    std.Coin
	gasWanted int64
}

// New creates a new static gas estimator
func New(gasFee std.Coin, gasWanted int64) *Estimator {
	return &Estimator{
		gasFee:    gasFee,
		gasWanted: gasWanted,
	}
}

func (e Estimator) EstimateGasFee() std.Coin {
	return e.gasFee
}

func (e Estimator) EstimateGasWanted(_ *std.Tx) int64 {
	return e.gasWanted
}
