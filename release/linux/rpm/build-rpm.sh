#!/usr/bin/env bash

#------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
#------------------------------------------------------------------------------

# Description:
#
# Build a rpm `sqlcmd` package. This script is intended to be run in a
# container with the respective distro/image laid down.
#
# Usage:
# $ build-rpm.sh

set -exv

: "${CLI_VERSION:?CLI_VERSION environment variable not set.}"
: "${CLI_VERSION_REVISION:?CLI_VERSION_REVISION environment variable not set.}"

yum update -y
yum install -y rpm-build

export LC_ALL=en_US.UTF-8
export REPO_ROOT_DIR=`cd $(dirname $0); cd ../../../; pwd`

rpmbuild -v -bb --clean ${REPO_ROOT_DIR}/release/linux/rpm/sqlcmd.spec && cp /root/rpmbuild/RPMS/x86_64/* /mnt/output
