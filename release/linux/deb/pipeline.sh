#!/usr/bin/env bash

#------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
#------------------------------------------------------------------------------

# Description:
#
# Instructions to be invoked under the build CI pipeline in AzureDevOps.
#
# Kickoff debian package build in docker and copy the .deb package artifact
# back to the local filesystem. The build pipeline can then save it as an
# artifact as it sees fit.
#
# Note: Intended to be ran under ubuntu.
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
# Example:
#
# export DISTRO=xenial
# export DISTRO_BASE_IMAGE=ubuntu:xenial
#
# $ pipeline.sh

set -exv

DISTRO=${DISTRO:=buster}
DISTRO_BASE_IMAGE=${DISTRO_BASE_IMAGE:=debian:buster}

: "${DISTRO:?DISTRO environment variable not set.}"
: "${DISTRO_BASE_IMAGE:?DISTRO_BASE_IMAGE environment variable not set.}"
: "${REPO_ROOT_DIR:=`cd $(dirname $0); cd ../../../; pwd`}"
DIST_DIR=${BUILD_STAGINGDIRECTORY:=${REPO_ROOT_DIR}/output/debian}

PIPELINE_WORKSPACE=${REPO_ROOT_DIR}

if [[ "${BUILD_OUTPUT}" != "" ]]; then
    cp ${BUILD_OUTPUT}/SqlcmdLinux-amd64/sqlcmd ${REPO_ROOT_DIR}/sqlcmd
fi

CLI_VERSION=${CLI_VERSION:=0.0.1}

echo "=========================================================="
echo "CLI_VERSION: ${CLI_VERSION}"
echo "CLI_VERSION_REVISION: ${CLI_VERSION_REVISION:=1}"
echo "Distribution: ${DISTRO}"
echo "Distribution Image: ${DISTRO_BASE_IMAGE}"
echo "=========================================================="

mkdir -p ${DIST_DIR} || exit 1

echo ${REPO_ROOT_DIR}

docker run --rm \
           -v "${REPO_ROOT_DIR}":/mnt/repo \
           -v "${DIST_DIR}":/mnt/output \
           -v "${PIPELINE_WORKSPACE}":/mnt/workspace \
           -e CLI_VERSION=${CLI_VERSION} \
           -e CLI_VERSION_REVISION=${CLI_VERSION_REVISION:=1}~${DISTRO} \
           "${DISTRO_BASE_IMAGE}" \
           /mnt/repo/release/linux/deb/build-pkg.sh
