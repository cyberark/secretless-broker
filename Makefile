.PHONY: docker_build_harness build test

./tmp/id_rsa:
	ssh-keygen -f ./tmp/id_rsa

docker_build_harness:
	docker-compose build secretless_dev 

build: ./tmp/id_rsa docker_build_harness
	docker-compose run --rm secretless_dev ./build/build.sh
	docker-compose build

test:
	./build/test.sh

all: build test
