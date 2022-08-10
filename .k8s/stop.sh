#!/bin/bash

# Convenience script to undeploy everything

kubectl delete svc/chord-be
kubectl delete deploy/chord-be
watch kubectl get all
