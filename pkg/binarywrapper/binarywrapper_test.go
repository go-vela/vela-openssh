// SPDX-License-Identifier: Apache-2.0

// binarywrapper_test was chosen to be blackbox testing of the wrapper
// to try and better understand the behavior that end consumers of the wrapper
// might experience as they're using this. That, and there isn't much "internal"
// to whitebox test anyhow.
package binarywrapper_test

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/go-vela/vela-openssh/pkg/binarywrapper"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"testing"
)

const (
	testMainEnvVar        = "env-var"
	testMainSuccessOutput = "success-output"
	testMainFailOutput    = "fail-output"
)

// TestMain is used so that we can mock calls to binaries that need
// to return specific output, or errors, exit codes, etc.
func TestMain(m *testing.M) {
	switch os.Getenv("GO_MAIN_TEST_CASE") {
	case "":
		os.Exit(m.Run())
	case testMainEnvVar:
		if strings.Contains(strings.Join(os.Args, ""), "$") {
			os.Exit(1)
		}

		os.Exit(0)
	case testMainSuccessOutput:
		if len(os.Args) != 4 {
			fmt.Printf("invalid os.Args: %s", strings.Join(os.Args, " "))
			os.Exit(2)
		}

		fmt.Println(os.Args[2])
		fmt.Fprint(os.Stderr, os.Args[3])
		os.Exit(0)
	case testMainFailOutput:
		if len(os.Args) != 4 {
			fmt.Printf("invalid os.Args: %s", strings.Join(os.Args, " "))
			os.Exit(3)
		}

		fmt.Println(os.Args[2])
		fmt.Fprint(os.Stderr, os.Args[3])
		os.Exit(4)
	}
}

type mockExecConfig struct {
	validationError string
	setupError      string
	binaryPath      string
	arguments       []string
	environment     map[string]string
}

func (m *mockExecConfig) Validate() error {
	if m.validationError != "" {
		return errors.New(m.validationError)
	}

	return nil
}

func (m *mockExecConfig) Setup() error {
	if m.setupError != "" {
		return errors.New(m.setupError)
	}

	return nil
}

func (m *mockExecConfig) Binary() string {
	return m.binaryPath
}

func (m *mockExecConfig) Arguments() []string {
	if len(m.arguments) > 0 {
		return m.arguments
	}

	return []string{}
}

func (m *mockExecConfig) Environment() map[string]string {
	if m.environment != nil {
		return m.environment
	}

	return map[string]string{}
}

func TestExecSuccess(t *testing.T) {
	tests := map[string]struct {
		config     *mockExecConfig
		execStyle  binarywrapper.ExecStyle
		wantStdOut string
		wantStdErr string
	}{
		"formats arguments with env vars before executing": {
			execStyle: binarywrapper.OSExecCommand,
			config: &mockExecConfig{
				binaryPath: os.Args[0],
				arguments:  []string{"$SOME_TEST"},
				environment: map[string]string{
					"GO_MAIN_TEST_CASE": testMainEnvVar,
					"SOME_TEST":         "Howdy!",
				},
			},
		},
		"OSExecCommand captures stdout and stderr of successful run": {
			execStyle: binarywrapper.OSExecCommand,
			config: &mockExecConfig{
				binaryPath: os.Args[0],
				arguments:  []string{"$VAR1", "$VAR2"},
				environment: map[string]string{
					"GO_MAIN_TEST_CASE": testMainSuccessOutput,
					"VAR1":              "stdout",
					"VAR2":              "stderr",
				},
			},
			wantStdOut: "stdout",
			wantStdErr: "stderr",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			p := binarywrapper.Plugin{}
			p.ExecStyle = test.execStyle
			p.PluginConfig = test.config

			if test.execStyle == binarywrapper.SyscallExec && len(test.wantStdOut)+len(test.wantStdErr) > 0 {
				t.Error("Exec() can't check output for the SyscallExec ExecStyle")
				t.FailNow()
			}

			if test.execStyle > 0 {
				p.ExecStyle = test.execStyle
			}

			// This is just so we can capture what output is coming back
			// from the logging in our exec method.
			var outputBuffer bytes.Buffer
			logrus.SetOutput(&outputBuffer)

			if err := p.Exec(); err != nil {
				t.Errorf("Exec() should not have raised error %q", err)
				t.FailNow()
			}

			if len(test.wantStdOut) > 0 && !strings.Contains(outputBuffer.String(), test.wantStdOut) {
				t.Errorf("Exec() mismatch stdout\ngot:    %s\nwanted: %s", outputBuffer.String(), test.wantStdOut)
				t.FailNow()
			}

			if len(test.wantStdErr) > 0 && !strings.Contains(outputBuffer.String(), test.wantStdErr) {
				t.Errorf("Exec() mismatch stderr\ngot:    %s\nwanted: %s", outputBuffer.String(), test.wantStdErr)
			}
		})
	}
}

