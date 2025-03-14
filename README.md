# Requirements
- Go
- Docker
- Protoc (Protocol Buffers Compiler)

# Usage Instruction

## Install protoc compiler for go
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## Running the docker compose
1. Start your docker
2. Create the docker containers

for linux and mac os
```bash
./setup.sh
```

for windows
```powershell
.\setup.bat
```
## Export kong routes
Wait for the kong container is fully set up, then run the command below

for linux and mac os
```bash
./kong.sh
```

for windows
```powershell
.\kong.bat
```

## Starting services
1. Go to the service you want to start
2. run this command
```bash
go run main.go
```

## Testing services
1. Go to the service you want to test
2. run this command
```bash
go test -v ./...
```

## Adding new service and route to api gateway
1. Go to the kong.bat and kong.sh
2. Add new curl commands for adding service and routes there.
3. run the command below

for linux and mac os
```bash
./kong.sh
```

for windows
```powershell
.\kong.bat
```
