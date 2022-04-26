FROM golang:alpine AS builder
WORKDIR /app

COPY ./src/gserve /app

RUN go mod init zookeeper
RUN go get github.com/samuel/go-zookeeper/zk

ENV CGO_ENABLED=0
RUN go build -o gserve /app/gserve.go /app/HbaseJSON.go

FROM scratch
WORKDIR /app
COPY --from=builder /app/gserve .
CMD ["/app/gserve"]