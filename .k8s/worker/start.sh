#!/bin/bash

# Convenience script to deploy everything

# gets latest image from dockerhub
export LATEST_WORKERS_IMAGE=$(curl -L --fail "https://hub.docker.com/v2/repositories/pw1124/chord-be-workers/tags/?page_size=1000" | \
        jq '.results | .[] | .name' -r | \
        sed 's/latest//' | \
        sort --version-sort | \
        tail -n 1)
export LATEST_YOUTUBE_IMAGE=$(curl -L --fail "https://hub.docker.com/v2/repositories/pw1124/youtube-dl-bin/tags/?page_size=1000" | \
        jq '.results | .[] | .name' -r | \
        sed 's/latest//' | \
        sort --version-sort | \
        tail -n 1)

printenv LATEST_WORKERS_IMAGE
printenv LATEST_YOUTUBE_IMAGE

mkdir -p ./processed-yamls/ && cat ./deploy.yaml | envsubst > ./processed-yamls/deploy.yaml

kubectl apply -f "../common/chord-be-common-cm.yaml"
kubectl apply -f "./processed-yamls/deploy.yaml"

watch kubectl get all -n chord
