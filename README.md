# Init Container for StorageOS

StorageOS requires Linux-IO (LIO) target. This init container will ensure that the kernel modules 
and configuration to use LIO are available in the kernel's host where StorageOS containers run.


If you experience any issue running the container, checkout the [os compatibility](https://docs.storageos.com/docs/reference/os_support) page.

## Docker

```
docker run --name enable_lio                  \
           --privileged                       \
           --rm                               \
           --cap-add=SYS_ADMIN                \
           -v /lib/modules:/lib/modules       \
           -v /sys:/sys:rshared               \
           storageos/init:0.1
```
