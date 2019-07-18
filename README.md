# Init Container for StorageOS

Init container to prepare the environment for StorageOS.


## Build

```
$ make image IMAGE=storageos/init:test
```


## Run it on host

Build the init container with `make image` and run it on the host with
`make run`.


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
