#!/usr/bin/env sh
#
# MIT License
#
# (C) Copyright [2019-2021] Hewlett Packard Enterprise Development LP
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
#

set -ex

echo "Running $1"

if [ "$1" = 'sls-init' ]; then
  # This directory has to exist either way, but hopefully a persistent storage is mounted here.
  mkdir -p /persistent_migrations

  # Make sure the migrations make their way to the persistent mounted storage.
  cp /migrations/*.sql /persistent_migrations/

  echo "Migrations copied to persistent location."
elif [ "$1" = 'sls-loader' ]; then
  if [ "${USE_S3_DNS_HACK}" = 'true' ]; then
    # Build up a list of all of the nameservers for k8s and the PIT/LiveCD nameservers.
    # So if k8s DNS is not aware of rgw-vip, then we will fall back to the PIT/LiveCD nameserver
    # which should be able to resolve the address.
    echo "resolv.conf from NCN:"
    echo "${PIT_NAMESERVER}"

    # Get a list of all of the nameservers avaiable:
    k8s_nameservers=$(cat /etc/resolv.conf     | grep nameserver | awk '{print $2}')
    pit_nameservers=$(echo "${PIT_NAMESERVER}" | grep nameserver | awk '{print $2}')
    all_nameservers=$(printf "%s\n%s\n" "${k8s_nameservers}" "${pit_nameservers}" )

    S3_HOSTNAME=$(basename ${S3_ENDPOINT})
    S3_IP=""

    if [ -z "${S3_DNS_LOOKUP_ATTEMPTS}" ]; then
      S3_DNS_LOOKUP_ATTEMPTS=10
    fi

    # Query each nameserver until we can determine an IP address
    for attempt in $(seq ${S3_DNS_LOOKUP_ATTEMPTS}); do
      # Try each nameserver
      echo "Lookup attempt $attempt of ${S3_DNS_LOOKUP_ATTEMPTS}"
      for nameserver in $all_nameservers; do
        echo "Using $nameserver to lookup $S3_HOSTNAME"
        S3_IP=$(dig +short "@$nameserver" "$S3_HOSTNAME" )
        if [ -n "$S3_IP" ]; then
          echo "Lookup succeeded: $S3_IP"
          break 2 # break out of both loops
        fi
        echo "Lookup failed"
      done
      echo "Sleeping 1 seconds before next lookup attempt"
      sleep 1
    done

    if [ -z "$S3_IP" ]; then
      echo "All lookup attempts failed, exiting"
      exit 1
    fi

    # HACK: We are always assuming that we need to use HTTPS, and there is no port
    # on the given S3_ENDPOINT value. If on vshasta, the DNS hack should be disabled
    export S3_ENDPOINT="https://${S3_IP}"
    echo "New S3_ENDPOINT: ${S3_ENDPOINT}"
  fi

  # If the loader is called then we have to do some prep work. First, pull the SLS file out of S3.
  sls-s3-downloader
  echo 'Configs downloaded from S3.'
fi

exec "$@"
