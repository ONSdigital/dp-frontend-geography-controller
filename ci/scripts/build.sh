#!/bin/bash -eux

cwd=$(pwd)

export GOPATH=$cwd/go

pushd $GOPATH/src/github.com/ONSdigital/dp-frontend-geography-controller
  make build && cp build/dp-frontend-geography-controller $cwd/build
  cp Dockerfile.concourse $cwd/build
popd
