#!/bin/bash -ex

cd test

exit_status="0"
for dir in */; do
  cd "$dir"
  if [[ -f start.sh ]]; then
    ./stop.sh || true
    ./start.sh
    set +e
    ./test.sh
    last_status="$?"
    if [[ "$exit_status" -eq "0" && "$last_status" -ne "0" ]]; then
      exit_status="$last_status"
    fi
    set -e
    ./stop.sh
  fi
  cd ..
done

exit $exit_status
