#!/bin/bash

set -e

function exit_with_msg {
  echo "${1}"
  exit 1
}

while [ $# -gt 0 ]; do
  case "${1}" in
    -e|--environment)
      ENVIRONMENT="${2}"
      shift
      ;;
    -h|--help)
      echo "Usage:"
      echo "$0 \\"
      echo "  -e|--environment <environment_name>"
      echo " [-h|--help]"
      echo "  -o|--owner <owner>"
      echo "  -p|--primaryregion <primary_region>"
      echo "  -s|--secondaryregion <secondary_region>"
      echo "  -t|--team <team>"
      exit 0
      ;;
    -o|--owner)
      OWNER="${2}"
      shift
      ;;
    -p|--primaryregion)
      PREGION="${2}"
      shift
      ;;
    -s|--secondaryregion)
      SREGION="${2}"
      shift
      ;;
    -t|--team)
      TEAM="${2}"
      shift
      ;;
    *)
      exit_with_msg "Error: Invalid argument '${1}'."
  esac
  shift
done

if [[ -f ../local.aws.yaml ]]; then
  PREFIX=$(cat ../local.aws.yaml | grep ^prefix | awk -F '[: #"]+' '{print $2}')
else
  PREFIX=$(cat ../aws.yaml | grep ^prefix | awk -F '[: #"]+' '{print $2}')
fi

[[ -z ${PREFIX} ]] && exit_with_msg "Can't locate deployment prefix. Exiting."
[[ ${#PREFIX} > 5 ]] && exit_with_msg "Prefix '${PREFIX}' is too long. Exiting."

[[ -z ${ENVIRONMENT} ]] && exit_with_msg "-e|--environment is a required parameter. Exiting."
[[ -z ${OWNER} ]] && exit_with_msg "-o|--owner is a required parameter. Exiting."
[[ -z ${PREGION} ]] && exit_with_msg "-p|--primaryregion is a required parameter. Exiting."
[[ -z ${SREGION} ]] && exit_with_msg "-s|--secondaryregion is a required parameter. Exiting."
[[ -z ${TEAM} ]] && exit_with_msg "-t|--team is a required parameter. Exiting."

echo "Deployment Owner: ${OWNER}"
echo "Environment: ${ENVIRONMENT}"
echo "Name Prefix: ${PREFIX}"
echo "Primary Region: ${PREGION}"
echo "Secondary Region: ${SREGION}"
echo "Support Team: ${TEAM}"

cp templates/env.tpl env.yaml
cp templates/region.tpl reg-primary/region.yaml
cp templates/region.tpl reg-secondary/region.yaml

sed -i -e "s:ENVIRONMENT:${ENVIRONMENT}:g" env.yaml
sed -i -e "s:OWNER:${OWNER}:g" env.yaml
sed -i -e "s:PREFIX:${PREFIX}:g" env.yaml
sed -i -e "s:PREGION:${PREGION}:g" env.yaml
sed -i -e "s:SREGION:${SREGION}:g" env.yaml
sed -i -e "s:REGION:${PREGION}:g" reg-primary/region.yaml
sed -i -e "s:ZONE:a:g" reg-primary/region.yaml
sed -i -e "s:REGION:${SREGION}:g" reg-secondary/region.yaml
sed -i -e "s:ZONE:a:g" reg-secondary/region.yaml
sed -i -e "s:TEAM:${TEAM}:g" env.yaml

aws configure set default.region ${PREGION}
aws configure set default.output json
