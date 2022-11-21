#!/usr/bin/env bash

#------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
#------------------------------------------------------------------------------

# Description:
#
# Instructions to be invoked under the build CI pipeline in AzureDevOps.
#
# Kickoff rpm package tests against versions:
#
# -----------------------------------
# centos:centos8
# centos:centos7
# -----------------------------------
# fedora:31
# fedora:30
# fedora:29
# -----------------------------------
# opensuse/leap:latest
# -----------------------------------
#
# Usage:
# $ pipeline-test.sh

set -exv

: "${REPO_ROOT_DIR:=`cd $(dirname $0); cd ../../../; pwd`}"

CLI_VERSION=${CLI_VERSION:=0.0.1}
CLI_VERSION_REVISION=${CLI_VERSION_REVISION:=1}

BUILD_ARTIFACTSTAGINGDIRECTORY=${BUILD_ARTIFACTSTAGINGDIRECTORY:=${REPO_ROOT_DIR}/output/rpm}

YUM_DISTRO_BASE_IMAGE=( centos:centos7 centos:centos8 fedora:29 fedora:30 fedora:31 )
YUM_DISTRO_SUFFIX=( el7 el7 fc29 fc29 fc29 )

ZYPPER_DISTRO_BASE_IMAGE=( opensuse/leap:latest )
ZYPPER_DISTRO_SUFFIX=( el7 )

echo "=========================================================="
echo "CLI_VERSION: ${CLI_VERSION}"
echo "CLI_VERSION_REVISION: ${CLI_VERSION_REVISION}"
echo "BUILD_ARTIFACTSTAGINGDIRECTORY: ${BUILD_ARTIFACTSTAGINGDIRECTORY}"
echo "Distribution: ${YUM_DISTRO_BASE_IMAGE} ${ZYPPER_DISTRO_BASE_IMAGE}"
echo "=========================================================="

# -- yum installs --
for i in ${!YUM_DISTRO_BASE_IMAGE[@]}; do
    image=${YUM_DISTRO_BASE_IMAGE[$i]}
    suffix=${YUM_DISTRO_SUFFIX[$i]}

    echo "=========================================================="
    echo "Test rpm package on ${image} .${suffix}"
    echo "=========================================================="
    rpmPkg=sqlcmd-${CLI_VERSION}-${CLI_VERSION_REVISION}.${suffix}.x86_64.rpm

    # Per: https://techglimpse.com/failed-metadata-repo-appstream-centos-8/
    # change the mirrors to vault.centos.org where they will be archived permanently
    mirrors=""
    if [[ "${image}" == "centos:centos8" ]]; then
        mirrors="cd /etc/yum.repos.d/ && \
           sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-* && \
           sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-* && \ "
    fi
     
    script="${mirrors}
           rpm --import https://packages.microsoft.com/keys/microsoft.asc && \
           yum update -y && \
           yum localinstall /mnt/artifacts/${rpmPkg} -y && \
           sqlcmd --help"

    docker pull ${image}
    docker run --rm -v ${BUILD_ARTIFACTSTAGINGDIRECTORY}:/mnt/artifacts \
               ${image} \
               /bin/bash -c "${script}"

    echo ""
done

# -- zypper installs --
for i in ${!ZYPPER_DISTRO_BASE_IMAGE[@]}; do
    image=${ZYPPER_DISTRO_BASE_IMAGE[$i]}
    suffix=${ZYPPER_DISTRO_SUFFIX[$i]}

    echo "=========================================================="
    echo "Test rpm package on ${image} .${suffix}"
    echo "=========================================================="
    rpmPkg=sqlcmd-${CLI_VERSION}-${CLI_VERSION_REVISION}.${suffix}.x86_64.rpm
    # If testing locally w/o signing, use `--allow-unsigned-rpm` but do not commit:
    # zypper --non-interactive install --allow-unsigned-rpm /mnt/artifacts/${rpmPkg} && \

    script="zypper --non-interactive install curl && \
            rpm -v --import https://packages.microsoft.com/keys/microsoft.asc && \
            zypper --non-interactive install /mnt/artifacts/${rpmPkg} && \
            sqlcmd --help"

    docker pull ${image}
    docker run --rm -v ${BUILD_ARTIFACTSTAGINGDIRECTORY}:/mnt/artifacts \
               ${image} \
               /bin/bash -c "${script}"

    echo ""
done
