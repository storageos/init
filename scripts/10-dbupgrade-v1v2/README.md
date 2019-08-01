# StorageoS DB Upgrade tool

storageos/node:1.4.0 changes the way in which StorageOS stores volume metadata in its internal database.
We call this new format, v2. Before running a 1.4.0 StorageOS node container it is necessary to upgrade
any existing v1 databases to the new v2 format. `storageos_dbupgrade_v1v2` does that.

Prior to running `storageos_dbupgrade_v1v2` the `NODE_IMAGE` environment variable must be set with the
intended StorageOS node container version, for example "storageos/node:1.4.0". `storageos_dbupgrade_v1v2`
uses this environment variable to work out whether or not to upgrade the database to the v2 format.
Obviously we don't want to upgrade the DB if we intend on running a version of StorageOS that's earlier
than 1.4.0.

It's safe to run `storageos_dbupgrade_v1v2` more than once as subsequent attempts to upgrade an already
upgraded DB will simply do nothing. If the `storageos_dbupgrade_v1v2` is killed mid-way through upgrading
a subsequent run will complete the upgrade. After a successful upgrade a backup of the original v1 database
will be kept at `/var/lib/storageos/data/db-old`.
