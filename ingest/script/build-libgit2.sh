#!/bin/sh

set -e

LIBGIT2_VER=${LIBGIT2_VER-"main"}
LIBGIT2_PATH=${HOME}/libgit2/libgit2-${LIBGIT2_VER}
GIT2GO_MAJOR_VER=${GIT2GO_MAJOR_VER-"v31"}
GIT2GO_VER=$(go list -f '{{.Version}}' -m github.com/libgit2/git2go/"${GIT2GO_MAJOR_VER}")

if echo "$LIBGIT2_VER" | grep -Eq '^([0-9]+)\.([0-9]+)\.([0-9]+)'; then
  DOWNLOAD_URL="https://codeload.github.com/libgit2/libgit2/tar.gz/v${LIBGIT2_VER}"
else
  DOWNLOAD_URL="https://codeload.github.com/libgit2/libgit2/tar.gz/${LIBGIT2_VER}"
fi

mkdir -p "${LIBGIT2_PATH}"
cd "${LIBGIT2_PATH}"/..
wget -O "${LIBGIT2_PATH}.tar.gz" "${DOWNLOAD_URL}"
tar -xzf "${LIBGIT2_PATH}.tar.gz"

if [ "$(uname)" = "Darwin" ]; then
  # macOS
  export SYSTEM_INSTALL_PREFIX=${SYSTEM_INSTALL_PREFIX-"/usr/local"}
fi

export VENDORED_PATH=${LIBGIT2_PATH}
export ROOT=${LIBGIT2_PATH}
wget -O - https://raw.githubusercontent.com/libgit2/git2go/"${GIT2GO_VER}"/script/build-libgit2.sh |
  sh -s -- "$@"

rm -rf "${LIBGIT2_PATH}"*
