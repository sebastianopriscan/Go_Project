FROM golang:latest

COPY go.mod /home/go.mod
COPY go.sum /home/go.sum
COPY src/server/ /home/src/server/
COPY src/protobuf/ /home/src/protobuf/

WORKDIR /home/

RUN go build src/server/server_main.go

CMD ["./server_main"]