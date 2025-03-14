@echo off

:: Create docker containers
docker compose up -d

:: List of services to loop through
set services=token pvc notebook

:: Loop through each service
for %%s in (%services%) do (
    cd "%%s-service"
    docker compose up -d
    protoc --go_out=. --go-grpc_out=. "api/%%s.proto"
    if not "%%s"=="token" (
        git clone https://github.com/kubeflow/manifests.git
        cd manifests
        git checkout 8634c24
        cd..
    )
    cd..
)