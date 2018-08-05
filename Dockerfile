FROM golang:latest as builder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY . /go/src/github.com/kokukuma/gcr-proxy/
WORKDIR /go/src/github.com/kokukuma/gcr-proxy/
RUN make

# runtime image
FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=builder /go/src/github.com/kokukuma/gcr-proxy/autocert /autocert
EXPOSE 8000
ENTRYPOINT ["/autocert"]

