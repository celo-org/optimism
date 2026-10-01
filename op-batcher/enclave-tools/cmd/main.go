package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/urfave/cli/v2"

	enclave_tools "github.com/ethereum-optimism/optimism/op-batcher/enclave-tools"
)

func main() {
	app := &cli.App{
		Name:        "enclave-tools",
		Usage:       "Build, register, and run enclave EIF images",
		Description: "A command-line interface for building, registering, and running enclave EIF (Enclave Image Format) images for the Optimism op-batcher.",
		Version:     "1.0.0",
		Commands: []*cli.Command{
			buildCommand(),
			buildEifCommand(),
			registerCommand(),
			isRegisteredCommand(),
			runCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func buildCommand() *cli.Command {
	return &cli.Command{
		Name:  "build",
		Usage: "Build enclave EIF image",
		Description: `Build a Docker image and then create an EIF (Enclave Image Format) file
with the op-batcher and specified arguments.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "op-root",
				Usage:    "Path to optimism root directory",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "tag",
				Usage:    "Docker tag for the EIF image",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "args",
				Usage: "Command-line arguments to op-batcher (comma-separated)",
			},
			&cli.UintFlag{
				Name:  "cpu-count",
				Usage: "Number of vCPUs to allocate to the enclave (affects PCR0)",
				Value: 2,
			},
			&cli.UintFlag{
				Name:  "memory-mb",
				Usage: "Memory in MiB to allocate to the enclave (affects PCR0)",
				Value: 4096,
			},
		},
		Action: buildAction,
	}
}

func buildEifCommand() *cli.Command {
	return &cli.Command{
		Name:  "build-eif",
		Usage: "Build EIF image from a pre-built app Docker image",
		Description: `Build an EIF (Enclave Image Format) image by wrapping a pre-built
op-batcher-enclave-app Docker image with Enclaver. Prints PCR measurements
to stdout in KEY=VALUE format (PCR0, PCR1, PCR2) for CI capture.

Example:
  enclave-tools build-eif \
    --app-image ghcr.io/espressosystems/optimism-espresso-integration/op-batcher-enclave-app:TAG \
    --eif-tag op-batcher-eif:TAG`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "app-image",
				Usage:    "Pre-built app Docker image to wrap (e.g. ghcr.io/org/op-batcher-enclave-app:tag)",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "eif-tag",
				Usage: "Docker tag for the resulting EIF image",
				Value: "op-batcher-eif:latest",
			},
			&cli.UintFlag{
				Name:  "cpu-count",
				Usage: "Number of vCPUs to allocate to the enclave (affects PCR0)",
				Value: 2,
			},
			&cli.UintFlag{
				Name:  "memory-mb",
				Usage: "Memory in MiB to allocate to the enclave (affects PCR0)",
				Value: 4096,
			},
		},
		Action: buildEifAction,
	}
}

func buildEifAction(c *cli.Context) error {
	appImage := c.String("app-image")
	eifTag := c.String("eif-tag")
	cpuCount := c.Uint("cpu-count")
	memoryMb := c.Uint("memory-mb")

	ctx := context.Background()
	slog.Info("Building EIF from pre-built app image...", "app-image", appImage, "eif-tag", eifTag, "cpu-count", cpuCount, "memory-mb", memoryMb)

	measurements, err := enclave_tools.BuildEifFromImage(ctx, appImage, eifTag, cpuCount, memoryMb)
	if err != nil {
		return fmt.Errorf("failed to build EIF: %w", err)
	}

	slog.Info("EIF build completed",
		"PCR0", measurements.PCR0,
		"PCR1", measurements.PCR1,
		"PCR2", measurements.PCR2)
	// Print measurements to stdout in KEY=VALUE format for CI capture
	fmt.Printf("PCR0=%s\nPCR1=%s\nPCR2=%s\n", measurements.PCR0, measurements.PCR1, measurements.PCR2)
	return nil
}

func registerCommand() *cli.Command {
	return &cli.Command{
		Name:  "register",
		Usage: "Register enclave PCR with verifier",
		Description: `Register the enclave's PCR0 measurement with the EspressoNitroTEEVerifier contract.
This allows the enclave to be trusted by the verification system.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "authenticator",
				Usage:    "BatchAuthenticator contract address",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "l1-url",
				Usage:    "L1 RPC URL",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "private-key",
				Usage:    "Private key for transaction signing (hex format)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "pcr0",
				Usage:    "PCR0 value in hex format",
				Required: true,
			},
		},
		Action: registerAction,
	}
}

func isRegisteredCommand() *cli.Command {
	return &cli.Command{
		Name:  "is-registered",
		Usage: "Check if enclave PCR is already registered",
		Description: `Check if the enclave's PCR0 measurement is already registered with the
EspressoNitroTEEVerifier contract. Exits with code 0 if registered, 1 if not.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "authenticator",
				Usage:    "BatchAuthenticator contract address",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "l1-url",
				Usage:    "L1 RPC URL",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "pcr0",
				Usage:    "PCR0 value in hex format",
				Required: true,
			},
		},
		Action: isRegisteredAction,
	}
}

func runCommand() *cli.Command {
	return &cli.Command{
		Name:  "run",
		Usage: "Launch/run the EIF",
		Description: `Launch the specified EIF image in a Docker container with the necessary
AWS Nitro Enclaves configuration.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "image",
				Usage:    "Name of the EIF image to run",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "args",
				Usage: "Command-line arguments to dynamically send to enclave (comma-separated)",
			},
		},
		Action: runAction,
	}
}

