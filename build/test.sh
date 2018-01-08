#!/bin/bash -ex

cd test

for dir in */; do
  cd $dir
  if [[ -f start.sh ]]; then
    ./start.sh
    ./test.sh || true
    ./stop.sh
  fi
  cd ..
done