func TestExecError(t *testing.T) {
	tests := map[string]struct {
		plugin     *binarywrapper.Plugin
		wantErr    error
		wantStdOut string
		wantStdErr string
	}{
		"returns error when no plugin configured": {
			wantErr: binarywrapper.ErrExec,
		},
		"returns error when Validate fails": {
			plugin: func() *binarywrapper.Plugin {
				p := binarywrapper.Plugin{
					PluginConfig: &mockExecConfig{
						validationError: "validation has failed",
					},
				}
				return &p
			}(),
			wantErr: binarywrapper.ErrValidation,
		},
		"returns error when Setup fails": {
			plugin: func() *binarywrapper.Plugin {
				p := binarywrapper.Plugin{
					PluginConfig: &mockExecConfig{
						setupError: "setup has failed",
					},
				}
				return &p
			}(),
			wantErr: binarywrapper.ErrSetup,
		},
		"returns error with unknown ExecStyle": {
			plugin: func() *binarywrapper.Plugin {
				p := binarywrapper.Plugin{
					ExecStyle:    -1,
					PluginConfig: &mockExecConfig{},
				}
				return &p
			}(),
			wantErr: binarywrapper.ErrUnknownExecStyle,
		},
		"SyscallExec returns with error if binary missing": {
			plugin: func() *binarywrapper.Plugin {
				p := binarywrapper.Plugin{
					ExecStyle:    binarywrapper.SyscallExec,
					PluginConfig: &mockExecConfig{binaryPath: "this-should-not-exist"},
				}
				return &p
			}(),
			wantErr: binarywrapper.ErrMissingBinary,
		},
		"OSExecCommand returns with error if binary missing": {
			plugin: func() *binarywrapper.Plugin {
				p := binarywrapper.Plugin{
					ExecStyle:    binarywrapper.OSExecCommand,
					PluginConfig: &mockExecConfig{binaryPath: "this-should-not-exist"},
				}
				return &p
			}(),
			wantErr: binarywrapper.ErrExec,
		},
		"OSExecCommand captures stdout and stderr of failed run": {
			plugin: func() *binarywrapper.Plugin {
				p := binarywrapper.Plugin{
					ExecStyle: binarywrapper.OSExecCommand,
					PluginConfig: &mockExecConfig{
						binaryPath: os.Args[0],
						arguments:  []string{"$VAR1", "$VAR2"},
						environment: map[string]string{
							"GO_MAIN_TEST_CASE": testMainFailOutput,
							"VAR1":              "stdout",
							"VAR2":              "stderr",
						},
					},
				}
				return &p
			}(),
			wantErr:    binarywrapper.ErrExec,
			wantStdOut: "stdout",
			wantStdErr: "stderr",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			// This is just so we can capture what output is coming back
			// from the logging in our exec method.
			var outputBuffer bytes.Buffer
			logrus.SetOutput(&outputBuffer)

			if err := test.plugin.Exec(); err == nil {
				t.Errorf("Exec() should have raised an error")
				t.FailNow()
			} else if test.wantErr != nil && err != nil && !errors.Is(err, test.wantErr) {
				t.Errorf("Exec() returned wrong error\ngot:    %s\nwanted: %s", err, test.wantErr)
				t.FailNow()
			}

			if len(test.wantStdOut) > 0 && !strings.Contains(outputBuffer.String(), test.wantStdOut) {
				t.Errorf("Exec() mismatch stdout\ngot:    %s\nwanted: %s", outputBuffer.String(), test.wantStdOut)
				t.FailNow()
			}

			if len(test.wantStdErr) > 0 && !strings.Contains(outputBuffer.String(), test.wantStdErr) {
				t.Errorf("Exec() mismatch stderr\ngot:    %s\nwanted: %s", outputBuffer.String(), test.wantStdErr)
			}
		})
	}
}
