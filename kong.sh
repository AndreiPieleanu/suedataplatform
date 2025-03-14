#!/bin/bash

# Add the services
# Token service
curl -i -X POST http://localhost:8001/services/ \
  --data "name=token-service" \
  --data "host=host.docker.internal" \
  --data "port=50051" \
  --data "protocol=grpc"

# Add a route
# Token route
curl -i -X POST http://localhost:8001/services/token-service/routes/ \
  --data "name=token-route" \
  --data "paths[]=/ " \
  --data "protocols[]=grpc"