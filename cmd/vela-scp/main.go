// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"net/mail"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"

	"github.com/go-vela/vela-openssh/internal/openssh"
	"github.com/go-vela/vela-openssh/internal/scp"
	"github.com/go-vela/vela-openssh/pkg/binarywrapper"
)

func main() {
	cmd := cli.Command{
		Name:      "vela-scp",
		Usage:     "Vela plugin wrapping the scp binary.",
		Copyright: "Copyright 2022 Target Brands, Inc. All rights reserved.",
		Authors: []any{
			&mail.Address{
				Name:    "Vela Admins",
				Address: "vela@target.com",
			},
		},
		// The version field looks gross but in practice is really only seen and used in integration tests
		// or when a plugin is misconfigured. We should log the version information of dependent binaries
		// to assist with debugging why a plugin might be failing to operate in a way users expect.
		Version: fmt.Sprintf("Plugin: %s - OpenSSH: %s - SSHPass: %s", openssh.PluginVersion, openssh.OpenSSHVersion, openssh.SSHPassVersion),
		Action:  run,
	}

	cmd.Flags = []cli.Flag{
		&cli.StringSliceFlag{
			Name:     "source",
			Usage:    "source parameter for scp (see manual 'man scp')",
			Required: true,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_SOURCE"),
				cli.EnvVar("SOURCE"),
				cli.File("/vela/parameters/vela-scp/source"),
				cli.File("/vela/secrets/vela-scp/source"),
			),
		},
		&cli.StringFlag{
			Name:     "target",
			Usage:    "target parameter for scp (see manual 'man scp')",
			Required: true,
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_TARGET"),
				cli.EnvVar("TARGET"),
				cli.File("/vela/parameters/vela-scp/target"),
				cli.File("/vela/secrets/vela-scp/target"),
			),
		},
		&cli.StringSliceFlag{
			Name:  "identity-file.path",
			Usage: "path to the identity file parameter for scp (see manual 'man scp')",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_IDENTITY_FILE_PATH"),
				cli.EnvVar("IDENTITY_FILE_PATH"),
				cli.EnvVar("PARAMETER_SSH_KEY_PATH"),
				cli.EnvVar("SSH_KEY_PATH"),
				cli.File("/vela/parameters/vela-scp/identity-file.path"),
				cli.File("/vela/secrets/vela-scp/identity-file.path"),
			),
		},
		&cli.StringFlag{
			Name:  "identity-file.contents",
			Usage: "contents of the identity-file (not the path, the real deal)",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_IDENTITY_FILE_CONTENTS"),
				cli.EnvVar("IDENTITY_FILE_CONTENTS"),
				cli.EnvVar("PARAMETER_SSH_KEY"),
				cli.EnvVar("SSH_KEY"),
				cli.File("/vela/parameters/vela-scp/identity-file.contents"),
				cli.File("/vela/secrets/vela-scp/identity-file.contents"),
			),
		},
		&cli.StringSliceFlag{
			Name:  "scp.flag",
			Usage: "any additional flags for scp can be specified here",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_SCP_FLAG"),
				cli.EnvVar("SCP_FLAG"),
				cli.File("/vela/parameters/vela-scp/scp.flag"),
				cli.File("/vela/secrets/vela-scp/scp.flag"),
			),
		},
		&cli.StringFlag{
			Name:  "sshpass.password",
			Usage: "password for use with destination target (used with sshpass)",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_SSHPASS_PASSWORD"),
				cli.EnvVar("PARAMETER_PASSWORD"),
				cli.EnvVar("SSHPASS_PASSWORD"),
				cli.EnvVar("PASSWORD"),
				cli.File("/vela/parameters/vela-scp/sshpass.password"),
				cli.File("/vela/secrets/vela-scp/sshpass.password"),
			),
		},
		&cli.StringFlag{
			Name:  "sshpass.passphrase",
			Usage: "passphrase for use with identity file (used with sshpass)",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_SSHPASS_PASSPHRASE"),
				cli.EnvVar("SSHPASS_PASSPHRASE"),
				cli.File("/vela/parameters/vela-scp/sshpass.passphrase"),
				cli.File("/vela/secrets/vela-scp/sshpass.passphrase"),
			),
		},
		&cli.StringSliceFlag{
			Name:  "sshpass.flag",
			Usage: "any additional flags for sshpass can be specified here)",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_SSHPASS_FLAG"),
				cli.EnvVar("SSHPASS_FLAG"),
				cli.File("/vela/parameters/vela-scp/sshpass.flag"),
				cli.File("/vela/secrets/vela-scp/sshpass.flag"),
			),
		},
		&cli.StringFlag{
			Name:  "ci",
			Usage: "set the CI environment (if $CI is set output tries to be friendlier)",
			Sources: cli.NewValueSourceChain(
				cli.EnvVar("PARAMETER_CI"),
				cli.EnvVar("CI"),
			),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(_ context.Context, c *cli.Command) error {
	if c.IsSet("ci") {
		logrus.SetFormatter(&logrus.TextFormatter{
			DisableColors: true,
			FullTimestamp: true,
		})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors:   true,
			FullTimestamp: false,
			PadLevelText:  true,
		})
	}

	if openssh.DirtyBuild {
		logrus.Warnf("binary built from modified commit %s", openssh.GitCommit)
	}

	logrus.WithFields(logrus.Fields{
		"code":            "https://github.com/go-vela/vela-openssh",
		"docs":            "https://go-vela.github.io/docs/plugins/registry/pipeline/scp",
		"registry":        "https://hub.docker.com/r/target/vela-scp",
		"commit":          openssh.GitCommit,
		"version-plugin":  openssh.PluginVersion,
		"version-openssh": openssh.OpenSSHVersion,
		"version-sshpass": openssh.SSHPassVersion,
	}).Info("Vela SCP Plugin")

	bp := binarywrapper.Plugin{
		PluginConfig: &scp.Config{
			Source:               c.StringSlice("source"),
			Target:               c.String("target"),
			IdentityFilePath:     c.StringSlice("identity-file.path"),
			IdentityFileContents: c.String("identity-file.contents"),
			SCPFlags:             c.StringSlice("scp.flag"),
			SSHPassword:          c.String("sshpass.password"),
			SSHPassphrase:        c.String("sshpass.passphrase"),
			SSHPASSFlags:         c.StringSlice("sshpass.flag"),
		},
	}

	return bp.Exec()
}
