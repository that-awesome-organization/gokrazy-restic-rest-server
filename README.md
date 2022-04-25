# Rest Server - gokrazy

This repository contains the package and [gokrazy](https://gokrazy.org) application for
[Restic's Rest Server](https://github.com/restic/rest-server)

# Disk Setup

Before running the `rest-server` we need to setup the disk so that it can be used directly.

## Directory

We need a directory inside the external disk which can be used. for example we can use a
directory named `restic/` which will have all the named directories by username and other
configurations.

```
cd /run/media/username/diskname
mkdir restic/
```

## Authentication

If rest-server is used in authentication mode the directory should have `.htpasswd` file with
bcrypt hashed password. This can be carried out by running following command

```
touch restic/.htpasswd
htpasswd -B restic/.htpasswd username
```

For all the users needed for authentication we can repeat the last command.

# gokrazy Setup

gokrazy does not support mounting a disk directly, so this project wraps the original arm64 binary
of restic's rest-server into `/usr/local/bin` with mount options.

Before we begin, we setup the directory where we can keep package files for gokrazy.

```
go mod init rpi4-gorilla                # this is an arbitrary name

go install github.com/dtchanpura/gokrazy-rest-server@latest
```

## Mount Specifications

This program uses three environment variables for this specification

* MNT_SOURCE: source device for mounting.
* MNT_TARGET: target directory where the device should be mounted.
* MNT_FSTYPE: filesystem type of the device, e.g. exfat, ext4, vfat, etc...

To use these variables there needs to be the following file in the directory where the gokr-packer
command is being run.

```
# in the same directory as go.mod

mkdir -p env/github.com/dtchanpura/gokrazy-rest-server/
cat << EOF > env/github.com/dtchanpura/gokrazy-rest-server/env.txt
MNT_SOURCE=/dev/sda1
MNT_TARGET=/perm/rest-server
MNT_FSTYPE=exfat
EOF
```

## Flags

This program transparently passes all arguments to rest-server command and also the stdout/stderr
from it, so the flags or arguments which we want to pass will be actually going in to rest-server

To add flags we can just create flags.txt same as the env.txt

```
mkdir -p flags/github.com/dtchanpura/gokrazy-rest-server/
cat < EOF > flags/github.com/dtchanpura/gokrazy-rest-server/flags.txt
--path
/perm/rest-server/restic
```

Here we are defining where the root of the rest-server should be available. Also in case if there's
need to add more arguments/flags just add it in this file separated by new lines.

## Other options

In some cases the command/program will run before the /perm/ directory is mounted and that will
make it fail and retry after a second. To avoid that, we can have a delay in startup using
waitforclock feature.

```
mkdir -p waitforclock/github.com/dtchanpura/gokrazy-rest-server/
touch waitforclock/github.com/dtchanpura/gokrazy-rest-server/waitforclock.txt
```

These commands will actually instruct gokrazy to start the rest-server only after the NTP has
synced the clock, which from my observations is also after mounting the /perm/ directory.
