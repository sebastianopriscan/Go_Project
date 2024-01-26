package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "sdcc_project.org/balancer/src/protobuf"
)

var graph pb.Graph
var servAddr string
var portString string

func createGraph() {

	nodes := [6]int64{1, 2, 3, 4, 5, 6}

	edges := [...]*pb.Edge{
		{StartNode: 1, EndNode: 3, Weight: 5},
		{StartNode: 1, EndNode: 6, Weight: 2},
		{StartNode: 1, EndNode: 5, Weight: 3},
		{StartNode: 2, EndNode: 3, Weight: 4},
		{StartNode: 2, EndNode: 4, Weight: 1},
		{StartNode: 3, EndNode: 6, Weight: 1},
		{StartNode: 4, EndNode: 5, Weight: 3},
		{StartNode: 4, EndNode: 6, Weight: 8},
	}

	graph = pb.Graph{Nodes: nodes[:], StartId: 1, EndId: 2, Edges: edges[:]}
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

		servAddr = extractValueFromEnv("SERVER_NAME")
		portString = extractValueFromEnv("BALANCER_PORT")

		_, convErr := strconv.Atoi(portString)

		if convErr != nil {
			log.Println("Error opening file")
			os.Exit(1)
		}

	} else {

		defer func() {
			if err := data.Close(); err != nil {
				panic(err)
			}
		}()

		reader := bufio.NewReader(data)

		var readErr error
		portString, readErr = reader.ReadString('\n')

		if readErr != nil {
			log.Println("Error in reading from config. Exiting...")
			os.Exit(1)
		}

		portString = portString[strings.Index(portString, ":")+2 : len(portString)-1]

		_, readErr = reader.ReadString('\n')

		if readErr != nil {
			log.Println("Error in reading from config. Exiting...")
			os.Exit(1)
		}

		_, readErr = reader.ReadString('\n')

		if readErr != nil {
			log.Println("Error in reading from config. Exiting...")
			os.Exit(1)
		}

		servAddr, readErr = reader.ReadString('\n')

		if readErr != nil && readErr != io.EOF {
			log.Println("Error in reading from config. Exiting...")
			os.Exit(1)
		}

		servAddr = servAddr[strings.Index(servAddr, ":")+2 : len(servAddr)-1]
	}

	insecure := grpc.WithTransportCredentials(insecure.NewCredentials())

	log.Println("Dialling " + servAddr + ":" + portString)
	conn, err := grpc.Dial(servAddr+":"+portString, insecure)
	if err != nil {
		log.Println("Error in parsing data. Exiting...")
		os.Exit(1)
	}
	defer conn.Close()
	client := pb.NewShortestPathClient(conn)

	log.Println("Created Stub")

	trials := rand.Intn(20) + 1

	createGraph()

	for i := 0; i < trials; i++ {
		ctx := context.Background()

		log.Println("Launching stub")
		result, err := client.SSSP(ctx, &graph)

		if err != nil {
			log.Println("Error")
		} else {
			for j := 0; j < len(result.GetNodeSequence()); j++ {
				log.Print(strconv.FormatInt(result.GetNodeSequence()[j], 10) + " ")
			}
		}
		log.Println("")
	}
}
