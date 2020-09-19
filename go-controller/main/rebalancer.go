package main

import (
	context "context"
	"log"
	"math"
	reflect "reflect"
	"time"

	grpc "google.golang.org/grpc"
)

// A Rebalancer stores load balancing data for two or more servers
type Rebalancer struct {
	serversHostPort    []string
	LastServerStatuses ServerStatuses
	NbStableIter       int // How many times in a row did the result stay the same ?
	MinNbStableIter    int // How many stable iterations are required before we assume this is stable?
}

// newRebalancer creates a new Rebalancer with a constant list of servers
func newRebalancer(serversHostPort []string, MinNbStableIter int) *Rebalancer {
	rebalancer := new(Rebalancer)
	rebalancer.serversHostPort = serversHostPort
	rebalancer.MinNbStableIter = MinNbStableIter
	return rebalancer
}

func (rebalancer *Rebalancer) run() {
	log.Printf("Fetching server data…")
	serverStatuses, err := rebalancer.retrieveData()
	if err != nil {
		log.Printf("Could not get all data: %v", err)

		// We can't assume stability, so we need to restart.
		rebalancer.NbStableIter = 0

		return
	}

	log.Printf("Current status = %v", serverStatuses)
	if reflect.DeepEqual(rebalancer.LastServerStatuses, serverStatuses) {
		rebalancer.NbStableIter++
		log.Printf("Results seems stable for %d times in a row!", rebalancer.NbStableIter)
	} else {
		log.Printf("Results unstable!")
		rebalancer.NbStableIter = 0
	}

	// No matter what happens next, the latest server status is now this one
	rebalancer.LastServerStatuses = serverStatuses

	if rebalancer.NbStableIter < rebalancer.MinNbStableIter {
		return
	}

	log.Printf("System seems stable enough, applying rebalancing algorithm")
	rebalanceResult := rebalancer.computeRebalance(serverStatuses)
	log.Printf("Rebalancing results: %v", rebalanceResult)

	log.Printf("Applying rebalancing results…")
	err = rebalancer.applyRebalanceResults(rebalanceResult)

	// No matter what, at this point, we need to restart
	rebalancer.NbStableIter = 0

	if err != nil {
		log.Printf("Error while applying results: %v", err)
	}
}

func (rebalancer *Rebalancer) retrieveData() (ServerStatuses, error) {
	resultMap := make(ServerStatuses)

	for _, serverAddr := range rebalancer.serversHostPort {
		log.Printf("About to connect to %s", serverAddr)

		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure())
		conn, err := grpc.Dial(serverAddr, opts...)
		if err != nil {
			log.Printf("Cannot connect, got error %v", err)
			return nil, err
		}
		defer conn.Close()

		client := NewLoadBalancingManagerClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		reply, err := client.GetClientStatus(ctx, &ClientRequest{})

		if err != nil {
			log.Printf("Could not get reply: %v", err)
			return nil, err
		}
		resultMap[serverAddr] = ServerStatus{reply.GetNbClients(), reply.GetMaxNbClients()}
	}
	return resultMap, nil
}

func (rebalancer Rebalancer) computeRebalance(statuses ServerStatuses) map[string]int32 {
	/*
		Rules:
		- at most one server can be limited, in case we got things wrong.
		- the limited server should be notified last, in case there is a communication issue
	*/
	nextStatus := make(map[string]int32)
	NbClients := int32(0)
	nbServers := int32(0)

	for k, status := range statuses {
		nextStatus[k] = -1
		NbClients += status.NbClients
		nbServers++
	}

	log.Printf("There are %v clients in total, and %v servers", NbClients, nbServers)

	// If there are no clients => nothing to do
	if NbClients == 0 {
		return nextStatus
	}

	clientMax := int32(0) // Max number of clients on a server
	serverMax := ""
	for k, status := range statuses {
		if status.NbClients > clientMax {
			serverMax = k
			clientMax = status.NbClients
		}
	}

	maxClientsAllowed := int32(math.Ceil(float64(NbClients) / float64(nbServers)))
	log.Printf("The biggest server is %v with %v clients", serverMax, clientMax)
	if clientMax > maxClientsAllowed {
		log.Printf("Instead, it should have %v clients", maxClientsAllowed)
		nextStatus[serverMax] = maxClientsAllowed
	} else {
		log.Printf("No rebalancing is required")
	}

	return nextStatus
}

func (rebalancer *Rebalancer) applyRebalanceResults(newStatuses map[string]int32) error {
	// Reminder: we expect all servers to be unbounded, except one.
	// Let's start by all the unbounded servers.
	for hostport, maxClients := range newStatuses {
		if maxClients >= 0 {
			continue
		}
		log.Printf("About to connect to %s", hostport)

		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure())
		conn, err := grpc.Dial(hostport, opts...)
		if err != nil {
			log.Printf("Cannot connect, got error %v", err)
			return err
		}
		defer conn.Close()

		client := NewLoadBalancingManagerClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err = client.SetMaxClients(ctx, &SetMaxClientsRequest{MaxNbClients: maxClients})

		if err != nil {
			log.Printf("Could not get reply: %v", err)
			return err
		}
	}

	// Now for the bounded clients (there should be only one)
	for hostport, maxClients := range newStatuses {
		if maxClients <= 0 {
			continue
		}
		log.Printf("About to connect to %s", hostport)

		var opts []grpc.DialOption
		opts = append(opts, grpc.WithInsecure())
		conn, err := grpc.Dial(hostport, opts...)
		if err != nil {
			log.Printf("Cannot connect, got error %v", err)
			return err
		}
		defer conn.Close()

		client := NewLoadBalancingManagerClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_, err = client.SetMaxClients(ctx, &SetMaxClientsRequest{MaxNbClients: maxClients})

		if err != nil {
			log.Printf("Could not get reply: %v", err)
			return err
		}
	}
	return nil
}
