#!/bin/bash

# Convenience script to deploy everything

# gets latest image from dockerhub
export LATEST_IMAGE=$(curl -L --fail "https://hub.docker.com/v2/repositories/pw1124/chord-be/tags/?page_size=1000" | \
        jq '.results | .[] | .name' -r | \
        sed 's/latest//' | \
        sort --version-sort | \
        tail -n 1)

printenv LATEST_IMAGE

mkdir -p ./processed-yamls/ && cat ./deploy.yaml | envsubst > ./processed-yamls/deploy.yaml

kubectl apply -f "../common/chord-be-common-cm.yaml"
kubectl apply -f "./processed-yamls/deploy.yaml"
kubectl apply -f "service.yaml"

watch kubectl get all
