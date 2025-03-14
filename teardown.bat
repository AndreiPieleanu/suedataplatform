@echo off

:: List of services to loop through
set services=token pvc notebook

:: Loop through each service
for %%s in (%services%) do (
    cd "%%s-service"
    docker compose down -v
    cd..
)

:: Create docker containers
docker compose down -v