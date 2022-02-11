package ssh

import (
	"errors"
	"strings"
	"testing"

	"github.com/go-vela/vela-openssh/internal/openssh"
	"github.com/go-vela/vela-openssh/internal/testutils"
	"github.com/spf13/afero"
)

var (
	mockCommand = []string{
		"whoami",
		"pwd",
		"ls ~",
	}
)

const (
	mockDestination      = "some-user@some-host:~"
	mockFormattedCommand = "whoami && pwd && ls ~"
)

func TestValidateSuccess(t *testing.T) {
	tests := map[string]Config{
		"returns no errors when properly configured": {
			Command:     mockCommand,
			Destination: mockDestination,
		},
		"returns no errors when using an SSH Password": {
			Command:     mockCommand,
			Destination: mockDestination,
			SSHPassword: testutils.MockSSHPassword,
		},
		"returns no errors when using an SSH Passphrase": {
			Command:       mockCommand,
			Destination:   mockDestination,
			SSHPassphrase: testutils.MockSSHPassphrase,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := test.Validate(); err != nil {
				t.Errorf("Validate() should not have raised error %q", err)
			}
		})
	}
}

func TestValidateErrors(t *testing.T) {
	tests := map[string]struct {
		config  Config
		wantErr error
	}{
		"with everything missing": {},
		"with destination missing": {
			wantErr: ErrMissingDestination,
		},
		"with command missing": {
			config: Config{
				Destination: mockDestination,
			},
			wantErr: ErrMissingCommand,
		},
		"with password and passphrase set": {
			config: Config{
				Command:       mockCommand,
				Destination:   mockDestination,
				SSHPassword:   testutils.MockSSHPassword,
				SSHPassphrase: testutils.MockSSHPassphrase,
			},
			wantErr: openssh.ErrAmbiguousAuth,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if err := test.config.Validate(); err == nil {
				t.Errorf("Validate() should have raised an error")
			} else if (test.wantErr != nil) && (err != test.wantErr) {
				t.Errorf("Validate() returned wrong error\ngot:    %s\nwanted: %s", err, test.wantErr)
			}
		})
	}
}

func TestSetupSuccess(t *testing.T) {
	tests := map[string]struct {
		config Config
		mockFS afero.Fs
	}{
		"sets default FS if not set": {
			config: Config{},
		},
		"can find binaries in common locations": {
			config: Config{},
			mockFS: testutils.CreateMockFiles(t, "./ssh", "/usr/local/bin/sshpass"),
		},
		"can find binaries in common locations pt2": {
			config: Config{},
			mockFS: testutils.CreateMockFiles(t, testutils.MockSSHPath, testutils.MockSSHPassPath),
		},
		"creates identity file from raw string and sets permissions and is default first identity file": {
			config: Config{
				IdentityFileContents: testutils.MockIdentityFileContents,
			},
			mockFS: testutils.CreateMockFiles(t, testutils.MockSSHPath, testutils.MockSSHPassPath),
		},
		"creates password file and saves temp location": {
			config: Config{
				SSHPassword: testutils.MockSSHPassword,
			},
			mockFS: testutils.CreateMockFiles(t, testutils.MockSSHPath, testutils.MockSSHPassPath),
		},
		"creates passphrase file and saves temp location": {
			config: Config{
				SSHPassphrase: testutils.MockSSHPassphrase,
			},
			mockFS: testutils.CreateMockFiles(t, testutils.MockSSHPath, testutils.MockSSHPassPath),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockFS != nil {
				test.config.fs = test.mockFS
			}

			if err := test.config.Setup(); err != nil && test.mockFS != nil {
				t.Errorf("Setup() should not have raised error %q", err)
				t.FailNow()
			}

			if test.config.fs == nil {
				t.Error("Setup() should have created default file system")
				t.FailNow()
			}

			if len(test.config.IdentityFileContents) > 0 {
				if len(test.config.IdentityFilePath) == 0 || !strings.Contains(test.config.IdentityFilePath[0], openssh.TempFileDirectory) {
					t.Error("Setup() did not add file first in the IdentityFile slice")
					t.FailNow()
				}

				testutils.ValidateMockFile(t, test.mockFS, test.config.IdentityFilePath[0], test.config.IdentityFileContents)
			}

			if test.config.SSHPassword != "" {
				if len(test.config.locationPasswordFile) == 0 {
					t.Error("Setup() did not set password file location")
					t.FailNow()
				}
				testutils.ValidateMockFile(t, test.mockFS, test.config.locationPasswordFile, test.config.SSHPassword)
			}

			if test.config.SSHPassphrase != "" {
				if len(test.config.locationPassphraseFile) == 0 {
					t.Error("Setup() did not set passphrase file location")
					t.FailNow()
				}
				testutils.ValidateMockFile(t, test.mockFS, test.config.locationPassphraseFile, test.config.SSHPassphrase)
			}
		})
	}
}

