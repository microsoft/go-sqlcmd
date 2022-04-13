# Debian Packaging Release

## Building the Debian package

Execute the following command from the root directory of this repository:

``` bash
./release/linux/debian/pipeline.sh
```

Output will be sent to `./output/debian`

## Dev Installation and Verification

``` bash
./release/linux/debian/pipeline-test.sh
```

## Release Install/Update/Uninstall Steps

> **Note:** Replace `{{HOST}}` and `{{CLI_VERSION}}` with the appropriate values.

### Install sqlcmd with apt (Ubuntu or Debian)

1. Download and install the signing key:

```bash
sudo curl -sL http://{{HOST}}/browse/repo/ubuntu/dpgswdist.v1.asc | gpg --dearmor | tee /etc/apt/trusted.gpg.d/dpgswdist.v1.asc.gpg > /dev/null
```

2. Add the sqlcmd repository information:

```bash
sudo echo "deb [trusted=yes arch=amd64] http://{{HOST}}/browse/repo/ubuntu/sqlcmd mssql main" | tee /etc/apt/sources.list.d/sqlcmd.list
```

3. Update repository information and install sqlcmd:

```bash
sudo apt-get update
sudo apt-get install sqlcmd
```

5. Verify installation success:

```bash
sqlcmd --help
```

### Update

1. Upgrade sqlcmd only:

```bash
sudo apt-get update && sudo apt-get install --only-upgrade -y sqlcmd
```

### Uninstall

1. Uninstall with apt-get remove:

```bash
sudo apt-get remove -y sqlcmd
```

2. Remove the sqlcmd repository information:

> Note: This step is not needed if you plan on installing sqlcmd in the future

```bash
sudo rm /etc/apt/sources.list.d/sqlcmd.list
```

3. Remove the signing key:

```bash
sudo rm /etc/apt/trusted.gpg.d/dpgswdist.v1.asc.gpg
```

4. Remove any unneeded dependencies that were installed with sqlcmd:

```bash
sudo apt autoremove
```
