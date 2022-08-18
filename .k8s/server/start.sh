#!/bin/bash

# Convenience script to deploy everything

# gets latest image from dockerhub
export LATEST_IMAGE=$(curl -L --fail "https://hub.docker.com/v2/repositories/pw1124/chord-be/tags/?page_size=1000" | \
        jq '.results | .[] | .name' -r | \
        sed 's/latest//' | \
        sort --version-sort | \
        tail -n 1)

printenv LATEST_IMAGE

mkdir ./processed-yamls/ && cat ./deploy.yml | envsubst > ./processed-yamls/deploy.yml

kubectl apply -f "./processed-yamls/deploy.yml"
kubectl apply -f "service.yml"

watch kubectl get all
