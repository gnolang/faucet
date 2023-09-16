package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/gnolang/faucet/config"
	"github.com/pelletier/go-toml"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type generateCfg struct {
	outputPath string
}

// newGenerateCmd creates the generate faucet command
func newGenerateCmd() *ffcli.Command {
	cfg := &generateCfg{}

	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	cfg.registerFlags(fs)

	return &ffcli.Command{
		Name:       "generate",
		ShortUsage: "generate [flags]",
		LongHelp:   "Generates and outputs the default Faucet configuration",
		FlagSet:    fs,
		Exec:       cfg.exec,
	}
}

// registerFlags registers the faucet root command flags
func (c *generateCfg) registerFlags(fs *flag.FlagSet) {
	fs.StringVar(
		&c.outputPath,
		"output-path",
		"./config.toml",
		"the path to output the TOML configuration file",
	)
}

// exec executes the faucet generate command
func (c *generateCfg) exec(_ context.Context, _ []string) error {
	if c.outputPath == "" {
		return errors.New("output path not set")
	}

	// Generate the default config
	cfg := config.DefaultConfig()

	// Write it out to a file
	encodedConfig, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("unable to encode config, %w", err)
	}

	// Create the output file
	outputFile, err := os.Create(c.outputPath)
	if err != nil {
		return fmt.Errorf("unable to create output file, %w", err)
	}

	defer outputFile.Close()

	// Write the config
	_, err = outputFile.Write(encodedConfig)
	if err != nil {
		return fmt.Errorf("unable to write output file, %w", err)
	}

	return nil
}
