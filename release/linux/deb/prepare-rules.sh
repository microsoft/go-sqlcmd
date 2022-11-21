#!/usr/bin/env bash

#------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
#------------------------------------------------------------------------------

# Description:
#
# Create the debian/directory for building the `sqlcmd` Debian package and
# the package rules.
#
# Usage:
#
# prepare-rules.sh DEBIAN-DIR SRC_DIR
#

set -evx

if [[ -z "$1" ]]
  then
    echo "No argument supplied for debian directory."
    exit 1
fi

if [[ -z "$2" ]]
  then
    echo "No argument supplied for source directory."
    exit 1
fi

TAB=$'\t'

debian_dir=$1
source_dir=$2

mkdir -p $debian_dir/source || exit 1

echo '1.0' > $debian_dir/source/format
echo '9' > $debian_dir/compat

cat > $debian_dir/changelog <<- EOM
sqlcmd (${CLI_VERSION}-${CLI_VERSION_REVISION:=1}) stable; urgency=low

  * Debian package release.

 -- sqlcmd tools team <dpgswdist@microsoft.com>  $(date -R)

EOM

cat > $debian_dir/control <<- EOM
Source: sqlcmd
Section: sql
Priority: extra
Maintainer: sqlcmd tools team <dpgswdist@microsoft.com>
Build-Depends: debhelper (>= 9)
Standards-Version: 3.9.5
Homepage: https://github.com/microsoft/go-sqlcmd

Package: sqlcmd
Architecture: all
Depends: \${shlibs:Depends}, \${misc:Depends}
Description: SQLCMD TOOLS CLI
 SQLCMD TOOLS CLI, a multi-platform command line experience for Microsoft SQL Server and Azure SQL.

EOM

cat > $debian_dir/copyright <<- EOM
Format: http://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: sqlcmd
Upstream-Contact: sqlcmd tools team <dpgswdist@microsoft.com>
Source: PUBLIC

Files: *
Copyright: Copyright (c) Microsoft Corporation
License: https://github.com/microsoft/go-sqlcmd/blob/main/LICENSE

MIT License

Copyright (c) Microsoft Corporation.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE

EOM

cat > $debian_dir/rules << EOM
#!/usr/bin/make -f

# Uncomment this to turn on verbose mode.
export DH_VERBOSE=1
export DH_OPTIONS=-v

%:
${TAB}dh \$@ --sourcedirectory $source_dir

override_dh_install:
${TAB}mkdir -p debian/sqlcmd/usr/bin/
${TAB}cp -r /opt/stage/sqlcmd debian/sqlcmd/usr/bin/sqlcmd
${TAB}chmod 0755 debian/sqlcmd/usr/bin/sqlcmd

override_dh_strip:
${TAB}dh_strip --exclude=_cffi_backend

EOM

cat $debian_dir/rules

# Debian rules should be executable
chmod 0755 $debian_dir/rules
