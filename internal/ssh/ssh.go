// Copyright (c) 2022 Target Brands, Inc. All rights reserved.
//
// Use of this source code is governed by the LICENSE file in this repository.

// This package wraps the ssh portion of the OpenSSH binaries to allow
// for executing scripts and commands on remote systems.

package ssh

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/afero"

	"github.com/go-vela/vela-openssh/internal/openssh"
)

var (
	// ErrMissingDestination is a returned when the plugin is missing the destination parameter.
	ErrMissingDestination = errors.New("missing destination parameter")

	// ErrMissingCommand is a returned when the plugin is missing the command parameter.
	ErrMissingCommand = errors.New("missing command parameter")
)

type Config struct {
	// Config from CLI/Env/External

	// Command is a required parameter containing all of the commands
	// to execute on the remote system
	Command []string

	// Destination is the machine where the plugin will execute the command.
	Destination string

	// IdentityFilePath is the path to the identity file to use
	// for authenticating against remote systems. You can specify
	// multiple files to use each one in turn if needed.
	IdentityFilePath []string

	// IdentityFileContents is a string representation of an identity
	// file. This is used since the plugin needs to consume secrets
	// and those sometimes come through as environmental variables
	// so we'll need to take the raw contents and place them into a file
	// so that the binaries later can use them.
	IdentityFileContents string

	// SSHFlags is for setting or overriding any sort of SSH features.
	SSHFlags []string

	// SSHPassword is used for authenticating against systems with a password
	// instead of an identity file. Since this is passed to sshpass it's
	// limited to only one password for whatever remote systems are in use.
	// Prefer to use identity files if possible since you can specify quite a bit more.
	SSHPassword string

	// SSHPassphrase is used for the identity files specified. Just like with SSHPassword
	// this is handed off to sshpass to use in authentication and it only supports
	// one parameter, so this passphrase needs to be viable for all identity files
	// that contain a passphrase. If needing multiple identity files be sure
	// they all use the same passphrase for the remote systems.
	SSHPassphrase string

	// SSHPASSFlags is for setting or overriding any sort of sshpass features.
	SSHPASSFlags []string

	// Internal flags & data
	fs                     afero.Fs
	locationSSHbinary      string
	locationSSHPASSbinary  string
	locationPassphraseFile string
	locationPasswordFile   string
}

// Validate checks some basic plugin configuration parameters
// to ensure everything is set that we need or expect and that
// and sort of conflicting parameters are sorted appropriately.
// Note that we're really not validating that the format of the
// parameters is correct since we'll just rely on surfacing those
// errors by using the binary itself. Why duplicate that validation
// logic when the binaries can do that for us?
func (c *Config) Validate() error {
	if len(c.Destination) == 0 {
		return ErrMissingDestination
	}

	if len(c.Command) == 0 {
		return ErrMissingCommand
	}

	if len(c.SSHPassword) > 0 && len(c.SSHPassphrase) > 0 {
		return openssh.ErrAmbiguousAuth
	}

	return nil
}

