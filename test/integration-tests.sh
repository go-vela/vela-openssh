#!/bin/bash
set -e

SCRIPT_PATH="$( cd "$(dirname "$0")"; pwd -P )"
cd "$SCRIPT_PATH"

FAILURE_MESSAGES=()

####################
# Build Containers #
####################
build_images() {
  echo "### Building Images"
  docker-compose -f ./docker-compose-scp.yml build &&
  docker-compose -f ./docker-compose-ssh.yml build
}

#######
# SCP #
#######
test_scp() {
  echo "### Testing vela-scp:local image"
  docker-compose -f docker-compose-scp.yml run --rm scp-plugin-password &&
  docker-compose -f docker-compose-scp.yml run --rm scp-plugin-passphrase
}

cleanup_scp() {
  echo "### Cleaning up SCP containers"
  docker-compose -f docker-compose-scp.yml stop -t 1
  docker-compose -f docker-compose-scp.yml down
}

#######
# SSH #
#######
test_ssh() {
  echo "### Testing vela-ssh:local image"
  docker-compose -f docker-compose-ssh.yml run --rm ssh-plugin-password &&
  docker-compose -f docker-compose-ssh.yml run --rm ssh-plugin-passphrase
}

cleanup_ssh() {
  echo "### Cleaning up SSH containers"
  docker-compose -f docker-compose-ssh.yml stop -t 1
  docker-compose -f docker-compose-ssh.yml down
}

##############
# Test Logic #
##############
if ! build_images; then
  printf "Unable to build integration test images"
  exit 1
fi

if ! test_scp; then
  FAILURE_MESSAGES+=("Failure testing SCP Images")
fi
cleanup_scp

if ! test_ssh; then
  FAILURE_MESSAGES+=("Failure testing SSH Images")
fi
cleanup_ssh

for MSG in "${FAILURE_MESSAGES[@]}"; do
  printf "%s\n" "$MSG"
done

if [ "${#FAILURE_MESSAGES[@]}" -gt 0 ]; then
  exit 2
fi
