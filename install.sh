#!/usr/bin/env bash

set -o errexit  \
    -o nounset  \
    -o pipefail

function check_for_go () {
    if [ -n "${GOROOT:-}" ] && which go >/dev/null 2>/dev/null; then
        echo "Found go compiler." >&2
    elif [[ -f "/etc/profile.d/modules.sh" ]]; then
        echo "No go compiler found, trying a module setup..." >&2
        source /etc/profile.d/modules.sh
        module purge
        module load gcc-libs
        module load compilers/go
    else
        echo "Could not get a go compiler, exiting..." >&2
        exit 1
    fi
}

echo "Checking go environment..." >&2
check_for_go

install_path="${INSTALL_PATH:-/shared/ucl/apps/cluster-bin}"

echo "Changing into \"$(dirname -- "$0")\"..." >&2
cd "$(dirname -- "$0")"

echo "Making temporary GOPATH..." >&2
export GOPATH
GOPATH="$(mktemp -t -d tmp-go-path.XXXXXXXX)"

echo "Linking current directory to correct place in GOPATH..." >&2
remote_url="$(git config --get remote.origin.url)"
if [[ "${remote_url:0:4}" == "git@" ]]; then
    remote_url="github.com/${remote_url#git@github.com:}"
    dir_for_remote="${remote_url%/*}"
else
    remote_url="${remote_url##https://}"
    dir_for_remote="${remote_url%/*}"
fi
mkdir -p "$GOPATH/src/$dir_for_remote"
ln -s "$(pwd)" "$GOPATH/src/$dir_for_remote/"

echo "Fetching dependencies..." >&2
./fetchdeps.sh

echo "Building..." >&2
./build.sh

echo "Installing to: $install_path" >&2
cp -vf bin/* "$install_path"/

echo "Deleting temporary GOPATH..." >&2
rm -Irf "${GOPATH}"

echo "Done." >&2
