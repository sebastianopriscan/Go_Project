package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "sdcc_project.org/balancer/src/protobuf"
)

var serverNum int = 0
var idx int = 0
var port int = 0
var serverStrings []string = nil

var clients []pb.ShortestPathClient = nil

type shortestPathServer struct {
	pb.UnimplementedShortestPathServer
}

func (s *shortestPathServer) SSSP(my_context context.Context, graph *pb.Graph) (*pb.Path, error) {

	ctx := context.Background()
	log.Printf("Balancing to replica %d\n", idx)
	path, err := clients[idx].SSSP(ctx, graph)

	idx = (idx + 1) % serverNum

	return path, err
}

func extractNumberFromEnv(variable string) int {

	valueString, present := os.LookupEnv(variable)

	if !present {
		log.Println("Bad configuration")
		os.Exit(1)
	}

	var convErr error
	value, convErr := strconv.Atoi(valueString)

	if convErr != nil {
		log.Println("Bad configuration")
		os.Exit(1)
	}

	return value
}

func extractValueFromEnv(variable string) string {

	valueString, present := os.LookupEnv(variable)

	if !present {
		log.Println("Bad configuration")
		os.Exit(1)
	}

	return valueString
}

func main() {

	data, err := os.Open("src/config")

	if err != nil {

		log.Println("Error opening file")

		port = extractNumberFromEnv("PORT")

		serverNum = extractNumberFromEnv("SERVER_NUM")

		serverFirstPort := extractValueFromEnv("SERVER_PORTS")

		log.Println(serverFirstPort)

		serverName := extractValueFromEnv("SERVER_NAME")

		serverNames := strings.Split(serverFirstPort, " ")

		if len(serverNames) != serverNum {
			log.Println("number of servers != number of ports!")
			os.Exit(1)
		}

		serverStrings = make([]string, serverNum)

		for i := 0; i < serverNum; i++ {

			_, convErr := strconv.Atoi(serverNames[i])

			if convErr != nil {
				log.Println("Format error")
				os.Exit(1)
			}
			serverStrings[i] = serverName + strconv.Itoa(i+1) + ":" + serverNames[i]
		}

	} else {
		defer func() {
			if err := data.Close(); err != nil {
				panic(err)
			}
		}()

		reader := bufio.NewReader(data)

		portString, readErr := reader.ReadString('\n')

		if readErr != nil {
			log.Println("Error in reading from config. Exiting...")
			os.Exit(1)
		}

		portString = portString[strings.Index(portString, ":")+2 : len(portString)-1]

		var convErr error
		port, convErr = strconv.Atoi(portString)

		if convErr != nil {
			log.Println("Error in reading from config. Exiting...")
			os.Exit(1)
		}

		serverNumString, readErr := reader.ReadString('\n')

		if readErr != nil {
			log.Println("Error in reading from config. Exiting...")
			os.Exit(1)
		}

		serverNumString = serverNumString[strings.Index(serverNumString, ":")+2 : len(serverNumString)-1]

		serverNum, convErr = strconv.Atoi(serverNumString)

		if convErr != nil {
			log.Println("Error in reading from config. Exiting...")
			os.Exit(1)
		}

		serversString, readErr := reader.ReadString('\n')

		if readErr != nil && readErr != io.EOF {
			log.Println("Error in reading from config. Exiting...")
			os.Exit(1)
		}

		serversString = serversString[strings.Index(serversString, ":")+2 : len(serversString)-1]

		serverStrings = strings.Split(serversString, " ")
	}

	clients = make([]pb.ShortestPathClient, serverNum)

	for i := 0; i < serverNum; i++ {

		nocreds := grpc.WithTransportCredentials(insecure.NewCredentials())

		log.Println("Dialling " + serverStrings[i])

		conn, err := grpc.Dial(serverStrings[i], nocreds)
		if err != nil {
			log.Println("Error in parsing data: " + err.Error() + ". Exiting...")
			os.Exit(1)
		}
		defer conn.Close()
		clients[i] = pb.NewShortestPathClient(conn)

		log.Println("Created Stub")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Println("Error in bining server. Exiting...")
		os.Exit(1)
	}

	log.Println("Created connection at" + fmt.Sprintf("localhost:%d", port))
	grpcServer := grpc.NewServer()
	pb.RegisterShortestPathServer(grpcServer, &shortestPathServer{})
	grpcServer.Serve(lis)
}
