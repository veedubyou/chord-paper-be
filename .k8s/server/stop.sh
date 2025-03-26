#!/bin/bash

# Convenience script to undeploy everything

kubectl delete svc/chord-be -n chord
kubectl delete deploy/chord-be -n chord
watch kubectl get all -n chord
