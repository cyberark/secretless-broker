function run_secretless() {
    go run ../cmd/secretless-broker -f ./secretless.yml
}

function run_client() {
    go run ./cmd/client :5433
}

function run_server() {
    go run ./cmd/server 5432
}