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
	cd src && go test -v ./... | \
	sed '/PASS/s//$(shell printf "\033[32mPASS\033[0m")/' | \
	sed '/FAIL/s//$(shell printf "\033[31mFAIL\033[0m")/'

test-count:
	cd src && go test ./... -v | grep -c RUN

test-coverage:
	cd src && go test ./... -cover

test-coverage-web:
	cd src && go test ./... -coverprofile=cover.out && go tool cover -html=cover.out && rm cover.out

update: 
	cd src && go get -u . && go mod tidy