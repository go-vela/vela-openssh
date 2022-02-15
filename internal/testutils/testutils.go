package testutils

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/afero"
)

// The constants here are used by both the SCP & SSH Plugins for testing.
const (
	MockSCPPath              = "/usr/bin/scp"
	MockSSHPath              = "/usr/bin/ssh"
	MockSSHPassPath          = "/usr/bin/sshpass" // #nosec G101
	MockSSHPassword          = "hunter2"
	MockSSHPassphrase        = "correct horse battery staple"
	MockIdentityFileContents = "-----BEGIN OPENSSH PRIVATE KEY-----"
	WantFilePermissions      = os.FileMode(0o600)
)

// CreateMockFiles will create an in memory file system
// along with any arbitrary named empty files for testing.
func CreateMockFiles(t *testing.T, filename ...string) afero.Fs {
	fs := &afero.MemMapFs{}
	for _, file := range filename {
		if _, err := fs.Create(file); err != nil {
			t.Errorf("fs.Create() should not have raised an error: %s", err)
			t.FailNow()
		}
	}
	return fs
}

// ValidateMockFile checks to ensure that particular files exist
// along with specific permissions and contents as a common action used in tests.
func ValidateMockFile(t *testing.T, fs afero.Fs, wantFilePath, wantFileContents string) {
	fileInfo, err := fs.Stat(wantFilePath)
	if err != nil {
		t.Errorf("should not have raised an error checking file: %s", err)
		t.FailNow()
	}

	if fileInfo.Mode() != WantFilePermissions {
		t.Errorf("did not create file with correct octal permissions: %o, wanted %o", fileInfo.Mode(), WantFilePermissions)
		t.FailNow()
	}

	fileContents, err := afero.ReadFile(fs, wantFilePath)
	if err != nil {
		t.Errorf("should not have raised an error reading file: %s", err)
		t.FailNow()
	}

	if string(fileContents) != wantFileContents {
		t.Errorf("did not create file with correct contents\ngot:\n%q\nwant:\n%q", fileContents, wantFileContents)
		t.FailNow()
	}
}

// FlattenArguments is a helper to smash together strings and
// string slices into a flat string slice for testing purposes.
func FlattenArguments(args ...interface{}) (flatArgs []string) {
	for _, arg := range args {
		switch v := arg.(type) {
		case []string:
			flatArgs = append(flatArgs, v...)
		case string:
			flatArgs = append(flatArgs, v)
		default:
		}
	}
	return
}

// ArgCompare will allow us to compare some command line arguments
// to validate they match what we'd expect. We can't directly use
// a reflect.DeepEqual because of some randomized file names/suffixes.
func ArgCompare(wantArgs, gotArgs []string) bool {
	if len(wantArgs) != len(gotArgs) {
		return false
	}

	for i, wa := range wantArgs {
		if !strings.Contains(gotArgs[i], wa) {
			return false
		}
	}

	return true
}
