package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"slices"
	"strconv"

	"google.golang.org/grpc"
	pb "sdcc_project.org/balancer/src/protobuf"
)

func checkGraphConsistency(graph *pb.Graph) bool {
	nodes := graph.Nodes
	start := graph.StartId
	end := graph.EndId
	edges := graph.Edges

	if !slices.Contains(nodes, start) {
		return false
	}
	if !slices.Contains(nodes, end) {
		return false
	}
	for i := 0; i < len(edges); i++ {
		if edges[i].StartNode == edges[i].EndNode {
			return false
		}
		if !slices.Contains(nodes, edges[i].StartNode) {
			return false
		}
		if !slices.Contains(nodes, edges[i].EndNode) {
			return false
		}
	}

	return true
}

type shortestPathServer struct {
	pb.UnimplementedShortestPathServer
}

type dijkstraNode struct {
	node int64
	dist int64
	prev int64
}

func getMinDistanceNode(nodes []*dijkstraNode) (*dijkstraNode, int) {
	min := (int64)(math.MaxInt64)
	var minNode *dijkstraNode
	var minNodeIdx int

	for i := 0; i < len(nodes); i++ {
		if nodes[i].dist < min {
			min = nodes[i].dist
			minNode = nodes[i]
			minNodeIdx = i
		}
	}

	return minNode, minNodeIdx
}

func isNodeContainedIndex(node int64, nodes []*dijkstraNode) int {

	for i := 0; i < len(nodes); i++ {
		if nodes[i].node == node {
			return i
		}
	}

	return -1
}

func getRemainingNeighboringEdges(graph *pb.Graph, node int64, nodes []*dijkstraNode) ([]*pb.Edge, []*dijkstraNode) {

	retVal := make([]*pb.Edge, 0)
	retVal2 := make([]*dijkstraNode, 0)
	for i := 0; i < len(graph.Nodes); i++ {
		if graph.Nodes[i] == node {
			continue
		}
		var idx = isNodeContainedIndex(graph.Nodes[i], nodes)
		if idx == -1 {
			continue
		}
		for j := 0; j < len(graph.Edges); j++ {
			edge := graph.Edges[j]

			if (edge.StartNode == node && edge.EndNode == graph.Nodes[i]) ||
				(edge.EndNode == node && edge.StartNode == graph.Nodes[i]) {
				retVal = append(retVal, edge)
				retVal2 = append(retVal2, nodes[idx])
			}
		}
	}

	return retVal, retVal2
}

func updateNode(node int64, newDist int64, newPrev int64, nodes []*dijkstraNode) {

	for i := 0; i < len(nodes); i++ {
		if nodes[i].node == node {
			nodes[i].dist = newDist
			nodes[i].prev = newPrev
		}
	}

}

func createPath(start int64, end int64, nodes []*dijkstraNode) *pb.Path {

	var path []int64 = make([]int64, 0)

	searched := end

	path = append(path, searched)

	for i := 0; i < len(nodes); i++ {
		if nodes[i].node == searched {
			searched = nodes[i].prev
			path = append(path, searched)
			if searched == start {
				break
			}
			i = 0
		}
	}

	var reversedPath []int64 = make([]int64, len(path))

	j := 0
	for i := len(path) - 1; i >= 0; i-- {
		reversedPath[j] = path[i]
		j++
	}

	return &pb.Path{NodeSequence: reversedPath}
}

func (s *shortestPathServer) SSSP(my_context context.Context, graph *pb.Graph) (*pb.Path, error) {

	if !checkGraphConsistency(graph) {
		return nil, errors.New("wrong graph format")
	}

	var nodes []*dijkstraNode = make([]*dijkstraNode, 0)

	var remainingNodes []*dijkstraNode = make([]*dijkstraNode, 0)

	for i := 0; i < len(graph.Nodes); i++ {
		if graph.Nodes[i] == graph.StartId {
			nodes = append(nodes, &dijkstraNode{graph.Nodes[i], 0, -1})
			remainingNodes = append(remainingNodes, &dijkstraNode{graph.Nodes[i], 0, -1})
		} else {
			nodes = append(nodes, &dijkstraNode{graph.Nodes[i], math.MaxInt, -1})
			remainingNodes = append(remainingNodes, &dijkstraNode{graph.Nodes[i], math.MaxInt, -1})
		}
	}

	for len(remainingNodes) > 0 {
		minNode, idx := getMinDistanceNode(remainingNodes)

		if idx == 0 {
			remainingNodes = remainingNodes[1:]
		} else if idx == len(nodes)-1 {
			remainingNodes = remainingNodes[0:idx]
		} else {
			remainingNodes = append(remainingNodes[0:idx], remainingNodes[(idx+1):]...)
		}

		if minNode.node == graph.EndId {
			break
		}

		remainingEdges, correspondingNodes := getRemainingNeighboringEdges(graph, minNode.node, remainingNodes)
		for i := 0; i < len(remainingEdges); i++ {
			alt := minNode.dist + remainingEdges[i].Weight

			if alt < correspondingNodes[i].dist {

				updateNode(correspondingNodes[i].node, alt, minNode.node, nodes)
				updateNode(correspondingNodes[i].node, alt, minNode.node, remainingNodes)
			}
		}
	}

	log.Println("Did a request")

	return createPath(graph.StartId, graph.EndId, nodes), nil
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

func main() {

	var port int

	if len(os.Args) == 2 {
		log.Println("Usage: ./prog port")

		var err error
		port, err = strconv.Atoi(os.Args[1])

		if err != nil {
			log.Println("Port is a number!")
			os.Exit(1)
		}
	} else {
		port = extractNumberFromEnv("PORT")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Println("Error in bining server: " + err.Error() + ". Exiting...")
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterShortestPathServer(grpcServer, &shortestPathServer{})
	grpcServer.Serve(lis)
}
