// SPDX-License-Identifier: Apache-2.0

// Package binarywrapper is a utility package that makes wrapping binaries a little easier
// as it aims to provide a common structure to use for converting binaries
// into plugins. Along the way it allows for some setup tasks, validation,
// and execution.
package binarywrapper

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
)

var (
	// ErrValidation is returned if the plugin fails validation.
	ErrValidation = errors.New("plugin failed validation")

	// ErrSetup is returned if the plugin fails to setup.
	ErrSetup = errors.New("plugin failed setup")

	// ErrUnknownExecStyle is returned when a plugin is configured with an unknown
	// execution style. Look at binarywrapper.ExecStyle for the available options.
	ErrUnknownExecStyle = errors.New("unknown ExecStyle, look at binarywrapper.ExecStyle for the available options")

	// ErrMissingBinary is returned when the binary referenced is not found.
	ErrMissingBinary = errors.New("missing binary")

	// ErrExec is returned for any generic execution based error.
	ErrExec = errors.New("execution error")
)

// PluginConfig holds the key methods required for a binarywrapper.Plugin to
// operate and enforces a common structure for wrapping binary plugins.
type PluginConfig interface {

	// Validate is responsible for checking all of the plugin's
	// configuration is valid before the plugin Setup() is executed.
	Validate() error

	// Setup is responsible to create all required files are in
	// place before any executable or shell actions are created.
	Setup() error

	// Binary should return the absolute path to the binary that should take
	// over when this plugin has been validated for execution. This should
	// not use any environmental variables like $HOME.
	Binary() string

	// Arguments should return the arguments to the binary when this plugin is executed.
	// If these include environmental variables like $HOME they'll
	// be expanded before execution by the shell.
	Arguments() []string

	// Environment should return any additional environmental variables
	// to use for the binary when this plugin is executed.
	Environment() map[string]string
}

// ExecStyle defines the types of execution paradims exists for the plugin.
type ExecStyle int

const (
	// SyscallExec sets the execution style such that when the binary is finally called
	// the system turns over all processing and resources over to said binary without
	// the overhead of Go still being in the way. What this means in practice is that
	// you're at the whim of how the binary behaves for logging, output, etc.
	SyscallExec ExecStyle = iota

	// OSExecCommand sets the execution style such that when the binary is called it is
	// done using a subprocess command. The output of the command is captured when finished
	// and not streamed, so if you desire streaming based output you should use SyscallExec for now.
	OSExecCommand
)

// Plugin holds the configuration required for a binarywrapper.Plugin to operate.
// We need a struct that implements the required binarywrapper.PluginConfig functions.
// It can also optionally set or override the execution style before the plugin is called.
type Plugin struct {
	ExecStyle
	PluginConfig
}

// Exec will call the plugin Validate, Setup and Exec methods
// This uses syscall.Exec to take over the processing.
// What this means is whatever binary is defined in the plugin will
// take over execution and if there are no errors this is the end of the
// relevant go code and handling required for the plugin. Think of this
// like the binary taking the place of the go code if the binary is found.
func (p *Plugin) Exec() error {
	if p == nil {
		return ErrExec
	}

	if err := p.Validate(); err != nil {
		return ErrValidation
	}

	if err := p.Setup(); err != nil {
		return ErrSetup
	}

	// Log some good debugging information here. There is a purposeful choice
	// here to NOT expand the arguments with environmental variables yet
	// as those might contain secrets or other information we don't want to leak.
	pluginArguments := p.Arguments()

	logrus.WithFields(logrus.Fields{
		"binary":    p.Binary(),
		"arguments": p.Arguments(),
	}).Info()

	// The subprocess call later expects that the first argument is always the binary
	// that is being called, so if the arguments don't contain the binary as the first
	// argument, slap it on the front and call it a day.
	if len(pluginArguments) == 0 || p.Binary() != pluginArguments[0] {
		pluginArguments = append([]string{p.Binary()}, pluginArguments...)
	}

	// Adopt any additional environmental variables from plugin
	// We set them in the OS environment so that we can us os.ExpandEnv
	// below, as well as placing this environment in with the binary when
	// execution happens further below.
	for k, v := range p.Environment() {
		os.Setenv(k, v)
	}

	// Using environmental variables like $HOME will be used literally
	// if we don't range over our arguments and expand them nicely.
	// We ideally don't do this inside of the plugin so we can log
	// all unexpanded arguments above.
	var expandedArgs []string
	for _, arg := range pluginArguments {
		expandedArgs = append(expandedArgs, os.ExpandEnv(arg))
	}

	// Having the option of execution styles allows users of this wrapper
	// to specify if they want the takeover style of syscall.Exec or the
	// subprocess behavior of exec.Command since they have their own nuances.
	if p.ExecStyle == OSExecCommand {
		var outBuffer, errorBuffer bytes.Buffer

		// #nosec G204
		cmd := exec.CommandContext(context.Background(), p.Binary(), expandedArgs...)
		cmd.Env = os.Environ()
		cmd.Stdout = &outBuffer
		cmd.Stderr = &errorBuffer

		if err := cmd.Run(); err != nil {
			if outBuffer.Len() > 0 {
				logrus.Info(outBuffer.String())
			}

			if errorBuffer.Len() > 0 {
				logrus.Error(errorBuffer.String())
			}

			return fmt.Errorf("%w: %w", ErrExec, err)
		}

		if outBuffer.Len() > 0 {
			logrus.Info(outBuffer.String())
		}

		if errorBuffer.Len() > 0 {
			logrus.Error(errorBuffer.String())
		}
	} else if p.ExecStyle == SyscallExec {
		// This portion of the code will replace the running go code with
		// whatever the binary by the specified plugin happens to be, but only
		// if the binary is found, otherwise it'll raise a file not found error.
		// If this does exist, and no other execve errors occur, we'll never reach
		// the return (or any code) past this function call (even in tests).
		// #nosec G204
		if err := syscall.Exec(p.Binary(), expandedArgs, os.Environ()); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("%w: %s", ErrMissingBinary, p.Binary())
			}

			return fmt.Errorf("%w: %w", ErrExec, err)
		}
	} else {
		return fmt.Errorf("%w: %d", ErrUnknownExecStyle, p.ExecStyle)
	}

	return nil
}
