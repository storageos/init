#!/bin/bash -e

# This script runs storageos as an independent docker container.

NODE_IMAGE="${NODE_IMAGE:-storageos/node:1.3.0}"

docker run -d --rm --name storageos \
    -e HOSTNAME \
    -e ADVERTISE_IP="127.0.1.1" \
    -e JOIN="127.0.1.1" \
    -e DESCRIPTION="test-stos" \
    -e CSI_ENDPOINT=unix://var/lib/storageos/csi.sock \
    -e CSI_VERSION=v1 \
    --net=host \
    --pid=host \
    --privileged \
    --cap-add SYS_ADMIN \
    --device /dev/fuse \
    -v /var/lib/storageos:/var/lib/storageos:rshared \
    -v /run/docker/plugins:/run/docker/plugins \
    -v /sys:/sys \
    $NODE_IMAGE server
