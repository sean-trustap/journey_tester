# Copyright 2026 Trustap. All rights reserved.
# Use of this source code is governed by an MIT
# licence that can be found in the LICENCE file.

# `$0 <proj> <proj-type> [--dev] [--docker-flag <flag>] [--uid <uid>]
# [--gid <gid>]` runs a command in the build environment for project `proj` of
# type `proj-type`.
#
# The `dev` argument runs the build environment in interactive mode with a new
# TTY and using the host network.
#
# The `--docker-flag <flag>` argument passes flag to the `docker run` command
# that starts the build environment.
#
# The `uid` and `gid` arguments run the command using the given user ID and
# group ID.

set -o errexit
set -o pipefail
set -o nounset

proj_name="journey_tester"

# We run the build as the local user (`id --user` and `id --group`) by default
# so that files will be created on the local filesystem with the correct
# permissions.
uid="$(id --user)"
gid="$(id --group)"
docker_flags=''
while true ; do
    case "${1:-}" in
        --dev)
            docker_flags="
                $docker_flags
                --interactive
                --tty
                --network=host
            "
            if [ ! -z "$SSH_AUTH_SOCK:-" ] ; then
                docker_flags="
                    $docker_flags
                    --volume=$SSH_AUTH_SOCK:/ssh_agent
                    --env=SSH_AUTH_SOCK=/ssh_agent
                "
            fi
            shift 1
            ;;
        --docker-flag)
            docker_flags="
                $docker_flags
                $2
            "
            shift 2
            ;;
        --uid)
            uid="$2"
            shift 2
            ;;
        --gid)
            gid="$2"
            shift 2
            ;;
        *)
            break
            ;;
    esac
done

org_name='trustap'
img_name="$org_name/$proj_name.build"

bash "scripts/docker_rbuild.sh" \
    "$img_name" \
    latest \
    --ssh=default \
    --build-arg=USER_ID="$uid" \
    --build-arg=GROUP_ID="$gid" \
    - \
    < build.Dockerfile

# See `scripts/docker_hostpath.sh` for details on `DOCKER_MOUNT_SRC`
# and `DOCKER_MOUNT_TGT`.
docker_mount_src="$(bash "scripts/docker_hostpath.sh" $(pwd))"

# We use named volumes to keep the Go caches between builds. We make them
# writable by anyone because we run the build as the local user (`id --user`
# and `id --group`), and the volumes are owned by root by default.
#
# We recursively set the permissions because `/go/pkg/mod` is initially
# created and populated by `root` when the image is built (e.g. by using `go
# get` to install some initial tools).
#
# NOTE The order of the `--volume` arguments matters, because Docker
# resolves them in order.
docker run \
    --rm \
    --user=root \
    --volume="${org_name}.${proj_name}.tmp_cache":/tmp/cache \
    --volume="${org_name}.${proj_name}.mod_cache":/go/pkg \
    --volume="${org_name}.go_download_cache":/go/pkg/mod/cache/download \
    "$img_name:latest" \
    chmod \
        --recursive \
        0777 \
        /tmp/cache \
        /go/pkg \
        /go/pkg/mod/cache/download

docker_mount_tgt="/go/src/github.com/$org_name/$proj_name"

# The group ID for the `docker` group on the host can be different from
# the `docker` group created inside the image when installing `docker`.
# If they're different, then a non-root user in the container can't access
# the host's `/var/run/docker.sock` socket because they're not in the
# correct group. For this reason, we get the exact group ID of the group
# that owns the socket and explicitly add the user to this group ID so that
# they can access the socket. We use the socket's owning group directly
# rather than looking up the `docker` group by name for stability.
host_docker_group_id=$(
    stat \
        --format=%g \
        /var/run/docker.sock
)

# We use `--init` so that the signals sent by commands like `docker stop` are
# handled as expected, which can be useful when cancelling worflows running in
# build pipelines and allow for faster termination of such processes.
#
# `XDG_CACHE_HOME` is used by Go to determine where the build cache should
# be.
docker run \
    $docker_flags \
    --rm \
    --init \
    --user="$uid":"$gid" \
    --env=XDG_CACHE_HOME=/tmp/cache \
    --volume=${org_name}.${proj_name}.tmp_cache:/tmp/cache \
    --volume=${org_name}.${proj_name}.mod_cache:/go/pkg \
    --volume=${org_name}.go_download_cache:/go/pkg/mod/cache/download \
    --volume="$docker_mount_src":"$docker_mount_tgt" \
    --workdir="$docker_mount_tgt" \
    --group-add="$host_docker_group_id" \
    --volume=/var/run/docker.sock:/var/run/docker.sock \
    --env=DOCKER_MOUNT_SRC="$docker_mount_src" \
    --env=DOCKER_MOUNT_TGT="$docker_mount_tgt" \
    "$img_name:latest" \
    "$@"
