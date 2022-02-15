package main

import (
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
		Action:  run,
		Version: openssh.PluginVersion.Semantic(),
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

	logrus.WithFields(logrus.Fields{
		"code":     "https://github.com/go-vela/vela-openssh",
		"docs":     "https://go-vela.github.io/docs/plugins/registry/ssh",
		"registry": "https://hub.docker.com/r/target/vela-ssh",
		"version":  openssh.PluginVersion.Semantic(),
		"commit":   openssh.PluginVersion.Metadata.GitCommit,
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
