# Copyright 2019-2020 Hewlett Packard Enterprise Development LP

# Dockerfile for building HMS SLS.

# Build base just has the packages installed we need.
FROM dtr.dev.cray.com/baseos/golang:1.14-alpine3.12 AS build-base

RUN set -ex \
    && apk update \
    && apk add build-base

FROM build-base AS base

# Copy all the necessary files to the image.
COPY cmd $GOPATH/src/stash.us.cray.com/HMS/hms-sls/cmd
COPY internal $GOPATH/src/stash.us.cray.com/HMS/hms-sls/internal
COPY vendor $GOPATH/src/stash.us.cray.com/HMS/hms-sls/vendor
COPY pkg $GOPATH/src/stash.us.cray.com/HMS/hms-sls/pkg

### Build Stage ###

FROM base AS builder

# Now build
RUN set -ex \
    && go build -v -i -o sls stash.us.cray.com/HMS/hms-sls/cmd/sls \
    && go build -v -i -o sls-init stash.us.cray.com/HMS/hms-sls/cmd/sls-init \
    && go build -v -i -o sls-loader stash.us.cray.com/HMS/hms-sls/cmd/sls-loader \
    && go build -v -i -o sls-s3-downloader stash.us.cray.com/HMS/hms-sls/cmd/sls-s3-downloader

### Final Stage ###

FROM dtr.dev.cray.com/baseos/alpine:3.12
LABEL maintainer="Cray, Inc."
STOPSIGNAL SIGTERM
EXPOSE 8376

# Setup environment variables.
ENV VAULT_ENABLED="true"
ENV VAULT_ADDR="http://cray-vault.vault:8200"
ENV VAULT_SKIP_VERIFY="true"
ENV VAULT_KEYPATH="secret/hms-creds"

# Default to latest schema version, this is overridden in the versioned chart.
ENV SCHEMA_VERSION latest

RUN set -ex \
    && apk update \
    && apk add --no-cache curl

# Copy files necessary for running/setup
COPY migrations /migrations
COPY configs /configs
COPY entrypoint.sh /

# Get sls from the builder stage.
COPY --from=builder /go/sls /usr/local/bin
COPY --from=builder /go/sls-init /usr/local/bin
COPY --from=builder /go/sls-loader /usr/local/bin
COPY --from=builder /go/sls-s3-downloader /usr/local/bin

ENTRYPOINT ["/entrypoint.sh"]
CMD sls