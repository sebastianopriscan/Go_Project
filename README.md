# The application

The application's RPC service is launching Dijkstra's algorithm over a provided graph

# Architecture

The architecture takes advantage of the HTTP/2 streams and the way gRPC uses them in Go. Being that, at each query, by default, a new Goroutine is lauched for every stream, the balancer just creates a client stub for each replica and executes, with a _Round Robin_ policy, the call to those various stubs.

# Project structure

- `src/protobuf` :

    Contains the `.proto` file containing the service definition and the `protoc` generated Go stubs
- `src/client` :

    Client main clode

- `src/server` :

    Server main code
- `src/balancer` :

    Balancer main code

# How to launch the application

## Local run

### Configuraiton

For configuration, edit `src/config` setting the balancer's port, the replicated servers number, the servers' addresses and ports (the `SERVERS_PORTS` property), and the balancer's address. Note: the number of server addresses has to be coherent with the number of replicated servers

### Running

From the base directory (not `src`!), launch the servers specifiying one of the ports precedently configured with

    go run src/server/server_main.go portnum

Then launch the balancer

    go run src/balancer/balancer_main.go

Then you're free to run as many clients as you like

    go run src/client/client_main.go

Every client will launch the same request a number of times between 1 and 20

## Docker Compose : basic

To run with docker compose a static configuration, run

    docker compose up -d

From the `.env` file, the default yaml will spawn a number of clients querying the balancer. Once again, the properties `SERVERS_NUM` and `SERVERS_PORTS` in `.env` have to be coherent, and the number of `server{i}` services in the `docker-compose.yml` and ports property should too.

Note : the balancer is exposed, so it can be invoked from a normal run, too (just check for `config`'s `BALANCER_PORT` consitency)

## Docker Compose: clever

Being that Docker Compose supports auto-balancing of replicas, the configuration in `docker-compose-clever.yml` doesn't use the balancer.

To run :

    docker compose -f docker-compose-clever.yml up -d

Feel free to edit the replicas number as you whish

# Monitoring functioning

The servers print a line for every request received, the clients print a line for every request sent, upon balancing, the balancer prints a line with the index of the chosen server