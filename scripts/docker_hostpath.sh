# Copyright 2026 Trustap. All rights reserved.
# Use of this source code is governed by an MIT
# licence that can be found in the LICENCE file.

# Copyright 2024-2026 Trustap. All rights reserved.
# Use of this source code is governed by an MIT
# licence that can be found in the LICENCE file.

# `$0 <path>` outputs the absolute path corresponding to `path` as it should
# exist on the Docker host's filesystem.
#
# `path` must be an absolute path.
#
# This script requires that `DOCKER_MOUNT_SRC` and `DOCKER_MOUNT_TGT` are set to
# the source and target of a bind-mounted volume in the container that this
# script is run from.
#
# This script is useful for bind-mounting directories from Docker containers
# into new containers.
#
# Example usage:
#
# 1. A build environment is run using the following:
#
#     docker run \
#         --name=build \
#         --group-add=docker \
#         --volume=$(pwd):/host \
#         --volume=/var/run/docker.sock:/var/run/docker.sock \
#         --env=DOCKER_MOUNT_SRC=$(pwd) \
#         --env=DOCKER_MOUNT_TGT=/host \
#         golang:1.14.3 \
#         bash
#
# 2. A user wants to perform a `docker run` on the `/host/src` directory in the
#    `build` container. To do this in a container-independent way, `bash
#    scripts/docker_hostpath.sh /host/src` can be used to discover where the
#    directory exists on the host (`$(pwd)/src`, in this case), and then
#    bind-mount that directory in the "sub-container":
#
#     docker run \
#         --name=build \
#         --group-add=docker \
#         --volume=$(pwd):/host \
#         golang:1.14.3 \
#         bash -c '\
#             docker run \
#                 --volume=$(bash scripts/docker_hostpath.sh /host/src):/src \
#                 golang:1.14.3 \
#                 bash
#         '
#
# The above will mount `/host/src` from the "outer container" to the "inner
# container" as expected.
#
# This approach can be made recursive by using `bash scripts/docker_hostpath.sh`
# for the `DOCKER_MOUNT_SRC` in successive containers. See the definition of
# `DOCKER_MOUNT_SRC` in `scripts/with_build_env.sh` for an example.
#
# Note that if `$DOCKER_MOUNT_SRC` or `$DOCKER_MOUNT_TGT` isn't defined
# (presumably because the script isn't being run in a container) then `path` is
# returned unchanged.

set -o errexit

if [ $# -ne 1 ] ; then
    echo "usage: $0 <path>" >&2
    exit
fi

path="$1"

if [ -z "$DOCKER_MOUNT_SRC" -o -z "$DOCKER_MOUNT_TGT" ] ; then
    echo "$path"
else
    rel_path=$(echo "$path" | sed "s@^$DOCKER_MOUNT_TGT@@")
    if [ "$path" = "$rel_path" ] ; then
        echo "couldn't get host path: '$path' isn't in '$DOCKER_MOUNT_TGT'" >&2
        exit 1
    fi
    echo "$DOCKER_MOUNT_SRC$rel_path"
fi
