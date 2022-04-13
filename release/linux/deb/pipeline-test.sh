#!/usr/bin/env bash

#------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
#------------------------------------------------------------------------------

# Description:
#
# Instructions to be invoked under the build CI pipeline in AzureDevOps.
#
# Kickoff Debian package tests against versions:
#
# Usage:
# -----------------------------------
# buster  - Debian 10
# stretch - Debian 9
# jessie  - Debian 8
# -----------------------------------
# focal  - Ubuntu 20.04
# bionic - Ubuntu 18.04
# xenial - Ubuntu 16.04
# -----------------------------------
#
# Usage:
# $ pipeline-test.sh

set -e #xv

: "${REPO_ROOT_DIR:=`cd $(dirname $0); cd ../../../; pwd`}"

CLI_VERSION=${CLI_VERSION:=0.0.1}
CLI_VERSION_REVISION=${CLI_VERSION_REVISION:=1}
BUILD_ARTIFACTSTAGINGDIRECTORY=${BUILD_ARTIFACTSTAGINGDIRECTORY:=${REPO_ROOT_DIR}/output/debian}

DISTROS=( buster buster ) # TODO: Should we validate against more distros?: buster stretch jessie bionic xenial focal
BASE_IMAGES=( debian:buster ubuntu:focal ) # TODO: debian:buster debian:stretch debian:jessie ubuntu:bionic ubuntu:xenial ubuntu:focal

echo "=========================================================="
echo "CLI_VERSION: ${CLI_VERSION}"
echo "CLI_VERSION_REVISION: ${CLI_VERSION_REVISION}"
echo "BUILD_ARTIFACTSTAGINGDIRECTORY: ${BUILD_ARTIFACTSTAGINGDIRECTORY}"
echo "Distribution: ${DISTROS}"
echo "=========================================================="

for i in ${!DISTROS[@]}; do
    echo "=========================================================="
    echo "Test debian package on ${DISTROS[$i]}"
    echo "=========================================================="

    debPkg=sqlcmd_${CLI_VERSION}-${CLI_VERSION_REVISION}~${DISTROS[$i]}_all.deb

    script="apt-get update && \
            dpkg -i /mnt/artifacts/${debPkg} && \
            apt-get -f install && \
            sqlcmd --help"

    docker pull ${BASE_IMAGES[$i]}
    docker run --rm -v ${BUILD_ARTIFACTSTAGINGDIRECTORY}:/mnt/artifacts \
               ${BASE_IMAGES[$i]} \
               /bin/bash -c "${script}"

    echo ""
done
