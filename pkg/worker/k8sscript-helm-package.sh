#!/usr/bin/env bash

# ${VAR:?"Message"} means it has to be set a cannot be empty
# ${VAR?"Message"} means it has to be set, but could be empty
# ${VAR:="fallback"} means it will use "fallback" if not set or empty

: ${CHART_PATH:?"Missing required Helm chart path"}
: ${CHART_VERSION?"Missing required Helm chart version"}
: ${CHART_REPO:?"Missing required .NET project path"}
: ${REG_USER:?"Missing required NuGet package version"}
: ${REG_PASS:?"Missing required NuGet package version"}

VERSION_FLAG=""

if [ ! -z "$CHART_VERSION" ]
then
    VERSION_FLAG="--version=$CHART_VERSION"
fi

echo "\$ helm package $CHART_PATH $VERSION_FLAG"
helm package "$CHART_PATH" "$VERSION_FLAG"

echo "\$ helm push *.tgz $CHART_REPO --insecure --username *REDACTED* --password *REDACTED*"
helm push *.tgz "$CHART_REPO" --insecure --username "$REG_USER" --password "$REG_PASS"
