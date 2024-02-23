go build -o ./build/proxy_server
GOOS=linux GOARCH=arm64 go build -o ./build/proxy_server_arm
GOOS=linux GOARCH=amd64 go build -o ./build/proxy_server_386