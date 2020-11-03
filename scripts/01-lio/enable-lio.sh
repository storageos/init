#!/bin/bash

set -e

function date_time_in_rfc3339() {
    local msg="$1"
    timestamp_utc=$(date -u --rfc-3339=seconds)
    echo $timestamp_utc $msg
}

function module_error_log() {
    local mod="$1"
    local mod_dir="$2"
    RE='\e[0;31m'
    NC='\e[0m' # No Color
    echo -e "${RE}ERROR: The kernel module $mod couldn't load properly. Please try to run${NC} modprobe $mod ${RE}. Once loaded, the directory $mod_dir should be accessible. Otherwise the module has not been loaded as expected.${NC}"
}

# Configfs can be built in the kernel, hence the module 
# initstate file will not exist. Even though, the mount
# is present and working
date_time_in_rfc3339 "Checking configfs"
if mount | grep -q "^configfs on /sys/kernel/config"; then
    date_time_in_rfc3339 "configfs mounted on sys/kernel/config"
else
    date_time_in_rfc3339 "configfs not mounted, checking if kmod is loaded"
    state_file=/sys/module/configfs/initstate
    if [ -f "$state_file" ] && grep -q live "$state_file"; then
        date_time_in_rfc3339 "configfs mod is loaded"
    else
        date_time_in_rfc3339 "configfs not loaded, executing: modprobe -b configfs"
        modprobe -b configfs
    fi

    if mount | grep -q configfs; then
        date_time_in_rfc3339 "configfs mounted"
    else
        date_time_in_rfc3339 "mounting configfs /sys/kernel/config"
        mount -t configfs configfs /sys/kernel/config
    fi
fi

target_dir=/sys/kernel/config/target
core_dir="$target_dir"/core
loop_dir="$target_dir"/loopback

# Enable a mod if not present
# /sys/module/$modname/initstate has got the word "live"
# in case the kernel module is loaded and running 
for mod in target_core_mod tcm_loop target_core_file uio target_core_user; do
    state_file=/sys/module/$mod/initstate
    if [ -f "$state_file" ] && grep -q live "$state_file"; then
        date_time_in_rfc3339 "Module $mod is running"
    else 
        date_time_in_rfc3339 "Module $mod is not running"
        date_time_in_rfc3339 "--> executing \"modprobe -b $mod\""
        if ! modprobe -b $mod; then
            # core_user and uio are not mandatory
            if [ "$mod" != "target_core_user" ] && [ "$mod" != "uio" ]; then
                exit 1
            else 
                date_time_in_rfc3339 "Couldn't enable $mod"
            fi
        fi
        # Enable module at boot
        mkdir -p /etc/modules-load.d
        [ ! -f /etc/modules-load.d/lio.conf ] && echo $mod >> /etc/modules-load.d/lio.conf # create file if doesn't exist
    fi
done

# Check if the modules loaded have its
# directories available on top of configfs

[ ! -d "$target_dir" ] && date_time_in_rfc3339 "$target_dir doesn't exist" && module_error_log "target_core_mod" "$target_dir"
[ ! -d "$core_dir" ]   && date_time_in_rfc3339 "$core_dir doesn't exist"   && module_error_log "target_core_file" "$core_dir"
[ ! -d "$loop_dir" ]   && date_time_in_rfc3339 "$loop_dir doesn't exist. Creating dir manually..." && mkdir $loop_dir

date_time_in_rfc3339 "LIO set up is ready!"
