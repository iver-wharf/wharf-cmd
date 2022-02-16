#!/usr/bin/env bash

# -u: exit when referencing undeclared variable
set -u

# ${VAR:?"Message"} means it has to be set a cannot be empty
# ${VAR?"Message"} means it has to be set, but could be empty
# ${VAR:="fallback"} means it will use "fallback" if not set or empty

: ${NUGET_TOKEN:?"Missing required NuGet token"}
: ${NUGET_REPO:?"Missing required NuGet repo"}
: ${NUGET_PROJECT_PATH:?"Missing required .NET project path"}
: ${NUGET_VERSION:?"Missing required NuGet package version"}
: ${NUGET_SKIP_DUP:="false"}

NUGET_SKIP_DUP_PARAM=""
if [ "$NUGET_SKIP_DUP" == "true" ]
then
    NUGET_SKIP_DUP_PARAM="--skip-duplicate"
fi

echo '$ dotnet pack "$NUGET_PROJECT_PATH" --output nugets/ /property:Version=$NUGET_VERSION'
if ! dotnet pack "$NUGET_PROJECT_PATH" --output nugets/ /property:"$NUGET_VERSION"
then
    echo "[!] Failed to build and pack the NuGet packages. Aborting" >&2
    exit 1
fi

# mkfifo solution found here: https://stackoverflow.com/a/61470435/3163818
mkfifo nuget_push_output

# Must push all NuGets one by one because the 500 status code issue
# will make the pushing abort half-way.
for nuget in nugets/*
do
    echo
    echo "$ dotnet nuget push \"$nuget\" --api-key *REDACTED* --source \"$NUGET_REPO\" $NUGET_SKIP_DUP_PARAM"

    # Store logs in nuget-push.log and output to the terminal
    tee nuget-push.log < nuget_push_output &
    TEE_PROCESS_ID=$!

    dotnet nuget push "$nuget" \
        --api-key "$NUGET_TOKEN" \
        --source "$NUGET_REPO" \
        "$NUGET_SKIP_DUP_PARAM" \
        > nuget_push_output 2>&1
    EXIT_STATUS=$?

    kill $TEE_PROCESS_ID

    # Some NuGet servers incorrectly returns HTTP status code 500 instead
    # of 409 on duplicates.
    # The --skip-duplicate flag only handles 409 status codes.
    # So we need this extra check via grep on the programs ouput.
    if grep -q 'The server is configured to not allow overwriting packages that already exist.' nuget-push.log
    then
        if [ "$NUGET_SKIP_DUP" != "true" ]
        then
            echo
            echo "[!] Detected duplicate, but NUGET_SKIP_DUP is not set to true. Aborting"
            exit 2
        else
            echo
            echo "[x] Detected duplicate, and NUGET_SKIP_DUP is set to true. Silently continuing."
        fi
    elif [ $EXIT_STATUS != 0 ]
    then
        echo "[!] Failed to push NuGet packages. Aborting" >&2
        exit 3
    fi
done
