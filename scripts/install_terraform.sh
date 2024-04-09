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

readonly TERRAFORM_INSTALL_DIR="/usr/local/bin"
mkdir -p "$TERRAFORM_INSTALL_DIR"

# Make sure we have write permissions to target directory before downloading
if [ ! -w "$TERRAFORM_INSTALL_DIR" ] ; then
	>&2 echo "User does not have write permission to folder: ${TERRAFORM_INSTALL_DIR}"
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

# Determine the version of terraform to install
if [[ -z VERSIONS_FILE ]]; then
  VERSIONS_FILE="${SCRIPT_DIR}/../versions.yaml"
fi
readonly VERSIONS_FILE
>&2 echo "Reading $VERSIONS_FILE"

readonly TERRAFORM_VERSION="$(cat $VERSIONS_FILE | grep '^terraform_install_version: ' | awk -F':' '{gsub(/^[[:space:]]*["]*|["]*[[:space:]]*$/,"",$2); print $2}')"
if [[ -z "$TERRAFORM_VERSION" ]]; then
	>&2 echo 'Unable to find version number'
	exit 1
fi

# Install terraform
readonly TERRAFORM_BIN="${TERRAFORM_INSTALL_DIR}/terraform"
cd "$(mktemp -d)"
wget "https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_${OS_IDENTIFIER}_amd64.zip" -O terraform.zip
rm -f "${TERRAFORM_BIN}" || echo "Terraform is not installed."
unzip terraform.zip -d "$TERRAFORM_INSTALL_DIR"
chmod +x "${TERRAFORM_BIN}"

# Cleanup
rm terraform.zip
