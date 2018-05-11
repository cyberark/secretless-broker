#!/bin/bash -ex

project_dir=$PWD
touch $project_dir/test/junit.output

go get -u github.com/jstemmer/go-junit-report

cd test

exit_status="0"
for dir in */; do
  cd "$dir"
  if [[ -f start.sh ]]; then
    [[ ! -f ./stop.sh ]] || ./stop.sh || true
    ./start.sh
    set +e
    ./test.sh | tee -a $project_dir/test/junit.output
    last_status="$?"
    if [[ "$exit_status" -eq "0" && "$last_status" -ne "0" ]]; then
      exit_status="$last_status"
    fi
    set -e
    [[ ! -f ./stop.sh ]] || ./stop.sh
  fi
  cd ..
done

cat $project_dir/test/junit.output | go-junit-report > $project_dir/test/junit.xml

exit $exit_status