func buildAction(c *cli.Context) error {
	opRoot := c.String("op-root")
	tag := c.String("tag")
	args := c.String("args")
	cpuCount := c.Uint("cpu-count")
	memoryMb := c.Uint("memory-mb")

	// Parse batcher arguments
	batcherArgs, err := ParseBatcherArgs(args)
	if err != nil {
		return fmt.Errorf("failed to parse batcher arguments: %w", err)
	}

	ctx := context.Background()
	slog.Info("Building enclave image...")
	measurements, err := enclave_tools.BuildBatcherImage(ctx, opRoot, tag, cpuCount, memoryMb, batcherArgs...)
	if err != nil {
		return fmt.Errorf("failed to build enclave image: %w", err)
	}

	slog.Info("Build completed successfully!")
	slog.Info("Measurements",
		"PCR0", measurements.PCR0,
		"PCR1", measurements.PCR1,
		"PCR2", measurements.PCR2)

	return nil
}

func registerAction(c *cli.Context) error {
	authenticatorAddr := c.String("authenticator")
	l1URL := c.String("l1-url")
	privateKey := c.String("private-key")
	pcr0 := c.String("pcr0")

	key, err := crypto.HexToECDSA(strings.TrimPrefix(privateKey, "0x"))
	if err != nil {
		return fmt.Errorf("invalid private key: %w", err)
	}

	// Parse authenticator address
	authAddr := common.HexToAddress(authenticatorAddr)
	if authAddr == (common.Address{}) {
		return fmt.Errorf("invalid authenticator address")
	}

	// Parse PCR0
	pcr0Bytes, err := hex.DecodeString(strings.TrimPrefix(pcr0, "0x"))
	if err != nil {
		return fmt.Errorf("failed to parse PCR0: %w", err)
	}

	ctx := context.Background()
	slog.Info("Registering enclave hash...")
	err = enclave_tools.RegisterEnclaveHash(ctx, authAddr, l1URL, key, pcr0Bytes)
	if err != nil {
		return fmt.Errorf("failed to register enclave hash: %w", err)
	}

	slog.Info("Enclave hash registered successfully!")
	return nil
}

func isRegisteredAction(c *cli.Context) error {
	authenticatorAddr := c.String("authenticator")
	l1URL := c.String("l1-url")
	pcr0 := c.String("pcr0")

	// Parse authenticator address
	authAddr := common.HexToAddress(authenticatorAddr)
	if authAddr == (common.Address{}) {
		return fmt.Errorf("invalid authenticator address")
	}

	// Parse PCR0
	pcr0Bytes, err := hex.DecodeString(strings.TrimPrefix(pcr0, "0x"))
	if err != nil {
		return fmt.Errorf("failed to parse PCR0: %w", err)
	}

	ctx := context.Background()
	slog.Info("Checking if enclave hash is registered...")
	isRegistered, err := enclave_tools.IsEnclaveHashRegistered(ctx, authAddr, l1URL, pcr0Bytes)
	if err != nil {
		return fmt.Errorf("failed to check registration: %w", err)
	}

	if isRegistered {
		slog.Info("Enclave hash is registered")
		fmt.Println("true")
		return nil
	} else {
		slog.Info("Enclave hash is NOT registered")
		fmt.Println("false")
		os.Exit(1)
		return nil
	}
}

func runAction(c *cli.Context) error {
	imageName := c.String("image")
	argsStr := c.String("args")

	// Parse arguments
	args, err := ParseBatcherArgs(argsStr)
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	ctx := context.Background()
	enclaverCli := &enclave_tools.EnclaverCli{}

	slog.Info("Starting enclave", "image", imageName)
	err = enclaverCli.RunEnclave(ctx, imageName, args)
	if err != nil {
		slog.Error("Failed to start enclave", "image", imageName, "error", err)
		return err
	}

	return nil
}

// ParseBatcherArgs parses comma-separated batcher arguments and validates them
func ParseBatcherArgs(argsStr string) ([]string, error) {
	if argsStr == "" {
		return []string{}, nil
	}

	args := strings.Split(argsStr, ",")
	var cleanedArgs []string

	for _, arg := range args {
		cleaned := strings.TrimSpace(arg)
		if cleaned == "" {
			continue // Skip empty args
		}
		cleanedArgs = append(cleanedArgs, cleaned)
	}

	return cleanedArgs, nil
}
