# Copyright 2026 Trustap. All rights reserved.
# Use of this source code is governed by an MIT
# licence that can be found in the LICENCE file.

# `$0 <env> <tmp-dir> [--dev] [--docker-flag <flag>*]` runs a command in the
# `env` test environment, and uses `tmp-dir` as a scratch space for temporary
# files.
#
# The `dev` argument runs the commnd in the mock environment in interactive mode
# with a new TTY.

set -o errexit
set -o pipefail
set -o nounset

main() {
    if [ $# -lt 2 ] ; then
        echo "usage: $0 <env> <tgt-dir>" >&2
        exit 1
    fi

    env="$1"
    tgt_dir="$2"
    shift 2

    proj_name="trustap_journey_tester"

    docker_flags=''
    while true ; do
        case "$1" in
            --dev)
                docker_flags="
                    $docker_flags
                    --docker-flag --interactive
                    --docker-flag --tty
                "
                shift 1
                ;;
            --docker-flag)
                docker_flags="
                    $docker_flags
                    --docker-flag $2
                "
                shift 2
                ;;
            *)
                break
                ;;
        esac
    done

    tgt_compose_confs_dir="$tgt_dir/docker_compose"
    mkdir --parents "$tgt_compose_confs_dir"

    proj_dir="test/envs"
    gen_docker_compose_confs \
        "$proj_dir" \
        "$tgt_compose_confs_dir"

    docker_compose() {
        docker_compose_with_env \
            "$proj_name" \
            "$proj_dir" \
            "$env" \
            "$tgt_compose_confs_dir" \
            "$@"
    }

    echo 'Starting the test environment...'
    echo ''
    docker_compose up \
            --detach \
            --no-build \
            --force-recreate \
            --wait \
            --wait-timeout 180 \
            2>&1 \
        | sed 's/^/     [init] /g' \
        || {
            echo ''
            echo 'Starting the test environment failed:' >&2
            echo '' >&2
            docker_compose logs \
                    2>&1 \
                | sed 's/^/    [logs] /g'
            echo ''
            echo 'Stopping the test environment...' >&2
            echo '' >&2
            docker_compose down \
                    2>&1 \
                | sed 's/^/    [down] /g'
            exit 1
        }

    output=''
    exit_code=0

    teardown() {
        exit_code="$?"

        echo "$output"

        echo "Stopping the test environment..."
        docker_compose down \
                --remove-orphans \
                --volumes \
                2>&1 \
            | sed 's/^/    [clean] /g'

        exit "$exit_code"
    }
    trap 'teardown' EXIT

    docker_compose logs \
        --follow \
        > "$tgt_dir/docker_compose_log.txt" \
        &

    echo "Running the command in the build environment..."
    time \
        bash scripts/with_build_env.sh \
            --docker-flag "--network=${proj_name}_default" \
            $docker_flags \
            "$@"
}

gen_docker_compose_confs() {
    # `$0 <src-dir> <tgt-dir>` generates new `docker_compose.yaml` files in
    # `tgt-dir` where relative paths inside the container are replaced with
    # absolute paths on the host to allow `docker compose` to be invoked from
    # within a Docker container, where these paths are then based on absolute
    # paths on the Docker server's host.

    if [ $# -ne 2 ] ; then
        echo "usage: $0 <src-dir> <tgt-dir>" >&2
        exit 1
    fi

    local src_dir="$1"
    local tgt_dir="$2"

    local host_abs_src_dir="$(bash scripts/docker_hostpath.sh $(pwd)/$src_dir)"
    for conf_env in 'common' 'live' 'mock' ; do
        local dir="$src_dir"
        if [ "$conf_env" != 'common' ] ; then
            dir="$dir/env_$conf_env"
        fi

        sed \
            --expression "s|source: \./|source: $host_abs_src_dir/|" \
            --expression "s|source: \.\./|source: $host_abs_src_dir/../|" \
            "$dir/docker_compose.yaml" \
            > "$tgt_dir/$conf_env.yaml"
    done
}

docker_compose_with_env() {
    if [ $# -lt 4 ] ; then
        echo "usage: $0 <proj-name> <proj-dir> <env> <docker-compose-confs-dir> [args...]" >&2
        exit 1
    fi

    local proj_name="$1"
    local proj_dir="$2"
    local env="$3"
    local docker_compose_confs_dir="$4"
    shift 4

    docker compose \
        --ansi=always \
        --project-name="$proj_name" \
        --project-directory="$proj_dir" \
        --file="$docker_compose_confs_dir/common.yaml" \
        --file="$docker_compose_confs_dir/$env.yaml" \
        "$@"
}

main "$@"
