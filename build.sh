#!/usr/bin/env bash

set -ex

pkgs=$(go list)

for goos in darwin freebsd linux openbsd windows
do
  for goarch in 386 amd64
  do
    mkdir -p build/$goos/$goarch
    (cd build/$goos/$goarch && GOOS=$goos GOARCH=$goarch go build $pkgs) &
  done
done

wait
