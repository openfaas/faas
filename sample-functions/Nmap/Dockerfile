FROM functions/alpine:latest

## nmap -A baidu.com
## when run nmap with `-A` argument
## if not install nmap-scripts will cause error:
##
##  NSE: failed to initialize the script engine:
##  could not locate nse_main.lua
##

RUN apk add --no-cache nmap nmap-scripts

ENV fprocess="xargs nmap"	ENV fprocess="xargs nmap"