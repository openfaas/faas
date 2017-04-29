#!/bin/sh
MASTER_REPO_URL=https://raw.githubusercontent.com/alexellis/faas/master
curl -O $MASTER_REPO_URL/docker-compose.yml
mkdir prometheus
cd prometheus
curl -O $MASTER_REPO_URL/prometheus/alert.rules
curl -O $MASTER_REPO_URL/prometheus/alertmanater.yml
curl -O $MASTER_REPO_URL/prometheus/prometheus.yml
cd
docker stack deploy func --compose-file docker-compose.yml
