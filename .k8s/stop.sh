#!/bin/bash

# Convenience script to undeploy everything

microk8s kubectl delete svc/chord-be
microk8s kubectl delete deploy/chord-be
watch microk8s kubectl get all
