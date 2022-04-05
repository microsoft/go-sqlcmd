#!/usr/bin/env bash
#------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
#------------------------------------------------------------------------------

# Description:
#
# Instructions to be invoked under the build CI pipeline in AzureDevOps.
#
# Kickoff rpm package build. The build pipeline can then save it as an
# artifact as it sees fit.
#
# Usage:
#
# foundation images: `centos:centos7|fedora:29`
#
# $ pipeline.sh

set -exv

: "${REPO_ROOT_DIR:=`cd $(dirname $0); cd ../../../; pwd`}"

if [[ "${BUILD_OUTPUT}" != "" ]]; then
    cp ${BUILD_OUTPUT}/SqlcmdLinuxAmd64/sqlcmd ${REPO_ROOT_DIR}/sqlcmd
fi

DIST_DIR=${BUILD_STAGINGDIRECTORY:=${REPO_ROOT_DIR}/output/rpm}
DISTRO_BASE_IMAGE=( centos:centos7 fedora:29 )

CLI_VERSION=${CLI_VERSION:=0.0.1}

echo "=========================================================="
echo "CLI_VERSION: ${CLI_VERSION}"
echo "CLI_VERSION_REVISION: ${CLI_VERSION_REVISION:=1}"
echo "Distribution Image: ${DISTRO_BASE_IMAGE}"
echo "Output location: ${DIST_DIR}"
echo "=========================================================="

mkdir -p ${DIST_DIR} || exit 1

for i in ${!DISTRO_BASE_IMAGE[@]}; do
    image=${DISTRO_BASE_IMAGE[$i]}

    echo "=========================================================="
    echo "Build rpm on ${image}"
    echo "=========================================================="

    docker run --rm \
               -v "${REPO_ROOT_DIR}":/mnt/repo \
               -v "${DIST_DIR}":/mnt/output \
               -v "${PIPELINE_WORKSPACE}":/mnt/workspace \
               -e CLI_VERSION=${CLI_VERSION} \
               -e CLI_VERSION_REVISION=${CLI_VERSION_REVISION:=1} \
               "${image}" \
               /mnt/repo/release/linux/rpm/build-rpm.sh
done
