# Generating a new access token
1. Setup the application following [this guide.](../README.md)
2. Run the application using following command
```bash
go run main.go
```
3. make a gRPC request to following url.
you can use postman or grpc curl for this
```bash
localhost:8000 token.Token/Login
```

note that the token will last for 1 day.