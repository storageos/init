# Init Container for StorageOS

StorageOS requires Linux-IO (LIO) target. This init container will ensure that the kernel modules 
and configuration to use LIO are available in the kernel's host where StorageOS containers run.

