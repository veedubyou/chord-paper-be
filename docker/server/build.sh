#!/bin/bash

# increment version manually, refactor later
docker build . --file ./Dockerfile-runner --tag pw1124/chord-be:runner-v0.1.2
docker build . --file ./Dockerfile-runner --tag pw1124/chord-be:runner-latest