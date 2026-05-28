# Copyright 2026 Trustap. All rights reserved.
# Use of this source code is governed by an MIT
# licence that can be found in the LICENCE file.

# `$0 <env> [args...]` runs journey tests for this project in the `env` test
# environment.

set -o errexit
set -o pipefail
set -o nounset

main() {
    if [ $# -lt 2 ] ; then
        echo "usage: $0 <env> <tgt-dir>" >&2
        exit 1
    fi

    local env="$1"
    local tgt_dir="$2"
    shift 2

    local tgt_confs_dir="$tgt_dir/confs"
    mkdir --parents "$tgt_confs_dir"

    populate_conf_files \
        "$env" \
        "$tgt_confs_dir"

    mv \
        "$tgt_confs_dir/journey_tester.yaml" \
        "$tgt_dir"

    local tgt_env_dir="$tgt_dir/env"
    mkdir "$tgt_env_dir"

    bash test/envs/scripts/with_test_env.sh \
        "$env" \
        "$tgt_env_dir" \
        go run cmd/journey_tester/main.go \
            -config="$tgt_dir/journey_tester.yaml" \
            -output=minimal \
            "$@"
}

populate_conf_files() {
    if [ $# -ne 2 ] ; then
        echo "usage: $0 <env> <tgt-dir>" >&2
        exit 1
    fi

    local env="$1"
    local tgt_dir="$2"

    go run tools/merge_yaml/main.go \
        "test/envs/configs/journey_tester.yaml" \
        "test/envs/env_${env}/journey_tester.yaml" \
        > "$tgt_dir/journey_tester.in.yaml"

    gomplate \
        --file "$tgt_dir/journey_tester.in.yaml" \
        --context .='test/envs/vars.env?type=application/x-env' \
        --out "$tgt_dir/journey_tester.yaml"
}

main "$@"
