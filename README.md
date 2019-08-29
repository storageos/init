# Init Container for StorageOS

[![Build Status](https://travis-ci.org/storageos/init.svg?branch=master)](https://travis-ci.org/storageos/init)

Init container to prepare the environment for StorageOS.


## Options

* `-scripts` - absolute path of the scripts directory.
* `-nodeImage` - StorageOS Node container image that the init container runs along. This should be used when running out of k8s.
* `-dsName` - StorageOS k8s DaemonSet name. Use when running within a k8s cluster.
* `-dsNamespace` - StorageOS k8s DaemonSet namespace. Use when running within a k8s cluster.


## Environment Variables

* `NODE_IMAGE` - StorageOS Node container image.
* `DAEMONSET_NAME` - StorageOS DaemonSet name.
* `DAEMONSET_NAMESPACE` - StorageOS DaemonSet namespace.


## Build

```
$ make image IMAGE=storageos/init:test
```


## Run it on host

Build the init container with `make image` and run it on the host with
`make run`.

Pass a StorageOS Node image and scripts directory as:

```
$ make run SCRIPTS_PATH=scripts/ NODE_IMAGE=storageos/node:1.4.0
```


## Script Framework

The script framework executes a set of scripts, performing any checks and
running the necessary script based on the host environment. The script's stdout
and stderr are written to the stdout and stderr of the init app. Container logs
should show all the logs of the individual scripts that ran. The exit status of
the scripts are used to determine initialization failure or success. Any
non-zero exit status are also logged as an event in the k8s pod events.

The scripts should be placed in the `scripts/` dir. The scripts are sorted for
execution based on their name and their parent directory name in lexical order.
The scripts must start with shebang (`#!/bin/bash` for bash scripts) and must
have executable permission(`chmod +x`).

Example scripts dir:
```
scripts
├── 01-script.sh
├── 05-foo
│   └── scriptx.sh
│   └── README.md
├── 07-scriptz.sh
└── 10-baz
    └── scripty.sh
    └── README.md
```

In the above example, the script execution order will be
```
01-script.sh, scriptx.sh, 07-scriptz.sh, scripty.sh
```

For documenting each script, they can be placed in a subdirectory along with a
markdown(.md) or a text file(.txt). These docs files are ignored.
