#!/bin/bash
#
# Copyright (C) 2023 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#

error() {
    printf '%s\n' "$1" >&2
    exit 1
}

cid_count="${cid_count:=0}"

echo "VOLUMES is: $VOLUMES"
echo "DOCKER image is: $DOCKER_IMAGE"
echo "RUN_PATH: $RUN_PATH"
echo "CONTAINER_NAME: $CONTAINER_NAME"
echo "DOCKER_CMD: $DOCKER_CMD"
echo "DOT_ENV_FILE: $DOT_ENV_FILE"

# Set RENDER_MODE=1 for demo purposes only
# RUN_MODE="-itd"
# if [ "$RENDER_MODE" == 1 ]
# then
# 	RUN_MODE="-it -e DISPLAY=$DISPLAY -v /tmp/.X11-unix:/tmp/.X11-unix"
# fi

containerNameInstance="$CONTAINER_NAME$cid_count"

# interpret any nested environment variables inside the VOLUMES if any
volFullExpand=$(eval echo "$VOLUMES")
echo "DEBUG: volFullExpand $volFullExpand"
DOCKER_CMD="${DOCKER_CMD:="/bin/bash"}"

echo
echo
echo  "DEBUG===================================== "
cat $DOT_ENV_FILE
echo
echo

# volFullExpand is docker volume command and meant to be words splitting
# shellcheck disable=2086
docker run --network host --user root --ipc=host \
--name "$containerNameInstance" \
--env-file "$DOT_ENV_FILE" \
-e CONTAINER_NAME="$containerNameInstance" \
$TARGET_USB_DEVICE \
$TARGET_GPU_DEVICE \
$volFullExpand \
"$DOCKER_IMAGE" \
bash -c '$DOCKER_CMD'
