FROM microsoft/dotnet:sdk

ADD https://github.com/alexellis/faas/releases/download/v0.5-alpha/fwatchdog /usr/bin
RUN chmod +x /usr/bin/fwatchdog

ENV DOTNET_CLI_TELEMETRY_OPTOUT 1

WORKDIR /root/
COPY src src
WORKDIR /root/src
RUN dotnet restore
RUN dotnet build

ENV fprocess="dotnet ./bin/Debug/netcoreapp1.1/root.dll"
EXPOSE 8080
CMD ["fwatchdog"]
