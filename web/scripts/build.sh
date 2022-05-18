#!/bin/sh

REACT_APP_GIT_SUMMARY=$(git describe --tags --dirty --always)
REACT_APP_GIT_COMMIT=$(git rev-parse --short HEAD)
REACT_APP_GIT_BRANCH=$(git symbolic-ref -q --short HEAD)
REACT_APP_BUILD_DATE=$(date)

export REACT_APP_GIT_SUMMARY
export REACT_APP_GIT_COMMIT
export REACT_APP_GIT_BRANCH
export REACT_APP_BUILD_DATE

relay-compiler && react-scripts "$1"
