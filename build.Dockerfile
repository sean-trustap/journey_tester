# Copyright 2026 Trustap. All rights reserved.
# Use of this source code is governed by an MIT
# licence that can be found in the LICENCE file.

FROM golang:1.26.3-bookworm

SHELL ["/bin/bash", "-o", "errexit", "-o", "pipefail", "-c"]

RUN \
    echo 'exec curl --fail --silent --show-error --location "$@"' \
        > /tmp/curl.sh

# We create the `/.docker` directory to store user configuration information
# when running as a non-`root` user that doesn't have a home directory, which is
# used by later versions of the Docker client.
RUN \
    bash /tmp/curl.sh 'https://get.docker.com' \
    | VERSION=29.1.1 sh \
    && mkdir /.docker \
    && chmod 0777 /.docker

RUN \
    apt-get update \
    && apt-get install \
        --assume-yes \
        python3 \
        python3-pip \
    && rm -rf /var/lib/apt/lists/*

# We use `--break-system-packages` to override PEP 668's externally-managed
# environment check, which Debian Bookworm enforces by default.
RUN \
    python3 -m pip install \
        --break-system-packages \
        --no-cache-dir \
        comment-style==0.1.1 \
        sqlfluff==3.3.1

RUN \
    bash /tmp/curl.sh 'https://just.systems/install.sh' \
    | bash \
        -s \
        -- \
        --tag 1.51.0 \
        --to /usr/local/bin

RUN \
    bash /tmp/curl.sh 'https://github.com/mikefarah/yq/releases/download/v4.44.6/yq_linux_amd64' \
        > /usr/local/bin/yq \
    && chmod +x /usr/local/bin/yq \
    && bash /tmp/curl.sh 'https://github.com/a8m/envsubst/releases/download/v1.1.0/envsubst-Linux-x86_64' \
        > /usr/local/bin/envsubst \
    && chmod +x /usr/local/bin/envsubst 

RUN \
    mkdir /tmp/migrate \
    && wget \
        --output-document /tmp/migrate/migrate.tar.gz \
        https://github.com/golang-migrate/migrate/releases/download/v4.15.1/migrate.linux-amd64.tar.gz \
    && cd /tmp/migrate \
    && tar \
        --extract \
        --file migrate.tar.gz \
    && install \
        -m 0755 \
        /tmp/migrate/migrate \
        /usr/local/bin \
    && rm -rf /tmp/migrate

COPY \
    --from=hairyhenderson/gomplate:v4.3.2 \
    /gomplate \
    /usr/local/bin/gomplate


RUN \
    go install github.com/air-verse/air@v1.61.7 \
    && go install github.com/go-swagger/go-swagger/cmd/swagger@v0.33.1 \
    && go install github.com/mgechev/revive@v1.3.7 \
    && go install golang.org/x/tools/cmd/goimports@v0.1.7 \
    && go install golang.org/x/vuln/cmd/govulncheck@v1.1.4 \
    && go install mvdan.cc/gofumpt@v0.10.0 \
    && curl \
        --fail \
        --silent \
        --show-error \
        --location \
        https://raw.githubusercontent.com/golangci/golangci-lint/v2.11.3/install.sh \
            | sh -s -- -b $(go env GOPATH)/bin v2.11.3 \
    && go clean -modcache
