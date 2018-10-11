#!/bin/bash

set -eo pipefail

project_dir=$PWD
rm -f $project_dir/test/bench.output
touch $project_dir/test/bench.output


pushd test
  exit_status="0"
  for dir in */; do
    pushd "$dir" 2>/dev/null
      # Bail if the folder doesn't have a bench_test file
      # Assumes folder has a start file if it has a bench_test file
      if [[ ! -f bench_test.go ]]; then
        popd 2>/dev/null
        continue
      fi

      # This bug in the current version of compose causes problems in
      # Jenkins:
      # https://github.com/docker/compose/issues/5929. docker-compose will
      # malfunction if it's run in a directory that has a name starting with
      # '_' or '-'. Until we get the fix, set COMPOSE_PROJECT_NAME
      COMPOSE_PROJECT_NAME="$(basename $PWD | sed 's/^[_-]*\(.*\)/\1/')"
      # Setup to allow compose to run in an isolated namespace
      export COMPOSE_PROJECT_NAME="${COMPOSE_PROJECT_NAME}_$(openssl rand -hex 3)"

      # Start the needed prerequisites. Assumes that start pre-cleans the env
      ./start

      # Run the tests
      set +e
        ./test -b | tee -a $project_dir/test/bench.output
        last_status="$?"

        # Only save first failure exit code
        if [[ "$exit_status" -eq "0" && "$last_status" -ne "0" ]]; then
          echo "ERROR: Detected a failure in the runner for $dir!"
          exit_status="$last_status"
        fi
      set -e

      # Clean up
      if [[ -f ./stop ]]; then
        ./stop
      fi
    popd
  done

  # Format the benchmark output
  rm -f $project_dir/test/bench.xml
  docker run --rm \
    -v $project_dir/test/:/secretless/test/output/ \
    secretless-dev \
    bash -exc "
      go get -u github.com/jstemmer/go-junit-report
      cat ./test/output/bench.output | go-junit-report > ./test/output/bench.xml
    "
popd || true

exit $exit_status