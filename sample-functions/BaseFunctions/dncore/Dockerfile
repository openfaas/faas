FROM openfaas/classic-watchdog:0.18.8 as watchdog

FROM mcr.microsoft.com/dotnet/core/sdk:2.1 as build

COPY --from=watchdog /fwatchdog /usr/bin/fwatchdog
RUN chmod +x /usr/bin/fwatchdog

ENV DOTNET_CLI_TELEMETRY_OPTOUT 1

WORKDIR /application/
COPY src src
WORKDIR /application/src
RUN dotnet restore
RUN dotnet build

FROM build as runner
RUN groupadd -g 1000 -r faas && useradd -r -g faas -u 1000 faas -m
USER 1000
ENV fprocess="dotnet ./bin/Debug/netcoreapp2.1/root.dll"
EXPOSE 8080
CMD ["fwatchdog"]
