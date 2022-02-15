## Description

This plugin is part of the [OpenSSH](https://www.openssh.com/) suite of plugins which enables you to use the `scp` binary in a Vela pipeline.

Source Code: https://github.com/go-vela/vela-openssh

Registry: https://hub.docker.com/r/target/vela-scp

## Usage

Because the plugin is a thin wrapper around the [`scp`](https://man.openbsd.org/scp) binary, the syntax and parameters follow from the [OpenSSH manual](https://man.openbsd.org/scp). The plugin will take care of some basic secrets identity management tasks for you, most importantly is that when an identity file is provided as a secret the plugin will place the file into the filesystem and change the permissions to match what the binary expects, and then add it to the list of identity files tried as part of authentication. Additionally, if using a password or passphrase for authentication or for unlocking an identity file, the [`sshpass`](https://linux.die.net/man/1/sshpass) binary will be used to provide those credentials without interactive user input.

> **NOTE:**
>
> Users should refrain from using latest as the tag for images.
>
> It is recommended to use a semantically versioned tag instead.

### Basic usage
```yaml
steps:
  - name: scp basic usage
    image: target/vela-scp:latest
    pull: always
    parameters:
      source:
        - ./relative/path/to/file/in/vela/workspace
        - a_user_name@some_remote_host_name:/absolute/path/on/remote/system
      target: a_different_user@some_other_host:~
```

### Using the `scp://` schema for non-standard ports
```diff
steps:
  - name: scp to non-standard port
    image: target/vela-scp:latest
    pull: always
    parameters:
      source:
        - ./relative/path/to/file/in/vela/workspace
        - a_user_name@some_remote_host_name:/absolute/path/on/remote/system
+     target: scp://a_different_user@some_remote_host_name:12345/path
```

### Passing additional `scp` flags
```diff
steps:
  - name: override default scp flags
    image: target/vela-scp:latest
    pull: always
    parameters:
      source:
        - ./relative/path/to/folder/in/vela/workspace
      target: scp://a_different_user@some_remote_host_name:12345/path
+     scp_flag:
+       - "-r"
+       - "-o StrictHostKeyChecking=yes"
```

### Using a password for authentication
```diff
steps:
  - name: password for authentication
    image: target/vela-scp:latest
    pull: always
+   secrets:
+     - source: my_non_user_account_password
+       target: sshpass_password
    parameters:
      source:
        - ./relative/path/to/file/in/vela/workspace
      target: scp://a_different_user@some_remote_host_name:12345/path
```

### Using an identity file (WITHOUT a passphrase) from an internal secret
```diff
steps:
  - name: identity file contents from an internal secret
    image: target/vela-scp:latest
    pull: always
+   secrets:
+     - source: my_non_user_account_id_rsa_file_contents
+       target: identity_file_contents
    parameters:
      source:
        - ./relative/path/to/file/in/vela/workspace
      target: scp://a_different_user@some_remote_host_name:12345/path
```

### Using an identity file (WITH a passphrase) from an internal secret
```diff
steps:
  - name: identity file contents with passphrase from an internal secret
    image: target/vela-scp:latest
    pull: always
+   secrets:
+     - source: my_non_user_account_id_rsa_file_contents
+       target: identity_file_contents
+     - source: my_non_user_account_id_rsa_passphrase
+       target: sshpass_passphrase
    parameters:
      source:
        - ./relative/path/to/file/in/vela/workspace
      target: scp://a_different_user@some_remote_host_name:12345/path
```

### Using an existing identity file from the workspace
```diff
steps:
  - name: identity file from the workspace
    image: target/vela-scp:latest
    pull: always
    parameters:
      source:
        - ./relative/path/to/file/in/vela/workspace
      target: scp://a_different_user@some_remote_host_name:12345/path
+     identity_file_path: ./some/workspace/location/id_rsa
```

### Using additional secrets in other parameters
```diff
steps:
  - name: additional secrets in other parameters
    image: target/vela-scp:latest
    pull: always
+   secrets:
+     - source: some_secret_user
+       target: secret_user
+     - source: some_secret_host
+       target: secret_host
+     - source: some_secret_port
+       target: secret_port
    parameters:
      source:
        - ./relative/path/to/file/in/vela/workspace
+     target: scp://$SECRET_USER@$SECRET_HOST:$SECRET_PORT/path
```

### Using the container without the plugin logic
```diff
steps:
  - name: override plugin logic to use scp directly
    image: target/vela-scp:latest
    pull: always
+   commands:
+     - scp -i ./some/existing/id_rsa ./relative/path/to/file/in/vela/workspace username@hostname:/path/
```

## Parameters & Secrets
> **NOTE:**
>
> The plugin supports reading all parameters via environment variables or files.
>
> Any values set from a file take precedence over values set from the environment.
>
> Don't confuse the source/target syntax of native secrets with the source and target parameters required for the scp binary.

| Name | Description | Required | Accepts Multiple Values? | Default | Environment Variables | File Paths |
| --- | --- | --- | --- | --- | --- | --- |
| `source` | The source option from the [`scp` manual](https://man.openbsd.org/scp). | :white_check_mark: | :white_check_mark: | | `PARAMETER_SOURCE`<br>`SOURCE` | `/vela/parameters/vela-scp/source`<br>`/vela/secrets/vela-scp/source` |
| `target` | The target option from the [`scp` manual](https://man.openbsd.org/scp). | :white_check_mark: | :x: | | `PARAMETER_TARGET`<br>`TARGET` | `/vela/parameters/vela-scp/target`<br>`/vela/secrets/vela-scp/target` |
| `identity_file_path` | A path for where the [`scp`](https://man.openbsd.org/scp) binary should look for existing identity files.<br>These are NOT auto created by the plugin as they must be created and managed by a user and only referenced here. | :x: | :white_check_mark: | | `PARAMETER_IDENTITY_FILE_PATH`<br>`IDENTITY_FILE_PATH`<br>`PARAMETER_SSH_KEY_PATH`<br>`SSH_KEY_PATH` | `/vela/parameters/vela-scp/identity-file.path`<br>`/vela/secrets/vela-scp/identity-file.path` |
| `identity_file_contents` | The raw contents of an identity file for use with [`scp`](https://man.openbsd.org/scp).<br>The plugin will take the raw contents and place it in a temporary location in the workspace with the correct permissions and inject it as an identity file to use during execution. | :x: | :x: | | `PARAMETER_IDENTITY_FILE_CONTENTS`<br>`IDENTITY_FILE_CONTENTS`<br>`PARAMETER_SSH_KEY`<br>`SSH_KEY` | `/vela/parameters/vela-scp/identity-file.contents`<br>`/vela/secrets/vela-scp/identity-file.contents` |
| `scp_flag` | Any additional options from the [`scp` manual](https://man.openbsd.org/scp).<br>These will override the default options and be placed between the identity file options and the source/target options at the end. | :x: | :white_check_mark: | `-o StrictHostKeyChecking=no`<br>`-o UserKnownHostsFile=/dev/null` | `PARAMETER_SCP_FLAG`<br>`SCP_FLAG` | `/vela/parameters/vela-scp/scp.flag`<br>`/vela/secrets/vela-scp/scp.flag` |
| `sshpass_password` | If any systems require a password for authentication it can be specified here, and the [`sshpass`](https://linux.die.net/man/1/sshpass) binary will be used in conjunction with [`scp`](https://man.openbsd.org/scp). | :x: | :x: | | `PARAMETER_SSHPASS_PASSWORD`<br>`PARAMETER_PASSWORD`<br>`SSHPASS_PASSWORD`<br>`PASSWORD` | `/vela/parameters/vela-scp/sshpass.password`<br>`/vela/secrets/vela-scp/sshpass.password` |
| `sshpass_passphrase` | If any identity files require a passphrase for authentication it can be specified here, and the [`sshpass`](https://linux.die.net/man/1/sshpass) binary will be used in conjunction with [`scp`](https://man.openbsd.org/scp). | :x: | :x: | | `PARAMETER_SSHPASS_PASSPHRASE`<br>`SSHPASS_PASSPHRASE` | `/vela/parameters/vela-scp/sshpass.passphrase`<br>`/vela/secrets/vela-scp/sshpass.passphrase` |
| `sshpass_flag` | Any additional options from the [`sshpass` manual](https://linux.die.net/man/1/sshpass). | :x: | :white_check_mark: | | `PARAMETER_SSHPASS_FLAG`<br>`SSHPASS_FLAG` | `/vela/parameters/vela-scp/sshpass.flag`<br>`/vela/secrets/vela-scp/sshpass.flag` |