// Setup will make sure all of the internal configuration of
// the plugin is set and ready to go along with any sorts of
// file system side effects and preparations are done.
func (c *Config) Setup() error {
	// This wouldn't be nil in testing situations but in
	// general it will be nil for most runtime scenarios.
	// This allows us to mock the filesystem in testing
	// but for live running of the plugin we'll use the
	// real file system.
	if c.fs == nil {
		c.fs = afero.NewOsFs()
	}

	// Pickup the SSH & sshpass binaries from whatever location
	// they might be currently installed into. Inside a plugin this
	// should stay static, but when debugging and running this
	// outside of the container it's nice if it picks up user binaries.
	for _, path := range openssh.BinSearchLocations {
		tempSSHPath := fmt.Sprintf("%s/ssh", path)
		tempSSHPassPath := fmt.Sprintf("%s/sshpass", path)

		if ok, _ := afero.Exists(c.fs, tempSSHPath); ok && len(c.locationSSHbinary) == 0 {
			c.locationSSHbinary = tempSSHPath
		}

		if ok, _ := afero.Exists(c.fs, tempSSHPassPath); ok && len(c.locationSSHPASSbinary) == 0 {
			c.locationSSHPASSbinary = tempSSHPassPath
		}
	}

	if c.locationSSHbinary == "" {
		return openssh.ErrMissingSSH
	}

	if c.locationSSHPASSbinary == "" {
		return openssh.ErrMissingSSHPASS
	}

	if c.IdentityFileContents != "" {
		filename, err := openssh.CreateRestrictedFile(c.fs, openssh.TempIdentityFilePrefix, c.IdentityFileContents)
		if err != nil {
			return err
		}

		c.IdentityFilePath = append([]string{filename}, c.IdentityFilePath...)
	}

	if c.SSHPassword != "" {
		filename, err := openssh.CreateRestrictedFile(c.fs, openssh.TempPasswordPrefix, c.SSHPassword)
		if err != nil {
			return err
		}

		c.locationPasswordFile = filename
	}

	if c.SSHPassphrase != "" {
		filename, err := openssh.CreateRestrictedFile(c.fs, openssh.TempPassphrasePrefix, c.SSHPassphrase)
		if err != nil {
			return err
		}

		c.locationPassphraseFile = filename
	}

	return nil
}

// Binary returns the system path location for either the ssh binary (by default)
// or the sshpass binary depending on if the plugin configuration requires
// the use of sshpass or not.
func (c *Config) Binary() string {
	if c.useSSHPass() {
		return c.locationSSHPASSbinary
	}

	return c.locationSSHbinary
}

// Arguments returns a string slice representation of all of the command
// line arguments required for the binary to work. If using sshpass parameters
// they will be placed at the start of the slice while all others float to the end.
// Think of these as the commands a user would normally manually type to use the binary.
func (c *Config) Arguments() []string {
	args := []string{}

	// sshpass expects to be first in the chain of commands called
	// so if we're using it, we'll need to bump all arguments to the end
	// and set any sshpass flags by the user before specifying the SSH binary.
	if c.useSSHPass() {
		if len(c.SSHPASSFlags) == 0 {
			args = append([]string{c.locationSSHPASSbinary}, openssh.DefaultSSHPassFlags...)
		} else {
			args = append([]string{c.locationSSHPASSbinary}, c.SSHPASSFlags...)
		}

		if len(c.SSHPassword) > 0 {
			args = append(args, "-f")
			args = append(args, c.locationPasswordFile)
		} else if len(c.SSHPassphrase) > 0 {
			args = append(args, "-Passphrase")
			args = append(args, "-f")
			args = append(args, c.locationPassphraseFile)
		}

		args = append(args, c.locationSSHbinary)
	} else {
		args = append(args, c.locationSSHbinary)
	}

	if len(c.SSHFlags) == 0 {
		args = append(args, openssh.DefaultSSHFlags...)
	} else {
		args = append(args, c.SSHFlags...)
	}

	if len(c.IdentityFilePath) > 0 {
		for _, file := range c.IdentityFilePath {
			args = append(args, "-i")
			args = append(args, file)
		}
	}

	args = append(args, c.Destination)

	args = append(args, strings.Join(c.Command, " && "))

	return args
}

// Environment returns a mapping of key/value strings representing any additional
// environmental variables a particular plugin might need. This plugin doesn't
// require anything in particular, but a few env vars are provided so that users
// can place that in their pipeline for diagnostic purposes.

func (c *Config) Environment() map[string]string {
	return map[string]string{
		"VELA_SSH_PLUGIN_VERSION": openssh.PluginVersion.Semantic(),
		"VELA_SSH_PLUGIN_COMMIT":  openssh.PluginVersion.Metadata.GitCommit,
	}
}

// useSSHPass returns true if the plugin configuration requires the use of the sshpass binary.
// This typically only happens if a password or passphrase is provided but if a user also wants
// to override the sshpass flags then we also will inject sshpass into the mix.

func (c *Config) useSSHPass() bool {
	return len(c.SSHPASSFlags)+
		len(c.SSHPassword)+
		len(c.SSHPassphrase) > 0
}
