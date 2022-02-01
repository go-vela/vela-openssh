## Description

This plugin is part of the [OpenSSH](https://www.openssh.com/) suite of plugins which enables you to use the `ssh` binary in a Vela pipeline.

Source Code: https://github.com/go-vela/vela-openssh

Registry: https://hub.docker.com/r/target/vela-ssh

## Usage

Because the plugin is a thin wrapper around the [`ssh`](https://man.openbsd.org/ssh) binary, the syntax and parameters follow from the [OpenSSH manual](https://man.openbsd.org/ssh). The plugin will take care of some basic secrets identity management tasks for you, most importantly is that when an identity file is provided as a secret the plugin will place the file into the filesystem and change the permissions to match what the binary expects, and then add it to the list of identity files tried as part of authentication. Additionally, if using a password or passphrase for authentication or for unlocking an identity file, the [`sshpass`](https://linux.die.net/man/1/sshpass) binary will be used to provide those credentials without interactive user input.

> **NOTE:**
>
> Users should refrain from using latest as the tag for images.
>
> It is recommended to use a semantically versioned tag instead.

### Basic usage with no-auth methods
```yaml
steps:
  - name: ssh using no-auth
    image: target/vela-ssh:latest
    pull: always
    parameters:
      destination: a_different_user@some_other_host
      command:
        - echo "Hello Vela!"
        - /some/path/to/a/remote/script.sh
```

### Using the `ssh://` schema for non-standard ports
```diff
steps:
  - name: ssh to non-standard port
    image: target/vela-ssh:latest
    pull: always
    parameters:
+     destination: ssh://a_different_user@some_remote_host_name:12345
      command:
        - echo "Hello Vela!"
```

### Passing additional `ssh` flags
```diff
steps:
  - name: override default ssh flags
    image: target/vela-ssh:latest
    pull: always
    parameters:
      destination: ssh://a_different_user@some_remote_host_name:12345
      command:
        - echo "Hello Vela!"
+     ssh_flag:
+       - "-o StrictHostKeyChecking=yes"
```

### Using a password for authentication
```diff
steps:
  - name: password for authentication
    image: target/vela-ssh:latest
    pull: always
+   secrets:
+     - source: my_non_user_account_password
+       target: sshpass_password
    parameters:
      destination: ssh://a_different_user@some_remote_host_name:12345
      command:
        - echo "Hello Vela!"
```

### Using an identity file (WITHOUT a passphrase) from an internal secret
```diff
steps:
  - name: identity file contents from an internal secret
    image: target/vela-ssh:latest
    pull: always
+   secrets:
+     - source: my_non_user_account_id_rsa_file_contents
+       target: identity_file_contents
    parameters:
      destination: ssh://a_different_user@some_remote_host_name:12345
      command:
        - echo "Hello Vela!"
```

### Using an identity file (WITH a passphrase) from an internal secret
```diff
steps:
  - name: identity file contents with passphrase from an internal secret
    image: target/vela-ssh:latest
    pull: always
+   secrets:
+     - source: my_non_user_account_id_rsa_file_contents
+       target: identity_file_contents
+     - source: my_non_user_account_id_rsa_passphrase
+       target: sshpass_passphrase
    parameters:
      destination: ssh://a_different_user@some_remote_host_name:12345
      command:
        - echo "Hello Vela!"
```

### Using an existing identity file from the workspace
```diff
steps:
  - name: identity file from the workspace
    image: target/vela-ssh:latest
    pull: always
    parameters:
      destination: ssh://a_different_user@some_remote_host_name:12345
      command:
        - echo "Hello Vela!"
+     identity_file_path: ./some/workspace/location/id_rsa
```

### Using additional secrets in other parameters
```diff
steps:
  - name: additional secrets in other parameters
    image: target/vela-ssh:latest
    pull: always
+   secrets:
+     - source: some_secret_user
+       target: secret_user
+     - source: some_secret_host
+       target: secret_host
+     - source: some_secret_port
+       target: secret_port
    parameters:
+     destination: ssh://$SECRET_USER@$SECRET_HOST:$SOME_SECRET_PORT
      command:
        - echo "Hello Vela!"
```

### Using the container without the plugin logic
```diff
steps:
  - name: override plugin logic to use ssh directly
    image: target/vela-ssh:latest
    pull: always
    # Note that this ISN'T part of the parameters section.
+   commands:
+     - ssh -i ./some/existing/id_rsa username@hostname some-bin
```

## Parameters & Secrets
> **NOTE:**
>
> The plugin supports reading all parameters via environment variables or files.
>
> Any values set from a file take precedence over values set from the environment.
>
> Don't confuse the `commands` parameter in a traditional Vela step with the `command` option required for the ssh binary.

| Name | Description | Required | Accepts Multiple Values? | Default | Environment Variables | File Paths |
| --- | --- | --- | --- | --- | --- | --- |
| `destination` | The destination option from the [`ssh` manual](https://man.openbsd.org/ssh). | :white_check_mark: | :x: | | `PARAMETER_DESTINATION`<br>`DESTINATION`<br>`PARAMETER_HOST` | `/vela/parameters/vela-ssh/destination`<br>`/vela/secrets/vela-ssh/destination` |
| `command` | The command option from the [`ssh` manual](https://man.openbsd.org/ssh). | :white_check_mark: | :white_check_mark: | | `PARAMETER_COMMAND`<br>`COMMAND`<br>`PARAMETER_SCRIPT`<br>`SCRIPT` | `/vela/parameters/vela-ssh/command`<br>`/vela/secrets/vela-ssh/command` |
| `identity_file_path` | A path for where the [`ssh`](https://man.openbsd.org/ssh) binary should look for existing identity files.<br>These are NOT auto created by the plugin as they must be created and managed by a user and only referenced here. | :x: | :white_check_mark: | | `PARAMETER_IDENTITY_FILE_PATH`<br>`IDENTITY_FILE_PATH`<br>`PARAMETER_SSH_KEY_PATH`<br>`SSH_KEY_PATH` | `/vela/parameters/vela-ssh/identity-file.path`<br>`/vela/secrets/vela-ssh/identity-file.path` |
| `identity_file_contents` | The raw contents of an identity file for use with [`ssh`](https://man.openbsd.org/ssh).<br>The plugin will take the raw contents and place it in a temporary location in the workspace with the correct permissions and inject it as an identity file to use during execution. | :x: | :x: | | `PARAMETER_IDENTITY_FILE_CONTENTS`<br>`IDENTITY_FILE_CONTENTS`<br>`PARAMETER_SSH_KEY`<br>`SSH_KEY` | `/vela/parameters/vela-ssh/identity-file.contents`<br>`/vela/secrets/vela-ssh/identity-file.contents` |
| `ssh_flag` | Any additional options from the [`ssh` manual](https://man.openbsd.org/ssh).<br>These will override the default options and be placed between the identity file options and the destination/command options at the end. | :x: | :white_check_mark: | `-o StrictHostKeyChecking=no`<br>`-o UserKnownHostsFile=/dev/null` | `PARAMETER_SSH_FLAG`<br>`SSH_FLAG` | `/vela/parameters/vela-ssh/ssh.flag`<br>`/vela/secrets/vela-ssh/ssh.flag` |
| `sshpass_password` | If any systems require a password for authentication it can be specified here, and the [`sshpass`](https://linux.die.net/man/1/sshpass) binary will be used in conjunction with [`ssh`](https://man.openbsd.org/ssh). | :x: | :x: | | `PARAMETER_SSHPASS_PASSWORD`<br>`PARAMETER_PASSWORD`<br>`SSHPASS_PASSWORD`<br>`PASSWORD` | `/vela/parameters/vela-ssh/sshpass.password`<br>`/vela/secrets/vela-ssh/sshpass.password` |
| `sshpass_passphrase` | If any identity files require a passphrase for authentication it can be specified here, and the [`sshpass`](https://linux.die.net/man/1/sshpass) binary will be used in conjunction with [`ssh`](https://man.openbsd.org/ssh). | :x: | :x: | | `PARAMETER_SSHPASS_PASSPHRASE`<br>`SSHPASS_PASSPHRASE` | `/vela/parameters/vela-ssh/sshpass.passphrase`<br>`/vela/secrets/vela-ssh/sshpass.passphrase` |
| `sshpass_flag` | Any additional options from the [`sshpass` manual](https://linux.die.net/man/1/sshpass). | :x: | :white_check_mark: | | `PARAMETER_SSHPASS_FLAG`<br>`SSHPASS_FLAG` | `/vela/parameters/vela-ssh/sshpass.flag`<br>`/vela/secrets/vela-ssh/sshpass.flag` |
