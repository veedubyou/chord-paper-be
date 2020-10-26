#!/bin/bash

# Convenience script to deploy everything

kubectl apply -f "deploy.yml"
kubectl apply -f "service.yml"
watch kubectl get all
