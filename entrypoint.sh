#!/usr/bin/env sh
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
    # Modify resolve.conf to append the PIT/LiveCD DNS server right after the kube DNS nameserver
    # So if k8s DNS is not aware of rgw-vip, then we will fall back to the PIT/LiveCD nameserver
    # which should be able to resolve the address.
    echo "${PIT_NAMESERVER}" >> /etc/resolv.conf

    echo "Modifed resolv.conf:"
    cat /etc/resolv.conf

    # Now use getent to get the IP address of the rgw-vip
    # HACK: We are always assuming that we need to use HTTPS, and there is no port
    # on the given S3_ENDPOINT value. If on vshasta, the DNS hack should be disabled
    export S3_ENDPOINT="https://$(getent hosts $(basename $S3_ENDPOINT) | awk '{print $1}')"
    echo "New S3_ENDPOINT: $S3_ENDPOINT"
  fi

  # If the loader is called then we have to do some prep work. First, pull the SLS file out of S3.
  sls-s3-downloader
  echo 'Configs downloaded from S3.'
fi

exec "$@"
