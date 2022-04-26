FROM golang:alpine AS builder
WORKDIR /app
COPY ./src/grproxy /app

RUN go mod init zookeeper
RUN go get github.com/samuel/go-zookeeper/zk

ENV CGO_ENABLED=0
RUN go build -o grproxy /app/grproxy.go

FROM scratch
WORKDIR /app
COPY --from=builder /app/grproxy .
CMD ["/app/grproxy"]
