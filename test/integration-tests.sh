#!/bin/bash
set -e

# These test names match with the container names in the respective
# docker compose files. This allows us to run each container step by step
# and record which ones fail so we can report on this later.
SCP_TESTS=(
  basic-usage-with-scp-schema
  additional-scp-flags
  password-auth
  passphrase-auth
  additional-secrets-in-params
  override-plugin
  ensure-version-info-set
)

SSH_TESTS=(
  basic-usage-with-ssh-schema
  additional-ssh-flags
  password-auth
  passphrase-auth
  additional-secrets-in-params
  override-plugin
  ensure-version-info-set
)

# Make sure we move into the folder where the integration tests
# are located so that we don't need to worry about rel vs. abs paths.
SCRIPT_PATH="$(
  cd "$(dirname "$0")"
  pwd -P
)"
cd "$SCRIPT_PATH"

# Global failure messages we'll append as tests are ran.
FAILURE_MESSAGES=()

setup() {
  local BINARY=$1
  printf "### Building images for %s tests\n" "$BINARY"
  if ! docker-compose -f "docker-compose-$BINARY.yml" build; then
    printf "❌ Unable to build integration test %s image" "$BINARY"
    exit 1
  fi
}

run_tests() {
  local BINARY=$1
  shift
  printf "### Testing vela-%s:local image\n" "$BINARY"
  for TEST_NAME in "$@"; do
    printf "#### Executing test '%s-%s'\n" "$BINARY" "$TEST_NAME"
    if ! docker-compose -f "docker-compose-$BINARY.yml" run --rm "$TEST_NAME"; then
      FAILURE_MESSAGES+=("❌ Failure testing '$BINARY-$TEST_NAME'")
    fi
  done
}

teardown() {
  local BINARY=$1
  printf "###Cleaning up %s containers\n" "$BINARY"
  docker-compose -f "docker-compose-$BINARY.yml" stop -t 1
  docker-compose -f "docker-compose-$BINARY.yml" down
}

##############
# Test Logic #
##############
setup "scp"
run_tests "scp" "${SCP_TESTS[@]}"
teardown "scp"

setup "ssh"
run_tests "ssh" "${SSH_TESTS[@]}"
teardown "ssh"

for MSG in "${FAILURE_MESSAGES[@]}"; do
  printf "%s\n" "$MSG"
done

if [ "${#FAILURE_MESSAGES[@]}" -gt 0 ]; then
  exit 2
fi

printf "✅ Integration tests passed successfully\n"
