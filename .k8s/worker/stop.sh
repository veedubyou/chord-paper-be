#!/bin/bash

# Convenience script to undeploy everything

kubectl delete -n chord deploy/chord-be-workers
watch kubectl get all -n chord
