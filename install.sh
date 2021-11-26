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
        module load compilers/go/1.16.5
    else
        echo "Could not get a go compiler, exiting..." >&2
        exit 1
    fi
}

echo "Checking go environment..." >&2
check_for_go

if [[ -n "${TRAVIS:-}" ]]; then
    install_part_path="$(mktemp -d)"
    install_path="$install_part_path/cluster-bin"
    mkdir -p "$install_path"
else
    install_path="${INSTALL_PATH:-/shared/ucl/apps/cluster-bin}"
fi

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

if [[ -n "${TRAVIS:-}" ]]; then
    echo "Making a tar file of binary artifacts... (Note: this is not currently pushed anywhere or deployed.)" >&2
    tar -C "$install_path/.." -cJf "cluster-bin-${TRAVIS_COMMIT:-NO_COMMIT_LABEL}.tar.xz" "cluster-bin"
fi

echo "Deleting temporary GOPATH..." >&2
rm -Irf "${GOPATH}"

echo "Done." >&2
