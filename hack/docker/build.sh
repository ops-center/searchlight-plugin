#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

GOPATH=$(go env GOPATH)
REPO_ROOT="$GOPATH/src/github.com/appscode/plugin-webhook"

IMG=webhook-plugin

build() {
    pushd $REPO_ROOT/hack/docker
    echo "building alpine based binary ..."
    env GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o webhook $REPO_ROOT/cmd/webhook/*
    local cmd="docker build -t aerokite/$IMG ."
    echo $cmd; $cmd
    rm -rf webhook
    popd
}

push() {
    echo "Push docker image..."
    local cmd="docker push aerokite/$IMG"
    echo $cmd; $cmd
}

RETVAL=0

if [ $# -eq 0 ]; then
    cmd=${DEFAULT_COMMAND:-build}
    $cmd
    exit $RETVAL
fi

case "$1" in
    build)
        build
        ;;
    push)
        push
        ;;
    *)
        echo $"Usage: $0 {build|push}"
        RETVAL=1
esac

exit $RETVAL
