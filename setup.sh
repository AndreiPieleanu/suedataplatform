#!/bin/bash

# Create docker containers
docker-compose up -d

# Create docker containers for token service
cd token-service
docker-compose up -d
cd ..
