#!/bin/bash

set -e

function log() {
    local msg="$1"
    timestamp_utc=$(date -u --rfc-3339=seconds | sed 's/ /T/')
    echo $timestamp_utc $msg
}

# For a directory containeing the cgroup slice information, return the value of
# pids.max, or 0 if set to "max". Return -1 exit code if the file doesn't exist.
function read_max_pids() {
    if [ ! -f ${1}/pids.max ]; then
        return -1
    fi
    local max_pids=$(<${1}/pids.max)
    if [ $max_pids == "max" ]; then
        echo 0
        return
    fi
    echo $max_pids
}

default_max_pids_limit=999999999
max_pids_limit=$default_max_pids_limit
dirprefix="/sys/fs/cgroup/pids"

for cg in $(grep :pids: /proc/self/cgroup); do
    # Parse out the slice field from the cgroup output.
    # <cgroup_id>:<subystem>:<slice>
    dirsuffix=$(echo "$cg" | awk -F\: '{print $3}')

    # The slice field can have a prefix that is not part of the directory path.
    # This must be stripped iteratively until we find the valid slice directory.
    while [ ! -d "${dirprefix}/${dirsuffix}" ]; do
        dirsuffix=${dirsuffix#*/}
    done
    dir="${dirprefix}/${dirsuffix}"

    # Start at the current cgroup and traverse up the directory hierarchy
    # reading max.pids in each.  The lowest value will be the effective max.pids
    # value.
    while [ -f "${dir}/pids.max" ]; do
        max_pids=$(read_max_pids "${dir}")
        if [[ $max_pids -gt 0 && $max_pids -lt $max_pids_limit ]]; then
            max_pids_limit=$max_pids
        fi
        dir="${dir}/.."
    done
done

# TBC: Don't fail if we can't determine limit.
if [ $max_pids_limit -eq $default_max_pids_limit ]; then
    log "WARNING: Unable to determine effective max.pids limit"
    exit 0
fi

# Fail if MINIMUM_MAX_PIDS_LIMIT is set and is greater than current limit.
if [[ -n "${MINIMUM_MAX_PIDS_LIMIT}" && $MINIMUM_MAX_PIDS_LIMIT -gt $max_pids_limit ]]; then
    log "ERROR: Effective max.pids limit ($max_pids_limit) less than MINIMUM_MAX_PIDS_LIMIT ($MINIMUM_MAX_PIDS_LIMIT)"
    exit 1
fi

if [ -n "${RECOMMENDED_MAX_PIDS_LIMIT}" ]; then
    if [ $RECOMMENDED_MAX_PIDS_LIMIT -gt $max_pids_limit ]; then
        log "WARNING: Effective max.pids limit ($max_pids_limit) less than RECOMMENDED_MAX_PIDS_LIMIT ($RECOMMENDED_MAX_PIDS_LIMIT)"
    else
        log "OK: Effective max.pids limit ($max_pids_limit) at least RECOMMENDED_MAX_PIDS_LIMIT ($RECOMMENDED_MAX_PIDS_LIMIT)"
    fi
    exit 0
fi

# No requirements set, just output current limit.
log "Effective max.pids limit: $max_pids_limit"
