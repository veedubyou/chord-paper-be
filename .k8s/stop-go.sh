#!/bin/bash

# Convenience script to undeploy everything

kubectl delete svc/chord-be-go
kubectl delete deploy/chord-be-go
watch kubectl get all
