#!/bin/sh
echo Building functions/api-key-protected:latest
docker build --no-cache -t functions/api-key-protected:latest .
