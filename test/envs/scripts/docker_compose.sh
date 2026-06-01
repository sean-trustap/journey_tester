# Copyright 2026 Trustap. All rights reserved.
# Use of this source code is governed by an MIT
# licence that can be found in the LICENCE file.

# `$0 <env> <tgt-dir> [args...]` runs `docker compose [args...]` relative to the
# `env` environment, and adapts `docker_compose.yaml` based on whether this
# command is being run within a Docker container, and uses `tmp-dir` as a
# scratch space for temporary files.
#
# `docker_compose.yaml` makes use of bind-mounted volumes, which are based on
# paths present on the Docker host. As such, if `docker compose` is run within a
# container, then the relative paths defined in the `docker_compose.yaml` will
# need to be updated to absolute paths defined on the Docker host. This script
# makes use of the `docker_hostpath.sh` script to get the absolute path to the
# mock environment directory on the Docker host, generates a copy of
# `docker_compose.yaml` with the relative paths replaced by absolute host paths,
# and runs `docker compose` with the new `docker_compose.yaml`. If this script
# isn't being run in a container then the existing `docker_compose.yaml` is used
# directly.

set -o errexit
set -o pipefail
set -o nounset

main() {
    if [ $# -lt 2 ] ; then
        echo "usage: $0 <env> <tgt-dir> [args...]" >&2
        exit 1
    fi

    env="$1"
    tgt_dir="$2"
    shift 2

    tgt_compose_confs_dir="$tgt_dir/docker_compose"
    mkdir --parents "$tgt_compose_confs_dir"

    proj_dir="test/envs"
    gen_docker_compose_confs \
        "$proj_dir" \
        "$tgt_compose_confs_dir"

    org_name="trustap"
    proj_name="journey_tester"
    compose_proj_name="${org_name}_${proj_name}_${env}_test"
    docker_compose_with_env \
        "$compose_proj_name" \
        "$proj_dir" \
        "$env" \
        "$tgt_compose_confs_dir" \
        "$@"
}

gen_docker_compose_confs() {
    # `$0 <src-dir> <tgt-dir>` generates new `docker_compose.yaml` files to
    # `tgt-dir` where relative paths inside the specifications are replaced with
    # absolute paths for the host to allow `docker compose` to be invoked from
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
            dir="$dir/${conf_env}_env"
        fi

        sed \
            --expression "s|source: \./|source: $host_abs_src_dir/|" \
            --expression "s|source: \.\./|source: $host_abs_src_dir/../|" \
            "$dir/docker_compose.yaml" \
            > "$tgt_dir/$conf_env.yml"
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
        --project-name="$proj_name" \
        --project-directory="$proj_dir" \
        --file="$docker_compose_confs_dir/common.yml" \
        --file="$docker_compose_confs_dir/$env.yml" \
        "$@"
}

main "$@"
