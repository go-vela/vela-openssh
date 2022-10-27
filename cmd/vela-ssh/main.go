// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"

	"github.com/go-vela/vela-openssh/internal/openssh"
	"github.com/go-vela/vela-openssh/internal/ssh"
	"github.com/go-vela/vela-openssh/pkg/binarywrapper"
)

func main() {
	app := &cli.App{
		Name:      "vela-ssh",
		Usage:     "Vela plugin wrapping the ssh binary.",
		Copyright: "Copyright (c) 2022 Target Brands, Inc. All rights reserved.",
		Authors: []*cli.Author{
			{
				Name:  "Vela Admins",
				Email: "vela@target.com",
			},
		},
		Action: run,
		// The version field looks gross but in practice is really only seen and used in integration tests
		// or when a plugin is misconfigured. We should log the version information of dependent binaries
		// to assist with debugging why a plugin might be failing to operate in a way users expect.
		Version: fmt.Sprintf("Plugin: %s - OpenSSH: %s - SSHPass: %s", openssh.PluginVersion, openssh.OpenSSHVersion, openssh.SSHPassVersion),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "destination",
				Usage:    "destination parameter for ssh (see manual 'man ssh')",
				EnvVars:  []string{"PARAMETER_DESTINATION", "DESTINATION", "PARAMETER_HOST"},
				FilePath: "/vela/parameters/vela-ssh/destination,/vela/secrets/vela-ssh/destination",
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:     "command",
				Usage:    "command to execute on remote system",
				EnvVars:  []string{"PARAMETER_COMMAND", "COMMAND", "PARAMETER_SCRIPT", "SCRIPT"},
				FilePath: "/vela/parameters/vela-ssh/command,/vela/secrets/vela-ssh/command",
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:     "identity-file.path",
				Usage:    "path to the identity file parameter for scp (see manual 'man scp')",
				EnvVars:  []string{"PARAMETER_IDENTITY_FILE_PATH", "IDENTITY_FILE_PATH", "PARAMETER_SSH_KEY_PATH", "SSH_KEY_PATH"},
				FilePath: "/vela/parameters/vela-ssh/identity-file.path,/vela/secrets/vela-ssh/identity-file.path",
			},
			&cli.StringFlag{
				Name:     "identity-file.contents",
				Usage:    "contents of the identity-file (not the path, the real deal)",
				EnvVars:  []string{"PARAMETER_IDENTITY_FILE_CONTENTS", "IDENTITY_FILE_CONTENTS", "PARAMETER_SSH_KEY", "SSH_KEY"},
				FilePath: "/vela/parameters/vela-ssh/identity-file.contents,/vela/secrets/vela-ssh/identity-file.contents",
			},
			&cli.StringSliceFlag{
				Name:     "ssh.flag",
				Usage:    "any additional flags for ssh can be specified here",
				EnvVars:  []string{"PARAMETER_SSH_FLAG", "SSH_FLAG"},
				FilePath: "/vela/parameters/vela-ssh/ssh.flag,/vela/secrets/vela-ssh/ssh.flag",
			},
			&cli.StringFlag{
				Name:     "sshpass.password",
				Usage:    "password for use with destination target (used with sshpass)",
				EnvVars:  []string{"PARAMETER_SSHPASS_PASSWORD", "PARAMETER_PASSWORD", "SSHPASS_PASSWORD", "PASSWORD"},
				FilePath: "/vela/parameters/vela-ssh/sshpass.password,/vela/secrets/vela-ssh/sshpass.password",
			},
			&cli.StringFlag{
				Name:     "sshpass.passphrase",
				Usage:    "passphrase for use with identity file (used with sshpass)",
				EnvVars:  []string{"PARAMETER_SSHPASS_PASSPHRASE", "SSHPASS_PASSPHRASE"},
				FilePath: "/vela/parameters/vela-ssh/sshpass.passphrase,/vela/secrets/vela-ssh/sshpass.passphrase",
			},
			&cli.StringSliceFlag{
				Name:     "sshpass.flag",
				Usage:    "any additional flags for sshpass can be specified here)",
				EnvVars:  []string{"PARAMETER_SSHPASS_FLAG", "SSHPASS_FLAG"},
				FilePath: "/vela/parameters/vela-ssh/sshpass.flag,/vela/secrets/vela-ssh/sshpass.flag",
			},
			&cli.StringFlag{
				Name:    "ci",
				Usage:   "set the CI environment (if $CI is set output tries to be friendlier)",
				EnvVars: []string{"PARAMETER_CI", "CI"},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
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
		"docs":            "https://go-vela.github.io/docs/plugins/registry/pipeline/ssh",
		"registry":        "https://hub.docker.com/r/target/vela-ssh",
		"commit":          openssh.GitCommit,
		"version-plugin":  openssh.PluginVersion,
		"version-openssh": openssh.OpenSSHVersion,
		"version-sshpass": openssh.SSHPassVersion,
	}).Info("Vela SSH Plugin")

	bp := binarywrapper.Plugin{
		PluginConfig: &ssh.Config{
			Destination:          c.String("destination"),
			Command:              c.StringSlice("command"),
			IdentityFilePath:     c.StringSlice("identity-file.path"),
			IdentityFileContents: c.String("identity-file.contents"),
			SSHFlags:             c.StringSlice("ssh.flag"),
			SSHPassword:          c.String("sshpass.password"),
			SSHPassphrase:        c.String("sshpass.passphrase"),
			SSHPASSFlags:         c.StringSlice("sshpass.flag"),
		},
	}

	return bp.Exec()
}
