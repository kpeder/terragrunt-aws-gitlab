#!/bin/bash

set -e

function exit_with_msg {
  echo "${1}"
  exit 1
}

while [ $# -gt 0 ]; do
  case "${1}" in
    -h|--help)
      echo "Usage:"
      echo "$0 \\"
      echo "  [-h : --help]"
      echo "  [<target directory>]"
      exit 0
      ;;
    *)
      [[ -d ${1} ]] || exit_with_msg "directory ${1} not found"
      TARGET=${1}
  esac
  shift
done

if [[ -z ${TARGET} ]]; then
    TARGET=.
    echo "Directory not specified; using current directory."
fi

find ${TARGET} -type d -name ".terragrunt-cache" -prune -exec rm -rf {} \;
find ${TARGET} -type f -name ".terraform.lock.hcl" -prune -exec rm -rf {} \;
