#!/bin/bash

# Convenience script to deploy everything

# gets latest image from dockerhub
export LATEST_GO_IMAGE=$(curl -L --fail "https://hub.docker.com/v2/repositories/pw1124/chord-be-go/tags/?page_size=1000" | \
        jq '.results | .[] | .name' -r | \
        sed 's/latest//' | \
        sort --version-sort | \
        tail -n 1)

export LATEST_RUST_IMAGE=$(curl -L --fail "https://hub.docker.com/v2/repositories/pw1124/chord-be-rust/tags/?page_size=1000" | \
        jq '.results | .[] | .name' -r | \
        sed 's/latest//' | \
        sort --version-sort | \
        tail -n 1)

printenv LATEST_GO_IMAGE
printenv LATEST_RUST_IMAGE

cat ./deploy.yml | envsubst > ./processed-yamls/deploy.yml

kubectl apply -f "./processed-yamls/deploy.yml"
kubectl apply -f "service.yml"
watch kubectl get all