func TestSetupErrors(t *testing.T) {
	tests := map[string]struct {
		config  Config
		mockFS  afero.Fs
		wantErr error
	}{
		"when ssh binary missing": {
			mockFS:  testutils.CreateMockFiles(t, testutils.MockSSHPassPath),
			wantErr: openssh.ErrMissingSSH,
		},
		"when sshpass binary missing": {
			mockFS:  testutils.CreateMockFiles(t, testutils.MockSSHPath),
			wantErr: openssh.ErrMissingSSHPASS,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockFS != nil {
				test.config.fs = test.mockFS
			}

			if err := test.config.Setup(); err == nil {
				t.Errorf("Setup() should have raised an error")
			} else if test.wantErr != nil && err != nil && !errors.Is(err, test.wantErr) {
				t.Errorf("Setup() returned wrong error\ngot:    %s\nwanted: %s", err, test.wantErr)
			}
		})
	}
}

func TestBinary(t *testing.T) {
	tests := map[string]struct {
		config  Config
		mockFS  afero.Fs
		wantSSH bool
	}{
		"uses ssh by default": {
			config:  Config{},
			mockFS:  testutils.CreateMockFiles(t, testutils.MockSSHPath, testutils.MockSSHPassPath),
			wantSSH: true,
		},
		"uses sshpass when sshpass flags set": {
			config: Config{
				SSHPASSFlags: []string{"-v"},
			},
			mockFS:  testutils.CreateMockFiles(t, testutils.MockSSHPath, testutils.MockSSHPassPath),
			wantSSH: false,
		},
		"uses sshpass when ssh password set": {
			config: Config{
				SSHPassword: testutils.MockSSHPassword,
			},
			mockFS:  testutils.CreateMockFiles(t, testutils.MockSSHPath, testutils.MockSSHPassPath),
			wantSSH: false,
		},
		"uses sshpass when ssh passphrase set": {
			config: Config{
				SSHPassphrase: testutils.MockSSHPassphrase,
			},
			mockFS:  testutils.CreateMockFiles(t, testutils.MockSSHPath, testutils.MockSSHPassPath),
			wantSSH: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if test.mockFS != nil {
				test.config.fs = test.mockFS
			}

			if err := test.config.Setup(); err != nil && test.mockFS != nil {
				t.Errorf("Setup() should not have raised error %q", err)
				t.FailNow()
			}

			if test.wantSSH && test.config.Binary() != testutils.MockSSHPath {
				t.Errorf("Binary() should have return ssh location")
			} else if !test.wantSSH && test.config.Binary() != testutils.MockSSHPassPath {
				t.Errorf("Binary() should have return sshpass location")
			}
		})
	}
}

