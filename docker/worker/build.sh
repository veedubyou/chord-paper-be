#!/bin/bash

# increment version manually, refactor later
docker build . --file ./Dockerfile-preinstall --tag pw1124/chord-be-workers:preinstall-v0.1.1
docker build . --file ./Dockerfile-preinstall --tag pw1124/chord-be-workers:preinstall-latest