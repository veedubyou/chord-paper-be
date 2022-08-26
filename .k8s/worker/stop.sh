#!/bin/bash

# Convenience script to undeploy everything

kubectl delete deploy/chord-be-workers
watch kubectl get all
