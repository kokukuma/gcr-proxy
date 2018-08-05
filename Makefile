all: setup build

run:
	go run cmd/https/main.go

setup:
	go get github.com/kokukuma/gcr-proxy/cmd/http
	go get github.com/kokukuma/gcr-proxy/cmd/https
	go get github.com/kokukuma/gcr-proxy/cmd/autocert
	go get github.com/kokukuma/gcr-proxy/proxy

build:
	go build -o http cmd/http/main.go
	go build -o https cmd/https/main.go
	go build -o autocert cmd/autocert/main.go

test:
	cd proxy; go test
