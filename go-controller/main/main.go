package main

import (
	"context"
	"log"
	"math"
	reflect "reflect"
	"time"

	"google.golang.org/grpc"
)

type serverStatus struct {
	nbClients    int32
	maxNbClients int32
}

func retrieveData() (map[string]serverStatus, error) {
	serverAddrs := []string{"localhost:50052", "192.168.39.101:31045", "localhost:50054"}
	resultMap := make(map[string]serverStatus)

	for _, serverAddr := range serverAddrs {
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
		resultMap[serverAddr] = serverStatus{reply.GetNbClients(), reply.GetMaxNbClients()}
	}
	return resultMap, nil
}

func rebalance(statuses map[string]serverStatus) map[string]int32 {
	/*
		Rules:
		- at most one server can be limited, in case we got things wrong.
		- the limited server should be notified last, in case there is a communication issue
	*/
	nextStatus := make(map[string]int32)
	nbClients := int32(0)
	nbServers := int32(0)

	for k, status := range statuses {
		nextStatus[k] = -1
		nbClients += status.nbClients
		nbServers++
	}

	log.Printf("There are %v clients in total, and %v servers", nbClients, nbServers)

	// If there are no clients => nothing to do
	if nbClients == 0 {
		return nextStatus
	}

	clientMax := int32(0) // Max number of clients on a server
	serverMax := ""
	for k, status := range statuses {
		if status.nbClients > clientMax {
			serverMax = k
			clientMax = status.nbClients
		}
	}

	maxClientsAllowed := math.Ceil(float64(nbClients) / float64(nbServers))
	log.Printf("The biggest server is %v with %v clients", serverMax, clientMax)
	log.Printf("Instead, it should have %v clients", maxClientsAllowed)
	nextStatus[serverMax] = int32(maxClientsAllowed)
	return nextStatus
}

func applyRebalanceResults(newStatuses map[string]int32) error {
	log.Printf("Applying rebalancing results…")

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

func main() {
	log.Println("Starting…")
	done := make(chan bool)
	ticker := time.NewTicker(2 * time.Second)

	prevMap := make(map[string]serverStatus)
	const minStableIter = 10
	nbStableIter := 0

	/*go*/
	func() {
		for {
			select {
			case <-ticker.C:
				log.Printf("Fetching server data…")
				resultMap, err := retrieveData()
				if err != nil {
					log.Printf("Could not get all data: %v", err)
				} else {
					log.Printf("All data = %v", resultMap)
					if reflect.DeepEqual(prevMap, resultMap) {
						nbStableIter++
						log.Printf("Results seems stable for %d times in a row!", nbStableIter)
					} else {
						log.Printf("Result unstable!")
						nbStableIter = 0
					}
					prevMap = resultMap

					if nbStableIter >= minStableIter {
						log.Printf("Enough stable iterations, can apply rebalancing algorithm")
						rebalanceResult := rebalance(resultMap)
						log.Printf("Rebalancing results: %v", rebalanceResult)
						err = applyRebalanceResults(rebalanceResult)
						// No matter what, at this point, we need to restart
						nbStableIter = 0

						if err != nil {
							log.Printf("Error while applying results: %v", err)
						}
					}
				}
			case <-done:
				return
			}
		}
	}()

	time.Sleep(2 * time.Second)
}
