FROM golang:latest

COPY go.mod /home/go.mod
COPY go.sum /home/go.sum
COPY src/balancer/ /home/src/balancer/
COPY src/protobuf/ /home/src/protobuf/

WORKDIR /home/

RUN go build src/balancer/balancer_main.go

CMD ["./balancer_main"]