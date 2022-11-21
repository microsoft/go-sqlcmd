#------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation.
# Licensed under the MIT license.
#------------------------------------------------------------------------------

# Example:
# docker run --rm microsoft/sqlcmd sqlcmd --help
#

FROM scratch
ARG BUILD_DATE
ARG PACKAGE_VERSION

LABEL maintainer="Microsoft" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.vendor="Microsoft" \
      org.label-schema.name="SQLCMD CLI" \
      org.label-schema.version=$PACKAGE_VERSION \
      org.label-schema.license="https://github.com/microsoft/go-sqlcmd/blob/main/LICENSE" \
      org.label-schema.description="The MSSQL SQLCMD CLI tool" \
      org.label-schema.url="https://github.com/microsoft/go-sqlcmd" \
      org.label-schema.usage="https://docs.microsoft.com/sql/tools/sqlcmd-utility" \
      org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.docker.cmd="docker run -it microsoft/sqlcmd:$PACKAGE_VERSION"

COPY ./sqlcmd /usr/bin/sqlcmd

CMD ["sqlcmd"]
