# Enable LIO

StorageOS requires Linux-IO (LIO) target. This script will ensure that the kernel modules 
and configuration to use LIO are available in the kernel's host where StorageOS containers run.


If you experience any issue running the container, checkout the [os compatibility](https://docs.storageos.com/docs/reference/os_support) page.

# Kernel modules enabled

The kernel modules enabled are the following.
- configfs
- tcm_loop
- target_core_mod
- target_core_file
- uio
- target_core_user

> Even though the modules uio and target_core_user are optional, they are highly recommended.
