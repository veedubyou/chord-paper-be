#!/bin/bash

# Convenience script to undeploy everything

kubectl delete svc/chord-be-rust
kubectl delete deploy/chord-be-rust
watch kubectl get all
