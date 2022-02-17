#!/bin/sh
set -e

BINARY="$1"

if [ -z "${OPENSSH_VERSION}" ]; then
  printf "OPENSSH_VERSION missing\n"
  exit 1
fi

if [ -z "${SSHPASS_VERSION}" ]; then
  printf "SSHPASS_VERSION missing\n"
  exit 2
fi

if "/bin/vela-$BINARY" -v | grep unknown; then
  printf "Version information isn't set\n"
  exit 3
fi
