FROM mcr.microsoft.com/mssql/server:2017-latest

RUN wget -q https://github.com/microsoft/go-sqlcmd/releases/download/v0.11.0/sqlcmd-v0.11.0-linux-arm64.tar.bz2 && \
    tar -xvjf sqlcmd-v0.11.0-linux-arm64.tar.bz2 && \
    chmod +x ./sqlcmd && \
    mv ./sqlcmd /usr/bin/sqlcmd && \
    rm sqlcmd-v0.11.0-linux-arm64.tar.bz2

ENV ACCEPT_EULA=Y
ENV MSSQL_PID=Developer
ENV SA_PASSWORD="SomeRandomPassw0rdForContainer"
ENV SQLCMDPASSWORD=${SA_PASSWORD}

COPY myscript.sql /app/
COPY myscript.sh /app/

CMD /app/myscript.sh
