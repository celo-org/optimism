package main

import (
	"fmt"

	"github.com/urfave/cli/v2"

	opservice "github.com/ethereum-optimism/optimism/op-service"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
)

const (
	ListenAddrFlagName        = "addr"
	PortFlagName              = "port"
	S3BucketFlagName          = "s3.bucket"
	S3EndpointFlagName        = "s3.endpoint"
	S3AccessKeyIDFlagName     = "s3.access-key-id"
	S3AccessKeySecretFlagName = "s3.access-key-secret"
	FileStorePathFlagName     = "file.path"
	GenericCommFlagName       = "generic-commitment"
)

const EnvVarPrefix = "OP_PLASMA_DA_SERVER"

func prefixEnvVars(name string) []string {
	return opservice.PrefixEnvVar(EnvVarPrefix, name)
}

var (
	ListenAddrFlag = &cli.StringFlag{
		Name:    ListenAddrFlagName,
		Usage:   "server listening address",
		Value:   "127.0.0.1",
		EnvVars: prefixEnvVars("ADDR"),
	}
	PortFlag = &cli.IntFlag{
		Name:    PortFlagName,
		Usage:   "server listening port",
		Value:   3100,
		EnvVars: prefixEnvVars("PORT"),
	}
	FileStorePathFlag = &cli.StringFlag{
		Name:    FileStorePathFlagName,
		Usage:   "path to directory for file storage",
		EnvVars: prefixEnvVars("FILESTORE_PATH"),
	}
	GenericCommFlag = &cli.BoolFlag{
		Name:    GenericCommFlagName,
		Usage:   "enable generic commitments for testing. Not for production use.",
		EnvVars: prefixEnvVars("GENERIC_COMMITMENT"),
		Value:   false,
	}
	S3BucketFlag = &cli.StringFlag{
		Name:    S3BucketFlagName,
		Usage:   "bucket name for S3 storage",
		EnvVars: prefixEnvVars("S3_BUCKET"),
	}
)

var requiredFlags = []cli.Flag{
	ListenAddrFlag,
	PortFlag,
}

var optionalFlags = []cli.Flag{
	FileStorePathFlag,
	S3BucketFlag,
	S3EndpointFlag,
	S3AccessKeyIDFlag,
	S3AccessKeySecretFlag,
	GenericCommFlag,
}

func init() {
	optionalFlags = append(optionalFlags, oplog.CLIFlags(EnvVarPrefix)...)
	Flags = append(requiredFlags, optionalFlags...)
}

// Flags contains the list of configuration options available to the binary.
var Flags []cli.Flag

type CLIConfig struct {
	FileStoreDirPath  string
	S3Bucket          string
	S3Endpoint        string
	S3AccessKeyID     string
	S3AccessKeySecret string
	UseGenericComm    bool
}

func ReadCLIConfig(ctx *cli.Context) CLIConfig {
	return CLIConfig{
		FileStoreDirPath:  ctx.String(FileStorePathFlagName),
		S3Bucket:          ctx.String(S3BucketFlagName),
		S3Endpoint:        ctx.String(S3EndpointFlagName),
		S3AccessKeyID:     ctx.String(S3AccessKeyIDFlagName),
		S3AccessKeySecret: ctx.String(S3AccessKeySecretFlagName),
		UseGenericComm:    ctx.Bool(GenericCommFlagName),
	}
}

func (c CLIConfig) Check() error {
	if !c.S3Enabled() && !c.FileStoreEnabled() {
		return fmt.Errorf("at least one storage backend must be enabled")
	}
	if c.S3Enabled() && c.FileStoreEnabled() {
		return fmt.Errorf("only one storage backend can be enabled")
	}
	return nil
}

func (c CLIConfig) S3Enabled() bool {
	return c.S3Bucket != ""
}

func (c CLIConfig) FileStoreEnabled() bool {
	return c.FileStoreDirPath != ""
}

func CheckRequired(ctx *cli.Context) error {
	for _, f := range requiredFlags {
		if !ctx.IsSet(f.Names()[0]) {
			return fmt.Errorf("flag %s is required", f.Names()[0])
		}
	}
	return nil
}
