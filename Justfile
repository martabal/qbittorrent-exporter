build:
	go build -o ../qbittorrent-exporter.out .

dev:
	go run .

dev-env:
	go run . -e

format:
	test -z $(gofmt -l .)

lint:
	golangci-lint run

release:
	git-cliff -l | wl-copy

test:
	gotestsum --format testname ./... -cover

test-coverage-web:
	gotestsum ./... -coverprofile=cover.out && go tool cover -html=cover.out && rm cover.out

update:
	go mod edit -toolchain=$(go version | awk '{print $3}') && go get -u . && go mod tidy
