build:
	go build -o ../qbittorrent-exporter.out .

dev:
	go run .

dev-env:
	go run . -e

format:
	test -z $(gofmt -l .)

lint:
	docker run --rm -v .:/app -w /app golangci/golangci-lint:latest golangci-lint run -v

release:
	git-cliff -l | wl-copy

test:
	go test ./... | \
	sed '/PASS/s//\x1b[32mPASS\x1b[0m/' | \
	sed '/FAIL/s//\x1b[31mFAIL\x1b[0m/'

test-count:
	go test ./... -v | grep -c RUN

test-coverage:
	go test ./... -cover

test-coverage-web:
	go test ./... -coverprofile=cover.out && go tool cover -html=cover.out && rm cover.out

update:
	go mod edit -toolchain=$(go version | awk '{print $3}') && go get -u . && go mod tidy
