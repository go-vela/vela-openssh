package openssh

import (
	"errors"
	"fmt"

	// We're not using anything from Kaniko plugin except the version.go file
	// so that we don't need to duplicate that bit of code and keep it in sync
	// across many different plugins with the same need. Ideally that version.go
	// would be in its own package somewhere so it wasn't tied to any particular plugin.
	// If this gets changed in the future, be sure to update the Makefile references
	// for the build-time injection of the Git Tag, Commit, etc.
	"github.com/go-vela/vela-kaniko/version"
	"github.com/spf13/afero"
)

var (
	// PluginVersion provides a common place to pull the plugin configuration from.
	PluginVersion = version.New()

	// ErrMissingSCP is returned when the scp binary isn't found in the locations below.
	ErrMissingSCP = errors.New("can't find scp binary")

	// ErrMissingSSH is returned when the ssh binary isn't found in the locations below.
	ErrMissingSSH = errors.New("can't find ssh binary")

	// ErrMissingSSHPASS is returned when the sshpass binary isn't found in the locations below.
	ErrMissingSSHPASS = errors.New("can't find sshpass binary")

	// ErrAmbiguousAuth is returned when both password and passphrase specified.
	ErrAmbiguousAuth = errors.New("can't use both password and passphrase for authentication")
)

// These constants are where the plugins should store the temporary files during execution.
const (
	TempFileDirectory      = "/tmp/"
	TempIdentityFilePrefix = "vela-plugin-openssh-identity-file-"
	TempPassphrasePrefix   = "vela-plugin-openssh-passphrase-file-" // #nosec G101
	TempPasswordPrefix     = "vela-plugin-openssh-password-file-"   // #nosec G101

	// Read-write only for the user who creates this file.
	TempFilePermissions = 0o600
)

var (
	// BinSearchLocations are the common binary locations to look for scp and sshpass
	// we could probably pick up the $PATH env var and then walk that looking around
	// but it's a bit more work than just assuming some sane defaults since we'll have
	// full control over how we construct and create the Dockerfile to hold this plugin.
	BinSearchLocations = []string{".", "/usr/local/bin", "/usr/bin", "/bin"}

	// DefaultSSHFlags makes the default behavior to not check host keys or save
	// them to the known hosts. This is because it'll typically ask for a user interaction
	// and that will break the plugin flow. If a user specifies their own flags these should
	// get overwritten.
	DefaultSSHFlags = []string{"-o StrictHostKeyChecking=no", "-o UserKnownHostsFile=/dev/null"}

	// DefaultSCPFlags uses the Default SSH flags because scp uses SSH under the covers and
	// benefits from the same default host checking behavior.
	DefaultSCPFlags = DefaultSSHFlags

	// DefaultSSHPassFlags is just like the SCP flags in that these are to aid with debugging
	// but if a user specifies any flags these will be disregarded.
	DefaultSSHPassFlags = []string{}
)

// CreateRestrictedFile will create a new file in a given location with a given prefix
// while ensuring it has the correct restricted permissions for the scp and ssh binaries to be happy.
func CreateRestrictedFile(fs afero.Fs, fileprefix string, contents string) (string, error) {
	file, err := afero.TempFile(fs, TempFileDirectory, fileprefix)
	if err != nil {
		return "", fmt.Errorf("couldn't create temporary file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write([]byte(contents)); err != nil {
		return "", fmt.Errorf("couldn't inject temporary file contents: %w", err)
	}

	if err := fs.Chmod(file.Name(), TempFilePermissions); err != nil {
		return "", fmt.Errorf("couldn't set file permissions: %w", err)
	}
	return file.Name(), nil
}
