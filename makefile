build: 
	go build -o ./qbittorrent-exporter.out ./src
dev : 
	go run ./src
dev-env : 
	go run ./src -e
format : 
	go fmt ./src
lint: 
	docker run --rm -v ./:/app -w /app golangci/golangci-lint:latest golangci-lint run -v
test: 
	go test -v ./src/tests
update: 
	go get -u ./src && go mod tidy