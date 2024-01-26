FROM golang:latest

COPY go.mod /home/go.mod
COPY go.sum /home/go.sum
COPY src/client/ /home/src/client/
COPY src/protobuf/ /home/src/protobuf/

WORKDIR /home/

RUN go build src/client/client_main.go

CMD ["./client_main"]