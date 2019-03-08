#!/bin/bash -e

get_first_pod_for_app() {
    kubectl get pods \
      --namespace "$2" \
      --selector=app="$1" \
      --output=jsonpath='{$.items[0].metadata.name}'
}

wait_for_app() {
  while kubectl get pods \
    --namespace "$2" \
    --selector=app="$1" \
    --output=jsonpath='{$.items[0].status.containerStatuses.*.ready}' \
      | grep -qv "false"
  do
    echo "Waiting for $1 to be ready"
    sleep 5
  done
  echo Ready!
}

# Usage: repeat_str 3 hi (returns hihihi)
function repeat_str() {
  local i
  for ((i=0; i<"$1"; i++)); do
    printf "%s" "$2"
  done
}

while ! docker-compose ps "${DB_TYPE}" | grep healthy > /dev/null 2>&1;
do
  >&2 printf '. '
  sleep 1
done

repeat() { local i n; n=$1; shift; for ((i=1; i<=n; i++)); do "$@"; done; }

# Usage:
# some_long_process &
# spinner "waiting for long process"
function spinner() {
    local info="$1"
    local pid=$!
    local delay=0.75
    local spin_chars="|/-\\"
    local spin_index=0
    while ps -p $pid > /dev/null; do
        local cur_char=${spin_chars:spin_index:1}
        spin_index=$(((spin_index+1) % ${#spin_chars}))
        printf " [%c]  $info" "$cur_char"

        sleep $delay
        local reset=$'\b\b\b\b\b\b'
        for ((i=1; i<=$(wc -c <<<"$info"); i++)); do
            reset+=$'\b'
        done
        printf "%s" "$reset"
    done
}

