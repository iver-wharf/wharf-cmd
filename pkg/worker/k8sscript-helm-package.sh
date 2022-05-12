#!/usr/bin/env bash

# -e: exit when any command fails
# -u: exit when referencing undeclared variable
# -o pipefail: exit when any commands fails in pipes
set -euo pipefail

# ${VAR:?"Message"} means it has to be set and cannot be empty
# ${VAR?"Message"} means it has to be set, but can be empty
# ${VAR:="fallback"} means it will use "fallback" if not set or empty

: ${CHART_PATH:?"Missing required Helm chart path"}
: ${CHART_VERSION?"Missing required Helm chart version"}
: ${CHART_REPO:?"Missing required Helm registry URL"}
: ${REG_USER:?"Missing required Helm registry username"}
: ${REG_PASS:?"Missing required Helm registry password"}

VERSION_FLAG=""

if [ ! -z "$CHART_VERSION" ]
then
    VERSION_FLAG="--version=$CHART_VERSION"
fi

echo "\$ helm package $CHART_PATH $VERSION_FLAG"
helm package "$CHART_PATH" "$VERSION_FLAG"

echo "\$ helm push *.tgz $CHART_REPO --insecure --username *REDACTED* --password *REDACTED*"
helm push *.tgz "$CHART_REPO" --insecure --username "$REG_USER" --password "$REG_PASS"
