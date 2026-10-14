package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"

	enclave_tools "github.com/ethereum-optimism/optimism/op-batcher/enclave-tools"
)

func main() {
	app := &cli.App{
		Name:        "enclave-tools",
		Usage:       "Build enclave EIF images",
		Description: "Builds AWS Nitro EIF (Enclave Image Format) images for the Optimism op-batcher.",
		Version:     "1.0.0",
		Commands: []*cli.Command{
			buildEifCommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
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
    --app-image ghcr.io/celo-org/optimism/op-batcher-enclave-app:TAG \
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
