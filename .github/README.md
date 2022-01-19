# vela-openssh

[![license](https://img.shields.io/crates/l/gl.svg)](../LICENSE)
[![GoDoc](https://godoc.org/github.com/go-vela/vela-openssh?status.svg)](https://godoc.org/github.com/go-vela/vela-openssh)
[![Go Report Card](https://goreportcard.com/badge/go-vela/vela-openssh)](https://goreportcard.com/report/go-vela/vela-openssh)
[![codecov](https://codecov.io/gh/go-vela/vela-openssh/branch/master/graph/badge.svg)](https://codecov.io/gh/go-vela/vela-openssh)

A set of Vela plugins designed to make common SSH and SCP actions easy to do within the Vela CI environment.

Internally, these plugins are wrappers around the [OpenSSH](https://www.openssh.com/) `scp` and `ssh` binaries. To assist in some functionality there is also an additional use of the [`sshpass`](https://linux.die.net/man/1/sshpass) utility.

## Documentation

For installation and usage, please [visit our docs](https://go-vela.github.io/docs).

## Contributing

We are always welcome to new pull requests!

Please see our [contributing](CONTRIBUTING.md) docs for further instructions.

## Support

We are always here to help!

Please see our [support](SUPPORT.md) documentation for further instructions.

## Copyright and License

```
Copyright (c) 2022 Target Brands, Inc.
```

[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0)
