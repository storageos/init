#!/bin/bash

set -e

function log() {
    local msg="$1"
    timestamp_utc=$(date -u --rfc-3339=ns | sed 's/ /T/; s/\(\....\).*\([+-]\)/\1\2/g')
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
log "Checking configfs"
if mount | grep -q "^configfs on /sys/kernel/config"; then
    log "configfs mounted on sys/kernel/config"
else
    log "configfs not mounted, checking if kmod is loaded"
    state_file=/sys/module/configfs/initstate
    if [ -f "$state_file" ] && grep -q live "$state_file"; then
        log "configfs mod is loaded"
    else
        log "configfs not loaded, executing: modprobe -b configfs"
        modprobe -b configfs
    fi

    if mount | grep -q configfs; then
        log "configfs mounted"
    else
        log "mounting configfs /sys/kernel/config"
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
        log "Module $mod is running"
    else 
        log "Module $mod is not running"
        log "--> executing \"modprobe -b $mod\""
        if ! modprobe -b $mod; then
            # core_user and uio are not mandatory
            if [ "$mod" != "target_core_user" ] && [ "$mod" != "uio" ]; then
                exit 1
            else 
                log "Couldn't enable $mod"
            fi
        fi
        # Enable module at boot
        mkdir -p /etc/modules-load.d
        [ ! -f /etc/modules-load.d/lio.conf ] && echo $mod >> /etc/modules-load.d/lio.conf # create file if doesn't exist
    fi
done

# Check if the modules loaded have its
# directories available on top of configfs

[ ! -d "$target_dir" ] && log "$target_dir doesn't exist" && module_error_log "target_core_mod" "$target_dir"
[ ! -d "$core_dir" ]   && log "$core_dir doesn't exist"   && module_error_log "target_core_file" "$core_dir"
[ ! -d "$loop_dir" ]   && log "$loop_dir doesn't exist. Creating dir manually..." && mkdir $loop_dir

log "LIO set up is ready!"
