# Contributing

## Getting Started

We'd love to accept your contributions to this project! If you are a first time contributor, please review our [Contributing Guidelines](https://go-vela.github.io/docs/community/contributing_guidelines/) before proceeding.

### Prerequisites

* [Review the commit guide we follow](https://chris.beams.io/posts/git-commit/#seven-rules) - ensure your commits follow our standards
* Review our [style guide](https://go-vela.github.io/docs/community/contributing_guidelines/#style-guide) to ensure your code is clean and consistent.
* [Docker](https://docs.docker.com/install/) - building block for local development
* [Docker Compose](https://docs.docker.com/compose/install/) - start up local development
* [Make](https://www.gnu.org/software/make/) - start up local development

### Setup

* [Fork](/fork) this repository

* Clone this repository to your workstation:

```bash
# Clone the project
git clone git@github.com:go-vela/vela-openssh.git $HOME/go-vela/vela-openssh
```

* Navigate to the repository code:

```bash
# Change into the project directory
cd $HOME/go-vela/vela-openssh
```

* Point the original code at your fork:

```bash
# Add a remote branch pointing to your fork
git remote add fork https://github.com/your_fork/vela-openssh
```

### Running Locally

* Navigate to the repository code:

```bash
# Change into the project directory
cd $HOME/go-vela/vela-openssh
```

* Build the repository code:

```bash
# Build the code with `make`
make build
```

* Run the repository code:

```bash
# Run the code with `make`
make run
```

### Development

* Navigate to the repository code:

```bash
# Change into the project directory
cd $HOME/go-vela/vela-openssh
```

* Write your code and tests to implement the changes you desire.

* Test the repository code (ensures your changes don't break existing functionality):

```bash
# execute the `test` target with `make`
make test
```

* Ensure your code meets the project standards:

```bash
# Clean the code with `make`
make clean
```

* Push to your fork:

```bash
# Push your code up to your fork
git push fork main
```

* Make sure to follow our [PR process](https://go-vela.github.io/docs/community/contributing_guidelines/#development-workflow) when opening a pull request

Thank you for your contribution!
