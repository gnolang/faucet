package root

import (
	"context"
	"flag"

	"github.com/gnolang/faucet/config"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/peterbourgon/ff/v3/fftoml"
)

const (
	configFlagName = "config"
	envPrefix      = "GNO_FAUCET"
)

// faucetCfg wraps the faucet
// startup command configuration
type faucetCfg struct {
	config.Config

	corsConfigPath string
}

// New creates the root faucet command
func New() *ffcli.Command {
	cfg := &faucetCfg{}

	fs := flag.NewFlagSet("start", flag.ExitOnError)
	registerFlags(cfg, fs)

	return &ffcli.Command{
		Name:       "start",
		ShortUsage: "start [flags]",
		LongHelp:   "Starts the Gno faucet service",
		FlagSet:    fs,
		Exec:       cfg.exec,
		Options: []ff.Option{
			// Allow using ENV variables
			ff.WithEnvVars(),
			ff.WithEnvVarPrefix(envPrefix),

			// Allow using TOML config files
			ff.WithConfigFileFlag(configFlagName),
			ff.WithConfigFileParser(fftoml.Parser),
		},
	}
}

// registerFlags registers the faucet root command flags
func registerFlags(cfg *faucetCfg, fs *flag.FlagSet) {
	// Config flag
	fs.String(
		configFlagName,
		"",
		"the path to the configuration file [TOML]",
	)

	// Top level flags
	fs.StringVar(
		&cfg.ListenAddress,
		"listen-address",
		"0.0.0.0:8545",
		"the IP:PORT URL for the faucet server",
	)

	fs.StringVar(
		&cfg.Remote,
		"remote",
		"http://127.0.0.1:26657",
		"the JSON-RPC URL of the Gno chain",
	)

	fs.StringVar(
		&cfg.ChainID,
		"chain-id",
		"dev",
		"the chain ID associated with the remote Gno chain",
	)

	fs.StringVar(
		&cfg.GasFee,
		"gas-fee",
		"1000000ugnot",
		"the static gas fee for the transaction",
	)

	fs.StringVar(
		&cfg.GasWanted,
		"gas-wanted",
		"100000",
		"the static gas wanted for the transaction",
	)

	fs.StringVar(
		&cfg.corsConfigPath,
		"cors-config",
		"",
		"the path to the CORS TOML configuration, if any",
	)
}

// exec executes the faucet start command
func (c *faucetCfg) exec(context.Context, []string) error {
	// TODO

	return nil
}
