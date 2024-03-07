package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/gnolang/faucet"
	tm2Client "github.com/gnolang/faucet/client/http"
	"github.com/gnolang/faucet/config"
	"github.com/gnolang/faucet/estimate/static"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/pelletier/go-toml"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/peterbourgon/ff/v3/fftoml"
	"go.uber.org/zap"
)

const (
	configFlagName = "config"
	envPrefix      = "GNO_FAUCET"
)

const (
	defaultGasFee    = "1000000ugnot"
	defaultGasWanted = "100000"
	defaultRemote    = "http://127.0.0.1:26657"
)

var remoteRegex = regexp.MustCompile(`^https?://[a-z\d.-]+(:\d+)?(?:/[a-z\d]+)*$`)

// faucetCfg wraps the faucet
// root command configuration
type faucetCfg struct {
	config *config.Config

	faucetConfigPath string
	remote           string
	gasFee           string
	gasWanted        string
}

// newRootCmd creates the root faucet command
func newRootCmd() *ffcli.Command {
	cfg := &faucetCfg{
		config: config.DefaultConfig(),
	}

	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	cfg.registerRootFlags(fs)

	return &ffcli.Command{
		Name:       "serve",
		ShortUsage: "serve [flags]",
		LongHelp:   "Serves the Gno faucet service",
		FlagSet:    fs,
		Exec:       cfg.exec,
		Options: []ff.Option{
			// Allow using ENV variables
			ff.WithEnvVars(),
			ff.WithEnvVarPrefix(envPrefix),

			// Allow using TOML config files
			ff.WithAllowMissingConfigFile(true),
			ff.WithConfigFileFlag(configFlagName),
			ff.WithConfigFileParser(fftoml.Parser),
		},
	}
}

// registerFlags registers the faucet root command flags
func (c *faucetCfg) registerRootFlags(fs *flag.FlagSet) {
	// Config flag
	fs.String(
		configFlagName,
		"",
		"the path to the command configuration file [TOML]",
	)

	// Top level flags
	fs.StringVar(
		&c.config.ListenAddress,
		"listen-address",
		config.DefaultListenAddress,
		"the IP:PORT URL for the faucet server",
	)

	fs.StringVar(
		&c.config.ChainID,
		"chain-id",
		config.DefaultChainID,
		"the chain ID associated with the remote Gno chain",
	)

	fs.StringVar(
		&c.config.Mnemonic,
		"mnemonic",
		"",
		"the mnemonic for faucet keys",
	)

	fs.Uint64Var(
		&c.config.NumAccounts,
		"num-accounts",
		config.DefaultNumAccounts,
		"the number of faucet accounts, based on the mnemonic",
	)

	fs.StringVar(
		&c.config.MaxSendAmount,
		"send-amount",
		config.DefaultMaxSendAmount,
		"the static max send amount per drip (native currency)",
	)

	fs.StringVar(
		&c.gasFee,
		"gas-fee",
		defaultGasFee,
		"the static gas fee for the transaction. Format: <AMOUNT>ugnot",
	)

	fs.StringVar(
		&c.gasWanted,
		"gas-wanted",
		defaultGasWanted,
		"the static gas wanted for the transaction. Format: <AMOUNT>ugnot",
	)

	fs.StringVar(
		&c.faucetConfigPath,
		"faucet-config",
		"",
		"the path to the faucet TOML configuration, if any",
	)

	fs.StringVar(
		&c.remote,
		"remote",
		defaultRemote,
		"the JSON-RPC URL of the Gno chain",
	)
}

// exec executes the faucet root command
func (c *faucetCfg) exec(_ context.Context, _ []string) error {
	// Read the faucet configuration, if any
	if c.faucetConfigPath != "" {
		faucetConfig, err := readFaucetConfig(c.faucetConfigPath)
		if err != nil {
			return fmt.Errorf("unable to read faucet config, %w", err)
		}

		c.config = faucetConfig
	}

	// Parse static gas values.
	// It is worth noting that this is temporary,
	// and will be removed once gas estimation is enabled
	// on Gno.land
	gasFee, err := std.ParseCoin(c.gasFee)
	if err != nil {
		return fmt.Errorf("invalid gas fee, %w", err)
	}

	gasWanted, err := strconv.ParseInt(c.gasWanted, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid gas wanted, %w", err)
	}

	// Validate the remote address
	if !remoteRegex.MatchString(c.remote) {
		return errors.New("invalid remote address")
	}

	// Create a new logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	// Create a new faucet with
	// static gas estimation
	f, err := faucet.NewFaucet(
		static.New(gasFee, gasWanted),
		tm2Client.NewClient(c.remote),
		faucet.WithLogger(newCommandLogger(logger)),
		faucet.WithConfig(c.config),
	)
	if err != nil {
		return fmt.Errorf("unable to create faucet, %w", err)
	}

	// Create a new waiter
	w := newWaiter()

	// Add the faucet service
	w.add(f.Serve)

	// Wait for the faucet to exit
	return w.wait()
}

// readFaucetConfig reads the faucet configuration
// from the specified path
func readFaucetConfig(path string) (*config.Config, error) {
	// Read the config file
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse it
	var faucetConfig config.Config

	if err := toml.Unmarshal(content, &faucetConfig); err != nil {
		return nil, err
	}

	return &faucetConfig, nil
}
