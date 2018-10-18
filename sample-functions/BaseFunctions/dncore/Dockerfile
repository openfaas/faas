FROM microsoft/dotnet:sdk

RUN apt-get update && apt-get install -y curl \
    && curl -sL https://github.com/openfaas/faas/releases/download/0.9.6/fwatchdog > /usr/bin/fwatchdog \
    && chmod +x /usr/bin/fwatchdog \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ENV DOTNET_CLI_TELEMETRY_OPTOUT 1

WORKDIR /root/
COPY src src
WORKDIR /root/src
RUN dotnet restore
RUN dotnet build

ENV fprocess="dotnet ./bin/Debug/netcoreapp1.1/root.dll"
EXPOSE 8080
CMD ["fwatchdog"]
