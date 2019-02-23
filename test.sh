#!/bin/bash

set -e
trap 'last_command=$current_command; current_command=$BASH_COMMAND' DEBUG
trap 'CMD=${last_command} RET=$?; if [[ $RET -ne 0 ]]; then echo "\"${CMD}\" command failed with exit code $RET."; fi' EXIT
SCRIPTPATH="$( cd "$(dirname "$0")" ; pwd -P )"
cd ${SCRIPTPATH}

go test . -count=1

# golint and gofmt always return 0... so we need to capture the output and test it
LINT=$(golint .)
if ! [ -z "${LINT}" ]; then
    echo ${LINT}
    exit 1
fi

FMT=$(gofmt -l -s *.go)
if ! [ -z ${FMT} ]; then
    echo ''
    echo "Some files have bad style. Run the following commands to fix them:"
    for LINE in ${FMT}
    do
        echo "  gofmt -s -w `pwd`/${LINE}"
    done
    echo ''
    exit 1
fi
