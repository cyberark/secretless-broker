here=$PWD
cd ..;
reflex -r "\.go$" -s -- bash -c "go run ./cmd/secretless-broker/main.go -watch -f '${here}/secretless.yml' -debug"
