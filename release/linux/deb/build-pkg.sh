#!/usr/bin/env bash

#------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
#------------------------------------------------------------------------------

# Description:
#
# Build a debian/ubuntu `sqlcmd` package. This script is intended to be ran in a
# container with the respective disto/image laid down.
#
# Usage:
# $ build-pkg.sh

set -exv

: "${CLI_VERSION:?CLI_VERSION environment variable not set.}"
: "${CLI_VERSION_REVISION:?CLI_VERSION_REVISION environment variable not set.}"

WORKDIR=`cd $(dirname $0); cd ../../../; pwd`

ls -la ${WORKDIR}

apt-get -y update || exit 1
export DEBIAN_FRONTEND=noninteractive
apt-get install -y \
  debhelper \
  dpkg-dev \
  locales || exit 1

# Locale
sed -i -e 's/# en_US.UTF-8 UTF-8/en_US.UTF-8 UTF-8/' /etc/locale.gen && \
    dpkg-reconfigure --frontend=noninteractive locales && \
    update-locale LANG=en_US.UTF-8

export LANG=en_US.UTF-8
export PATH=$PATH

# Verify
chmod u+x /mnt/workspace/sqlcmd
/mnt/workspace/sqlcmd --help

mkdir /opt/stage
cp /mnt/workspace/sqlcmd /opt/stage/sqlcmd

# Create create directory for debian build
mkdir -p ${WORKDIR}/debian
${WORKDIR}/release/linux/deb/prepare-rules.sh ${WORKDIR}/debian ${WORKDIR}

cd ${WORKDIR}
dpkg-buildpackage -us -uc

ls ${WORKDIR} -R

debPkg=${WORKDIR}/../sqlcmd_${CLI_VERSION}-${CLI_VERSION_REVISION:=1}_all.deb
cp ${debPkg} /mnt/output/
