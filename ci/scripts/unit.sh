#!/bin/bash -eux

cwd=$(pwd)

pushd $cwd/dp-frontend-geography-controller
  make test
popd
