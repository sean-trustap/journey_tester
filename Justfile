# Copyright 2026 Trustap. All rights reserved.
# Use of this source code is governed by an MIT
# licence that can be found in the LICENCE file.

tgt_name := 'journey_tester'
build_conf_dir := "configs/build"

tgt_dir := 'target'
tgt_artfs_dir := tgt_dir / 'artfs'
tgt_deps_dir := tgt_dir / 'deps'
tgt_gen_dir := tgt_dir / 'gen'
tgt_tmp_dir := tgt_dir / 'tmp'
tgt_test_dir := tgt_tmp_dir / 'test'
tgt_jrn_test_dir := tgt_test_dir / 'journey'
tgt := tgt_artfs_dir / tgt_name

# We list the sub packages to be tested explicitly so we can skip the source
# files that may be in `tgt_dir`.
src_dirs := "cmd internal pkg tools test"

# List available recipes.
default:
    just --list

# The checks are ordered in terms of estimated runtime, from quickest to
# slowest, so that failures should be found as quickly as possible.
#
# Run all non-journey tests.
check: check_builds check_db_migrations check_style check_lint check_unit

# Verify that code builds.
check_builds:
    make '{{tgt_gen_dir}}/test_server'

# Run style checks.
check_style: check_go_style

# Run style checks for Go files.
check_go_style:
    @# `gofmt` returns 0 even if formatting issues were found. However, it only
    @# produces output if issues were found so we fail the check if any output
    @# was produced.
    @#
    @# We check the formatting with `gofmt` after `gofumpt` and `goimports` to
    @# avoid drift between the different formatters; `gofmt` is the canonical
    @# representation.
    ! (gofumpt -d {{src_dirs}} | grep '')
    ! (goimports -d {{src_dirs}} | grep '')
    ! (gofmt -s -d -r 'interface{} -> any' {{src_dirs}} | grep '')
    comment_style '{{build_conf_dir}}/comment_style.yaml'

# Check for semantic issues in source code.
check_lint: check_go_lint check_sql_lint

# Check for semantic issues in Go files.
check_go_lint:
    @# `revive` is supported by `golangci-lint`, but we run `revive` directly
    @# here because it's unclear how to disable specific `revive` rules via the
    @# `golangci-lint` configuration.
    @#
    @# TODO There is an overlap between some linters that `golangci-lint`
    @# provides and `revive`. `revive` should be used to replace these where
    @# possible, for consistency, and because `revive` generally runs faster
    @# than `golangci-lint`.
    revive \
        -config='{{build_conf_dir}}/revive.toml' \
        -formatter=plain \
        cmd/... \
        pkg/... \
        tools/...
    golangci-lint run \
        --config='{{build_conf_dir}}/golangci.yaml' \
        cmd/... \
        pkg/... \
        tools/...

# Check for semantic issues in SQL files.
check_sql_lint:
    sqlfluff lint \
        --dialect=mysql \
        --config='{{build_conf_dir}}/sqlfluff.cfg' \
        assets/db_migrations/*
    sqlfluff lint \
        --dialect=mysql \
        --config='{{build_conf_dir}}/sqlfluff.cfg' \
        test/envs/configs/dbs.sql

# Run unit tests.
check_unit run_pattern=".*":
    go test \
        -run '{{run_pattern}}' \
        ./cmd/... \
        ./pkg/... \
        ./tools/...

check_db_migrations:
    #!/usr/bin/env sh

    duplicates=$(
        ls assets/db_migrations/ \
            | cut \
                --delimiter='_' \
                --fields=1 \
        | uniq --repeated
    )
    if [ -n "$duplicates" ]; then
        echo "Found duplicate prefixes for db migrations: $(echo $duplicates)." >&2
        exit 1
    fi

# Run journey tests.
check_jrns env="mock" run_pattern=".*":
    #!/usr/bin/env sh
    set -o errexit

    rm -rf '{{tgt_jrn_test_dir}}'

    if [ '{{env}}' = 'mock' -o '{{env}}' = 'live' ] ; then
        args='-groups=live_test_env,live_test_env_only'
        if [ '{{env}}' = 'mock' ] ; then
            args='-skip-groups=live_test_env_only'
        fi
        bash test/envs/scripts/run_journey_tests.sh \
                '{{env}}' \
                '{{tgt_jrn_test_dir}}' \
                -skip-groups=live_test_env_only \
                -pre-delete \
                -create \
                -run='{{run_pattern}}' \
                -delete \
            || if [ "$?" -eq 2 ] ; then
                grep  \
                    --invert-match \
                    db.trustap.local_1 \
                    '{{tgt_jrn_test_dir}}/env/docker_compose_log.txt'
            fi
    elif [ '{{env}}' = 'local' ] ; then
        go run cmd/journey_tester/main.go \
            -config='configs/journey_tester.yaml' \
            -output=minimal \
            -pre-delete \
            -create \
            -run='{{run_pattern}}' \
            -delete
    else
        echo "'{{env}}' isn't a recognised environment" >&2
        exit 1
    fi

# Format files.
fmt: fmt_go

# Format Go files.
fmt_go:
    gofumpt -w .

# TODO Add recipes to interact with an active test environment.
