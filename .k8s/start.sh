#!/bin/bash

# Convenience script to deploy everything

microk8s kubectl apply -f "deploy.yml"
microk8s kubectl apply -f "service.yml"
watch microk8s kubectl get all
