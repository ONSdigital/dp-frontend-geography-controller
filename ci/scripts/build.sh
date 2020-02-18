#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-frontend-geography-controller
  make build && cp build/dp-frontend-geography-controller $cwd/build
  cp Dockerfile.concourse $cwd/build
popd
