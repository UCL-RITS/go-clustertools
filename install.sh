#!/usr/bin/env bash

set -o errexit  \
    -o nounset  \
    -o pipefail

function check_for_go () {
    if [ -n "${GOROOT:-}" ] && which go >/dev/null 2>/dev/null; then
        echo "Found go compiler." >&2
    else
        echo "Could not get a go compiler, exiting..." >&2
        exit 1
    fi
}    

echo "Checking go environment..." >&2
check_for_go

install_path="${INSTALL_PATH:-/shared/ucl/apps/cluster-bin}"

echo "Making temporary GOPATH..." >&2
export GOPATH
GOPATH="$(mktemp -t -d tmp-go-path.XXXXXXXX)"

echo "Fetching dependencies..." >&2
./fetchdeps.sh

echo "Building..." >&2
./build.sh

echo "Installing to: $install_path" >&2
cp -vf bin/* "$install_path"/

echo "Deleting temporary GOPATH..." >&2
rm -Irf "${GOPATH}"

echo "Done." >&2
