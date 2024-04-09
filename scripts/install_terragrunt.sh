#!/bin/bash
set -euo pipefail

function exit_with_msg {
  echo "${1}"
  exit 1
}

while [ $# -gt 0 ]; do
  case "${1}" in
    -h|--help)
      echo "Usage:"
      echo "$0 \\"
      echo " [-h|--help]"
      echo "  -v|--version-file <path>"
      exit 0
      ;;
    -v|--version-file)
      VERSIONS_FILE="${2}"
      shift
      ;;
    *)
      exit_with_msg "Error: Invalid argument '${1}'."
  esac
  shift
done

readonly TERRAGRUNT_INSTALL_DIR="/usr/local/bin"
mkdir -p "$TERRAGRUNT_INSTALL_DIR"

# Make sure we have write permissions to target directory before downloading
if [ ! -w "$TERRAGRUNT_INSTALL_DIR" ] ; then
  >&2 echo "User does not have write permission to folder: ${TERRAGRUNT_INSTALL_DIR}"
  exit 1
fi

# Get the directory where the script is located
readonly SCRIPT_DIR="$(dirname $0)"

# Get the operating system identifier.
# May be one of "linux", "darwin", "freebsd" or "openbsd".
OS_IDENTIFIER="${1:-}"
if [[ -z "$OS_IDENTIFIER" ]]; then
  # POSIX compliant OS detection
  OS_IDENTIFIER=$(uname -s | tr '[:upper:]' '[:lower:]')
  >&2 echo "Detected OS Identifier: ${OS_IDENTIFIER}"
fi
readonly OS_IDENTIFIER

# Determine the version of terragrunt to install
if [[ -z VERSIONS_FILE ]]; then
  VERSIONS_FILE="${SCRIPT_DIR}/../versions.yaml"
fi
readonly VERSIONS_FILE
>&2 echo "Reading $VERSIONS_FILE"

readonly TERRAGRUNT_VERSION="$(cat $VERSIONS_FILE | grep '^terragrunt_install_version: ' | awk -F':' '{gsub(/^[[:space:]]*["]*|["]*[[:space:]]*$/,"",$2); print $2}')"
if [[ -z "$TERRAGRUNT_VERSION" ]]; then
  >&2 echo 'Unable to find version number'
  exit 1
fi

# Install terragrunt
readonly TERRAGRUNT_BIN="$TERRAGRUNT_INSTALL_DIR/terragrunt"
cd "$(mktemp -d)"
wget "https://github.com/gruntwork-io/terragrunt/releases/download/v${TERRAGRUNT_VERSION}/terragrunt_${OS_IDENTIFIER}_amd64" -O terragrunt
rm -f "$TERRAGRUNT_BIN" || echo "Terragrunt is not installed."
cp terragrunt "$TERRAGRUNT_BIN"
chmod +x "$TERRAGRUNT_BIN"

# Cleanup
rm terragrunt
