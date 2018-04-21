#!/bin/bash
set -eou pipefail

GOPATH=$(go env GOPATH)
PKG=github.com/appscode/searchlight-plugin
REPO_ROOT="$GOPATH/src/$PKG"

DOCKER_REGISTRY=appscode
IMG=searchlight-plugin-go

build() {
    pushd $REPO_ROOT/searchlight-plugin-go/hack/docker
    echo "building alpine based binary ..."
    # GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o webhook $REPO_ROOT/cmd/webhook/*
    docker run                                                              \
        --rm                                                                \
        -u $(id -u):$(id -g)                                                \
        -v /tmp:/.cache                                                     \
        -v "$REPO_ROOT:/go/src/$PKG"                                        \
        -w "/go/src/$PKG/searchlight-plugin-go"                             \
        -e GOOS=linux                                                       \
        -e GOARCH=amd64                                                     \
        -e CGO_ENABLED=0                                                    \
        golang:1.10.0-alpine                                                \
        go build -a -installsuffix cgo -o hack/docker/webhook main.go
    local cmd="docker build -t $DOCKER_REGISTRY/$IMG ."
    echo $cmd; $cmd
    rm -rf webhook
    popd
}

push() {
    echo "Push docker image..."
    local cmd="docker push $DOCKER_REGISTRY/$IMG"
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
