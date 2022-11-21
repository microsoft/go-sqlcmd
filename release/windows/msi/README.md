# Windows MSI

This document provides instructions on creating the MSI.

## Prerequisites

1. WIX Toolset
2. Turn on the '.NET Framework 3.5' Windows Feature (required for WIX Toolset)
3. Install [WIX Toolset build tools](http://wixtoolset.org/releases/) if not already installed
4. Install [Microsoft Build Tools](https://www.microsoft.com/en-us/download/details.aspx?id=48159)

## Building

1. Set the `CLI_VERSION` environment variable
2. Run `release\windows\msi\scripts\pipeline.cmd`
3. The unsigned MSI will be in the `.\output\msi` folder

> **Note:** For `building step 1.` above set both env-vars to the same version-tag for the immediate, this will consolidated in the future.

## Release Install/Update/Uninstall Steps

> **Note:** Replace `{{HOST}}` and `{{CLI_VERSION}}` with the appropriate values.

### Install `Sqlcmd Tools` on Windows

The MSI distributable is used for installing or updating the `Sqlcmd Tools` CLI on Windows. 

[Download the MSI Installer](http://{{HOST}}/sqlcmd-{{CLI_VERSION}}.msi)

When the installer asks if it can make changes to your computer, click the `Yes` box.

### Uninstall

You can uninstall the `SqlCmd Tools` from the Windows _Apps and Features_ list. To uninstall:

| Platform      | Instructions                                           |
| ------------- |--------------------------------------------------------|
| Windows 10	| Start > Settings > Apps                                |
| Windows 8     | Start > Control Panel > Programs > Uninstall a program |

The program to uninstall is listed as **Sqlcmd Tools** . Select this application, then click the `Uninstall` button.
