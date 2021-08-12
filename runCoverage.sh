#!/usr/bin/env bash
# MIT License
#
# (C) Copyright [2021] Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.

test_result=0

# It's possible we don't have docker-compose, so if necessary bring our own.
docker_compose_exe=$(command -v docker-compose)
if ! [[ -x "$docker_compose_exe" ]]; then
    if ! [[ -x "./docker-compose" ]]; then
        echo "Getting docker-compose..."
        curl -L "https://github.com/docker/compose/releases/download/1.24.1/docker-compose-$(uname -s)-$(uname -m)" \
        -o ./docker-compose

        if [[ $? -ne 0 ]]; then
            echo "Failed to fetch docker-compose!"
            exit 1
        fi

        chmod +x docker-compose
    fi
    docker_compose_exe="./docker-compose"
fi

# Print executions
set -x

# Need to setup a working DB environment
PROJECT=$RANDOM
NETWORK_NAME="${PROJECT}_default"
INIT_CONTAINER_NAME="${PROJECT}_sls-init_1"

${docker_compose_exe} --project-name $PROJECT -f docker-compose.testing.yaml up -d --build
if [[ $? -ne 0 ]]; then
    echo "Failed to setup environment!"
    exit 1
fi

docker wait $INIT_CONTAINER_NAME

# Build the build base image (if it's not already)
docker build -t cray/sls-base --target base .

# Run the coverage.
DOCKER_BUILDKIT=0 docker build --network $NETWORK_NAME -t cray/sls-unit-testing -f Dockerfile.coverage --no-cache .
build_result=$?
if [ $build_result -ne 0 ]; then
  echo "Coverage tests failed!"
  test_result=$build_result
else
 echo "Coverage tests passed!"
fi

# Cleanup.
${docker_compose_exe} --project-name $PROJECT -f docker-compose.testing.yaml down

exit $test_result
