build: 
	cd src && go build -o ../qbittorrent-exporter.out .

dev : 
	cd src && go run .

dev-env : 
	cd src && go run . -e

format : 
	cd src && test -z $(gofmt -l .)

lint: 
	docker run --rm -v ./src:/app -w /app golangci/golangci-lint:latest golangci-lint run -v

test: 
	cd src && go test -v ./...

test-count:
	cd src && go test ./... -v | grep -c RUN

test-coverage:
	cd src && go test ./... -coverprofile=cover.out && go tool cover -html=cover.out && rm cover.out

update: 
	cd src && go get -u . && go mod tidy