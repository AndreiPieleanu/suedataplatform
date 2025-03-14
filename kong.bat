@echo off

:: Token service
curl -i -X POST http://localhost:8001/services/ ^
  --data "name=token-service"^
  --data "host=host.docker.internal"^
  --data "port=50051"^
  --data "protocol=grpc"

:: Token route
curl -i -X POST http://localhost:8001/services/token-service/routes/ ^
  --data "name=token-route" ^
  --data "paths=/token"^
  --data "protocols[]=grpc"

:: Add the pvc service
curl -i -X POST http://localhost:8001/services/ ^
  --data "name=pvc-service"^
  --data "host=host.docker.internal"^
  --data "port=50052"^
  --data "protocol=grpc"

:: Add pvc route
curl -i -X POST http://localhost:8001/services/pvc-service/routes/ ^
  --data "name=pvc-route" ^
  --data "paths=/pvcservice"^
  --data "protocols[]=grpc"

:: Add jwt for pvc route
curl -i -X POST http://localhost:8001/routes/pvc-route/plugins/ ^
  --data "name=jwt" ^
  --data "config.uri_param_names[]=paramName_2.2.x"

:: Add the notebook service
curl -i -X POST http://localhost:8001/services/ ^
  --data "name=notebook-service"^
  --data "host=host.docker.internal"^
  --data "port=50053"^
  --data "protocol=grpc"

:: Add notebook route
curl -i -X POST http://localhost:8001/services/notebook-service/routes/ ^
  --data "name=notebook-route" ^
  --data "paths=/notebookservice"^
  --data "protocols[]=grpc"

:: Add jwt for notebook route
curl -i -X POST http://localhost:8001/routes/notebook-route/plugins/ ^
  --data "name=jwt" ^
  --data "config.uri_param_names[]=paramName_2.2.x"
