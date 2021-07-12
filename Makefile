NAME ?= hms-sls
VERSION ?= $(shell cat .version)

all: image unittest coverage snyk

image:
	docker build --pull ${DOCKER_ARGS} --tag '${NAME}:${VERSION}' .

unittest:
	./runUnitTest.sh

coverage:
	./runCoverage.sh

snyk:
	./runSnyk.sh

