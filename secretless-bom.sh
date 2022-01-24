#!/usr/bin/env bash

usage() {
  echo "Usage: $0 <tools directory> <project directory> <optional tool options>"
  echo "  <tools directory> - Location of release-tools"
  echo "  <project directory> - Absolute path to project directory containing Gemfile and Gemfile.lock"
  echo "  <optional tool options> - Any additional options to pass to the cyclonedx-ruby command"
  exit 1
}

if [[ $# -le 0 ]]; then
    usage
fi

while [[ $# -gt 0 ]]; do
    case "$1" in
    -h|--help)
        usage
        ;;
    -c|--compose)
        echo "compose"
        shift
        ;;
    -e|--entrypoint)
        export ENTRYPOINT=$1
        shift
        ;;
    -i|--image)
        export IMAGE=$1
        shift
        ;;
    -t|--tools)
        export TOOLS_DIR=$1
        shift
        ;;
    -m|--main)
        export MAIN=$1
        shift
        ;;
    -o|--output)
        export OUT=$1
        shift
        ;;
    esac
done

#TODO: Make script and mount that script into the docker run command.
#TODO: tools dir gets put in with the main go repo (ignoring .git of tools dir). 

if [[ -z ${ENTRYPOINT} && -z ${IMAGE} ]]; then
    #docker run --rm -v "$(pwd)":"/output" -v "$(pwd)/.git":"/secretless/.git" --entrypoint /secretless/bomgen.sh secretless-dev
    #docker run --rm -v "$(pwd)":"/output" --entrypoint "${ENTRYPOINT}" "${IMAGE}"
    #docker run --rm -v "$(pwd)":"/output" -v "$(pwd)":"bomgen.sh" --entrypoint "${ENTRYPOINT}" "${IMAGE}"
    docker run \
        --rm \
        --volume "$(pwd)":"$(pwd)" \
        --workdir "$(pwd)" \
        --entrypoint \
            "${TOOLS_DIR}/bom/go/gobomgen" \
            "${MAIN}" "${OUT}"
fi