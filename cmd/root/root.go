package root

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
)

const (
	configFlagName = "config"
	defaultRemote  = "http://127.0.0.1:26657"
	envPrefix      = "GNO_FAUCET"
)

var remoteRegex = regexp.MustCompile(`^https?:\/\/(?:w{1,3}\.)?[^\s.]+(?:\.[a-z]+)*(?::\d+)(?![^<]*(?:<\/\w+>|\/?>))$`)

// faucetCfg wraps the faucet
// root command configuration
type faucetCfg struct {
	corsConfigPath string
	remote         string

	config.Config
}

// New creates the root faucet command
func New() *ffcli.Command {
	cfg := &faucetCfg{}

	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	registerFlags(cfg, fs)

	return &ffcli.Command{
		Name:       "serve",
		ShortUsage: "serve [flags]",
		LongHelp:   "Starts the Gno faucet service",
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
		config.DefaultListenAddress,
		"the IP:PORT URL for the faucet server",
	)

	fs.StringVar(
		&cfg.remote,
		"remote",
		defaultRemote,
		"the JSON-RPC URL of the Gno chain",
	)

	fs.StringVar(
		&cfg.ChainID,
		config.DefaultChainID,
		"dev",
		"the chain ID associated with the remote Gno chain",
	)

	fs.StringVar(
		&cfg.Mnemonic,
		"mnemonic",
		config.DefaultMnemonic,
		"the mnemonic for faucet keys",
	)

	fs.Uint64Var(
		&cfg.NumAccounts,
		"num-accounts",
		config.DefaultNumAccounts,
		"the number of faucet accounts, based on the mnemonic",
	)

	fs.StringVar(
		&cfg.SendAmount,
		"send-amount",
		config.DefaultSendAmount,
		"the static send amount (native currency)",
	)

	fs.StringVar(
		&cfg.GasFee,
		"gas-fee",
		config.DefaultGasFee,
		"the static gas fee for the transaction",
	)

	fs.StringVar(
		&cfg.GasWanted,
		"gas-wanted",
		config.DefaultGasWanted,
		"the static gas wanted for the transaction",
	)

	fs.StringVar(
		&cfg.corsConfigPath,
		"cors-config",
		"",
		"the path to the CORS TOML configuration, if any",
	)
}

// exec executes the faucet root command
func (c *faucetCfg) exec(context.Context, []string) error {
	// Read the CORS configuration, if any
	if c.corsConfigPath != "" {
		corsConfig, err := readCORSConfig(c.corsConfigPath)
		if err != nil {
			return fmt.Errorf("unable to read CORS config, %w", err)
		}

		c.CORSConfig = corsConfig
	}

	// Parse static gas values.
	// It is worth noting that this is temporary,
	// and will be removed once gas estimation is enabled
	// on Gno.land
	gasFee, err := std.ParseCoin(c.GasFee)
	if err != nil {
		return fmt.Errorf("invalid gas fee, %w", err)
	}

	gasWanted, err := strconv.ParseInt(c.GasWanted, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid gas wanted, %w", err)
	}

	// Validate the remote URL
	// validate the remote address
	if !remoteRegex.MatchString(c.remote) {
		return errors.New("invalid remote address")
	}

	// Create a new faucet with
	// static gas estimation
	f, err := faucet.NewFaucet(
		static.New(gasFee, gasWanted),
		tm2Client.NewClient(defaultRemote),
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

// readCORSConfig reads the CORS configuration
// from the specified path
func readCORSConfig(path string) (*config.CORS, error) {
	// Read the config file
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse it
	var corsConfig config.CORS

	if err := toml.Unmarshal(content, &corsConfig); err != nil {
		return nil, err
	}

	return &corsConfig, nil
}
