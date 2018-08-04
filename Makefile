all: setup build

run:
	go run cmd/main.go

setup:
	go get github.com/kokukuma/gcr-proxy/cmd
	go get github.com/kokukuma/gcr-proxy/proxy

build:
	go build -o app cmd/main.go

test:
	cd proxy; go test