func TestArguments(t *testing.T) {
	tests := map[string]struct {
		config      Config
		wantCommand []string
	}{
		"basic ssh usage": {
			config: Config{
				Command:     mockCommand,
				Destination: mockDestination,
			},
			wantCommand: testutils.FlattenArguments(
				testutils.MockSSHPath,
				openssh.DefaultSSHFlags,
				mockDestination,
				mockFormattedCommand,
			),
		},
		"basic sshpass usage": {
			config: Config{
				Command:      mockCommand,
				Destination:  mockDestination,
				SSHPASSFlags: []string{"-h"},
			},
			wantCommand: testutils.FlattenArguments(
				testutils.MockSSHPassPath,
				"-h",
				testutils.MockSSHPath,
				openssh.DefaultSSHFlags,
				mockDestination,
				mockFormattedCommand,
			),
		},
		"custom ssh flags override defaults": {
			config: Config{
				Command:     mockCommand,
				Destination: mockDestination,
				SSHFlags:    []string{"-h"},
			},
			wantCommand: testutils.FlattenArguments(
				testutils.MockSSHPath,
				"-h",
				mockDestination,
				mockFormattedCommand,
			),
		},
		"custom sshpass flags override defaults": {
			config: Config{
				Command:      mockCommand,
				Destination:  mockDestination,
				SSHPASSFlags: []string{"-v"},
			},
			wantCommand: testutils.FlattenArguments(
				testutils.MockSSHPassPath,
				"-v",
				testutils.MockSSHPath,
				openssh.DefaultSSHFlags,
				mockDestination,
				mockFormattedCommand,
			),
		},
		"ssh password sets file path": {
			config: Config{
				Command:     mockCommand,
				Destination: mockDestination,
				SSHPassword: testutils.MockSSHPassword,
			},
			wantCommand: testutils.FlattenArguments(
				testutils.MockSSHPassPath,
				openssh.DefaultSSHPassFlags,
				"-f", "/tmp/vela-plugin-openssh-password-file-",
				testutils.MockSSHPath,
				openssh.DefaultSSHFlags,
				mockDestination,
				mockFormattedCommand,
			),
		},
		"ssh passphrase sets file path": {
			config: Config{
				Command:       mockCommand,
				Destination:   mockDestination,
				SSHPassphrase: testutils.MockSSHPassphrase,
			},
			wantCommand: testutils.FlattenArguments(
				testutils.MockSSHPassPath,
				openssh.DefaultSSHPassFlags,
				"-Passphrase",
				"-f", "/tmp/vela-plugin-openssh-passphrase-file-",
				testutils.MockSSHPath,
				openssh.DefaultSSHFlags,
				mockDestination,
				mockFormattedCommand,
			),
		},
		"multiple identity files set with identity contents and ssh flags": {
			config: Config{
				Command:              mockCommand,
				Destination:          mockDestination,
				IdentityFilePath:     []string{"~/.ssh/id_rsa", "$HOME/.ssh/id_dsa"},
				IdentityFileContents: testutils.MockIdentityFileContents,
				SSHFlags:             []string{"-o", "StrictHostKeyChecking=no", "-o", "UserKnownHostsFile=/dev/null"},
			},
			wantCommand: testutils.FlattenArguments(
				testutils.MockSSHPath,
				"-o", "StrictHostKeyChecking=no",
				"-o", "UserKnownHostsFile=/dev/null",
				"-i", "/tmp/vela-plugin-openssh-identity-file-",
				"-i", "~/.ssh/id_rsa",
				"-i", "$HOME/.ssh/id_dsa",
				mockDestination,
				mockFormattedCommand,
			),
		},
		"everything all at once": {
			config: Config{
				Command:              mockCommand,
				Destination:          mockDestination,
				IdentityFilePath:     []string{"~/.ssh/id_rsa", "$HOME/.ssh/id_dsa"},
				IdentityFileContents: testutils.MockIdentityFileContents,
				SSHFlags:             []string{"-o", "StrictHostKeyChecking=yes"},
				SSHPassphrase:        testutils.MockSSHPassphrase,
				SSHPASSFlags:         []string{"-v"},
			},
			wantCommand: testutils.FlattenArguments(
				testutils.MockSSHPassPath,
				"-v",
				"-Passphrase",
				"-f", "/tmp/vela-plugin-openssh-passphrase-file-",
				testutils.MockSSHPath,
				"-o", "StrictHostKeyChecking=yes",
				"-i", "/tmp/vela-plugin-openssh-identity-file-",
				"-i", "~/.ssh/id_rsa",
				"-i", "$HOME/.ssh/id_dsa",
				mockDestination,
				mockFormattedCommand,
			),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			test.config.fs = testutils.CreateMockFiles(t, testutils.MockSSHPath, testutils.MockSSHPath, testutils.MockSSHPassPath)

			if err := test.config.Validate(); err != nil {
				t.Errorf("Validate() should not have raised error %q", err)
				t.FailNow()
			}

			if err := test.config.Setup(); err != nil {
				t.Errorf("Setup() should not have raised error %q", err)
				t.FailNow()
			}

			if !testutils.ArgCompare(test.wantCommand, test.config.Arguments()) {
				t.Errorf("arguments mismatched\ngot:    %s\nwanted: %s", test.config.Arguments(), test.wantCommand)
			}
		})
	}
}

func TestEnvironment(t *testing.T) {
	c := &Config{}
	env := c.Environment()

	if len(env) == 0 {
		t.Errorf("Environment() should not be empty")
		t.FailNow()
	}

	if len(env["VELA_SSH_PLUGIN_VERSION"]) == 0 {
		t.Errorf("Environment() VELA_SSH_PLUGIN_VERSION should be set")
		t.FailNow()
	}
}
