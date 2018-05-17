#!/bin/bash -ex

project_dir=$PWD
rm -f $project_dir/test/junit.output
touch $project_dir/test/junit.output

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

rm -f $project_dir/test/junit.xml
docker run --rm \
  -v $project_dir/test/:/go/src/github.com/conjurinc/secretless/test/output/ \
  secretless-dev bash -c "
    go get -u github.com/jstemmer/go-junit-report
    cat ./test/output/junit.output | go-junit-report > ./test/output/junit.xml
  "

exit $exit_status
